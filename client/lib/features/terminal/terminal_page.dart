import 'dart:async';
import 'dart:typed_data';

import 'package:flutter/material.dart';

import 'package:hitest/core/config/api_config.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/core/network/term_screen.dart';
import 'package:hitest/core/network/terminal_client.dart';
import 'package:hitest/core/theme/app_design_tokens.dart';

/// 终端页：连目标机器的 PTY，自己渲染 VT/ANSI。
///
/// 渲染 = scrollback（被顶出屏幕的历史）+ 当前固定网格。整块放进纵向滚动视图，
/// 新输出自动贴到底部，用户往上滑即可翻历史；列宽按屏幕宽度铺满，不再横向滚动。
class TerminalPage extends StatefulWidget {
  const TerminalPage({super.key});

  @override
  State<TerminalPage> createState() => _TerminalPageState();
}

class _TerminalPageState extends State<TerminalPage> {
  final _client = TerminalClient();
  final _inputController = TextEditingController();
  final _scrollController = ScrollController();

  TermScreen? _screen;
  TerminalStatus _status = TerminalStatus.idle;
  String? _message;
  StreamSubscription<TermFrameEvent>? _frameSub;
  StreamSubscription<TerminalStatus>? _statusSub;

  double _cellW = 7;
  double _cellH = 15;
  bool _cellMeasured = false;
  int _cols = 80;
  int _rows = 24;
  bool _stickToBottom = true;

  // 最多渲染的历史行数，避免长会话 build 成本失控。
  static const int _maxHistoryRows = 800;

  static const TextStyle _mono = TextStyle(
    fontFamily: 'Menlo',
    fontFamilyFallback: ['Courier', 'Courier New', 'monospace'],
    fontSize: 12,
    height: 1.3,
  );
  static const Color _defaultFg = Color(0xFFE8E8E8);
  static const Color _screenBg = Color(0xFF101014);

  @override
  void initState() {
    super.initState();
    _frameSub = _client.frames.listen(_onFrame);
    _statusSub = _client.status.listen((s) {
      if (mounted) setState(() => _status = s);
    });
    _scrollController.addListener(_onScroll);
    WidgetsBinding.instance.addPostFrameCallback((_) => _connect());
  }

  @override
  void dispose() {
    _frameSub?.cancel();
    _statusSub?.cancel();
    _scrollController.dispose();
    _client.dispose();
    _inputController.dispose();
    super.dispose();
  }

  bool get _connected => _client.isOpen;

  Future<void> _connect() async {
    final config = await ApiConfigManager.getCurrentConfig();
    if (!mounted) return;
    if (config == null) {
      setState(() => _message = context.l10n.noServerSelectedDescription);
      return;
    }
    _measureCell();
    setState(() {
      _screen = TermScreen(cols: _cols, rows: _rows);
      _message = null;
    });
    await _client.open(config: config, cols: _cols, rows: _rows);
  }

  void _measureCell() {
    if (_cellMeasured) return;
    final painter = TextPainter(
      text: TextSpan(text: 'M' * 20, style: _mono),
      textDirection: TextDirection.ltr,
    )..layout();
    if (painter.width > 0) {
      _cellW = painter.width / 20;
      _cellH = painter.height;
      _cellMeasured = true;
    }
  }

  ({int cols, int rows}) _gridFor(Size size) {
    if (size.width <= 0 || size.height <= 0 || !_cellMeasured) {
      return (cols: _cols, rows: _rows);
    }
    final cols = (size.width / _cellW).floor().clamp(20, 400);
    final rows = (size.height / _cellH).floor().clamp(6, 200);
    return (cols: cols, rows: rows);
  }

  void _applyGrid(Size size) {
    final screen = _screen;
    if (screen == null) return;
    final grid = _gridFor(size);
    if (grid.cols == _cols && grid.rows == _rows) return;
    _cols = grid.cols;
    _rows = grid.rows;
    screen.resize(_cols, _rows);
    if (_connected) _client.resize(_cols, _rows);
    setState(() {});
  }

