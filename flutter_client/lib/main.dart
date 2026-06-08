import 'dart:isolate';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

import 'generated/fugo/v1/fugo.pb.dart';
import 'events.dart';
import 'fugo_renderer.dart';
import 'grpc_isolate.dart';

final _fugoRendererKey = GlobalKey<FugoRendererState>();

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await windowManager.ensureInitialized();

  const windowOptions = WindowOptions(
    size: Size(800, 600),
    center: true,
    title: 'Fugo',
  );

  await windowManager.waitUntilReadyToShow(windowOptions, () async {
    await windowManager.show();
    await windowManager.focus();
  });

  final receivePort = ReceivePort();
  await Isolate.spawn(grpcIsolateEntry, receivePort.sendPort);

  var firstMessage = true;

  receivePort.listen((message) {
    if (firstMessage) {
      firstMessage = false;
      setEventSendPort(message as SendPort);
      runApp(FugoApp(rendererKey: _fugoRendererKey));

      return;
    }

    if (message is List<int>) {
      try {
        final payload = RenderPayload.fromBuffer(Uint8List.fromList(message));
        final state = _fugoRendererKey.currentState;
        if (state == null) return;

        if (payload.hasFullTree()) {
          state.applyFullTree(payload.fullTree);
        } else if (payload.hasPatches()) {
          state.applyPatches(payload.patches);
        }
      } catch (e, stack) {
        print('[fugo] error in message handler: $e');
        print('[fugo] stack: $stack');
      }
    }
  });
}
