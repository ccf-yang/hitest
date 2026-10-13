import 'package:flutter/material.dart';
import 'package:hitest/config/app_router.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/core/theme/app_design_tokens.dart';
import 'package:hitest/core/utils/platform_utils.dart';
import 'package:hitest/features/settings/screens/theme_settings_page.dart';
import 'package:hitest/features/shell/shell_navigation.dart';

class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    if (PlatformUtils.isDesktop(context)) {
      return const _SettingsPageDesktop();
    }
    return const _SettingsPageMobile();
  }
}

class _SettingsPageMobile extends StatelessWidget {
  const _SettingsPageMobile();

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    final canPop = Navigator.of(context).canPop();

    return Scaffold(
      appBar: AppBar(
        leading: canPop
            ? IconButton(
                icon: const Icon(Icons.arrow_back),
                onPressed: () => Navigator.of(context).maybePop(),
                tooltip: MaterialLocalizations.of(context).backButtonTooltip,
              )
            : null,
        title: Text(l10n.settingsPageTitle),
      ),
      body: const _SettingsBody(),
    );
  }
}

class _SettingsPageDesktop extends StatelessWidget {
  const _SettingsPageDesktop();

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Scaffold(
      backgroundColor: scheme.surface,
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 800),
          child: ColoredBox(
            color: scheme.surface,
            child: const _SettingsBody(),
          ),
        ),
      ),
    );
  }
}

class _SettingsBody extends StatelessWidget {
  const _SettingsBody();

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;

    return ListView(
      padding: AppDesignTokens.pagePadding,
      children: [
        Text(l10n.settingsGeneral,
            style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: AppDesignTokens.spacingSm),
        Card(
          child: Column(
            children: [
              ListTile(
                leading: const Icon(Icons.color_lens_outlined),
                title: Text(l10n.settingsTheme),
                trailing: const Icon(Icons.chevron_right),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const ThemeSettingsPage(),
                    ),
                  );
                },
              ),
            ],
          ),
        ),
        const SizedBox(height: AppDesignTokens.spacingLg),
        Text(l10n.settingsAppSectionTitle,
            style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: AppDesignTokens.spacingSm),
        Card(
          child: Column(
            children: [
              ListTile(
                leading: const Icon(Icons.dns_outlined),
                title: Text(l10n.settingsServerManagement),
                subtitle: Text(l10n.settingsServerManagementSubtitle),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => openRouteRespectingShell(context, AppRoutes.server),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
