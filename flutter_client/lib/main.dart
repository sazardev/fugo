import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';

import 'package:file_selector/file_selector.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:local_notifier/local_notifier.dart';
import 'package:window_manager/window_manager.dart';

import 'generated/fugo/v1/fugo.pb.dart'
    hide MainAxisSize, MainAxisAlignment, CrossAxisAlignment;
import 'events.dart';
import 'fugo_renderer.dart';
import 'grpc_isolate.dart';
import 'registry.dart' show hexToColor, requestFocus;

final _fugoRendererKey = GlobalKey<FugoRendererState>();
final _messengerKey = GlobalKey<ScaffoldMessengerState>();
final _navigatorKey = GlobalKey<NavigatorState>();

// _shortcutBindings is the current app-wide keyboard shortcut set, replaced
// wholesale whenever Go sends a ShortcutsCommand (see applyShortcutsCommand).
Set<String> _shortcutBindings = {};

void applyShortcutsCommand(ShortcutsCommand cmd) {
  _shortcutBindings = cmd.bindings.toSet();
}

// _handleKeyEvent is registered globally via HardwareKeyboard, independent of
// the focus tree, so a shortcut fires no matter which widget (if any) has
// focus. Caveat: it runs alongside normal focus-based key handling (e.g. a
// focused TextField still receives the character), so binding a shortcut to
// a bare printable key with no modifier while a text field is focused would
// both type the character AND fire the shortcut — in practice this only
// matters for modifier-less bindings, which most real shortcuts (ctrl+s,
// escape, ctrl+z) aren't.
bool _handleKeyEvent(KeyEvent event) {
  if (event is! KeyDownEvent) return false;
  if (_isModifierKey(event.logicalKey)) return false;

  final label = _keyLabel(event.logicalKey);
  if (label == null) return false;

  final parts = <String>[];
  if (HardwareKeyboard.instance.isControlPressed) parts.add('ctrl');
  if (HardwareKeyboard.instance.isAltPressed) parts.add('alt');
  if (HardwareKeyboard.instance.isShiftPressed) parts.add('shift');
  if (HardwareKeyboard.instance.isMetaPressed) parts.add('meta');
  parts.add(label);

  final binding = parts.join('+');
  if (!_shortcutBindings.contains(binding)) return false;

  sendEvent(ClientEvent(
    nodeId: '',
    eventType: 'shortcut',
    eventData: binding.codeUnits,
  ));

  return true;
}

bool _isModifierKey(LogicalKeyboardKey key) {
  // Not `const`: LogicalKeyboardKey overrides ==/hashCode, which the analyzer
  // disallows in const collection literals.
  final modifiers = {
    LogicalKeyboardKey.control,
    LogicalKeyboardKey.controlLeft,
    LogicalKeyboardKey.controlRight,
    LogicalKeyboardKey.alt,
    LogicalKeyboardKey.altLeft,
    LogicalKeyboardKey.altRight,
    LogicalKeyboardKey.shift,
    LogicalKeyboardKey.shiftLeft,
    LogicalKeyboardKey.shiftRight,
    LogicalKeyboardKey.meta,
    LogicalKeyboardKey.metaLeft,
    LogicalKeyboardKey.metaRight,
  };

  return modifiers.contains(key);
}

// _keyLabel normalizes a logical key to the lowercase token Go's bindings use
// (e.g. LogicalKeyboardKey.keyS -> "s", arrowUp -> "up"); null for keys with
// no normalized name (not meant to be bindable).
String? _keyLabel(LogicalKeyboardKey key) {
  // Not `const` — see _isModifierKey.
  final named = {
    LogicalKeyboardKey.escape: 'escape',
    LogicalKeyboardKey.enter: 'enter',
    LogicalKeyboardKey.tab: 'tab',
    LogicalKeyboardKey.delete: 'delete',
    LogicalKeyboardKey.backspace: 'backspace',
    LogicalKeyboardKey.space: 'space',
    LogicalKeyboardKey.arrowUp: 'up',
    LogicalKeyboardKey.arrowDown: 'down',
    LogicalKeyboardKey.arrowLeft: 'left',
    LogicalKeyboardKey.arrowRight: 'right',
    LogicalKeyboardKey.f1: 'f1',
    LogicalKeyboardKey.f2: 'f2',
    LogicalKeyboardKey.f3: 'f3',
    LogicalKeyboardKey.f4: 'f4',
    LogicalKeyboardKey.f5: 'f5',
  };
  if (named.containsKey(key)) return named[key];

  final label = key.keyLabel;
  if (label.length == 1) return label.toLowerCase();

  return null;
}

// _ResizeReporter debounces window-resize notifications back to Go (a resize
// drag fires many raw metric changes per second) — there is no other channel
// for Go to learn the client's actual viewport size, since the widget tree is
// built before anything is measured.
class _ResizeReporter with WidgetsBindingObserver {
  Timer? _debounce;

