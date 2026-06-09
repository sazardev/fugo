import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

import 'generated/fugo/v1/fugo.pb.dart' as proto;
import 'events.dart';
import 'icons_gen.dart';

class WidgetRegistry {
  Widget build(BuildContext context, proto.WidgetNode node, List<Widget> children) {
    switch (node.type) {
      case proto.WidgetType.TEXT:
        return _buildText(context, node);
      case proto.WidgetType.CONTAINER:
        return _buildContainer(node, children);
      case proto.WidgetType.COLUMN:
        return _buildColumn(node, children);
      case proto.WidgetType.CENTER:
        return _buildCenter(children);
      case proto.WidgetType.BUTTON:
        return _buildButton(context, node);
      case proto.WidgetType.ROW:
        return _buildRow(node, children);
      case proto.WidgetType.STACK:
        return _buildStack(children);
      case proto.WidgetType.EXPANDED:
        return _buildExpanded(node, children);
      case proto.WidgetType.PADDING:
        return _buildPadding(node, children);
      case proto.WidgetType.SIZEDBOX:
        return _buildSizedBox(node, children);
      case proto.WidgetType.IMAGE:
        return _buildImage(node);
      case proto.WidgetType.TEXTFIELD:
        return _buildTextField(node);
      case proto.WidgetType.POSITIONED:
        return _buildPositioned(node, children);
      case proto.WidgetType.CHECKBOX:
        return _buildCheckbox(node);
      case proto.WidgetType.SWITCH_WIDGET:
        return _buildSwitch(node);
      case proto.WidgetType.SLIDER:
        return _buildSlider(node);
      case proto.WidgetType.LISTVIEW:
        return _buildListView(node, children);
      case proto.WidgetType.ANIMATEDCONTAINER:
        return _buildAnimatedContainer(node, children);
      case proto.WidgetType.ICON:
        return _buildIcon(node);
      case proto.WidgetType.DIVIDER:
        return _buildDivider(node);
      case proto.WidgetType.WRAP:
        return _buildWrap(node, children);
      case proto.WidgetType.GRIDVIEW:
        return _buildGridView(node, children);
      case proto.WidgetType.ANIMATEDOPACITY:
        return _buildAnimatedOpacity(node, children);
      case proto.WidgetType.SCROLLVIEW:
        return _buildScrollView(node, children);
      case proto.WidgetType.GESTUREDETECTOR:
        return _buildGestureDetector(node, children);
      case proto.WidgetType.ALIGN:
        return _buildAlign(node, children);
      case proto.WidgetType.RADIO:
        return _buildRadio(context, node);
      case proto.WidgetType.DROPDOWN:
        return _buildDropdown(node);
      case proto.WidgetType.ANIMATEDPOSITIONED:
        return _buildAnimatedPositioned(node, children);
      case proto.WidgetType.WINDOWDRAGAREA:
        return _buildWindowDragArea(children);
      case proto.WidgetType.CARD:
        return _buildCard(node, children);
      case proto.WidgetType.SCAFFOLD:
        return _buildScaffold(node, children);
      case proto.WidgetType.APPBAR:
        return _buildAppBar(node, children);
      case proto.WidgetType.NAVIGATIONBAR:
        return _buildNavigationBar(node);
      case proto.WidgetType.TABS:
        return _buildTabs(node, children);
      case proto.WidgetType.TOOLTIP:
        return _buildTooltip(node, children);
      case proto.WidgetType.BADGE:
        return _buildBadge(node, children);
      case proto.WidgetType.AVATAR:
        return _buildAvatar(node);
      case proto.WidgetType.SEGMENTEDBUTTON:
        return _buildSegmentedButton(node);
      case proto.WidgetType.ASPECTRATIO:
        return _buildAspectRatio(node, children);
      case proto.WidgetType.CLIPRRECT:
        return _buildClipRRect(node, children);
      case proto.WidgetType.FITTEDBOX:
        return _buildFittedBox(children);
      case proto.WidgetType.FLEXIBLE:
        return _buildFlexible(node, children);
      case proto.WidgetType.EXPANSIONTILE:
        return _buildExpansionTile(node, children);
      case proto.WidgetType.POPUPMENU:
        return _buildPopupMenu(node);
      case proto.WidgetType.RICHTEXT:
        return _buildRichText(node);
      case proto.WidgetType.DATATABLE:
        return _buildDataTable(node);
      case proto.WidgetType.STEPPER:
        return _buildStepper(node, children);
      case proto.WidgetType.FLOATINGACTIONBUTTON:
        return _buildFab(node);
      case proto.WidgetType.LISTTILE:
        return _buildListTile(node);
      case proto.WidgetType.CHIP:
        return _buildChip(node);
      case proto.WidgetType.PROGRESS:
        return _buildProgress(node);
      default:
        return const SizedBox.shrink();
    }
  }

