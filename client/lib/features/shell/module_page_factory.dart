import 'package:flutter/material.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/features/monitoring/monitoring_page.dart';
import 'package:hitest/features/server/server_list_page.dart';
import 'package:hitest/features/shell/models/client_module.dart';
import 'package:hitest/features/shell/widgets/no_server_selected_state.dart';
import 'package:hitest/pages/settings/settings_page.dart';

/// 构建 shell 容器内的模块页面。
Widget buildShellModulePage(
  BuildContext context, {
  required ClientModule module,
  required String? serverId,
  bool useStableModuleKey = false,
}) {
  switch (module) {
    case ClientModule.servers:
      return const ServerListPage();
    case ClientModule.monitoring:
      if (serverId == null) {
        return NoServerSelectedState(moduleName: context.l10n.serverModuleMonitoring);
      }
      final monitoringPage = const MonitoringPage();
      return useStableModuleKey
          ? monitoringPage
          : KeyedSubtree(
              key: ValueKey('monitoring:$serverId'),
              child: monitoringPage,
            );
    case ClientModule.settings:
      return const SettingsPage();
  }
}
