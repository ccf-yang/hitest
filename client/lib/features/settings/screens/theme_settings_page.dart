import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/core/theme/theme_controller.dart';

class ThemeSettingsPage extends StatelessWidget {
  const ThemeSettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    final themeController = context.watch<ThemeController>();

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.settingsTheme),
      ),
      body: ListView(
        children: [
          RadioGroup<ThemeMode>(
            groupValue: themeController.themeMode,
            onChanged: (value) {
              if (value != null) {
                themeController.updateThemeMode(value);
              }
            },
            child: Column(
              children: [
                RadioListTile<ThemeMode>(
                  title: Text(l10n.themeSystem),
                  value: ThemeMode.system,
                ),
                RadioListTile<ThemeMode>(
                  title: Text(l10n.themeLight),
                  value: ThemeMode.light,
                ),
                RadioListTile<ThemeMode>(
                  title: Text(l10n.themeDark),
                  value: ThemeMode.dark,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
