import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:hitest/config/app_router.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/core/theme/app_design_tokens.dart';
import 'package:hitest/features/server/server_models.dart';
import 'package:hitest/features/server/server_provider.dart';
import 'package:hitest/features/server/widgets/server_detail_header_card.dart';
import 'package:hitest/features/server/widgets/server_detail_section_card.dart';
import 'package:hitest/features/shell/controllers/current_server_controller.dart';
import 'package:hitest/features/quick/quick_commands_page.dart';
import 'package:hitest/features/shell/shell_navigation.dart';
import 'package:hitest/features/shell/widgets/server_switcher_action.dart';

class ServerDetailPage extends StatelessWidget {
  const ServerDetailPage({
    super.key,
    required this.server,
  });

  final ServerCardViewModel server;

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    final scheme = Theme.of(context).colorScheme;
    final currentServerId =
        context.watch<CurrentServerController>().currentServerId;
    final serverProvider = context.watch<ServerProvider>();
    final activeServer = serverProvider.servers.firstWhere(
      (item) => item.config.id == currentServerId,
      orElse: () => server,
    );

    return Scaffold(
      backgroundColor: scheme.surface,
      appBar: AppBar(
        title: Text(l10n.serverDetailTitle),
        actions: [
          ServerSwitcherAction(onChanged: () => serverProvider.load()),
          IconButton(
            tooltip: l10n.commonDelete,
            icon: const Icon(Icons.delete_outline),
            onPressed: () => _deleteServer(context, activeServer),
          ),
        ],
      ),
      body: ColoredBox(
        color: scheme.surface,
        child: ListView(
          padding: AppDesignTokens.pagePadding,
          children: [
            ServerDetailHeaderCard(
              server: activeServer,
              onRefresh: () => _refreshServer(context),
            ),
            const SizedBox(height: AppDesignTokens.spacingLg),
            ServerDetailSectionCard(
              title: l10n.commonMore,
              items: _buildServerModuleItems(context),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _refreshServer(BuildContext context) async {
    final l10n = context.l10n;
    final provider = context.read<ServerProvider>();
    await provider.load();
    await provider.loadMetrics();
    if (!context.mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.commonRefresh)),
    );
  }

  List<ServerDetailSectionItem> _buildServerModuleItems(BuildContext context) {
    final l10n = context.l10n;
    return [
      ServerDetailSectionItem(
        title: l10n.serverModuleMonitoring,
        icon: Icons.monitor_heart_outlined,
        onTap: () => openRouteRespectingShell(context, AppRoutes.monitoring),
      ),
      ServerDetailSectionItem(
        title: l10n.quickCommandsTitle,
        icon: Icons.terminal_outlined,
        onTap: () => Navigator.push(
          context,
          MaterialPageRoute(builder: (_) => const QuickCommandsPage()),
        ),
      ),
      ServerDetailSectionItem(
        title: l10n.settingsPageTitle,
        icon: Icons.settings_outlined,
        onTap: () => openRouteRespectingShell(context, AppRoutes.settings),
      ),
    ];
  }

  Future<void> _deleteServer(
    BuildContext context,
    ServerCardViewModel target,
  ) async {
    final l10n = context.l10n;
    final confirmed = await showDialog<bool>(
          context: context,
          builder: (dialogContext) => AlertDialog(
            title: Text(l10n.serverDeleteConfirmTitle),
            content: Text(l10n.serverDeleteConfirmMessage(target.config.name)),
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
    if (!confirmed || !context.mounted) {
      return;
    }

    final serverProvider = context.read<ServerProvider>();
    final currentServerController = context.read<CurrentServerController>();
    final messenger = ScaffoldMessenger.of(context);
    final navigator = Navigator.of(context);
    final successMessage = l10n.serverDeleteSuccess(target.config.name);

    try {
      // Close the page first to avoid widget tree issues during deletion
      navigator.pop();

      await serverProvider.delete(target.config.id);
      await currentServerController.refresh();

      if (!context.mounted) {
        return;
      }
      messenger.showSnackBar(
        SnackBar(content: Text(successMessage)),
      );
    } catch (error) {
      if (!context.mounted) {
        return;
      }
      messenger.showSnackBar(
        SnackBar(
          content: Text(l10n.serverDeleteFailed(error.toString())),
        ),
      );
    }
  }
}
