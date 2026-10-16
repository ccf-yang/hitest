import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:hitest/core/config/api_config.dart';
import 'package:hitest/core/network/hitest_websocket_connector.dart';

/// 终端帧类型，与服务端 `internal/modules/terminal/terminal.go` 保持一致：
/// WS 消息 = `[1B kind][payload]`；open 参数是首条 JSON 文本消息。
class TermFrame {
  static const int data = 0;
  static const int resize = 1;
  static const int close = 2;
}

enum TerminalStatus { idle, connecting, open, closed, failed }

/// hitest 终端会话客户端。
///
/// 协议（同 Web 前端）：
///   1. 连 `/api/v1/{local|m/{id}}/terminal/ws`
///   2. 首条消息发 JSON `{cols, rows, cmd}`
///   3. 之后双向都是二进制帧 `[1B kind][payload]`
///      0=数据 1=resize(4B: cols u16, rows u16) 2=关闭
class TerminalClient {
  TerminalClient();

  WebSocket? _socket;
  StreamSubscription<dynamic>? _sub;

  final _frames = StreamController<TermFrameEvent>.broadcast();
  final _status = StreamController<TerminalStatus>.broadcast();

  Stream<TermFrameEvent> get frames => _frames.stream;
  Stream<TerminalStatus> get status => _status.stream;

  TerminalStatus _current = TerminalStatus.idle;
  TerminalStatus get current => _current;
  bool get isOpen => _current == TerminalStatus.open;

  void _setStatus(TerminalStatus s) {
    _current = s;
    _status.add(s);
  }

  Future<void> open({
    required ApiConfig config,
    required int cols,
    required int rows,
    String cmd = '/bin/bash -l',
  }) async {
    await close();
    if (cols < 1 || rows < 1) return;
    _setStatus(TerminalStatus.connecting);
    try {
      final socket = await OnePanelWebSocketConnector.connect(
        config: config,
        path: '/api/v1/terminal/ws',
      );
      _socket = socket;
      socket.add(jsonEncode({'cols': cols, 'rows': rows, 'cmd': cmd}));
      _sub = socket.listen(
        _onMessage,
        onError: (Object e) {
          _frames.add(TermFrameEvent.error(_describe(e)));
          _setStatus(TerminalStatus.failed);
        },
        onDone: () {
          _frames.add(TermFrameEvent.close());
          _setStatus(TerminalStatus.closed);
        },
        cancelOnError: true,
      );
      _setStatus(TerminalStatus.open);
    } on WebSocketException catch (e) {
      _setStatus(TerminalStatus.failed);
      _frames.add(TermFrameEvent.error(_describe(e)));
    } on HandshakeException catch (e) {
      _setStatus(TerminalStatus.failed);
      _frames.add(TermFrameEvent.error(_describe(e)));
    } catch (e) {
      _setStatus(TerminalStatus.failed);
      _frames.add(TermFrameEvent.error(e.toString()));
    }
  }

  void _onMessage(dynamic message) {
    if (message is String) {
      // 本机 pty 启动失败时服务端直接写一段裸文本错误
      _frames.add(TermFrameEvent.error(message));
      return;
    }
    final bytes = _asBytes(message);
    if (bytes == null || bytes.isEmpty) return;
    final kind = bytes[0];
    final payload = bytes.sublist(1);
    switch (kind) {
      case TermFrame.data:
        if (payload.isNotEmpty) _frames.add(TermFrameEvent.data(payload));
      case TermFrame.close:
        _frames.add(TermFrameEvent.close());
        _setStatus(TerminalStatus.closed);
      default:
      // 其余方向（服务端不会发 resize）忽略
    }
  }

  static Uint8List? _asBytes(dynamic message) {
    if (message is Uint8List) return message;
    if (message is List<int>) return Uint8List.fromList(message);
    return null;
  }

  void write(String text) {
    final socket = _socket;
    if (socket == null || text.isEmpty) return;
    final payload = utf8.encode(text);
    final frame = Uint8List(1 + payload.length);
    frame[0] = TermFrame.data;
    frame.setRange(1, frame.length, payload);
    socket.add(frame);
  }

  void resize(int cols, int rows) {
    final socket = _socket;
    if (socket == null || cols < 1 || rows < 1) return;
    final frame = Uint8List(5);
    frame[0] = TermFrame.resize;
    ByteData.view(frame.buffer)
        ..setUint16(1, cols, Endian.big)
        ..setUint16(3, rows, Endian.big);
    socket.add(frame);
  }

  Future<void> close() async {
    final socket = _socket;
    _socket = null;
    await _sub?.cancel();
    _sub = null;
    if (socket != null) {
      try {
        socket.add(Uint8List.fromList([TermFrame.close]));
      } catch (_) {
        // 已经关了
      }
      try {
        await socket.close();
      } catch (_) {
        // 忽略关闭异常
      }
    }
    if (_current == TerminalStatus.open) _setStatus(TerminalStatus.closed);
  }

  Future<void> dispose() async {
    await close();
    await _frames.close();
    await _status.close();
  }

  static String _describe(Object e) {
    final text = e.toString();
    if (text.contains('Connection refused') ||
        text.contains('Failed host lookup')) {
      return '连不上服务端，检查地址与机器 ID';
    }
    return text;
  }
}

class TermFrameEvent {
  const TermFrameEvent._(this.kind, [this.bytes, this.message = '']);

  final TermFrameEventKind kind;
  final Uint8List? bytes;
  final String message;

  factory TermFrameEvent.data(Uint8List bytes) =>
      TermFrameEvent._(TermFrameEventKind.data, bytes);
  factory TermFrameEvent.close() =>
      TermFrameEvent._(TermFrameEventKind.close);
  factory TermFrameEvent.error(String message) =>
      TermFrameEvent._(TermFrameEventKind.error, null, message);
}

enum TermFrameEventKind { data, close, error }
