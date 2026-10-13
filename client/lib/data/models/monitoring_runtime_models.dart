class MonitorSetting {
  final int? interval;
  final int? retention;
  final bool? enabled;
  final String? defaultNetwork;
  final String? defaultIO;

  const MonitorSetting({
    this.interval,
    this.retention,
    this.enabled,
    this.defaultNetwork,
    this.defaultIO,
  });

  factory MonitorSetting.fromJson(Map<String, dynamic> json) {
    final intervalRaw = json['monitorInterval'] ?? json['MonitorInterval'];
    final retentionRaw = json['monitorStoreDays'] ?? json['MonitorStoreDays'];
    final statusRaw = json['monitorStatus'] ?? json['MonitorStatus'];
    final defaultNetworkRaw = json['defaultNetwork'] ?? json['DefaultNetwork'];
    final defaultIORaw = json['defaultIO'] ?? json['DefaultIO'];

    bool? enabled;
    if (statusRaw is bool) {
      enabled = statusRaw;
    } else if (statusRaw != null) {
      final normalized = statusRaw.toString().toLowerCase();
      if (normalized == 'enable' ||
          normalized == 'enabled' ||
          normalized == 'true' ||
          normalized == '1') {
        enabled = true;
      } else if (normalized == 'disable' ||
          normalized == 'disabled' ||
          normalized == 'false' ||
          normalized == '0') {
        enabled = false;
      }
    }

    return MonitorSetting(
      interval: int.tryParse(intervalRaw?.toString() ?? ''),
      retention: int.tryParse(retentionRaw?.toString() ?? ''),
      enabled: enabled,
      defaultNetwork: defaultNetworkRaw?.toString(),
      defaultIO: defaultIORaw?.toString(),
    );
  }
}