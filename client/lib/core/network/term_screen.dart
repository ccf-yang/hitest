/// 终端屏幕格子（最小 VT/ANSI 实现）。
///
/// 只实现移动端运维够用的一族：CSI 光标移动/定位、ED/EL 擦除、SGR（16/256/RGB）、
/// CR/LF/BS/TAB、滚动区、IL/DL、DEC 私有模式与 OSC 吞掉不画。
/// 已知不做的：宽字符（中文占 2 列）、alt screen、鼠标上报、右键插入粘贴 —
/// 所以 vim/htop 这类全屏程序显示会不准，日常看日志、跑命令、top 快照够用。
class TermCell {
  const TermCell(this.ch, this.fg, this.bg, this.bold, this.italic, this.underline);
  final String ch;
  final int fg; // -1 = 默认色
  final int bg; // -1 = 透明
  final bool bold;
  final bool italic;
  final bool underline;
}

class TermStyle {
  int fg = -1;
  int bg = -1;
  bool bold = false;
  bool faint = false;
  bool italic = false;
  bool underline = false;
  bool reverse = false;

  TermStyle clone() => TermStyle()
    ..fg = fg
    ..bg = bg
    ..bold = bold
    ..faint = faint
    ..italic = italic
    ..underline = underline
    ..reverse = reverse;
}

class TermScreen {
  TermScreen({required int cols, required int rows})
      : _cols = cols,
        _rows = rows,
        _lines = List.generate(rows, (_) => _blankLine(cols));

  static const int scrollbackLimit = 2000;

  int _cols;
  int _rows;
  final List<List<TermCell>> _lines;
  final List<List<TermCell>> _scrollback = [];

  int _cx = 0;
  int _cy = 0;
  int _top = 0;
  int _bottom = 0;
  TermStyle _style = TermStyle();
  bool _wrapNext = false;
  bool _appCursorKeys = false;
  bool _bracketedPaste = false;
  int _tabStep = 8;

  int get cols => _cols;
  int get rows => _rows;
  int get cursorCol => _cx;
  int get cursorRow => _cy;
  bool get appCursorKeys => _appCursorKeys;
  bool get bracketedPaste => _bracketedPaste;
  List<List<TermCell>> get lines => List.unmodifiable(_lines);
  List<List<TermCell>> get history => List.unmodifiable(_scrollback);
  int get historyLength => _scrollback.length;

  static List<TermCell> _blankLine(int cols) =>
      List.generate(cols, (_) => const TermCell(' ', -1, -1, false, false, false));

  /// 追加一段收到的字节。
  void feed(List<int> bytes) {
    _decoder.add(bytes);
  }

  late final Utf8LineDecoder _decoder = Utf8LineDecoder(_onChar);

  void _onChar(int rune) {
    switch (rune) {
      case 0x1b:
        _state = _EscState.escape;
        return;
      case 0x0d: // CR
        _cx = 0;
        _wrapNext = false;
        return;
      case 0x0a: // LF
      case 0x0b: // VT
      case 0x0c: // FF
        _lineFeed(1);
        _wrapNext = false;
        return;
      case 0x08: // BS
        if (_cx > 0) _cx--;
        _wrapNext = false;
        return;
      case 0x09: // TAB
        _wrapNext = false;
        final next = ((_cx ~/ _tabStep) + 1) * _tabStep;
        _cx = next < _cols ? next : _cols - 1;
        return;
      case 0x07: // BEL
        return;
      case 0x00:
        return;
      default:
        break;
    }

    switch (_state) {
      case _EscState.normal:
        _putRune(rune);
      case _EscState.escape:
        _handleEscape(rune);
      case _EscState.csi:
        _handleCsi(rune);
      case _EscState.osc:
        _handleOsc(rune);
    }
  }

  _EscState _state = _EscState.normal;
  final List<int> _params = [];
  int _curParam = -1;
  String _private = '';

