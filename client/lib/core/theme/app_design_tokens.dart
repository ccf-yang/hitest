import 'package:flutter/material.dart';

class AppDesignTokens {
  const AppDesignTokens._();

  static const double radiusSm = 10;
  static const double radiusMd = 14;
  static const double radiusLg = 20;

  static const double spacingXs = 4;
  static const double spacingSm = 8;
  static const double spacingMd = 12;
  static const double spacingLg = 16;
  static const double spacingXl = 24;

  static const Duration motionFast = Duration(milliseconds: 150);
  static const Duration motionNormal = Duration(milliseconds: 240);

  static const Color brand = Color(0xFF0C9B4B);
  static const Color success = Color(0xFF16A34A);
  static const Color warning = Color(0xFFF59E0B);
  static const Color danger = Color(0xFFDC2626);

  /// 排版刻度（字号）：与 Material 3 的 role 对应，仅覆盖本应用用到的档位。
  static const double fontDisplay = 28;
  static const double fontTitle = 22;
  static const double fontHeading = 16;
  static const double fontBody = 14;
  static const double fontCaption = 12;
  static const double fontMicro = 10;

  static const EdgeInsets pagePadding = EdgeInsets.symmetric(
    horizontal: spacingLg,
    vertical: spacingLg,
  );
}