  @override
  void didChangeMetrics() {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 250), () {
      final view = WidgetsBinding.instance.platformDispatcher.views.first;
      final size = view.physicalSize / view.devicePixelRatio;

      sendEvent(ClientEvent(
        nodeId: '',
        eventType: 'resize',
        eventData:
            '${size.width.toStringAsFixed(0)}x${size.height.toStringAsFixed(0)}'
                .codeUnits,
      ));
    });
  }
}

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
      final result = await openFile(
        acceptedTypeGroups: cmd.extensions.isNotEmpty
            ? [XTypeGroup(extensions: cmd.extensions)]
            : const [],
      );
      _replyHost(requestId, result?.path ?? '');
      break;
    case HostOp.HOST_FILE_SAVE:
      final location = await getSaveLocation(
        suggestedName: cmd.defaultName.isNotEmpty ? cmd.defaultName : null,
        acceptedTypeGroups: cmd.extensions.isNotEmpty
            ? [XTypeGroup(extensions: cmd.extensions)]
            : const [],
      );
      _replyHost(requestId, location?.path ?? '');
      break;
    case HostOp.HOST_NOTIFICATION:
      LocalNotification(title: cmd.title, body: cmd.text).show();
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
        final id = cmd.requestId.toInt();
        if (ctx == null) {
          if (cmd.actions.isNotEmpty) _replyHost(id, '');
          break;
        }
        // cmd.actions.isNotEmpty renders custom action buttons (e.g.
        // "Cancel"/"Delete") instead of the default single "OK"; whichever
        // one is tapped becomes the dialog's pop result, and the single
        // .then() below is the one place that replies — so an action button
        // is never double-replied (once by itself, once by dismissal).
        showDialog<String>(
          context: ctx,
          builder: (c) => AlertDialog(
            title: cmd.title.isNotEmpty ? Text(cmd.title) : null,
            content: cmd.message.isNotEmpty ? Text(cmd.message) : null,
            actions: cmd.actions.isNotEmpty
                ? cmd.actions
                    .map((a) => TextButton(
                          onPressed: () => Navigator.of(c).pop(a),
                          child: Text(a),
                        ))
                    .toList()
                : [
                    TextButton(
                      onPressed: () => Navigator.of(c).pop(),
                      child: const Text('OK'),
                    ),
                  ],
          ),
        ).then((result) {
          if (cmd.actions.isNotEmpty) _replyHost(id, result ?? '');
        });
      }
      break;
    case OverlayOp.OVERLAY_BOTTOMSHEET:
      {
        final ctx = _navigatorKey.currentContext;
        final id = cmd.requestId.toInt();
        if (ctx == null) {
          if (cmd.actions.isNotEmpty) _replyHost(id, '');
          break;
        }
        showModalBottomSheet<String>(
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
                if (cmd.actions.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 16),
                    child: Wrap(
                      spacing: 8,
                      children: cmd.actions
                          .map((a) => FilledButton(
                                onPressed: () => Navigator.of(c).pop(a),
                                child: Text(a),
                              ))
                          .toList(),
                    ),
                  ),
              ],
            ),
          ),
        ).then((result) {
          if (cmd.actions.isNotEmpty) _replyHost(id, result ?? '');
        });
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
  await localNotifier.setup(appName: title);
  final width = double.tryParse(Platform.environment['FUGO_WIDTH'] ?? '') ?? 800;
  final height = double.tryParse(Platform.environment['FUGO_HEIGHT'] ?? '') ?? 600;

  // Material 3 theme, seeded by Go (FUGO_THEME_SEED / FUGO_THEME_BRIGHTNESS).
  final seedColor = hexToColor(
    Platform.environment['FUGO_THEME_SEED'] ?? '#2563EB',
  );
  final brightness = Platform.environment['FUGO_THEME_BRIGHTNESS'] == 'dark'
      ? Brightness.dark
      : Brightness.light;
  final followSystemTheme = Platform.environment['FUGO_THEME_FOLLOW_SYSTEM'] == '1';
  final fontFamily = Platform.environment['FUGO_THEME_FONT_FAMILY'];

  final windowOptions = WindowOptions(
    size: Size(width, height),
    center: true,
    title: title,
  );

  await windowManager.waitUntilReadyToShow(windowOptions, () async {
    await windowManager.show();
    await windowManager.focus();
  });

  HardwareKeyboard.instance.addHandler(_handleKeyEvent);
  WidgetsBinding.instance.addObserver(_ResizeReporter());

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
        followSystem: followSystemTheme,
        fontFamily: fontFamily,
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

        if (payload.hasShortcuts()) {
          applyShortcutsCommand(payload.shortcuts);
          return;
        }

        if (payload.hasFocus()) {
          requestFocus(payload.focus.nodeId.toInt());
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