  void _handleEscape(int rune) {
    switch (rune) {
      case 0x5b: // [
        _params.clear();
        _curParam = -1;
        _private = '';
        _state = _EscState.csi;
      case 0x5d: // ]  OSC
        _state = _EscState.osc;
      case 0x37: // 7  save cursor
        _savedX = _cx;
        _savedY = _cy;
        _savedStyle = _style.clone();
        _state = _EscState.normal;
      case 0x38: // 8  restore cursor
        _cx = _savedX;
        _cy = _savedY;
        _style = _savedStyle.clone();
        _state = _EscState.normal;
      case 0x63: // c  RIS 全清
        _resetFull();
        _state = _EscState.normal;
      case 0x44: // D  index
        _lineFeed(1);
        _state = _EscState.normal;
      case 0x4d: // M  reverse index
        _reverseIndex(1);
        _state = _EscState.normal;
      case 0x45: // E  next line
        _cx = 0;
        _lineFeed(1);
        _state = _EscState.normal;
      default:
        // (B/(0 之类的字符集切换、=> 之类未支持序列)直接吞掉
        _state = _EscState.normal;
    }
  }

  int _savedX = 0;
  int _savedY = 0;
  TermStyle _savedStyle = TermStyle();

  void _handleOsc(int rune) {
    // 标题之类不画, BEL 或 ST(Esc \) 结束
    if (rune == 0x07) _state = _EscState.normal;
    if (rune == 0x5c) _state = _EscState.normal; // '\' 需先经 escape 分支, 这里兜住
  }

  void _handleCsi(int rune) {
    if (rune >= 0x30 && rune <= 0x39) {
      _curParam = (_curParam < 0 ? 0 : _curParam) * 10 + (rune - 0x30);
      return;
    }
    if (rune == 0x3b) {
      _params.add(_curParam < 0 ? 0 : _curParam);
      _curParam = -1;
      return;
    }
    if (rune >= 0x3c && rune <= 0x3f) {
      _private = String.fromCharCode(rune); // '<' '=' '>' '?'
      return;
    }
    if (rune >= 0x20 && rune <= 0x2f) {
      if (rune == 0x3a) return; // 子参数分隔, 不支持
      _intermediate = String.fromCharCode(rune);
      return;
    }
    if (rune >= 0x40 && rune <= 0x7e) {
      _params.add(_curParam < 0 ? 0 : _curParam);
      _curParam = -1;
      _dispatchCsi(String.fromCharCode(rune));
      _state = _EscState.normal;
      return;
    }
    // 异常字节：丢掉当前序列, 回到普通状态
    _state = _EscState.normal;
  }

  // ignore: unused_field
  String _intermediate = '';

  int _p(int index, int fallback) =>
      index < _params.length ? (_params[index] == 0 ? fallback : _params[index]) : fallback;

  void _dispatchCsi(String cmd) {
    if (_private == '?') {
      switch (cmd) {
        case 'h':
        case 'l':
          final on = cmd == 'h';
          for (final p in _params) {
            switch (p) {
              case 1:
                _appCursorKeys = on;
              case 2026:
                _bracketedPaste = on;
              default:
                break; // 其余私有模式（光标可见、换行、alt screen…）忽略
            }
          }
          return;
      }
    }
    switch (cmd) {
      case 'A':
        _moveCursor(0, -_p(0, 1));
      case 'B':
        _moveCursor(0, _p(0, 1));
      case 'C':
        _moveCursor(_p(0, 1), 0);
      case 'D':
        _moveCursor(-_p(0, 1), 0);
      case 'E':
        _cy += _p(0, 1);
        _cx = 0;
      case 'F':
        _cy -= _p(0, 1);
        _cx = 0;
      case 'G':
        _cx = _clampCol(_p(0, 1) - 1);
      case 'd':
        _cy = _clampRow(_p(0, 1) - 1);
      case 'H':
      case 'f':
        _cx = _clampCol(_p(1, 1) - 1);
        _cy = _clampRow(_p(0, 1) - 1);
      case 'J':
        _eraseDisplay(_p(0, 0));
      case 'K':
        _eraseLine(_p(0, 0));
      case 'L':
        _insertLines(_p(0, 1));
      case 'M':
        _deleteLines(_p(0, 1));
      case 'P':
        _deleteChars(_p(0, 1));
      case '@':
        _insertChars(_p(0, 1));
      case 'X':
        _eraseChars(_p(0, 1));
      case 'S':
        _scrollUp(_p(0, 1));
      case 'T':
        _scrollDown(_p(0, 1));
      case 'r':
        _top = _clampRow(_p(0, 1) - 1);
        _bottom = _clampRow(_p(1, _rows) - 1);
        if (_bottom <= _top) {
          _top = 0;
          _bottom = _rows - 1;
        }
        _cx = 0;
        _cy = _top;
      case 'm':
        _applySgr();
      case 'g':
        _tabStep = _p(0, 0) == 3 ? 8 : _tabStep;
      case 't':
        break; // 窗口操作, 忽略
      default:
        break;
    }
  }

