import 'package:flutter/material.dart';
import 'package:hitest/core/config/api_config.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/features/monitoring/monitoring_page.dart';
import 'package:hitest/features/server/server_form_page.dart';
import 'package:hitest/features/server/server_list_page.dart';
import 'package:hitest/features/shell/models/client_module.dart';
import 'package:hitest/features/shell/shell_navigation.dart';
import 'package:hitest/ui/routing/ui_route_host.dart';
import 'package:hitest/ui/routing/route_registry.dart';
import 'package:hitest/ui/routing/ui_target.dart';
import 'package:hitest/ui/routing/ui_target_resolver.dart';
import 'package:hitest/pages/settings/settings_page.dart';

class AppRoutes {
  static const String splash = '/';
  static const String home = '/home';
  static const String server = '/server';
  static const String serverConfig = '/server-config';
  static const String serverSelection = '/server-selection';
  static const String serverDetail = '/server-detail';
  static const String monitoring = '/monitoring';
  static const String settings = '/settings';
}

class AppRouter {
  static bool _routeRegistryInitialized = false;

  static Route<dynamic> generateRoute(RouteSettings settings) {
    _ensureRouteRegistryInitialized();
    final registryRoute = _buildRouteFromRegistry(settings);
    if (registryRoute != null) {
      return registryRoute;
    }

    switch (settings.name) {
      case AppRoutes.splash:
        return MaterialPageRoute(builder: (_) => const SplashPage());
      case AppRoutes.home:
        return MaterialPageRoute(
          builder: (_) => UiRouteHost(settings: settings),
        );
      case AppRoutes.server:
      case AppRoutes.serverSelection:
        return MaterialPageRoute(builder: (_) => const ServerListPage());
      case AppRoutes.serverConfig:
        return MaterialPageRoute(builder: (_) => const ServerFormPage());
      case AppRoutes.serverDetail:
        return MaterialPageRoute(builder: (_) => const ServerListPage());
      case AppRoutes.monitoring:
        return MaterialPageRoute(builder: (_) => const MonitoringPage());
      case AppRoutes.settings:
        return MaterialPageRoute(builder: (_) => const SettingsPage());
      default:
        return MaterialPageRoute(builder: (_) => const NotFoundPage());
    }
  }

  static void _ensureRouteRegistryInitialized() {
    if (_routeRegistryInitialized) {
      return;
    }
    RouteRegistry.registerAll(_buildRouteRegistryEntries());
    _routeRegistryInitialized = true;
  }

  static Route<dynamic>? _buildRouteFromRegistry(RouteSettings settings) {
    final entry = RouteRegistry.lookup(settings.name);
    if (entry == null) {
      return null;
    }

    return MaterialPageRoute(
      settings: settings,
      builder: (context) {
        final target = UiTargetResolver.resolve(context);
        final builder = entry.resolve(target);
        return builder(context, settings);
      },
    );
  }

  static Map<String, RouteEntry> _buildRouteRegistryEntries() {
    return <String, RouteEntry>{
      AppRoutes.splash: RouteEntry(
        defaultBuilder: (_, __) => const SplashPage(),
      ),
      AppRoutes.home: RouteEntry(
        defaultBuilder: (_, settings) => UiRouteHost(settings: settings),
      ),
      AppRoutes.server: _shellAwareModuleEntry(
        routeName: AppRoutes.server,
        defaultBuilder: (_, __) => const ServerListPage(),
      ),
      AppRoutes.serverSelection: _shellAwareModuleEntry(
        routeName: AppRoutes.serverSelection,
        defaultBuilder: (_, __) => const ServerListPage(),
      ),
      AppRoutes.serverConfig: _shellAwareModuleEntry(
        routeName: AppRoutes.serverConfig,
        defaultBuilder: (_, __) => const ServerFormPage(),
      ),
      AppRoutes.serverDetail: _shellAwareModuleEntry(
        routeName: AppRoutes.serverDetail,
        defaultBuilder: (_, __) => const ServerListPage(),
      ),
      AppRoutes.settings: _shellAwareModuleEntry(
        routeName: AppRoutes.settings,
        defaultBuilder: (_, __) => const SettingsPage(),
      ),
      AppRoutes.monitoring: _shellAwareModuleEntry(
        routeName: AppRoutes.monitoring,
        defaultBuilder: (_, __) => const MonitoringPage(),
      ),
    };
  }

  static RouteEntry _shellAwareModuleEntry({
    required String routeName,
    required UiRouteBuilder defaultBuilder,
  }) {
    final shellTarget = shellRouteTargetForRoute(routeName);
    if (shellTarget == null) {
      return RouteEntry(defaultBuilder: defaultBuilder);
    }

    Widget desktopBuilder(BuildContext context, RouteSettings settings) {
      final arguments = <String, dynamic>{
        'module': shellTarget.module.storageId,
      };
      if (shellTarget.embedRouteInShell) {
        arguments['route'] = routeName;
        if (settings.arguments != null) {
          arguments['routeArgs'] = settings.arguments;
        }
      }
      return UiRouteHost(
        settings: RouteSettings(
          name: AppRoutes.home,
          arguments: arguments,
        ),
      );
    }

    return RouteEntry(
      defaultBuilder: defaultBuilder,
      platformOverrides: {
        UiPlatformKind.desktopWindows: desktopBuilder,
        UiPlatformKind.desktopMacos: desktopBuilder,
        UiPlatformKind.desktopLinux: desktopBuilder,
      },
    );
  }

  static int readInitialIndex(Object? arguments) {
    if (arguments is int) {
      return arguments;
    }

    if (arguments is Map<String, dynamic>) {
      return (arguments['tab'] as int?) ?? 0;
    }

    return 0;
  }

  static String? readInitialModuleId(Object? arguments) {
    if (arguments is String) {
      return arguments;
    }

    if (arguments is Map<String, dynamic>) {
      return arguments['module'] as String?;
    }

    return null;
  }
}

class SplashPage extends StatefulWidget {
  const SplashPage({super.key});

  @override
  State<SplashPage> createState() => _SplashPageState();
}

class _SplashPageState extends State<SplashPage> {
  @override
  void initState() {
    super.initState();
    _bootstrap();
  }

  Future<void> _bootstrap() async {
    await Future.delayed(const Duration(milliseconds: 300));
    if (!mounted) {
      return;
    }

    final configs = await ApiConfigManager.getConfigs();
    if (!mounted) {
      return;
    }

    if (configs.isEmpty) {
      Navigator.pushReplacementNamed(context, AppRoutes.serverConfig);
      return;
    }

    Navigator.pushReplacementNamed(context, AppRoutes.home);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    return Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.dns_outlined, size: 72),
            const SizedBox(height: 16),
            Text(l10n.appName, style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 24),
            const CircularProgressIndicator(),
            const SizedBox(height: 12),
            Text(l10n.commonLoading),
          ],
        ),
      ),
    );
  }
}

class NotFoundPage extends StatelessWidget {
  const NotFoundPage({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.notFoundTitle)),
      body: Center(child: Text(l10n.notFoundDesc)),
    );
  }
}
