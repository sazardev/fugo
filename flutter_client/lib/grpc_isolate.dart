import 'dart:async';
import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';

import 'package:grpc/grpc.dart';

import 'generated/fugo/v1/fugo.pbgrpc.dart';

const _reconnectDelay = Duration(milliseconds: 500);
const _maxBackoff = Duration(seconds: 5);

void grpcIsolateEntry(SendPort mainSendPort) {
  final receivePort = ReceivePort();
  mainSendPort.send(receivePort.sendPort);

  final addr = Platform.environment['FUGO_ADDR'] ?? '127.0.0.1:9510';
  print('[fugo] grpc isolate started (addr: $addr)');

  connect(addr, mainSendPort, receivePort);
}

Future<void> connect(
  String addr,
  SendPort mainSendPort,
  ReceivePort receivePort,
) async {
  // A ReceivePort is a single-subscription stream: it can be listened to only
  // once for the lifetime of the isolate. Re-listening on each reconnect throws
  // "Bad state: Stream has already been listened to", which broke reconnection
  // (and hence hot reload). So subscribe ONCE here and forward decoded events
  // to whichever per-connection controller is currently active.
  StreamController<ClientEvent>? active;
  receivePort.listen((message) {
    if (message is List<int>) {
      final event = ClientEvent.fromBuffer(Uint8List.fromList(message));
      print('[fugo] isolate: received event ${event.eventType} node=${event.nodeId}');
      active?.add(event);
    }
  });

  var delay = _reconnectDelay;
  while (true) {
    final eventController = StreamController<ClientEvent>.broadcast();
    active = eventController;

    var connected = false;
    try {
      connected = await _runStream(addr, mainSendPort, eventController.stream);
    } catch (e) {
      print('[fugo] stream error: $e');
    } finally {
      active = null;
      await eventController.close();
    }

    // After a session that actually connected (e.g. the server restarted for a
    // hot reload), reconnect promptly; only back off when we never got through.
    delay = connected
        ? _reconnectDelay
        : (delay < _maxBackoff ? delay * 2 : _maxBackoff);
    print('[fugo] reconnecting in ${delay.inMilliseconds}ms...');
    await Future.delayed(delay);
  }
}

// _runStream opens one render stream and pumps payloads to the main isolate
// until it ends or errors. It returns whether the connection was established
// (at least one payload received), so the caller can reconnect quickly after a
// hot reload instead of backing off. It never throws.
Future<bool> _runStream(
  String addr,
  SendPort mainSendPort,
  Stream<ClientEvent> events,
) async {
  final parts = addr.split(':');
  final host = parts[0];
  final port = int.parse(parts[1]);

  final channel = ClientChannel(
    host,
    port: port,
    options: const ChannelOptions(
      credentials: ChannelCredentials.insecure(),
    ),
  );

  final client = FugoRenderClient(channel);

  final token = Platform.environment['FUGO_TOKEN'];
  final options = (token != null && token.isNotEmpty)
      ? CallOptions(metadata: {'x-fugo-token': token})
      : null;

  var connected = false;
  try {
    final stream = client.renderStream(events, options: options);
    final authNote = (token != null && token.isNotEmpty) ? ' (authenticated)' : '';
    print('[fugo] connected to $addr$authNote');

    await for (final payload in stream) {
      connected = true;
      mainSendPort.send(payload.writeToBuffer());
    }
    print('[fugo] stream ended cleanly');
  } on GrpcError catch (e) {
    print('[fugo] grpc error: $e');
  } catch (e) {
    print('[fugo] stream error: $e');
  } finally {
    await channel.shutdown();
    print('[fugo] disconnected');
  }

  return connected;
}