  // ---------- 绘制 ----------

  /// 本地回显：远程 pty 处于 canonical 模式（逐行回显）时，输入字符在回车前
  /// 不会从服务端回来，这里直接画到光标处；回车后服务端会重画整行，不冲突。
  void localEcho(String text) {
    for (final rune in text.runes) {
      if (rune < 0x20 || rune == 0x7f) continue;
      _putRune(rune);
    }
  }

  void _putRune(int rune) {
    final ch = String.fromCharCode(rune);
    if (_wrapNext) {
      _wrapNext = false;
      _cx = 0;
      _lineFeed(1);
    }
    if (_cx >= _cols) {
      _cx = _cols - 1;
    }
    _setCell(_cy, _cx, TermCell(ch, _effectiveFg, _effectiveBg, _style.bold, _style.italic, _style.underline));
    if (_cx == _cols - 1) {
      _wrapNext = true; //  delayed wrap
    } else {
      _cx++;
    }
  }

  int get _effectiveFg => _style.reverse ? _style.bg : _style.fg;
  int get _effectiveBg => _style.reverse ? _style.fg : _style.bg;

  void _setCell(int row, int col, TermCell cell) {
    final line = _lines[_clampRow(row)];
    line[_clampCol(col)] = cell;
  }

  int _clampCol(int c) => c < 0 ? 0 : (c >= _cols ? _cols - 1 : c);
  int _clampRow(int r) => r < 0 ? 0 : (r >= _rows ? _rows - 1 : r);

  void _moveCursor(int dx, int dy) {
    _cx += dx;
    if (_cx < 0) _cx = 0;
    if (_cx >= _cols) _cx = _cols - 1;
    _cy += dy;
    if (_cy < 0) _cy = 0;
    if (_cy >= _rows) _cy = _rows - 1;
    _wrapNext = false;
  }

  void _lineFeed(int n) {
    for (var i = 0; i < n; i++) {
      if (_cy == _bottom) {
        _scrollUp(1);
      } else if (_cy < _rows - 1) {
        _cy++;
      }
    }
  }

  void _reverseIndex(int n) {
    for (var i = 0; i < n; i++) {
      if (_cy == _top) {
        _scrollDown(1);
      } else if (_cy > 0) {
        _cy--;
      }
    }
  }

  void _scrollUp(int n) {
    for (var i = 0; i < n; i++) {
      final removed = _lines.removeAt(_top);
      if (_top == 0 && _bottom == _rows - 1) {
        _scrollback.add(removed);
        if (_scrollback.length > scrollbackLimit) _scrollback.removeAt(0);
      }
      _lines.insert(_bottom, _blankLine(_cols));
    }
    _cx = _cx.clamp(0, _cols - 1);
  }

  void _scrollDown(int n) {
    for (var i = 0; i < n; i++) {
      _lines.removeAt(_bottom);
      _lines.insert(_top, _blankLine(_cols));
    }
  }

  void _insertLines(int n) {
    if (_cy < _top || _cy > _bottom) return;
    for (var i = 0; i < n; i++) {
      _lines.insert(_cy, _blankLine(_cols));
      _lines.removeAt(_bottom);
    }
  }

