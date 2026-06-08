import 'package:flutter/material.dart';

import 'generated/fugo/v1/fugo.pb.dart' as proto;
import 'events.dart';

class WidgetRegistry {
  Widget build(proto.WidgetNode node, List<Widget> children) {
    switch (node.type) {
      case proto.WidgetType.TEXT:
        return _buildText(node);
      case proto.WidgetType.CONTAINER:
        return _buildContainer(node, children);
      case proto.WidgetType.COLUMN:
        return _buildColumn(children);
      case proto.WidgetType.CENTER:
        return _buildCenter(children);
      case proto.WidgetType.BUTTON:
        return _buildButton(node);
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
      default:
        return const SizedBox.shrink();
    }
  }

  Widget _buildText(proto.WidgetNode node) {
    final props = proto.TextProps.fromBuffer(node.props);
    return Text(
      props.value,
      style: TextStyle(
        fontSize: props.hasFontSize() ? props.fontSize : 14,
        color: props.hasColor() ? hexToColor(props.color) : Colors.white,
      ),
    );
  }

  Widget _buildButton(proto.WidgetNode node) {
    final props = proto.ButtonProps.fromBuffer(node.props);

    return GestureDetector(
      onTap: () {
        sendEvent(proto.ClientEvent(
          nodeId: node.id.toString(),
          eventType: 'onClick',
        ));
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        decoration: BoxDecoration(
          color: props.hasBgColor() ? hexToColor(props.bgColor) : const Color(0xFF3B82F6),
          borderRadius: BorderRadius.circular(
            props.hasBorderRadius() ? props.borderRadius : 8,
          ),
        ),
        child: Text(
          props.label,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 14,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }

  Widget _buildContainer(proto.WidgetNode node, List<Widget> children) {
    final props = proto.ContainerProps.fromBuffer(node.props);
    final child = children.isNotEmpty ? children.first : const SizedBox.shrink();

    return Container(
      color: props.hasBgColor() ? hexToColor(props.bgColor) : null,
      padding: props.hasPadding()
          ? EdgeInsets.all(props.padding)
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

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: TextField(
        obscureText: props.obscure,
        style: TextStyle(
          fontSize: props.hasFontSize() ? props.fontSize : 14,
          color: Colors.white,
        ),
        decoration: InputDecoration(
          hintText: props.placeholder,
          hintStyle: const TextStyle(color: Colors.grey),
          filled: true,
          fillColor: const Color(0xFF2A2A2A),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: BorderSide.none,
          ),
        ),
        onChanged: (value) {
          sendEvent(proto.ClientEvent(
            nodeId: node.id.toString(),
            eventType: 'onChange',
            eventData: value.codeUnits,
          ));
        },
      ),
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

    return AnimatedContainer(
      duration: duration,
      curve: props.hasCurve() ? _mapCurve(props.curve) : Curves.ease,
      color: props.hasBgColor() ? hexToColor(props.bgColor) : null,
      padding: props.hasPadding() ? EdgeInsets.all(props.padding) : null,
      decoration: props.hasBorderRadius()
          ? BoxDecoration(
              borderRadius: BorderRadius.circular(props.borderRadius),
            )
          : null,
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
