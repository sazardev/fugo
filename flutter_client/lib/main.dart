import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

import 'generated/fugo/v1/fugo.pb.dart';
import 'events.dart';
import 'fugo_renderer.dart';
import 'grpc_isolate.dart';

final _fugoRendererKey = GlobalKey<FugoRendererState>();

// applyWindowCommand applies a runtime window-control command from Go via the
// OS window manager (driven by WindowController on the Go side).
Future<void> applyWindowCommand(WindowCommand cmd) async {
  switch (cmd.op) {
    case WindowOp.WINDOW_SET_TITLE:
      await windowManager.setTitle(cmd.title);
      break;
    case WindowOp.WINDOW_SET_SIZE:
      await windowManager.setSize(Size(cmd.width, cmd.height));
      break;
    case WindowOp.WINDOW_MINIMIZE:
      await windowManager.minimize();
      break;
    case WindowOp.WINDOW_MAXIMIZE:
      await windowManager.maximize();
      break;
    case WindowOp.WINDOW_CENTER:
      await windowManager.center();
      break;
    case WindowOp.WINDOW_FULLSCREEN:
      await windowManager.setFullScreen(cmd.flag);
      break;
    default:
      break;
  }
}

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await windowManager.ensureInitialized();

  final title = Platform.environment['FUGO_TITLE'] ?? 'Fugo';
  final width = double.tryParse(Platform.environment['FUGO_WIDTH'] ?? '') ?? 800;
  final height = double.tryParse(Platform.environment['FUGO_HEIGHT'] ?? '') ?? 600;

  final windowOptions = WindowOptions(
    size: Size(width, height),
    center: true,
    title: title,
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

        if (payload.hasWindow()) {
          applyWindowCommand(payload.window);
          return;
        }

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
