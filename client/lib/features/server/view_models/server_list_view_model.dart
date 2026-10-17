import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/core/presentation/safe_change_notifier.dart';
import 'package:hitest/core/utils/platform_utils.dart';
import 'package:hitest/features/server/server_form_page.dart';
import 'package:hitest/features/server/server_models.dart';
import 'package:hitest/features/server/server_provider.dart';
import 'package:hitest/features/shell/controllers/current_server_controller.dart';

class ServerListViewModel extends ChangeNotifier with SafeChangeNotifier {
  ServerListViewModel({
    required ServerProvider serverProvider,
  }) : _serverProvider = serverProvider {
    _searchController.addListener(_onSearchChanged);
  }

  final ServerProvider _serverProvider;
  final TextEditingController _searchController = TextEditingController();

  final GlobalKey addKey = GlobalKey();

  TextEditingController get searchController => _searchController;
  ServerProvider get serverProvider => _serverProvider;

  void init(BuildContext context) {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _serverProvider.load().then((_) {
        if (_serverProvider.servers.isNotEmpty) {
          _serverProvider.loadMetrics();
        }
      });
    });
    // Add listener but prevent recursive updates
    _serverProvider.addListener(_onProviderChanged);
  }

  void _onProviderChanged() {
    // Just notify listeners without triggering coach marks
    // Coach marks will be shown only on initial load
    notifyListeners();
  }

  @override
  void dispose() {
    _serverProvider.removeListener(_onProviderChanged);
    _searchController.dispose();
    super.dispose();
  }

  void _onSearchChanged() {
    notifyListeners();
  }

  List<ServerCardViewModel> get filteredServers {
    final query = _searchController.text.trim().toLowerCase();
    if (query.isEmpty) {
      return _serverProvider.servers;
    }
    return _serverProvider.servers.where((item) {
      return item.config.name.toLowerCase().contains(query) ||
          item.config.url.toLowerCase().contains(query);
    }).toList();
  }

  Future<void> refresh() async {
    await _serverProvider.load();
    if (_serverProvider.servers.isNotEmpty) {
      await _serverProvider.loadMetrics();
    }
  }

  Future<void> openAddServer(BuildContext context) async {
    final currentServer = context.read<CurrentServerController>();
    final bool? result;
    if (PlatformUtils.isDesktop(context)) {
      result = await showDialog<bool>(
        context: context,
        builder: (dialogContext) => Dialog(
          insetPadding: const EdgeInsets.all(24),
          clipBehavior: Clip.antiAlias,
          child: ConstrainedBox(
            constraints: const BoxConstraints(
              maxWidth: 720,
              maxHeight: 820,
            ),
            child: const ServerFormPage(),
          ),
        ),
      );
    } else {
      result = await Navigator.push<bool>(
        context,
        MaterialPageRoute(builder: (_) => const ServerFormPage()),
      );
    }

    if (result == true) {
      await _serverProvider.load();
      if (context.mounted) {
        await currentServer.refresh();
      }
    }
  }

  /// 卡片点击 = 把这台设为「当前机器」，之后所有 tab（监控/文件/终端）都作用于它。
  /// 不再跳详情页——功能入口都在底部 tab，详情页是多余的壳。
  Future<void> selectAsCurrent(
      BuildContext context, ServerCardViewModel item) async {
    if (item.isCurrent) return;
    final l10n = context.l10n;
    final messenger = ScaffoldMessenger.of(context);
    await context.read<CurrentServerController>().selectServer(item.config.id);
    await _serverProvider.load();
    messenger.showSnackBar(
      SnackBar(content: Text(l10n.serverSwitched(item.config.name))),
    );
  }

  Future<void> deleteServer(
      BuildContext context, ServerCardViewModel item) async {
    final l10n = context.l10n;
    final confirmed = await showDialog<bool>(
          context: context,
          builder: (dialogContext) => AlertDialog(
            title: Text(l10n.serverDeleteConfirmTitle),
            content: Text(l10n.serverDeleteConfirmMessage(item.config.name)),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: Text(l10n.commonCancel),
              ),
              FilledButton(
                onPressed: () => Navigator.pop(dialogContext, true),
                child: Text(l10n.commonDelete),
              ),
            ],
          ),
        ) ??
        false;

    if (!confirmed || !context.mounted) return;

    final currentServerController = context.read<CurrentServerController>();
    final messenger = ScaffoldMessenger.of(context);
    final successMessage = l10n.serverDeleteSuccess(item.config.name);

    try {
      await _serverProvider.delete(item.config.id);
      await currentServerController.refresh();
      if (!context.mounted) return;
      messenger.showSnackBar(SnackBar(content: Text(successMessage)));
    } catch (error) {
      if (!context.mounted) return;
      messenger.showSnackBar(
        SnackBar(content: Text(l10n.serverDeleteFailed(error.toString()))),
      );
    }
  }
}