  void _onFrame(TermFrameEvent event) {
    switch (event.kind) {
      case TermFrameEventKind.data:
        final screen = _screen;
        if (screen == null) return;
        screen.feed(Uint8List.fromList(event.bytes!));
        if (mounted) {
          setState(() {});
          if (_stickToBottom) _scheduleScrollToBottom();
        }
      case TermFrameEventKind.close:
        if (mounted) setState(() => _message = context.l10n.terminalClosed);
      case TermFrameEventKind.error:
        if (mounted) {
          setState(
            () => _message = context.l10n.terminalFailed(event.message),
          );
        }
    }
  }

  void _onScroll() {
    if (!_scrollController.hasClients) return;
    final pos = _scrollController.position;
    _stickToBottom = pos.pixels >= pos.maxScrollExtent - 24;
  }

  void _scheduleScrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.jumpTo(_scrollController.position.maxScrollExtent);
      }
    });
  }

  // 远程 pty 处于 echo 模式，输入的字符会由服务端回显，这里不再本地回显，
  // 否则每个字符会画两遍。
  void _sendText(String text) {
    if (text.isEmpty || !_connected) return;
    _client.write(text);
  }

  void _submitLine() {
    if (!_connected) return;
    final text = _inputController.text;
    _inputController.clear();
    _sendText('$text\r');
    _stickToBottom = true;
  }

  String _arrow(String letter) {
    // application cursor keys 模式（vim 之类打开）下编码不同
    return (_screen?.appCursorKeys ?? false) ? '\x1bO$letter' : '\x1b[$letter';
  }

  // ---------- UI ----------

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '${l10n.terminalTitle} · ${_statusTitle(l10n)}',
              style: const TextStyle(fontSize: 15),
            ),
            Text(
              '$_cols×$_rows',
              style: const TextStyle(fontSize: 11),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: Icon(_connected ? Icons.link_off : Icons.link),
            tooltip:
                _connected ? l10n.terminalDisconnect : l10n.terminalReconnect,
            onPressed: () async {
              if (_connected) {
                await _client.close();
              } else {
                await _connect();
              }
            },
          ),
        ],
      ),
      body: Column(
        children: [
          if (_message != null)
            Container(
              width: double.infinity,
              color: Theme.of(context).colorScheme.errorContainer,
              padding: const EdgeInsets.symmetric(
                horizontal: AppDesignTokens.spacingMd,
                vertical: AppDesignTokens.spacingXs,
              ),
              child: Text(
                _message!,
                style: TextStyle(
                  fontSize: 12,
                  color: Theme.of(context).colorScheme.onErrorContainer,
                ),
              ),
            ),
          Expanded(
            child: LayoutBuilder(
              builder: (context, constraints) {
                WidgetsBinding.instance.addPostFrameCallback(
                  (_) => _applyGrid(constraints.biggest),
                );
                return GestureDetector(
                  behavior: HitTestBehavior.opaque,
                  onTap: () => FocusScope.of(context).unfocus(),
                  child: Container(
                    color: _screenBg,
                    padding: const EdgeInsets.all(4),
                    child: _buildScreen(),
                  ),
                );
              },
            ),
          ),
          _buildInputBar(context),
          _buildQuickKeys(context),
        ],
      ),
    );
  }

  Widget _buildScreen() {
    final screen = _screen;
    if (screen == null) {
      return const Center(
        child: CircularProgressIndicator(color: Colors.white),
      );
    }
    final history = screen.history;
    final histStart = history.length > _maxHistoryRows
        ? history.length - _maxHistoryRows
        : 0;
    final rows = <List<TermCell>>[
      ...history.sublist(histStart),
      ...screen.lines,
    ];
    return SingleChildScrollView(
      controller: _scrollController,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          for (final cells in rows) _buildRow(cells),
        ],
      ),
    );
  }

  Widget _buildRow(List<TermCell> cells) {
    return SizedBox(
      height: _cellH,
      child: Text.rich(
        TextSpan(children: _buildRowSpans(cells)),
        style: _mono.copyWith(color: _defaultFg, height: 1.0),
      ),
    );
  }

  List<InlineSpan> _buildRowSpans(List<TermCell> cells) {
    final spans = <InlineSpan>[];
    final sb = StringBuffer();
    int? fgKey;
    int? bgKey;
    bool? bold;

    void flush() {
      if (sb.isEmpty) return;
      spans.add(
        TextSpan(
          text: sb.toString(),
          style: TextStyle(
            color: fgKey == null ? _defaultFg : ansiColor(fgKey),
            backgroundColor: bgKey == null ? null : ansiColor(bgKey),
            fontWeight: bold == true ? FontWeight.bold : null,
          ),
        ),
      );
      sb.clear();
    }

    for (final cell in cells) {
      final cellFg = cell.fg < 0 ? null : cell.fg;
      final cellBg = cell.bg < 0 ? null : cell.bg;
      if (cellFg != fgKey || cellBg != bgKey || cell.bold != bold) {
        flush();
        fgKey = cellFg;
        bgKey = cellBg;
        bold = cell.bold;
      }
      sb.write(cell.ch);
    }
    flush();
    if (spans.isEmpty) {
      spans.add(TextSpan(text: ' ' * cells.length));
    }
    return spans;
  }

  Widget _buildInputBar(BuildContext context) {
    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(8, 6, 8, 0),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: _inputController,
                enabled: _connected,
                textInputAction: TextInputAction.send,
                autocorrect: false,
                enableSuggestions: false,
                keyboardType: TextInputType.visiblePassword,
                onSubmitted: (_) => _submitLine(),
                decoration: const InputDecoration(
                  isDense: true,
                  border: OutlineInputBorder(),
                ),
              ),
            ),
            const SizedBox(width: 6),
            IconButton.filledTonal(
              onPressed: _connected ? _submitLine : null,
              icon: const Icon(Icons.keyboard_return),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildQuickKeys(BuildContext context) {
    final l10n = context.l10n;
    Widget key(String label, String Function() payload) {
      return Expanded(
        child: OutlinedButton(
          onPressed: _connected ? () => _sendText(payload()) : null,
          style: OutlinedButton.styleFrom(
            padding: const EdgeInsets.symmetric(vertical: 8),
            minimumSize: Size.zero,
            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          ),
          child: Text(label, style: const TextStyle(fontSize: 12)),
        ),
      );
    }

    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(8, 6, 8, 8),
        child: Row(
          children: [
            key('Ctrl-C', () => '\x03'),
            const SizedBox(width: 4),
            key('Tab', () => '\t'),
            const SizedBox(width: 4),
            key('↑', () => _arrow('A')),
            const SizedBox(width: 4),
            key('↓', () => _arrow('B')),
            const SizedBox(width: 4),
            key('←', () => _arrow('D')),
            const SizedBox(width: 4),
            key('→', () => _arrow('C')),
            const SizedBox(width: 4),
            key(l10n.terminalClear, () => 'clear\r'),
          ],
        ),
      ),
    );
  }

  String _statusTitle(AppLocalizations l10n) {
    return switch (_status) {
      TerminalStatus.idle => l10n.terminalNotConnected,
      TerminalStatus.connecting => l10n.terminalConnecting,
      TerminalStatus.open => l10n.terminalConnect,
      TerminalStatus.closed => l10n.terminalClosed,
      TerminalStatus.failed => l10n.terminalNotConnected,
    };
  }
}

/// ANSI 颜色索引 → 颜色。0-15 是 16 色，16-255 是 256 色，
/// 带 0x1000000 标记的是 SGR;2 真彩（值本身就是 ARGB）。
Color? ansiColor(int index) {
  if (index >= 0x1000000) return Color(index);
  const palette = <int>[
    0xFF1C1C1C, 0xFFC93A42, 0xFF32A142, 0xFFD19E31, //
    0xFF3A78C2, 0xFFB757A8, 0xFF3BAEC0, 0xFFD5D5D5, //
    0xFF6B6B6B, 0xFFFF6B71, 0xFF5AEA6E, 0xFFF9DF6F, //
    0xFF79AFFF, 0xFFE67DE6, 0xFF7EE2F1, 0xFFFFFFFF, //
  ];
  if (index < 16) return Color(palette[index]);
  if (index < 232) {
    final i = index - 16;
    return Color.fromARGB(
      255,
      (i ~/ 36) * 51,
      ((i % 36) ~/ 6) * 51,
      (i % 6) * 51,
    );
  }
  final v = (8 + (index - 232) * 10).clamp(0, 255);
  return Color.fromARGB(255, v, v, v);
}
