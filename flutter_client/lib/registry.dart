import 'package:flutter/material.dart';

import 'generated/fugo/v1/fugo.pb.dart';
import 'events.dart';

class WidgetRegistry {
  Widget build(WidgetNode node, List<Widget> children) {
    switch (node.type) {
      case WidgetType.TEXT:
        return _buildText(node);
      case WidgetType.CONTAINER:
        return _buildContainer(node, children);
      case WidgetType.COLUMN:
        return _buildColumn(children);
      case WidgetType.CENTER:
        return _buildCenter(children);
      case WidgetType.BUTTON:
        return _buildButton(node);
      default:
        return const SizedBox.shrink();
    }
  }

  Widget _buildText(WidgetNode node) {
    final props = TextProps.fromBuffer(node.props);
    return Text(
      props.value,
      style: TextStyle(
        fontSize: props.hasFontSize() ? props.fontSize : 14,
        color: props.hasColor() ? hexToColor(props.color) : Colors.white,
      ),
    );
  }

  Widget _buildButton(WidgetNode node) {
    final props = ButtonProps.fromBuffer(node.props);

    return GestureDetector(
      onTap: () {
        sendEvent(ClientEvent(
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
          style: TextStyle(
            color: Colors.white,
            fontSize: props.hasFontSize() ? props.fontSize : 14,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }

  Widget _buildContainer(WidgetNode node, List<Widget> children) {
    final props = ContainerProps.fromBuffer(node.props);
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
