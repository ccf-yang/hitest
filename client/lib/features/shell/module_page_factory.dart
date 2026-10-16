import 'package:flutter/material.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/features/files/files_page.dart';
import 'package:hitest/features/monitoring/monitoring_page.dart';
import 'package:hitest/features/server/server_list_page.dart';
import 'package:hitest/features/shell/models/client_module.dart';
import 'package:hitest/features/shell/widgets/no_server_selected_state.dart';
import 'package:hitest/features/terminal/terminal_page.dart';
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
      return _serverScoped(
        context: context,
        module: module,
        serverId: serverId,
        useStableModuleKey: useStableModuleKey,
        child: const MonitoringPage(),
      );
    case ClientModule.files:
      return _serverScoped(
        context: context,
        module: module,
        serverId: serverId,
        useStableModuleKey: useStableModuleKey,
        child: const FilesPage(),
      );
    case ClientModule.terminal:
      return _serverScoped(
        context: context,
        module: module,
        serverId: serverId,
        useStableModuleKey: useStableModuleKey,
        child: const TerminalPage(),
      );
    case ClientModule.settings:
      return const SettingsPage();
  }
}

/// 依赖服务器的模块统一处理「未选服务器」，并把 serverId 编进 Key,
/// 切换机器时强制重建 State, 免得页面带着上一台机器的缓存。
Widget _serverScoped({
  required BuildContext context,
  required ClientModule module,
  required String? serverId,
  required bool useStableModuleKey,
  required Widget child,
}) {
  if (serverId == null) {
    return NoServerSelectedState(moduleName: module.label(context.l10n));
  }
  return useStableModuleKey
      ? child
      : KeyedSubtree(
          key: ValueKey('${module.storageId}:$serverId'),
          child: child,
        );
}
