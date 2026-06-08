import 'dart:async';
import 'dart:isolate';

import 'generated/fugo/v1/fugo.pb.dart';

SendPort? _sendPort;

final _immediateEvents = {'onClick', 'onTap', 'onSubmit', 'onLongPress'};
final _debounceTimers = <String, Timer>{};
const _debounceInterval = Duration(milliseconds: 16);

void setEventSendPort(SendPort port) {
  _sendPort = port;
}

void sendEvent(ClientEvent event) {
  if (_sendPort == null) {
    print('[fugo] events: sendPort is null, dropping event ${event.eventType}');

    return;
  }

  if (_immediateEvents.contains(event.eventType)) {
    print('[fugo] events: send immediate ${event.eventType} node=${event.nodeId}');
    _sendPort!.send(event.writeToBuffer());

    return;
  }

  final key = '${event.nodeId}_${event.eventType}';
  _debounceTimers[key]?.cancel();
  _debounceTimers[key] = Timer(_debounceInterval, () {
    _debounceTimers.remove(key);
    print('[fugo] events: send debounced ${event.eventType} node=${event.nodeId}');
    _sendPort?.send(event.writeToBuffer());
  });
}
