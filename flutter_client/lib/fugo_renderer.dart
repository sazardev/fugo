import 'package:flutter/material.dart';

import 'generated/fugo/v1/fugo.pb.dart';
import 'events.dart';
import 'registry.dart';

class FugoApp extends StatelessWidget {
  final GlobalKey<FugoRendererState> rendererKey;

  const FugoApp({super.key, required this.rendererKey});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      home: Scaffold(
        backgroundColor: const Color(0xFF1E1E1E),
        body: FugoRenderer(key: rendererKey),
      ),
    );
  }
}

class FugoRenderer extends StatefulWidget {
  const FugoRenderer({super.key});

  @override
  FugoRendererState createState() => FugoRendererState();
}

class FugoRendererState extends State<FugoRenderer> {
  final Map<int, WidgetNode> _widgetMap = {};
  int? _rootId;
  final _registry = WidgetRegistry();

  void applyFullTree(WidgetTree tree) {
    setState(() {
      _widgetMap.clear();
      for (final node in tree.nodes) {
        _widgetMap[node.id] = node;
      }
      _rootId = tree.root;
    });
  }

  void applyPatches(PatchList patchList) {
    setState(() {
      for (final patch in patchList.patches) {
        switch (patch.op) {
          case PatchOp.PATCH_CREATE:
            if (patch.hasNode()) {
              _widgetMap[patch.node.id] = patch.node;
            }
            break;
          case PatchOp.PATCH_UPDATE:
            if (patch.hasProps() && _widgetMap.containsKey(patch.nodeId)) {
              _widgetMap[patch.nodeId]!.props = patch.props;
            }
            break;
          case PatchOp.PATCH_DELETE:
            _deleteRecursive(patch.nodeId);
            break;
          case PatchOp.PATCH_REPLACE:
            _deleteRecursive(patch.nodeId);
            if (patch.hasNode()) {
              _widgetMap[patch.node.id] = patch.node;
            }
            break;
          case PatchOp.PATCH_REORDER:
            if (_widgetMap.containsKey(patch.nodeId) && patch.children.isNotEmpty) {
              _widgetMap[patch.nodeId]!.children.clear();
              _widgetMap[patch.nodeId]!.children.addAll(patch.children);
            }
            break;
        }
      }
    });
  }

  void _deleteRecursive(int nodeId) {
    final node = _widgetMap.remove(nodeId);
    if (node != null) {
      for (final childId in node.children) {
        _deleteRecursive(childId);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final rootId = _rootId;
    if (rootId == null) {
      return const Center(
        child: CircularProgressIndicator(color: Colors.white),
      );
    }
    return _buildNode(rootId);
  }

  Widget _buildNode(int id) {
    final node = _widgetMap[id];
    if (node == null) return const SizedBox.shrink();

    final children = node.children
        .map((childId) => _buildNode(childId))
        .toList();

    return _registry.build(node, children);
  }
}
