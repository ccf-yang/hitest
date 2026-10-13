import 'package:flutter/material.dart';
import 'package:hitest/core/services/app_preferences_service.dart';

class ThemeController extends ChangeNotifier {
  ThemeController({AppPreferencesService? preferencesService})
      : _preferencesService = preferencesService ?? AppPreferencesService();

  final AppPreferencesService _preferencesService;

  ThemeMode _themeMode = ThemeMode.system;
  bool _loaded = false;

  ThemeMode get themeMode => _themeMode;
  bool get loaded => _loaded;

  Future<void> load() async {
    _themeMode = await _preferencesService.loadThemeMode();
    _loaded = true;
    notifyListeners();
  }

  Future<void> updateThemeMode(ThemeMode mode) async {
    if (_themeMode == mode) return;
    _themeMode = mode;
    await _preferencesService.saveThemeMode(mode);
    notifyListeners();
  }
}