  void _deleteLines(int n) {
    if (_cy < _top || _cy > _bottom) return;
    for (var i = 0; i < n; i++) {
      _lines.removeAt(_cy);
      _lines.insert(_bottom, _blankLine(_cols));
    }
  }

  void _deleteChars(int n) {
    final line = _lines[_cy];
    for (var i = 0; i < n; i++) {
      if (_cx < line.length) {
        line.removeAt(_cx);
        line.add(const TermCell(' ', -1, -1, false, false, false));
      }
    }
    _lines[_cy] = line;
  }

  void _insertChars(int n) {
    final line = _lines[_cy];
    for (var i = 0; i < n; i++) {
      if (_cx < line.length) {
        line.insert(_cx, const TermCell(' ', -1, -1, false, false, false));
        line.removeAt(line.length - 1);
      }
    }
    _lines[_cy] = line;
  }

  void _eraseChars(int n) {
    final line = _lines[_cy];
    for (var i = 0; i < n && _cx + i < line.length; i++) {
      line[_cx + i] = const TermCell(' ', -1, -1, false, false, false);
    }
    _lines[_cy] = line;
  }

  void _eraseDisplay(int mode) {
    switch (mode) {
      case 0:
        _eraseLine(0);
        for (var r = _cy + 1; r < _rows; r++) {
          _lines[r] = _blankLine(_cols);
        }
      case 1:
        _eraseLine(1);
        for (var r = 0; r < _cy; r++) {
          _lines[r] = _blankLine(_cols);
        }
      case 2:
      case 3:
        for (var r = 0; r < _rows; r++) {
          _lines[r] = _blankLine(_cols);
        }
        _cx = 0;
        _cy = 0;
    }
  }

  void _eraseLine(int mode) {
    final line = _lines[_cy];
    final blank = const TermCell(' ', -1, -1, false, false, false);
    switch (mode) {
      case 0:
        for (var c = _cx; c < line.length; c++) {
          line[c] = blank;
        }
      case 1:
        for (var c = 0; c <= _cx && c < line.length; c++) {
          line[c] = blank;
        }
      case 2:
        _lines[_cy] = _blankLine(_cols);
    }
    _lines[_cy] = line;
  }

  void _applySgr() {
    if (_params.isEmpty || (_params.length == 1 && _params[0] == 0)) {
      _style = TermStyle();
      return;
    }
    var i = 0;
    // 逐个处理, 参数列表里 _p(i) 依赖 _params 顺序
    while (i < _params.length) {
      final code = _params[i];
      switch (code) {
        case 0:
          _style = TermStyle();
        case 1:
          _style.bold = true;
        case 2:
          _style.faint = true;
        case 3:
          _style.italic = true;
        case 4:
          _style.underline = true;
        case 22:
          _style.bold = false;
          _style.faint = false;
        case 23:
          _style.italic = false;
        case 24:
          _style.underline = false;
        case 7:
          _style.reverse = true;
        case 27:
          _style.reverse = false;
        case 39:
          _style.fg = -1;
        case 49:
          _style.bg = -1;
        case 38:
          i = _read256(i, true);
        case 48:
          i = _read256(i, false);
        default:
          if (code >= 30 && code <= 37) {
            _style.fg = code - 30;
          } else if (code >= 40 && code <= 47) {
            _style.bg = code - 40;
          } else if (code >= 90 && code <= 97) {
            _style.fg = code - 90 + 8;
          } else if (code >= 100 && code <= 107) {
            _style.bg = code - 100 + 8;
          }
      }
      i++;
    }
  }

  /// 处理 `38;5;N` / `38;2;r;g;b`，返回消费到的下标。
  int _read256(int i, bool isFg) {
    if (i + 1 < _params.length) {
      final mode = _params[i + 1];
      if (mode == 5 && i + 2 < _params.length) {
        final index = _params[i + 2];
        if (isFg) {
          _style.fg = index;
        } else {
          _style.bg = index;
        }
        return i + 2;
      }
      if (mode == 2 && i + 4 < _params.length) {
        final r = _params[i + 2], g = _params[i + 3], b = _params[i + 4];
        final packed = 0x1000000 | (r << 16) | (g << 8) | b;
        if (isFg) {
          _style.fg = packed;
        } else {
          _style.bg = packed;
        }
        return i + 4;
      }
    }
    return i;
  }

