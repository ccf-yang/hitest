import 'package:flutter/material.dart';
import 'package:hitest/l10n/generated/app_localizations.dart';

const List<ClientModule> kPinnableClientModules = [
  ClientModule.monitoring,
];

enum ClientModule {
  servers,
  monitoring,
  files,
  terminal,
  settings,
}

ClientModule? clientModuleFromId(String? id) {
  if (id == null) return null;
  for (final module in ClientModule.values) {
    if (module.storageId == id) {
      return module;
    }
  }
  return null;
}

extension ClientModuleX on ClientModule {
  String get storageId {
    switch (this) {
      case ClientModule.servers:
        return 'servers';
      case ClientModule.monitoring:
        return 'monitoring';
      case ClientModule.files:
        return 'files';
      case ClientModule.terminal:
        return 'terminal';
      case ClientModule.settings:
        return 'settings';
    }
  }

  bool get requiresServer {
    switch (this) {
      case ClientModule.servers:
      case ClientModule.settings:
        return false;
      case ClientModule.monitoring:
      case ClientModule.files:
      case ClientModule.terminal:
        return true;
    }
  }

  bool get pinnable => kPinnableClientModules.contains(this);

  bool get experimental {
    switch (this) {
      case ClientModule.servers:
      case ClientModule.monitoring:
      case ClientModule.files:
      case ClientModule.terminal:
      case ClientModule.settings:
        return false;
    }
  }

  IconData get icon {
    switch (this) {
      case ClientModule.servers:
        return Icons.dns_outlined;
      case ClientModule.monitoring:
        return Icons.monitor_heart_outlined;
      case ClientModule.files:
        return Icons.folder_outlined;
      case ClientModule.terminal:
        return Icons.terminal_outlined;
      case ClientModule.settings:
        return Icons.settings_outlined;
    }
  }

  IconData get selectedIcon {
    switch (this) {
      case ClientModule.servers:
        return Icons.dns;
      case ClientModule.monitoring:
        return Icons.monitor_heart;
      case ClientModule.files:
        return Icons.folder;
      case ClientModule.terminal:
        return Icons.terminal;
      case ClientModule.settings:
        return Icons.settings;
    }
  }

  String label(AppLocalizations l10n) {
    switch (this) {
      case ClientModule.servers:
        return l10n.navServer;
      case ClientModule.monitoring:
        return l10n.serverModuleMonitoring;
      case ClientModule.files:
        return l10n.navFiles;
      case ClientModule.terminal:
        return l10n.navTerminal;
      case ClientModule.settings:
        return l10n.settingsPageTitle;
    }
  }
}