  Widget _buildText(BuildContext context, proto.WidgetNode node) {
    final props = proto.TextProps.fromBuffer(node.props);
    final base = Theme.of(context).textTheme.bodyMedium ?? const TextStyle();
    return Text(
      props.value,
      textAlign: props.hasTextAlign() ? _mapTextAlign(props.textAlign) : null,
      style: base.copyWith(
        fontSize: props.hasFontSize() ? props.fontSize : null,
        color: props.hasColor() ? hexToColor(props.color) : null,
        fontWeight:
            props.hasFontWeight() ? _mapFontWeight(props.fontWeight) : null,
      ),
    );
  }

  Widget _buildButton(BuildContext context, proto.WidgetNode node) {
    final props = proto.ButtonProps.fromBuffer(node.props);
    final onPressed = props.enabled
        ? () => sendEvent(proto.ClientEvent(
              nodeId: node.id.toString(),
              eventType: 'onClick',
            ))
        : null;
    final style = _buttonStyle(props);
    final hasIcon = props.icon.isNotEmpty;
    final icon = hasIcon ? Icon(_mapIconData(props.icon)) : null;
    final label = Text(props.label);

    switch (props.variant) {
      case proto.ButtonVariant.BUTTON_ICON:
        return IconButton(
          onPressed: onPressed,
          icon: Icon(_mapIconData(props.icon)),
          style: style,
        );
      case proto.ButtonVariant.BUTTON_FILLED_TONAL:
        return hasIcon
            ? FilledButton.tonalIcon(
                onPressed: onPressed, icon: icon!, label: label, style: style)
            : FilledButton.tonal(
                onPressed: onPressed, style: style, child: label);
      case proto.ButtonVariant.BUTTON_OUTLINED:
        return hasIcon
            ? OutlinedButton.icon(
                onPressed: onPressed, icon: icon!, label: label, style: style)
            : OutlinedButton(onPressed: onPressed, style: style, child: label);
      case proto.ButtonVariant.BUTTON_TEXT:
        return hasIcon
            ? TextButton.icon(
                onPressed: onPressed, icon: icon!, label: label, style: style)
            : TextButton(onPressed: onPressed, style: style, child: label);
      case proto.ButtonVariant.BUTTON_ELEVATED:
        return hasIcon
            ? ElevatedButton.icon(
                onPressed: onPressed, icon: icon!, label: label, style: style)
            : ElevatedButton(onPressed: onPressed, style: style, child: label);
      default:
        return hasIcon
            ? FilledButton.icon(
                onPressed: onPressed, icon: icon!, label: label, style: style)
            : FilledButton(onPressed: onPressed, style: style, child: label);
    }
  }

