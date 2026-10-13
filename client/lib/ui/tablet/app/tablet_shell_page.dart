import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/features/shell/controllers/current_server_controller.dart';
import 'package:hitest/features/shell/models/client_module.dart';
import 'package:hitest/features/shell/module_page_factory.dart';
import 'package:hitest/features/shell/widgets/server_switcher_action.dart';
import 'package:hitest/ui/routing/ui_target.dart';

class TabletShellPage extends StatefulWidget {
  const TabletShellPage({
    super.key,
    this.initialIndex = 0,
    this.initialModuleId,
    required this.tabletKind,
  });

  final int initialIndex;
  final String? initialModuleId;
  final TabletKind tabletKind;

  @override
  State<TabletShellPage> createState() => _TabletShellPageState();
}

class _TabletShellPageState extends State<TabletShellPage> {
  static const List<ClientModule> _modules = [
    ClientModule.servers,
    ClientModule.monitoring,
    ClientModule.settings,
  ];

  late ClientModule _selectedModule;

  @override
  void initState() {
    super.initState();
    _selectedModule = clientModuleFromId(widget.initialModuleId) ??
        _moduleFromIndex(widget.initialIndex);
  }

  ClientModule _moduleFromIndex(int index) {
    if (index < 0 || index >= _modules.length) {
      return ClientModule.servers;
    }
    return _modules[index];
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<CurrentServerController>(
      builder: (context, currentServer, _) {
        final modules = _modules;
        final selectedModule =
            (!currentServer.hasServer && _selectedModule.requiresServer)
                ? ClientModule.servers
                : _selectedModule;
        final selectedIndex = modules.contains(selectedModule)
            ? modules.indexOf(selectedModule)
            : 0;
        final disabledColor =
            Theme.of(context).colorScheme.onSurfaceVariant.withValues(alpha: 0.42);

        return Scaffold(
          body: Row(
            children: [
              NavigationRail(
                extended: false,
                selectedIndex: selectedIndex,
                onDestinationSelected: (index) {
                  final module = modules[index];
                  if (!currentServer.hasServer && module.requiresServer) {
                    ServerSwitcherAction.showServerPicker(context);
                    return;
                  }
                  setState(() {
                    _selectedModule = module;
                  });
                },
                labelType: NavigationRailLabelType.all,
                useIndicator: true,
                indicatorColor: Theme.of(context).colorScheme.primaryContainer,
                destinations: [
                  for (final module in modules)
                    NavigationRailDestination(
                      icon: Icon(
                        module.icon,
                        color: !module.requiresServer || currentServer.hasServer
                            ? null
                            : disabledColor,
                      ),
                      selectedIcon: Icon(
                        module.selectedIcon,
                        color: !module.requiresServer || currentServer.hasServer
                            ? null
                            : disabledColor,
                      ),
                      label: Text(module.label(context.l10n)),
                    ),
                ],
              ),
              const VerticalDivider(thickness: 1, width: 1),
              Expanded(
                child: buildShellModulePage(
                  context,
                  module: selectedModule,
                  serverId: currentServer.currentServerId,
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
