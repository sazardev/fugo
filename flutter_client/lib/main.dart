import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:grpc/grpc.dart';
import 'package:window_manager/window_manager.dart';

import 'generated/fugo/v1/fugo.pbgrpc.dart';
import 'generated/fugo/v1/fugo.pb.dart';
import 'events.dart';
import 'fugo_renderer.dart';

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

  final addr = Platform.environment['FUGO_ADDR'] ?? '127.0.0.1:9510';
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

  runApp(FugoApp(rendererKey: _fugoRendererKey));

  final responseStream = client.renderStream(eventStream);

  try {
    await for (final payload in responseStream) {
      final state = _fugoRendererKey.currentState;
      if (state == null) continue;

      if (payload.hasFullTree()) {
        state.applyFullTree(payload.fullTree);
      } else if (payload.hasPatches()) {
        state.applyPatches(payload.patches);
      }
    }
  } catch (e) {
    debugPrint('[fugo] stream closed: $e');
  }
}