  // _buttonStyle returns explicit overrides only when the Go side set them;
  // otherwise null, so the Material 3 ColorScheme styles the button natively.
  ButtonStyle? _buttonStyle(proto.ButtonProps props) {
    final hasBg = props.bgColor.isNotEmpty;
    final hasRadius = props.borderRadius > 0;
    final hasFont = props.fontSize > 0;
    if (!hasBg && !hasRadius && !hasFont) {
      return null;
    }
    return ButtonStyle(
      backgroundColor:
          hasBg ? WidgetStatePropertyAll(hexToColor(props.bgColor)) : null,
      textStyle: hasFont
          ? WidgetStatePropertyAll(TextStyle(fontSize: props.fontSize))
          : null,
      shape: hasRadius
          ? WidgetStatePropertyAll(RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(props.borderRadius)))
          : null,
    );
  }

  Widget _buildContainer(proto.WidgetNode node, List<Widget> children) {
    final props = proto.ContainerProps.fromBuffer(node.props);
    final child = children.isNotEmpty ? children.first : const SizedBox.shrink();
    final color = props.hasBgColor() ? hexToColor(props.bgColor) : null;
    final radius = props.hasBorderRadius() ? props.borderRadius : 0.0;
    final decorated = color != null || radius > 0;
    final hasPadding = props.padTop != 0 ||
        props.padRight != 0 ||
        props.padBottom != 0 ||
        props.padLeft != 0;
    return Container(
      padding: hasPadding
          ? EdgeInsets.fromLTRB(
              props.padLeft,
              props.padTop,
              props.padRight,
              props.padBottom,
            )
          : null,
      decoration: decorated
          ? BoxDecoration(
              color: color,
              borderRadius:
                  radius > 0 ? BorderRadius.circular(radius) : null,
            )
          : null,
      child: child,
    );
  }

  Widget _buildColumn(proto.WidgetNode node, List<Widget> children) {
    final props = proto.ColumnProps.fromBuffer(node.props);

    return Column(
      mainAxisSize: _mapMainAxisSize(props.mainAxisSize),
      mainAxisAlignment: _mapMainAlign(props.mainAlignment),
      crossAxisAlignment: _mapCrossAlign(props.crossAlignment),
      children: children,
    );
  }

  Widget _buildCenter(List<Widget> children) {
    final child = children.isNotEmpty ? children.first : const SizedBox.shrink();
    return Center(child: child);
  }

  Widget _buildRow(proto.WidgetNode node, List<Widget> children) {
    final props = proto.RowProps.fromBuffer(node.props);

    return Row(
      mainAxisSize: _mapMainAxisSize(props.mainAxisSize),
      mainAxisAlignment: _mapMainAlign(props.mainAlignment),
      crossAxisAlignment: _mapCrossAlign(props.crossAlignment),
      children: children,
    );
  }

  Widget _buildStack(List<Widget> children) {
    return Stack(children: children);
  }

  Widget _buildExpanded(proto.WidgetNode node, List<Widget> children) {
    final props = proto.ExpandedProps.fromBuffer(node.props);
    final child = children.isNotEmpty ? children.first : const SizedBox.shrink();

    return Expanded(
      flex: props.hasFlex() ? props.flex : 1,
      child: child,
    );
  }

  Widget _buildPadding(proto.WidgetNode node, List<Widget> children) {
    final props = proto.PaddingProps.fromBuffer(node.props);
    final child = children.isNotEmpty ? children.first : const SizedBox.shrink();

    return Padding(
      padding: EdgeInsets.fromLTRB(
        props.left,
        props.top,
        props.right,
        props.bottom,
      ),
      child: child,
    );
  }

  Widget _buildSizedBox(proto.WidgetNode node, List<Widget> children) {
    final props = proto.SizedBoxProps.fromBuffer(node.props);
    final child = children.isNotEmpty ? children.first : null;

    return SizedBox(
      width: props.hasWidth() ? props.width : null,
      height: props.hasHeight() ? props.height : null,
      child: child,
    );
  }

  Widget _buildImage(proto.WidgetNode node) {
    final props = proto.ImageProps.fromBuffer(node.props);

    return Image.network(
      props.src,
      width: props.hasWidth() ? props.width : null,
      height: props.hasHeight() ? props.height : null,
      errorBuilder: (_, _, _) => const Icon(Icons.broken_image),
    );
  }

  Widget _buildTextField(proto.WidgetNode node) {
    final props = proto.TextFieldProps.fromBuffer(node.props);
    return TextField(
      obscureText: props.obscure,
      style: props.hasFontSize() ? TextStyle(fontSize: props.fontSize) : null,
      decoration: InputDecoration(
        hintText: props.placeholder,
        filled: true,
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
      ),
      onChanged: (value) => sendEvent(proto.ClientEvent(
        nodeId: node.id.toString(),
        eventType: 'onChange',
        eventData: value.codeUnits,
      )),
    );
  }

  Widget _buildPositioned(proto.WidgetNode node, List<Widget> children) {
    final props = proto.PositionedProps.fromBuffer(node.props);
    final child = children.isNotEmpty ? children.first : const SizedBox.shrink();

    return Positioned(
      left: props.hasLeft() ? props.left : null,
      top: props.hasTop() ? props.top : null,
      right: props.hasRight() ? props.right : null,
      bottom: props.hasBottom() ? props.bottom : null,
      width: props.hasWidth() ? props.width : null,
      height: props.hasHeight() ? props.height : null,
      child: child,
    );
  }

  Widget _buildCheckbox(proto.WidgetNode node) {
    final props = proto.CheckboxProps.fromBuffer(node.props);

    return Checkbox(
      value: props.checked,
      onChanged: (value) {
        sendEvent(proto.ClientEvent(
          nodeId: node.id.toString(),
          eventType: 'onChange',
          eventData: (value == true ? '1' : '0').codeUnits,
        ));
      },
    );
  }

  Widget _buildSwitch(proto.WidgetNode node) {
    final props = proto.SwitchProps.fromBuffer(node.props);

    return Switch(
      value: props.value,
      onChanged: (value) {
        sendEvent(proto.ClientEvent(
          nodeId: node.id.toString(),
          eventType: 'onChange',
          eventData: (value ? '1' : '0').codeUnits,
        ));
      },
    );
  }

  Widget _buildSlider(proto.WidgetNode node) {
    final props = proto.SliderProps.fromBuffer(node.props);

    return Slider(
      value: props.hasValue() ? props.value : 0,
      min: props.hasMin() ? props.min : 0,
      max: props.hasMax() ? props.max : 100,
      onChanged: (value) {
        sendEvent(proto.ClientEvent(
          nodeId: node.id.toString(),
          eventType: 'onChange',
          eventData: value.toStringAsFixed(2).codeUnits,
        ));
      },
    );
  }

  Widget _buildListView(proto.WidgetNode node, List<Widget> children) {
    final props = proto.ListViewProps.fromBuffer(node.props);

    if (props.hasItemExtent()) {
      return ListView.builder(
        itemExtent: props.itemExtent,
        itemCount: children.length,
        itemBuilder: (context, index) => children[index],
      );
    }

    return ListView(children: children);
  }

  Widget _buildAnimatedContainer(proto.WidgetNode node, List<Widget> children) {
    final props = proto.AnimatedContainerProps.fromBuffer(node.props);
    final child = children.isNotEmpty ? children.first : const SizedBox.shrink();
    final duration = Duration(milliseconds: props.hasDurationMs() ? props.durationMs : 200);
    final radius = props.hasBorderRadius() ? props.borderRadius : 0.0;
    return AnimatedContainer(
      duration: duration,
      curve: props.hasCurve() ? _mapCurve(props.curve) : Curves.ease,
      padding: props.hasPadding() ? EdgeInsets.all(props.padding) : null,
      decoration: BoxDecoration(
        color: props.hasBgColor() ? hexToColor(props.bgColor) : null,
        borderRadius: radius > 0 ? BorderRadius.circular(radius) : null,
      ),
      child: child,
    );
  }

  MainAxisSize _mapMainAxisSize(proto.MainAxisSize size) {
    if (size == proto.MainAxisSize.MAIN_MIN) return MainAxisSize.min;
    return MainAxisSize.max;
  }

  MainAxisAlignment _mapMainAlign(proto.MainAxisAlignment align) {
    switch (align) {
      case proto.MainAxisAlignment.MAIN_END:
        return MainAxisAlignment.end;
      case proto.MainAxisAlignment.MAIN_CENTER:
        return MainAxisAlignment.center;
      case proto.MainAxisAlignment.MAIN_SPACE_BETWEEN:
        return MainAxisAlignment.spaceBetween;
      case proto.MainAxisAlignment.MAIN_SPACE_AROUND:
        return MainAxisAlignment.spaceAround;
      case proto.MainAxisAlignment.MAIN_SPACE_EVENLY:
        return MainAxisAlignment.spaceEvenly;
      default:
        return MainAxisAlignment.start;
    }
  }

  CrossAxisAlignment _mapCrossAlign(proto.CrossAxisAlignment align) {
    switch (align) {
      case proto.CrossAxisAlignment.CROSS_END:
        return CrossAxisAlignment.end;
      case proto.CrossAxisAlignment.CROSS_CENTER:
        return CrossAxisAlignment.center;
      case proto.CrossAxisAlignment.CROSS_STRETCH:
        return CrossAxisAlignment.stretch;
      default:
        return CrossAxisAlignment.start;
    }
  }

  // _mapFontWeight maps the numeric 100..900 weight to Flutter's FontWeight
  // (whose values run w100..w900 at indices 0..8).
  FontWeight _mapFontWeight(int weight) {
    final index = ((weight ~/ 100) - 1).clamp(0, 8);
    return FontWeight.values[index];
  }

  TextAlign _mapTextAlign(int align) {
    switch (align) {
      case 1:
        return TextAlign.center;
      case 2:
        return TextAlign.right;
      default:
        return TextAlign.left;
    }
  }

  Widget _buildIcon(proto.WidgetNode node) {
    final props = proto.IconProps.fromBuffer(node.props);

    return Icon(
      _mapIconData(props.name),
      size: props.hasSize() ? props.size : 24,
      color: props.hasColor() ? hexToColor(props.color) : Colors.white,
    );
  }

  Widget _buildDivider(proto.WidgetNode node) {
    final props = proto.DividerProps.fromBuffer(node.props);

    return Divider(
      thickness: props.hasThickness() ? props.thickness : 1,
      color: props.hasColor() ? hexToColor(props.color) : Colors.grey,
    );
  }

  Widget _buildWrap(proto.WidgetNode node, List<Widget> children) {
    final props = proto.WrapProps.fromBuffer(node.props);

    return Wrap(
      spacing: props.hasSpacing() ? props.spacing : 0,
      runSpacing: props.hasRunSpacing() ? props.runSpacing : 0,
      children: children,
    );
  }

  Widget _buildGridView(proto.WidgetNode node, List<Widget> children) {
    final props = proto.GridViewProps.fromBuffer(node.props);

    return GridView.count(
      crossAxisCount: props.hasCrossAxisCount() ? props.crossAxisCount : 2,
      childAspectRatio: props.hasChildAspectRatio()
          ? props.childAspectRatio
          : 1,
      children: children,
    );
  }

  Widget _buildAnimatedOpacity(proto.WidgetNode node, List<Widget> children) {
    final props = proto.AnimatedOpacityProps.fromBuffer(node.props);
    final child = children.isNotEmpty ? children.first : const SizedBox.shrink();
    final duration = Duration(
      milliseconds: props.hasDurationMs() ? props.durationMs : 200,
    );

    return AnimatedOpacity(
      opacity: props.hasOpacity() ? props.opacity : 1,
      duration: duration,
      child: child,
    );
  }

  Widget _buildScrollView(proto.WidgetNode node, List<Widget> children) {
    final props = proto.ScrollViewProps.fromBuffer(node.props);

    return SingleChildScrollView(
      scrollDirection: props.scrollDirection == 1
          ? Axis.horizontal
          : Axis.vertical,
      child: children.isNotEmpty ? children.first : const SizedBox.shrink(),
    );
  }

  Widget _buildGestureDetector(proto.WidgetNode node, List<Widget> children) {
    return GestureDetector(
      onTap: () => sendEvent(proto.ClientEvent(
        nodeId: node.id.toString(),
        eventType: 'onTap',
      )),
      child: children.isNotEmpty ? children.first : const SizedBox.shrink(),
    );
  }

  Widget _buildAlign(proto.WidgetNode node, List<Widget> children) {
    final props = proto.AlignProps.fromBuffer(node.props);

    return Align(
      alignment: Alignment(props.x, props.y),
      child: children.isNotEmpty ? children.first : const SizedBox.shrink(),
    );
  }

  Widget _buildRadio(BuildContext context, proto.WidgetNode node) {
    final props = proto.RadioProps.fromBuffer(node.props);
    final selected = props.value == props.groupValue;

    return GestureDetector(
      onTap: () {
        sendEvent(proto.ClientEvent(
          nodeId: node.id.toString(),
          eventType: 'onChange',
          eventData: props.value.codeUnits,
        ));
      },
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            selected
                ? Icons.radio_button_checked
                : Icons.radio_button_unchecked,
            color: selected ? Theme.of(context).colorScheme.primary : null,
          ),
          const SizedBox(width: 8),
          Text(props.label),
        ],
      ),
    );
  }

  Widget _buildDropdown(proto.WidgetNode node) {
    final props = proto.DropdownProps.fromBuffer(node.props);

    return DropdownButton<String>(
      value: (props.value.isNotEmpty && props.items.contains(props.value))
          ? props.value
          : null,
      items: props.items
          .map((i) => DropdownMenuItem<String>(value: i, child: Text(i)))
          .toList(),
      onChanged: (v) {
        if (v != null) {
          sendEvent(proto.ClientEvent(
            nodeId: node.id.toString(),
            eventType: 'onChange',
            eventData: v.codeUnits,
          ));
        }
      },
    );
  }

  Widget _buildAnimatedPositioned(proto.WidgetNode node, List<Widget> children) {
    final props = proto.AnimatedPositionedProps.fromBuffer(node.props);
    final child = children.isNotEmpty ? children.first : const SizedBox.shrink();
    final duration = Duration(
      milliseconds: props.hasDurationMs() ? props.durationMs : 200,
    );

    return AnimatedPositioned(
      left: props.hasLeft() ? props.left : null,
      top: props.hasTop() ? props.top : null,
      right: props.hasRight() ? props.right : null,
      bottom: props.hasBottom() ? props.bottom : null,
      width: props.hasWidth() ? props.width : null,
      height: props.hasHeight() ? props.height : null,
      duration: duration,
      curve: props.hasCurve() ? _mapCurve(props.curve) : Curves.ease,
      child: child,
    );
  }

  Widget _buildWindowDragArea(List<Widget> children) {
    return DragToMoveArea(
      child: children.isNotEmpty ? children.first : const SizedBox.shrink(),
    );
  }

  Widget _buildCard(proto.WidgetNode node, List<Widget> children) {
    final props = proto.CardProps.fromBuffer(node.props);
    Widget? child = children.isNotEmpty ? children.first : null;
    if (props.padding > 0 && child != null) {
      child = Padding(padding: EdgeInsets.all(props.padding), child: child);
    }

    return Card(
      elevation: props.elevation > 0 ? props.elevation : null,
      shape: props.borderRadius > 0
          ? RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(props.borderRadius))
          : null,
      child: child,
    );
  }

  Widget _buildScaffold(proto.WidgetNode node, List<Widget> children) {
    final props = proto.ScaffoldProps.fromBuffer(node.props);

    // Children arrive in order (each present only when its flag is set): body,
    // app bar, FAB, drawer, bottom navigation bar.
    var i = 0;
    final body = i < children.length ? children[i++] : null;

    PreferredSizeWidget? appBar;
    if (props.hasAppBar && i < children.length) {
      final w = children[i++];
      appBar = w is PreferredSizeWidget ? w : null;
    }

    final fab = (props.hasFab && i < children.length) ? children[i++] : null;
    final drawer =
        (props.hasDrawer && i < children.length) ? children[i++] : null;
    final bottomBar =
        (props.hasBottomBar && i < children.length) ? children[i++] : null;

    return Scaffold(
      appBar: appBar,
      body: body,
      floatingActionButton: fab,
      drawer: drawer != null ? Drawer(child: drawer) : null,
      bottomNavigationBar: bottomBar,
    );
  }

  Widget _buildTabs(proto.WidgetNode node, List<Widget> children) {
    final props = proto.TabsProps.fromBuffer(node.props);
    final n = props.labels.length;
    if (n == 0) {
      return const SizedBox.shrink();
    }

    return DefaultTabController(
      length: n,
      initialIndex: props.initialIndex.clamp(0, n - 1),
      child: Column(
        children: [
          TabBar(tabs: [for (final l in props.labels) Tab(text: l)]),
          Expanded(child: TabBarView(children: children)),
        ],
      ),
    );
  }

  Widget _buildNavigationBar(proto.WidgetNode node) {
    final props = proto.NavigationBarProps.fromBuffer(node.props);

    final destinations = <NavigationDestination>[];
    for (var k = 0; k < props.labels.length; k++) {
      final icon = k < props.icons.length ? props.icons[k] : '';
      destinations.add(NavigationDestination(
        icon: Icon(_mapIconData(icon)),
        label: props.labels[k],
      ));
    }
    // A Material NavigationBar requires at least two destinations.
    while (destinations.length < 2) {
      destinations.add(
        const NavigationDestination(icon: Icon(Icons.circle), label: ''),
      );
    }

    return NavigationBar(
      selectedIndex: props.selectedIndex.clamp(0, destinations.length - 1),
      destinations: destinations,
      onDestinationSelected: (idx) => sendEvent(proto.ClientEvent(
        nodeId: node.id.toString(),
        eventType: 'onChange',
        eventData: idx.toString().codeUnits,
      )),
    );
  }

  Widget _buildAppBar(proto.WidgetNode node, List<Widget> children) {
    final props = proto.AppBarProps.fromBuffer(node.props);

    // The leading widget (when has_leading) comes first, then the actions.
    var i = 0;
    Widget? leading;
    if (props.hasLeading && i < children.length) {
      leading = children[i++];
    }
    final actions = children.sublist(i);

    return AppBar(
      title: Text(props.title),
      centerTitle: props.centerTitle,
      leading: leading,
      actions: actions.isEmpty ? null : actions,
      backgroundColor:
          props.bgColor.isNotEmpty ? hexToColor(props.bgColor) : null,
    );
  }

  Widget _buildFab(proto.WidgetNode node) {
    final props = proto.FabProps.fromBuffer(node.props);
    void onPressed() => sendEvent(proto.ClientEvent(
          nodeId: node.id.toString(),
          eventType: 'onClick',
        ));
    final icon = Icon(_mapIconData(props.icon));
    // A unique hero tag per node lets an app show more than one FAB (e.g. an
    // increment and a decrement button) without the default-tag Hero collision.
    final tag = 'fugo-fab-${node.id}';

    if (props.label.isNotEmpty) {
      return FloatingActionButton.extended(
        heroTag: tag,
        onPressed: onPressed,
        icon: icon,
        label: Text(props.label),
      );
    }

    return FloatingActionButton(
      heroTag: tag,
      onPressed: onPressed,
      mini: props.mini,
      child: icon,
    );
  }

  Widget _buildListTile(proto.WidgetNode node) {
    final props = proto.ListTileProps.fromBuffer(node.props);

    return ListTile(
      title: Text(props.title),
      subtitle: props.subtitle.isNotEmpty ? Text(props.subtitle) : null,
      leading: props.leadingIcon.isNotEmpty
          ? Icon(_mapIconData(props.leadingIcon))
          : null,
      trailing: props.trailingIcon.isNotEmpty
          ? Icon(_mapIconData(props.trailingIcon))
          : null,
      onTap: () => sendEvent(proto.ClientEvent(
        nodeId: node.id.toString(),
        eventType: 'onTap',
      )),
    );
  }

  Widget _buildChip(proto.WidgetNode node) {
    final props = proto.ChipProps.fromBuffer(node.props);

    return FilterChip(
      label: Text(props.label),
      selected: props.selected,
      onSelected: (v) => sendEvent(proto.ClientEvent(
        nodeId: node.id.toString(),
        eventType: 'onTap',
        eventData: (v ? '1' : '0').codeUnits,
      )),
      onDeleted: props.deletable
          ? () => sendEvent(proto.ClientEvent(
                nodeId: node.id.toString(),
                eventType: 'onDeleted',
              ))
          : null,
    );
  }

  Widget _buildProgress(proto.WidgetNode node) {
    final props = proto.ProgressProps.fromBuffer(node.props);
    final value = props.value >= 0 ? props.value : null;

    return props.linear
        ? LinearProgressIndicator(value: value)
        : CircularProgressIndicator(value: value);
  }

  // _mapIconData resolves an fg icon name (fg.Icons.Home -> 'home') to its
  // Flutter IconData via the generated table (see cmd/gen-icons).
  Widget _buildTooltip(proto.WidgetNode node, List<Widget> children) {
    final props = proto.TooltipProps.fromBuffer(node.props);
    final child =
        children.isNotEmpty ? children.first : const SizedBox.shrink();

    return Tooltip(message: props.message, child: child);
  }

  Widget _buildBadge(proto.WidgetNode node, List<Widget> children) {
    final props = proto.BadgeProps.fromBuffer(node.props);
    final child =
        children.isNotEmpty ? children.first : const SizedBox.shrink();

    return Badge(
      label: props.label.isNotEmpty ? Text(props.label) : null,
      child: child,
    );
  }

  Widget _buildAvatar(proto.WidgetNode node) {
    final props = proto.AvatarProps.fromBuffer(node.props);

    Widget? child;
    if (props.text.isNotEmpty) {
      child = Text(props.text);
    } else if (props.icon.isNotEmpty) {
      child = Icon(_mapIconData(props.icon));
    }

    return CircleAvatar(
      radius: props.radius > 0 ? props.radius : null,
      backgroundColor:
          props.bgColor.isNotEmpty ? hexToColor(props.bgColor) : null,
      child: child,
    );
  }

  Widget _buildSegmentedButton(proto.WidgetNode node) {
    final props = proto.SegmentedButtonProps.fromBuffer(node.props);

    final segments = <ButtonSegment<String>>[];
    for (var k = 0; k < props.values.length; k++) {
      final label = k < props.labels.length ? props.labels[k] : props.values[k];
      segments.add(
        ButtonSegment<String>(value: props.values[k], label: Text(label)),
      );
    }
    if (segments.isEmpty) {
      return const SizedBox.shrink();
    }

    final selected =
        props.selected.isNotEmpty && props.values.contains(props.selected)
            ? props.selected
            : props.values.first;

    return SegmentedButton<String>(
      segments: segments,
      selected: {selected},
      onSelectionChanged: (s) {
        if (s.isNotEmpty) {
          sendEvent(proto.ClientEvent(
            nodeId: node.id.toString(),
            eventType: 'onChange',
            eventData: s.first.codeUnits,
          ));
        }
      },
    );
  }

  Widget _buildAspectRatio(proto.WidgetNode node, List<Widget> children) {
    final props = proto.AspectRatioProps.fromBuffer(node.props);
    final child =
        children.isNotEmpty ? children.first : const SizedBox.shrink();

    return AspectRatio(
      aspectRatio: props.ratio > 0 ? props.ratio : 1,
      child: child,
    );
  }

  Widget _buildClipRRect(proto.WidgetNode node, List<Widget> children) {
    final props = proto.ClipRRectProps.fromBuffer(node.props);
    final child =
        children.isNotEmpty ? children.first : const SizedBox.shrink();

    return ClipRRect(
      borderRadius: BorderRadius.circular(props.radius),
      child: child,
    );
  }

  Widget _buildFittedBox(List<Widget> children) {
    return FittedBox(
      child: children.isNotEmpty ? children.first : const SizedBox.shrink(),
    );
  }

  Widget _buildFlexible(proto.WidgetNode node, List<Widget> children) {
    final props = proto.FlexibleProps.fromBuffer(node.props);
    final child =
        children.isNotEmpty ? children.first : const SizedBox.shrink();

    return Flexible(flex: props.flex > 0 ? props.flex : 1, child: child);
  }

  Widget _buildExpansionTile(proto.WidgetNode node, List<Widget> children) {
    final props = proto.ExpansionTileProps.fromBuffer(node.props);

    return ExpansionTile(
      title: Text(props.title),
      subtitle: props.subtitle.isNotEmpty ? Text(props.subtitle) : null,
      leading: props.leadingIcon.isNotEmpty
          ? Icon(_mapIconData(props.leadingIcon))
          : null,
      initiallyExpanded: props.initiallyExpanded,
      children: children,
    );
  }

  Widget _buildPopupMenu(proto.WidgetNode node) {
    final props = proto.PopupMenuProps.fromBuffer(node.props);

    return PopupMenuButton<String>(
      icon: Icon(
        props.icon.isNotEmpty ? _mapIconData(props.icon) : Icons.more_vert,
      ),
      itemBuilder: (_) => [
        for (var k = 0; k < props.values.length; k++)
          PopupMenuItem<String>(
            value: props.values[k],
            child: Text(k < props.labels.length ? props.labels[k] : props.values[k]),
          ),
      ],
      onSelected: (v) => sendEvent(proto.ClientEvent(
        nodeId: node.id.toString(),
        eventType: 'onSelected',
        eventData: v.codeUnits,
      )),
    );
  }

  Widget _buildRichText(proto.WidgetNode node) {
    final props = proto.RichTextProps.fromBuffer(node.props);

    return Text.rich(TextSpan(
      children: [
        for (final s in props.spans)
          TextSpan(
            text: s.text,
            style: TextStyle(
              color: s.color.isNotEmpty ? hexToColor(s.color) : null,
              fontSize: s.fontSize > 0 ? s.fontSize : null,
              fontWeight: s.bold ? FontWeight.bold : null,
            ),
          ),
      ],
    ));
  }

  Widget _buildDataTable(proto.WidgetNode node) {
    final props = proto.DataTableProps.fromBuffer(node.props);

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: DataTable(
        columns: [for (final c in props.columns) DataColumn(label: Text(c))],
        rows: [
          for (final r in props.rows)
            DataRow(cells: [for (final cell in r.cells) DataCell(Text(cell))]),
        ],
      ),
    );
  }

  Widget _buildStepper(proto.WidgetNode node, List<Widget> children) {
    final props = proto.StepperProps.fromBuffer(node.props);

    final steps = <Step>[];
    for (var k = 0; k < props.titles.length; k++) {
      steps.add(Step(
        title: Text(props.titles[k]),
        content: k < children.length ? children[k] : const SizedBox.shrink(),
        isActive: k == props.active,
      ));
    }
    if (steps.isEmpty) {
      return const SizedBox.shrink();
    }

    return Stepper(
      currentStep: props.active.clamp(0, steps.length - 1),
      steps: steps,
      onStepTapped: (i) => sendEvent(proto.ClientEvent(
        nodeId: node.id.toString(),
        eventType: 'onChange',
        eventData: i.toString().codeUnits,
      )),
    );
  }

  IconData _mapIconData(String name) => materialIcons[name] ?? Icons.help_outline;

  Curve _mapCurve(String name) {
    switch (name) {
      case 'linear':
        return Curves.linear;
      case 'easeIn':
        return Curves.easeIn;
      case 'easeOut':
        return Curves.easeOut;
      case 'easeInOut':
        return Curves.easeInOut;
      default:
        return Curves.ease;
    }
  }
}

Color hexToColor(String hex) {
  final buffer = StringBuffer();
  if (hex.startsWith('#')) {
    buffer.write('FF');
    buffer.write(hex.substring(1));
  } else {
    buffer.write('FF');
    buffer.write(hex);
  }
  final intVal = int.parse(buffer.toString(), radix: 16);
  return Color(intVal);
}
