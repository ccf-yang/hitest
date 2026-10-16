import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';

import 'package:hitest/api/v2/files_v2.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/core/network/api_client_manager.dart';
import 'package:hitest/core/services/file_save_service.dart';
import 'package:hitest/core/theme/app_design_tokens.dart';
import 'package:hitest/shared/widgets/operations/module_empty_state_widget.dart';
import 'package:hitest/shared/widgets/operations/module_error_state_widget.dart';

/// 文件页：浏览/预览/上传/下载/新建目录/重命名/删除目标机器上的文件。
/// 全部走服务端 files 模块的裸 HTTP 端点，大文件流式，无分块协议。
class FilesPage extends StatefulWidget {
  const FilesPage({super.key, this.initialPath = '/'});

  final String initialPath;

  @override
  State<FilesPage> createState() => _FilesPageState();
}

class _FilesPageState extends State<FilesPage> {
  late String _path = widget.initialPath;
  List<FileEntry>? _entries;
  String? _error;
  bool _busy = false;

  Future<FilesV2Api> _ensureApi() async {
    final client = await ApiClientManager().getCurrentClient();
    return FilesV2Api(client);
  }

  Future<void> _load([String? path]) async {
    final target = path ?? _path;
    setState(() {
      _error = null;
      _busy = true;
    });
    try {
      final api = await _ensureApi();
      final entries = await api.list(target);
      if (!mounted) return;
      setState(() {
        _path = target;
        _entries = entries;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = e.toString());
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) _load();
    });
  }

  /// 串行执行一个远端操作：失败只弹一条 snackbar，成功后调用方自行刷新。
  Future<void> _run(Future<void> Function() action) async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      await action();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(context.l10n.filesOpFailed(e.toString()))),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  // ---------- 操作 ----------

  Future<void> _mkdir() async {
    final name = await _promptText(
      title: context.l10n.filesNewFolder,
      hint: context.l10n.filesNewFolderHint,
    );
    if (name == null || name.trim().isEmpty) return;
    await _run(() async {
      final api = await _ensureApi();
      await api.mkdir(joinPath(_path, name.trim()));
      await _load();
    });
  }

  Future<void> _rename(FileEntry entry) async {
    final name = await _promptText(
      title: context.l10n.filesRename,
      hint: context.l10n.filesRenameHint,
      initial: entry.name,
    );
    if (name == null || name.trim().isEmpty || name.trim() == entry.name) {
      return;
    }
    await _run(() async {
      final api = await _ensureApi();
      await api.rename(
        entry.path,
        joinPath(parentPath(entry.path), name.trim()),
      );
      await _load();
    });
  }

  Future<void> _delete(FileEntry entry) async {
    final l10n = context.l10n;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(entry.name),
        content: Text(l10n.filesDeleteConfirm(entry.name)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(l10n.commonCancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(l10n.commonDelete),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    await _run(() async {
      final api = await _ensureApi();
      await api.delete(entry.path);
      await _load();
    });
  }

  Future<void> _download(FileEntry entry) async {
    await _run(() async {
      final api = await _ensureApi();
      final bytes = await api.downloadToBytes(entry.path);
      final result = await FileSaveService().saveFile(
        fileName: entry.name,
        bytes: bytes,
      );
      if (result.success && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              '${context.l10n.filesDownloaded}\n${result.filePath ?? ''}',
            ),
          ),
        );
      }
    });
  }

  Future<void> _preview(FileEntry entry) async {
    String? content;
    await _run(() async {
      final api = await _ensureApi();
      content = await api.readPreview(entry.path, entry.size);
    });
    if (!mounted || content == null) return;
    final text = (content ?? '').isNotEmpty
        ? content!
        : (entry.size > FilesV2Api.previewLimit
            ? context.l10n.filesPreviewTooLarge
            : context.l10n.quickCommandsNoOutput);
    await showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(entry.name, maxLines: 1, overflow: TextOverflow.ellipsis),
        content: SizedBox(
          width: double.maxFinite,
          height: MediaQuery.of(ctx).size.height * 0.5,
          child: SingleChildScrollView(
            child: SelectableText(
              text,
              style: const TextStyle(fontFamily: 'Menlo', fontSize: 12),
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text(context.l10n.commonClose),
          ),
        ],
      ),
    );
  }

  Future<void> _upload() async {
    final picked = await FilePicker.platform.pickFiles(
      allowMultiple: true,
      withData: true,
    );
    if (picked == null || picked.files.isEmpty) return;
    await _run(() async {
      final api = await _ensureApi();
      for (final file in picked.files) {
        final bytes = file.bytes;
        if (bytes == null) continue;
        await api.uploadBytes(joinPath(_path, file.name), bytes);
      }
      await _load();
    });
  }

  Future<String?> _promptText({
    required String title,
    required String hint,
    String? initial,
  }) {
    final controller = TextEditingController(text: initial);
    return showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(title),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: InputDecoration(hintText: hint),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text(ctx.l10n.commonCancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, controller.text),
            child: Text(ctx.l10n.commonConfirm),
          ),
        ],
      ),
    ).whenComplete(controller.dispose);
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
            Text(l10n.filesTitle, style: const TextStyle(fontSize: 16)),
            Text(
              _path,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(fontSize: 12),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.arrow_upward),
            tooltip: l10n.filesUp,
            onPressed: _busy ? null : () => _load(parentPath(_path)),
          ),
          IconButton(
            icon: const Icon(Icons.home_outlined),
            tooltip: l10n.filesHome,
            onPressed: _busy ? null : () => _load('/'),
          ),
          IconButton(
            icon: const Icon(Icons.create_new_folder_outlined),
            tooltip: l10n.filesNewFolder,
            onPressed: _busy ? null : _mkdir,
          ),
          IconButton(
            icon: const Icon(Icons.upload_file_outlined),
            tooltip: l10n.filesPickToUpload,
            onPressed: _busy ? null : _upload,
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: l10n.commonRefresh,
            onPressed: _busy ? null : _load,
          ),
        ],
      ),
      body: _buildBody(context),
    );
  }

  Widget _buildBody(BuildContext context) {
    final l10n = context.l10n;
    if (_error != null && _entries == null) {
      return ModuleErrorStateWidget(message: _error!, onRetry: _load);
    }
    if (_entries == null) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_entries!.isEmpty) {
      return ModuleEmptyStateWidget(
        title: l10n.filesEmpty,
        description: _path,
        actionLabel: l10n.filesPickToUpload,
        onAction: _busy ? null : _upload,
      );
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView.builder(
        padding: const EdgeInsets.only(bottom: AppDesignTokens.spacingXl),
        itemCount: _entries!.length,
        itemBuilder: (context, index) => _entryTile(_entries![index]),
      ),
    );
  }

  Widget _entryTile(FileEntry entry) {
    final l10n = context.l10n;
    return ListTile(
      leading: Icon(_iconFor(entry)),
      title: Text(entry.name, maxLines: 1, overflow: TextOverflow.ellipsis),
      subtitle: Text(
        '${entry.isDir ? '-' : entry.sizeLabel} · '
        '${_formatTime(entry.modified)} · ${entry.mode}',
        style: const TextStyle(fontSize: 12),
      ),
      trailing: PopupMenuButton<String>(
        enabled: !_busy,
        onSelected: (value) {
          switch (value) {
            case 'preview':
              _preview(entry);
            case 'download':
              _download(entry);
            case 'rename':
              _rename(entry);
            case 'delete':
              _delete(entry);
          }
        },
        itemBuilder: (context) => [
          if (!entry.isDir && FilesV2Api.isProbablyText(entry))
            PopupMenuItem(value: 'preview', child: Text(l10n.filesPreview)),
          if (!entry.isDir)
            PopupMenuItem(value: 'download', child: Text(l10n.filesDownloaded)),
          PopupMenuItem(value: 'rename', child: Text(l10n.filesRename)),
          PopupMenuItem(value: 'delete', child: Text(l10n.commonDelete)),
        ],
      ),
      onTap: entry.isDir ? () => _load(entry.path) : () => _preview(entry),
    );
  }

  IconData _iconFor(FileEntry entry) {
    if (entry.isDir) return Icons.folder_outlined;
    if (entry.mode.contains('x')) return Icons.settings_ethernet_outlined;
    return Icons.description_outlined;
  }

  String _formatTime(DateTime dt) {
    String two(int v) => v.toString().padLeft(2, '0');
    return '${dt.year}-${two(dt.month)}-${two(dt.day)} '
        '${two(dt.hour)}:${two(dt.minute)}';
  }
}
