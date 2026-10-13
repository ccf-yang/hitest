import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:provider/provider.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:hitest/config/app_router.dart';
import 'package:hitest/core/services/app_settings_controller.dart';
import 'package:hitest/core/services/logger/logger_service.dart';
import 'package:hitest/core/theme/theme_controller.dart';
import 'package:hitest/core/theme/app_theme.dart';
import 'package:hitest/l10n/generated/app_localizations.dart';

import 'package:hitest/core/config/api_config_migration.dart';

// Feature Providers
import 'features/server/server_provider.dart';
import 'features/shell/controllers/current_server_controller.dart';
import 'features/monitoring/monitoring_provider.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();


  appLogger.init();
  try {
    await appLogger.loadPreferences();
    await appLogger.applyReleaseChannelPolicy();
  } catch (error, stackTrace) {
    appLogger.wWithPackage(
      'main',
      'Failed to load logger preferences, using default log level',
      error: error,
      stackTrace: stackTrace,
    );
  }

  // Migrate API keys from SharedPreferences fallback to secure storage
  try {
    await ApiConfigMigration.runMigrationIfNeeded();
  } catch (error, stackTrace) {
    appLogger.wWithPackage(
      'main',
      'Failed to run API config migration',
      error: error,
      stackTrace: stackTrace,
    );
  }

  FlutterError.onError = (details) {
    appLogger.eWithPackage(
      'main',
      'Flutter Error',
      error: details.exception,
      stackTrace: details.stack,
    );
  };

  WidgetsBinding.instance.platformDispatcher.onError = (error, stack) {
    appLogger.eWithPackage(
      'main',
      'Unhandled Error',
      error: error,
      stackTrace: stack,
    );
    return true;
  };

  // Initialize Hive
  await Hive.initFlutter();

  runApp(
    MultiProvider(
      providers: [
        // App Settings
        ChangeNotifierProvider(
          create: (_) => AppSettingsController(),
        ),
        ChangeNotifierProvider(
          create: (_) => ThemeController(),
        ),
        // Server Management
        ChangeNotifierProvider(
          create: (_) => ServerProvider(),
        ),
        ChangeNotifierProvider(
          create: (_) => CurrentServerController(),
        ),
        ChangeNotifierProvider(
          create: (_) => MonitoringProvider(),
        ),
      ],
      child: const AppBootstrap(child: MyApp()),
    ),
  );
}

class AppBootstrap extends StatefulWidget {
  const AppBootstrap({
    super.key,
    required this.child,
  });

  final Widget child;

  @override
  State<AppBootstrap> createState() => _AppBootstrapState();
}

class _AppBootstrapState extends State<AppBootstrap>
    with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      if (!mounted) return;
      await context.read<AppSettingsController>().load();
      if (!mounted) return;
      await context.read<ThemeController>().load();
      if (!mounted) return;
      await context.read<CurrentServerController>().load();
    });
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (!mounted) {
      return;
    }
    context.read<MonitoringProvider>().onAppLifecycleChanged(state);
    if (state == AppLifecycleState.paused) {
      context.read<ServerProvider>().pauseMetricsAutoRefresh();
    } else if (state == AppLifecycleState.resumed) {
      context.read<ServerProvider>().resumeMetricsAutoRefresh();
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer2<AppSettingsController, ThemeController>(
      builder: (context, settings, themeController, _) {
        return MaterialApp(
          title: 'hitest Client',
          debugShowCheckedModeBanner: false,
          theme: AppTheme.getLightTheme(),
          darkTheme: AppTheme.getDarkTheme(),
          themeMode: themeController.themeMode,
          locale: settings.locale,
          localizationsDelegates: const [
            AppLocalizations.delegate,
            GlobalMaterialLocalizations.delegate,
            GlobalCupertinoLocalizations.delegate,
            GlobalWidgetsLocalizations.delegate,
          ],
          supportedLocales: const [
            Locale('zh'),
          ],
          onGenerateRoute: AppRouter.generateRoute,
          initialRoute: AppRoutes.splash,
        );
      },
    );
  }
}
