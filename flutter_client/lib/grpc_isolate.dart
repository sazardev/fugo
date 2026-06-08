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
  var delay = _reconnectDelay;

  while (true) {
    try {
      await _runStream(addr, mainSendPort, receivePort);
    } catch (e) {
      print('[fugo] stream error: $e');
    }

    delay = delay < _maxBackoff ? delay * 2 : _maxBackoff;
    await Future.delayed(delay);
    print('[fugo] reconnecting in ${delay.inMilliseconds}ms...');
  }
}

Future<void> _runStream(
  String addr,
  SendPort mainSendPort,
  ReceivePort receivePort,
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
  final eventController = StreamController<ClientEvent>.broadcast();

  final eventSub = receivePort.listen((message) {
    if (message is List<int>) {
      final event = ClientEvent.fromBuffer(Uint8List.fromList(message));
      print('[fugo] isolate: received event ${event.eventType} node=${event.nodeId}');
      eventController.add(event);
    }
  });

  try {
    final stream = client.renderStream(eventController.stream);
    print('[fugo] connected to $addr');

    await for (final payload in stream) {
      mainSendPort.send(payload.writeToBuffer());
    }
    print('[fugo] stream ended cleanly');
  } on GrpcError catch (e) {
    print('[fugo] grpc error: $e');
    rethrow;
  } finally {
    await eventSub.cancel();
    await eventController.close();
    channel.shutdown();
    print('[fugo] disconnected');
  }
}
