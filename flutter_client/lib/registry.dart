import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

import 'generated/fugo/v1/fugo.pb.dart' as proto;
import 'events.dart';

class WidgetRegistry {
  Widget build(BuildContext context, proto.WidgetNode node, List<Widget> children) {
    switch (node.type) {
      case proto.WidgetType.TEXT:
        return _buildText(context, node);
      case proto.WidgetType.CONTAINER:
        return _buildContainer(node, children);
      case proto.WidgetType.COLUMN:
        return _buildColumn(children);
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
    final radius = props.hasBorderRadius() ? props.borderRadius : 12.0;
    return FilledButton(
      style: FilledButton.styleFrom(
        backgroundColor: props.hasBgColor() ? hexToColor(props.bgColor) : null,
        textStyle: props.hasFontSize() ? TextStyle(fontSize: props.fontSize) : null,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(radius)),
      ),
      onPressed: () => sendEvent(proto.ClientEvent(
        nodeId: node.id.toString(),
        eventType: 'onClick',
      )),
      child: Text(props.label),
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

  Widget _buildColumn(List<Widget> children) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      crossAxisAlignment: CrossAxisAlignment.center,
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

  IconData _mapIconData(String name) {
    switch (name) {
      case 'home':
        return Icons.home;
      case 'settings':
        return Icons.settings;
      case 'search':
        return Icons.search;
      case 'add':
        return Icons.add;
      case 'delete':
        return Icons.delete;
      case 'edit':
        return Icons.edit;
      case 'close':
        return Icons.close;
      case 'check':
        return Icons.check;
      case 'arrow_back':
        return Icons.arrow_back;
      case 'arrow_forward':
        return Icons.arrow_forward;
      case 'menu':
        return Icons.menu;
      case 'person':
        return Icons.person;
      case 'favorite':
        return Icons.favorite;
      case 'share':
        return Icons.share;
      case 'info':
        return Icons.info;
      case 'warning':
        return Icons.warning;
      case 'star':
        return Icons.star;
      case 'play_arrow':
        return Icons.play_arrow;
      case 'pause':
        return Icons.pause;
      case 'refresh':
        return Icons.refresh;
      default:
        return Icons.circle;
    }
  }

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
