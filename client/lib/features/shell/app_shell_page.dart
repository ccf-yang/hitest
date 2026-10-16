import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/features/shell/controllers/current_server_controller.dart';
import 'package:hitest/features/shell/models/client_module.dart';
import 'package:hitest/features/shell/module_page_factory.dart';
import 'package:hitest/features/shell/widgets/server_switcher_action.dart';
import 'package:hitest/widgets/navigation/app_bottom_navigation_bar.dart';

const List<ClientModule> kShellModules = [
  ClientModule.servers,
  ClientModule.monitoring,
  ClientModule.files,
  ClientModule.terminal,
  ClientModule.settings,
];

class AppShellPage extends StatefulWidget {
  const AppShellPage({
    super.key,
    this.initialIndex = 0,
    this.initialModuleId,
  });

  final int initialIndex;
  final String? initialModuleId;

  @override
  State<AppShellPage> createState() => _AppShellPageState();
}

class _AppShellPageState extends State<AppShellPage> {
  late ClientModule _selectedModule;

  @override
  void initState() {
    super.initState();
    _selectedModule = clientModuleFromId(widget.initialModuleId) ??
        _moduleFromIndex(widget.initialIndex);
  }

  ClientModule _moduleFromIndex(int index) {
    if (index < 0 || index >= kShellModules.length) {
      return ClientModule.servers;
    }
    return kShellModules[index];
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<CurrentServerController>(
      builder: (context, currentServer, _) {
        final modules = kShellModules;
        final selectedModule =
            (!currentServer.hasServer && _selectedModule.requiresServer)
                ? ClientModule.servers
                : _selectedModule;
        final selectedIndex = modules.contains(selectedModule)
            ? modules.indexOf(selectedModule)
            : 0;

        final body = buildShellModulePage(
          context,
          module: selectedModule,
          serverId: currentServer.currentServerId,
        );

        final items = [
          for (final module in modules)
            AppNavigationBarItem(
              icon: module.icon,
              selectedIcon: module.selectedIcon,
              label: module.label(context.l10n),
              enabled: !module.requiresServer || currentServer.hasServer,
            ),
        ];

        return Scaffold(
          body: body,
          bottomNavigationBar: AppBottomNavigationBar(
            currentIndex: selectedIndex,
            onTap: (index) => _handleModuleSelection(
              modules[index],
              currentServer.hasServer,
            ),
            items: items,
          ),
        );
      },
    );
  }

  Future<void> _handleModuleSelection(
      ClientModule module, bool hasServer) async {
    if (!hasServer && module.requiresServer) {
      await ServerSwitcherAction.showServerPicker(context);
      return;
    }
    if (!mounted) return;
    setState(() {
      _selectedModule = module;
    });
  }
}
