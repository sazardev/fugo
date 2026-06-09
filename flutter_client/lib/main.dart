import 'dart:convert';
import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
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

// applyHostCommand fulfills an out-of-band host-service request from Go
// (clipboard access, native file dialog) and, for requests that expect a reply,
// sends the result back as a "host" ClientEvent keyed by the request id.
Future<void> applyHostCommand(HostCommand cmd) async {
  final requestId = cmd.requestId.toInt();
  switch (cmd.op) {
    case HostOp.HOST_CLIPBOARD_WRITE:
      await Clipboard.setData(ClipboardData(text: cmd.text));
      break;
    case HostOp.HOST_CLIPBOARD_READ:
      final data = await Clipboard.getData(Clipboard.kTextPlain);
      _replyHost(requestId, data?.text ?? '');
      break;
    case HostOp.HOST_FILE_OPEN:
      final result = await FilePicker.pickFiles(
        dialogTitle: cmd.text.isNotEmpty ? cmd.text : null,
        type: cmd.extensions.isNotEmpty ? FileType.custom : FileType.any,
        allowedExtensions: cmd.extensions.isNotEmpty ? cmd.extensions : null,
      );
      _replyHost(requestId, result?.files.single.path ?? '');
      break;
    case HostOp.HOST_FILE_SAVE:
      final path = await FilePicker.saveFile(
        dialogTitle: cmd.text.isNotEmpty ? cmd.text : null,
        fileName: cmd.defaultName.isNotEmpty ? cmd.defaultName : null,
        type: cmd.extensions.isNotEmpty ? FileType.custom : FileType.any,
        allowedExtensions: cmd.extensions.isNotEmpty ? cmd.extensions : null,
      );
      _replyHost(requestId, path ?? '');
      break;
    default:
      break;
  }
}

// _replyHost sends the result of a host request back to Go. A request id of 0
// means fire-and-forget (e.g. a clipboard write), so no reply is sent.
void _replyHost(int requestId, String result) {
  if (requestId == 0) return;
  sendEvent(ClientEvent(
    nodeId: requestId.toString(),
    eventType: 'host',
    eventData: utf8.encode(result),
  ));
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

        if (payload.hasHost()) {
          applyHostCommand(payload.host);
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
