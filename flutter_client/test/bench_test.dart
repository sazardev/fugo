// Dart-side performance benchmarks for the render client's hot path: decoding
// the protobuf wire format (which happens for every payload) and building
// Flutter widgets from nodes via the registry. These mirror the Go engine
// benchmarks (engine/differ_bench_test.go) so Phase G has coverage on both
// sides of the wire.
//
// Run with: flutter test test/bench_test.dart
//
// The printed ns/op numbers are the useful output; the assertions are loose
// ceilings that only fire on a catastrophic regression, so they stay stable on
// shared CI runners rather than chasing wall-clock noise.
import 'dart:typed_data';

import 'package:fixnum/fixnum.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:fugo_flutter_client/generated/fugo/v1/fugo.pb.dart';
import 'package:fugo_flutter_client/registry.dart';

/// Builds a flat WidgetTree of [n] Text nodes under a single Column root,
/// matching the shape the Go side emits.
WidgetTree makeTree(int n) {
  final nodes = <WidgetNode>[];
  final rootChildren = <int>[];

  for (var i = 1; i <= n; i++) {
    final props = TextProps(value: 'item $i', fontSize: 14, color: '#FFFFFF');
    nodes.add(WidgetNode(
      id: i + 1,
      type: WidgetType.TEXT,
      props: props.writeToBuffer(),
    ));
    rootChildren.add(i + 1);
  }

  nodes.insert(
    0,
    WidgetNode(id: 1, type: WidgetType.COLUMN, children: rootChildren),
  );

  return WidgetTree(root: 1, nodes: nodes);
}

/// times runs [body] [iterations] times and reports nanoseconds per op.
int times(String label, int iterations, void Function() body) {
  // Warm up the JIT / fill caches before measuring.
  for (var i = 0; i < 100; i++) {
    body();
  }

  final sw = Stopwatch()..start();
  for (var i = 0; i < iterations; i++) {
    body();
  }
  sw.stop();

  final nsPerOp = (sw.elapsedMicroseconds * 1000) ~/ iterations;
  debugPrint('[bench] $label: $nsPerOp ns/op ($iterations iterations)');
  return nsPerOp;
}

void main() {
  test('decode full tree (1000 nodes)', () {
    final bytes = Uint8List.fromList(
      RenderPayload(fullTree: makeTree(1000)).writeToBuffer(),
    );

    final nsPerOp = times('RenderPayload.fromBuffer(1000 nodes)', 2000, () {
      RenderPayload.fromBuffer(bytes);
    });

    // Generous ceiling: a 1000-node payload decode should stay well under a
    // handful of frames even on a slow runner.
    expect(nsPerOp, lessThan(50 * 1000 * 1000)); // 50ms
  });

  test('decode patch batch (100 UPDATE patches)', () {
    final patches = <Patch>[];
    for (var i = 1; i <= 100; i++) {
      final props = TextProps(value: 'updated $i', fontSize: 16);
      patches.add(Patch(
        op: PatchOp.PATCH_UPDATE,
        nodeId: i + 1,
        props: props.writeToBuffer(),
      ));
    }
    final bytes = Uint8List.fromList(
      RenderPayload(patches: PatchList(patches: patches, seqNum: Int64(1)))
          .writeToBuffer(),
    );

    final nsPerOp = times('RenderPayload.fromBuffer(100 patches)', 5000, () {
      RenderPayload.fromBuffer(bytes);
    });

    expect(nsPerOp, lessThan(20 * 1000 * 1000)); // 20ms
  });

  test('decode single TextProps', () {
    final bytes = Uint8List.fromList(
      TextProps(value: 'hello', fontSize: 14, color: '#FFFFFF', fontWeight: 700)
          .writeToBuffer(),
    );

    final nsPerOp = times('TextProps.fromBuffer', 50000, () {
      TextProps.fromBuffer(bytes);
    });

    expect(nsPerOp, lessThan(1 * 1000 * 1000)); // 1ms
  });

  testWidgets('registry builds a Text node', (tester) async {
    final registry = WidgetRegistry();
    final node = WidgetNode(
      id: 2,
      type: WidgetType.TEXT,
      props: TextProps(value: 'hello', fontSize: 14, color: '#FFFFFF')
          .writeToBuffer(),
    );

    late BuildContext ctx;
    await tester.pumpWidget(MaterialApp(
      home: Builder(builder: (context) {
        ctx = context;
        return const SizedBox.shrink();
      }),
    ));

    final nsPerOp = times('WidgetRegistry.build(Text)', 20000, () {
      registry.build(ctx, node, const []);
    });

    expect(nsPerOp, lessThan(5 * 1000 * 1000)); // 5ms
  });
}
