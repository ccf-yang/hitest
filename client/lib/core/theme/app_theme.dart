import 'package:flutter/material.dart';
import 'app_design_tokens.dart';

class AppTheme {
  const AppTheme._();

  /// 由给定 ColorScheme 构建主题（供内部/复用）。
  static ThemeData create(ColorScheme scheme) {
    return _buildTheme(scheme);
  }

  static ThemeData getLightTheme() {
    return _buildTheme(_resolveScheme(Brightness.light));
  }

  static ThemeData getDarkTheme() {
    return _buildTheme(_resolveScheme(Brightness.dark));
  }

  static ColorScheme _resolveScheme(Brightness brightness) {
    return ColorScheme.fromSeed(
      seedColor: AppDesignTokens.brand,
      brightness: brightness,
    );
  }

  /// 以设计令牌覆盖 Material 3 默认排版刻度。
  static TextTheme _buildTextTheme(ColorScheme scheme) {
    final base = ThemeData(colorScheme: scheme, useMaterial3: true).textTheme;
    return base.copyWith(
      headlineSmall: base.headlineSmall?.copyWith(
        fontSize: AppDesignTokens.fontDisplay,
        fontWeight: FontWeight.w700,
      ),
      titleLarge: base.titleLarge?.copyWith(
        fontSize: AppDesignTokens.fontTitle,
      ),
      titleMedium: base.titleMedium?.copyWith(
        fontSize: AppDesignTokens.fontHeading,
        fontWeight: FontWeight.w600,
      ),
      bodyMedium: base.bodyMedium?.copyWith(
        fontSize: AppDesignTokens.fontBody,
      ),
      bodySmall: base.bodySmall?.copyWith(
        fontSize: AppDesignTokens.fontCaption,
      ),
    );
  }

  static ThemeData _buildTheme(ColorScheme scheme) {
    return ThemeData(
      useMaterial3: true,
      colorScheme: scheme,
      scaffoldBackgroundColor: scheme.surface,
      textTheme: _buildTextTheme(scheme),
      appBarTheme: AppBarTheme(
        centerTitle: false,
        elevation: 0,
        backgroundColor: scheme.surface,
        foregroundColor: scheme.onSurface,
      ),
      cardTheme: CardThemeData(
        elevation: 1,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(AppDesignTokens.radiusLg),
          side: BorderSide(color: scheme.outlineVariant),
        ),
        color: scheme.surface,
      ),
      navigationBarTheme: NavigationBarThemeData(
        elevation: 1,
        indicatorColor: scheme.primaryContainer,
        backgroundColor: scheme.surface,
      ),
      navigationRailTheme: NavigationRailThemeData(
        backgroundColor: scheme.surface,
        indicatorColor: scheme.primaryContainer,
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: scheme.surfaceContainerHighest.withValues(alpha: 0.5),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(AppDesignTokens.radiusMd),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(AppDesignTokens.radiusMd),
          borderSide: BorderSide(color: scheme.outlineVariant),
        ),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(AppDesignTokens.radiusMd),
          ),
          minimumSize: const Size(0, 44),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(AppDesignTokens.radiusMd),
          ),
          minimumSize: const Size(0, 44),
        ),
      ),
      chipTheme: ChipThemeData(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(AppDesignTokens.radiusSm),
        ),
      ),
    );
  }
}
