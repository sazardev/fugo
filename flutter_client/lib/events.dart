import 'dart:async';

import 'generated/fugo/v1/fugo.pb.dart';

final _eventController = StreamController<ClientEvent>.broadcast();

void sendEvent(ClientEvent event) {
  _eventController.add(event);
}

Stream<ClientEvent> get eventStream => _eventController.stream;
