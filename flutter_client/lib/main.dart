import 'dart:convert';
import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:window_manager/window_manager.dart';

import 'generated/fugo/v1/fugo.pb.dart'
    hide MainAxisSize, MainAxisAlignment, CrossAxisAlignment;
import 'events.dart';
import 'fugo_renderer.dart';
import 'grpc_isolate.dart';
import 'registry.dart' show hexToColor;

final _fugoRendererKey = GlobalKey<FugoRendererState>();
final _messengerKey = GlobalKey<ScaffoldMessengerState>();
final _navigatorKey = GlobalKey<NavigatorState>();

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

// applyOverlayCommand shows a transient overlay (snackbar or dialog) requested
// by Go, using the app-level messenger/navigator keys so it works without a
// widget build context.
void applyOverlayCommand(OverlayCommand cmd) {
  switch (cmd.op) {
    case OverlayOp.OVERLAY_SNACKBAR:
      _messengerKey.currentState
          ?.showSnackBar(SnackBar(content: Text(cmd.message)));
      break;
    case OverlayOp.OVERLAY_DIALOG:
      {
        final ctx = _navigatorKey.currentContext;
        if (ctx != null) {
          showDialog<void>(
            context: ctx,
            builder: (c) => AlertDialog(
              title: cmd.title.isNotEmpty ? Text(cmd.title) : null,
              content: cmd.message.isNotEmpty ? Text(cmd.message) : null,
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(c).pop(),
                  child: const Text('OK'),
                ),
              ],
            ),
          );
        }
      }
      break;
    case OverlayOp.OVERLAY_BOTTOMSHEET:
      {
        final ctx = _navigatorKey.currentContext;
        if (ctx != null) {
          showModalBottomSheet<void>(
            context: ctx,
            builder: (c) => Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (cmd.title.isNotEmpty)
                    Text(cmd.title, style: Theme.of(c).textTheme.titleLarge),
                  if (cmd.message.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.only(top: 8),
                      child: Text(cmd.message),
                    ),
                ],
              ),
            ),
          );
        }
      }
      break;
    case OverlayOp.OVERLAY_DATE_PICKER:
      {
        final id = cmd.requestId.toInt();
        final ctx = _navigatorKey.currentContext;
        if (ctx != null) {
          showDatePicker(
            context: ctx,
            initialDate: DateTime.now(),
            firstDate: DateTime(1900),
            lastDate: DateTime(2100),
          ).then((d) => _replyHost(
              id, d != null ? d.toIso8601String().split('T').first : ''));
        } else {
          _replyHost(id, '');
        }
      }
      break;
    case OverlayOp.OVERLAY_TIME_PICKER:
      {
        final id = cmd.requestId.toInt();
        final ctx = _navigatorKey.currentContext;
        if (ctx != null) {
          showTimePicker(context: ctx, initialTime: TimeOfDay.now()).then((t) {
            final s = t != null
                ? '${t.hour.toString().padLeft(2, '0')}:${t.minute.toString().padLeft(2, '0')}'
                : '';
            _replyHost(id, s);
          });
        } else {
          _replyHost(id, '');
        }
      }
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

  // Material 3 theme, seeded by Go (FUGO_THEME_SEED / FUGO_THEME_BRIGHTNESS).
  final seedColor = hexToColor(
    Platform.environment['FUGO_THEME_SEED'] ?? '#2563EB',
  );
  final brightness = Platform.environment['FUGO_THEME_BRIGHTNESS'] == 'dark'
      ? Brightness.dark
      : Brightness.light;

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
      runApp(FugoApp(
        rendererKey: _fugoRendererKey,
        messengerKey: _messengerKey,
        navigatorKey: _navigatorKey,
        seedColor: seedColor,
        brightness: brightness,
      ));

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

        if (payload.hasOverlay()) {
          applyOverlayCommand(payload.overlay);
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
