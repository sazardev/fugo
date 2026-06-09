import 'package:flutter/material.dart';

import 'generated/fugo/v1/fugo.pb.dart';
import 'registry.dart';

class FugoApp extends StatelessWidget {
  final GlobalKey<FugoRendererState> rendererKey;
  final GlobalKey<ScaffoldMessengerState> messengerKey;
  final GlobalKey<NavigatorState> navigatorKey;
  final Color seedColor;
  final Brightness brightness;

  const FugoApp({
    super.key,
    required this.rendererKey,
    required this.messengerKey,
    required this.navigatorKey,
    required this.seedColor,
    required this.brightness,
  });

  @override
  Widget build(BuildContext context) {
    // The Material 3 ColorScheme is derived from the seed + brightness that Go
    // sends (FUGO_THEME_SEED / FUGO_THEME_BRIGHTNESS). We flatten the
    // seed-tinted surfaces to a clean neutral background for a minimal look —
    // the seed still colors interactive elements (buttons, FAB, switches).
    final scheme = ColorScheme.fromSeed(
      seedColor: seedColor,
      brightness: brightness,
    );
    final background = brightness == Brightness.light
        ? Colors.white
        : const Color(0xFF111315);
    final theme = ThemeData(
      useMaterial3: true,
      colorScheme: scheme.copyWith(surface: background),
      scaffoldBackgroundColor: background,
      appBarTheme: AppBarTheme(
        backgroundColor: background,
        surfaceTintColor: Colors.transparent,
        scrolledUnderElevation: 0,
      ),
    );

    // The outer surface is a plain Material (not a Scaffold) so an app that
    // uses fg.Scaffold isn't nested inside a second Scaffold — nested Scaffolds
    // misroute a FloatingActionButton's gestures. A Scaffold-less app still
    // gets a Material ancestor (for ink/buttons) and the themed background.
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      scaffoldMessengerKey: messengerKey,
      navigatorKey: navigatorKey,
      theme: theme,
      home: Material(
        color: background,
        child: SafeArea(child: FugoRenderer(key: rendererKey)),
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

  // Root widget types that already occupy the whole viewport. A root of any
  // other (intrinsically-sized) type is wrapped in a Center so simple content
  // — e.g. a bare Column — lands in the middle of the window automatically.
  static const _fillTypes = <WidgetType>{
    WidgetType.SCAFFOLD,
    WidgetType.CONTAINER,
    WidgetType.LISTVIEW,
    WidgetType.GRIDVIEW,
    WidgetType.STACK,
    WidgetType.ALIGN,
    WidgetType.CENTER,
    WidgetType.SCROLLVIEW,
    WidgetType.EXPANDED,
    WidgetType.ANIMATEDCONTAINER,
    WidgetType.TABS,
  };

  @override
  Widget build(BuildContext context) {
    final rootId = _rootId;
    if (rootId == null) {
      return const Center(child: CircularProgressIndicator());
    }

    final root = _buildNode(rootId);
    final rootType = _widgetMap[rootId]?.type;
    if (rootType != null && _fillTypes.contains(rootType)) {
      return root;
    }
    return Center(child: root);
  }

  Widget _buildNode(int id) {
    final node = _widgetMap[id];
    if (node == null) {
      print('[fugo] _buildNode: node $id not found in map');
      return const SizedBox.shrink();
    }

    final children = node.children
        .map((childId) => _buildNode(childId))
        .toList();

    try {
      return _registry.build(context, node, children);
    } catch (e, stack) {
      print('[fugo] _buildNode error for id=$id type=${node.type}: $e');

      return const SizedBox.shrink();
    }
  }
}