  void _resetFull() {
    for (var r = 0; r < _rows; r++) {
      _lines[r] = _blankLine(_cols);
    }
    _scrollback.clear();
    _cx = 0;
    _cy = 0;
    _top = 0;
    _bottom = _rows - 1;
    _style = TermStyle();
    _wrapNext = false;
    _state = _EscState.normal;
    _appCursorKeys = false;
    _bracketedPaste = false;
  }

  void resize(int cols, int rows) {
    if (cols < 1 || rows < 1) return;
    if (cols != _cols) {
      for (var r = 0; r < _lines.length; r++) {
        _lines[r] = _resizeLine(_lines[r], cols);
      }
      _cols = cols;
    }
    if (rows != _rows) {
      while (_lines.length < rows) {
        _lines.add(_blankLine(_cols));
      }
      while (_lines.length > rows) {
        _lines.removeLast();
      }
      _rows = rows;
    }
    _top = 0;
    _bottom = _rows - 1;
    _cx = _cx.clamp(0, _cols - 1);
    _cy = _cy.clamp(0, _rows - 1);
  }

  List<TermCell> _resizeLine(List<TermCell> line, int cols) {
    final out = List<TermCell>.from(line);
    while (out.length < cols) {
      out.add(const TermCell(' ', -1, -1, false, false, false));
    }
    if (out.length > cols) out.removeRange(cols, out.length);
    return out;
  }

  /// 供 UI 直接渲染的一行富文本（忽略全空尾列）。
  String lineText(int row) {
    if (row < 0 || row >= _lines.length) return '';
    final cells = _lines[row];
    var end = cells.length;
    while (end > 0 && cells[end - 1].ch == ' ') {
      end--;
    }
    final sb = StringBuffer();
    for (var i = 0; i < end; i++) {
      sb.write(cells[i].ch);
    }
    return sb.toString();
  }

  String plainText() {
    final sb = StringBuffer();
    for (final line in _lines) {
      var text = line.map((c) => c.ch).join();
      while (text.endsWith(' ')) {
        text = text.substring(0, text.length - 1);
      }
      sb.writeln(text);
    }
    return sb.toString();
  }
}

enum _EscState { normal, escape, csi, osc }

/// UTF-8 增量解码：跨帧的多字节字符不会断。
class Utf8LineDecoder {
  Utf8LineDecoder(this._sink);
  final void Function(int rune) _sink;

  final List<int> _pending = [];

  void add(List<int> bytes) {
    if (_pending.isEmpty && _isAscii(bytes)) {
      for (final b in bytes) {
        _sink(b);
      }
      return;
    }
    final all = <int>[..._pending, ...bytes];
    var consumed = 0;
    var i = 0;
    while (i < all.length) {
      final b = all[i];
      if (b < 0x80) {
        _sink(b);
        i++;
        consumed = i;
        continue;
      }
      final len = (b & 0xE0) == 0xC0
          ? 2
          : (b & 0xF0) == 0xE0
              ? 3
              : (b & 0xF8) == 0xF0
                  ? 4
                  : 1;
      if (i + len > all.length) break; // 等下一帧
      var rune = b & (0xFF >> (len + 1));
      for (var k = 1; k < len; k++) {
        rune = (rune << 6) | (all[i + k] & 0x3F);
      }
      _sink(rune);
      i += len;
      consumed = i;
    }
    _pending
      ..clear()
      ..addAll(all.sublist(consumed.clamp(0, all.length)));
    if (_pending.length > 8) {
      // 明显不是合法 UTF-8 序列, 丢弃避免无限增长
      for (final p in _pending) {
        _sink(p < 0x80 ? p : 0xFFFD);
      }
      _pending.clear();
    }
  }

  static bool _isAscii(List<int> bytes) {
    for (final b in bytes) {
      if (b >= 0x80) return false;
    }
    return true;
  }
}
