import 'package:flutter/material.dart';
import 'package:hitest/config/app_router.dart';
import 'package:hitest/core/utils/platform_utils.dart';
import 'package:hitest/features/shell/models/client_module.dart';

class ShellRouteTarget {
  const ShellRouteTarget({
    required this.module,
    required this.embedRouteInShell,
  });

  final ClientModule module;
  final bool embedRouteInShell;
}

ClientModule? shellModuleForRoute(String route) {
  return shellRouteTargetForRoute(route)?.module;
}

ShellRouteTarget? shellRouteTargetForRoute(String route) {
  switch (route) {
    case AppRoutes.server:
    case AppRoutes.serverSelection:
      return const ShellRouteTarget(
        module: ClientModule.servers,
        embedRouteInShell: false,
      );
    case AppRoutes.serverConfig:
    case AppRoutes.serverDetail:
      return const ShellRouteTarget(
        module: ClientModule.servers,
        embedRouteInShell: true,
      );
    case AppRoutes.monitoring:
      return const ShellRouteTarget(
        module: ClientModule.monitoring,
        embedRouteInShell: false,
      );
    case AppRoutes.settings:
      return const ShellRouteTarget(
        module: ClientModule.settings,
        embedRouteInShell: false,
      );
  }
  return null;
}

Future<void> openRouteRespectingShell(
  BuildContext context,
  String route, {
  Object? arguments,
}) {
  final target = shellRouteTargetForRoute(route);
  if (target != null && PlatformUtils.isDesktop(context)) {
    final shellArguments = <String, dynamic>{
      'module': target.module.storageId,
    };
    if (target.embedRouteInShell) {
      shellArguments['route'] = route;
      if (arguments != null) {
        shellArguments['routeArgs'] = arguments;
      }
    }

    return Navigator.of(context).pushReplacementNamed(
      AppRoutes.home,
      arguments: shellArguments,
    );
  }

  return Navigator.of(context).pushNamed(route, arguments: arguments);
}
