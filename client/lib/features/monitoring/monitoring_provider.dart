import 'dart:async';
import 'dart:ui';
import 'package:flutter/foundation.dart';
import 'package:hitest/core/presentation/safe_change_notifier.dart';
import '../../data/repositories/monitor_repository.dart';
import '../../data/models/monitoring_runtime_models.dart';
import 'monitoring_service.dart';
import '../../core/services/logger/logger_service.dart';

const String _monitoringProviderPackage =
    'features.monitoring.monitoring_provider';

/// 监控数据状态
class MonitoringData {
  final bool isLoading;
  final bool isRefreshing;
  final String? error;
  final MonitorMetricsSnapshot? currentMetrics;
  final MonitorTimeSeries? cpuTimeSeries;
  final MonitorTimeSeries? memoryTimeSeries;
  final MonitorTimeSeries? loadTimeSeries;
  final MonitorTimeSeries? ioTimeSeries;
  final MonitorTimeSeries? networkTimeSeries;
  final MonitorSetting? settings;
  final List<String> ioOptions;
  final List<String> networkOptions;
  final String selectedIO;
  final String selectedNetwork;
  final DateTime? lastUpdated;

  const MonitoringData({
    this.isLoading = false,
    this.isRefreshing = false,
    this.error,
    this.currentMetrics,
    this.cpuTimeSeries,
    this.memoryTimeSeries,
    this.loadTimeSeries,
    this.ioTimeSeries,
    this.networkTimeSeries,
    this.settings,
    this.ioOptions = const ['all'],
    this.networkOptions = const ['all'],
    this.selectedIO = 'all',
    this.selectedNetwork = 'all',
    this.lastUpdated,
  });

  MonitoringData copyWith({
    bool? isLoading,
    bool? isRefreshing,
    String? error,
    MonitorMetricsSnapshot? currentMetrics,
    MonitorTimeSeries? cpuTimeSeries,
    MonitorTimeSeries? memoryTimeSeries,
    MonitorTimeSeries? loadTimeSeries,
    MonitorTimeSeries? ioTimeSeries,
    MonitorTimeSeries? networkTimeSeries,
    MonitorSetting? settings,
    List<String>? ioOptions,
    List<String>? networkOptions,
    String? selectedIO,
    String? selectedNetwork,
    DateTime? lastUpdated,
  }) {
    return MonitoringData(
      isLoading: isLoading ?? this.isLoading,
      isRefreshing: isRefreshing ?? this.isRefreshing,
      error: error,
      currentMetrics: currentMetrics ?? this.currentMetrics,
      cpuTimeSeries: cpuTimeSeries ?? this.cpuTimeSeries,
      memoryTimeSeries: memoryTimeSeries ?? this.memoryTimeSeries,
      loadTimeSeries: loadTimeSeries ?? this.loadTimeSeries,
      ioTimeSeries: ioTimeSeries ?? this.ioTimeSeries,
      networkTimeSeries: networkTimeSeries ?? this.networkTimeSeries,
      settings: settings ?? this.settings,
      ioOptions: ioOptions ?? this.ioOptions,
      networkOptions: networkOptions ?? this.networkOptions,
      selectedIO: selectedIO ?? this.selectedIO,
      selectedNetwork: selectedNetwork ?? this.selectedNetwork,
      lastUpdated: lastUpdated ?? this.lastUpdated,
    );
  }
}

/// 监控数据 Provider
///
/// 直连后端拉取一次全量趋势 + 当前指标，交给定时器周期刷新。
class MonitoringProvider extends ChangeNotifier with SafeChangeNotifier {
  MonitoringProvider({MonitoringService? service})
      : _service = service ?? MonitoringService();

  final MonitoringService _service;
  Timer? _refreshTimer;

  MonitoringData _data = const MonitoringData();

  final Duration _refreshInterval = const Duration(seconds: 30);
  final Duration _timeRange = const Duration(hours: 1);
  bool _autoRefreshEnabled = false;
  bool _isPolling = false;

  MonitoringData get data => _data;
  Duration get refreshInterval => _refreshInterval;
  bool get autoRefreshEnabled => _autoRefreshEnabled;
  Duration get timeRange => _timeRange;

  String _normalizeMonitorOption(String? value, List<String> options) {
    if (options.isEmpty) {
      return 'all';
    }
    if (value != null && value.isNotEmpty && options.contains(value)) {
      return value;
    }
    for (final option in options) {
      if (option != 'all') {
        return option;
      }
    }
    return options.first;
  }

  Future<void> _loadOptionsAndSettings() async {
    try {
      final settings = await _service.getSetting();
      final ioOptions = await _service.getIOOptions();
      final networkOptions = await _service.getNetworkOptions();
      final safeIoOptions = ioOptions.isEmpty ? const ['all'] : ioOptions;
      final safeNetworkOptions =
          networkOptions.isEmpty ? const ['all'] : networkOptions;
      _data = _data.copyWith(
        settings: settings,
        ioOptions: safeIoOptions,
        networkOptions: safeNetworkOptions,
        selectedIO: _normalizeMonitorOption(
          _data.selectedIO != 'all' ? _data.selectedIO : settings?.defaultIO,
          safeIoOptions,
        ),
        selectedNetwork: _normalizeMonitorOption(
          _data.selectedNetwork != 'all'
              ? _data.selectedNetwork
              : settings?.defaultNetwork,
          safeNetworkOptions,
        ),
      );
    } catch (e, stackTrace) {
      appLogger.wWithPackage(
        _monitoringProviderPackage,
        'Load monitor options failed',
        error: e,
        stackTrace: stackTrace,
      );
    }
  }

  Future<void> selectIOOption(String value) async {
    if (value == _data.selectedIO) {
      return;
    }
    _data = _data.copyWith(selectedIO: value);
    notifyListeners();
    await load(silent: true);
  }

  Future<void> selectNetworkOption(String value) async {
    if (value == _data.selectedNetwork) {
      return;
    }
    _data = _data.copyWith(selectedNetwork: value);
    notifyListeners();
    await load(silent: true);
  }

  void _recordError(
    String action,
    Object error, {
    StackTrace? stackTrace,
  }) {
    appLogger.eWithPackage(
      _monitoringProviderPackage,
      '$action failed',
      error: error,
      stackTrace: stackTrace,
    );
    _data = _data.copyWith(
      isLoading: false,
      isRefreshing: false,
      error: '$action失败: $error',
    );
    notifyListeners();
  }

  /// 生命周期变化处理：暂停时停表，恢复时重启并补一次刷新。
  void onAppLifecycleChanged(AppLifecycleState state) {
    if (state == AppLifecycleState.paused ||
        state == AppLifecycleState.inactive) {
      _stopTimer();
      return;
    }
    if (state == AppLifecycleState.resumed && _autoRefreshEnabled) {
      _startTimer();
      unawaited(load(silent: true));
    }
  }

  /// 启用/禁用自动刷新
  void toggleAutoRefresh(bool enabled) {
    _autoRefreshEnabled = enabled;
    if (enabled) {
      _startTimer();
      unawaited(load(silent: true));
    } else {
      _stopTimer();
    }
    notifyListeners();
  }

  void _startTimer() {
    _stopTimer();
    _refreshTimer = Timer.periodic(_refreshInterval, (_) {
      unawaited(load(silent: true));
    });
  }

  void _stopTimer() {
    _refreshTimer?.cancel();
    _refreshTimer = null;
  }

  @override
  void dispose() {
    _stopTimer();
    super.dispose();
  }

  /// 加载所有监控数据
  Future<void> load({bool silent = false}) async {
    if (_isPolling) {
      return;
    }
    _isPolling = true;
    try {
      if (!silent) {
        _data = _data.copyWith(isLoading: true);
        notifyListeners();
      } else {
        _data = _data.copyWith(isRefreshing: true);
      }

      if (_data.settings == null) {
        await _loadOptionsAndSettings();
      }

      final result = await _service.getMonitorData(
        io: _data.selectedIO,
        network: _data.selectedNetwork,
        duration: _timeRange,
      );

      _data = _data.copyWith(
        isLoading: false,
        isRefreshing: false,
        currentMetrics: result.current,
        cpuTimeSeries: result.timeSeries['cpu'],
        memoryTimeSeries: result.timeSeries['memory'],
        loadTimeSeries: result.timeSeries['load'],
        ioTimeSeries: result.timeSeries['io'],
        networkTimeSeries: result.timeSeries['network'],
        lastUpdated: DateTime.now(),
      );
      notifyListeners();
    } catch (e, stackTrace) {
      _recordError('加载监控数据', e, stackTrace: stackTrace);
    } finally {
      _isPolling = false;
    }
  }

  /// 刷新数据
  Future<void> refresh() async {
    await load(silent: true);
  }

  /// 清除错误
  void clearError() {
    _data = _data.copyWith(error: null);
    notifyListeners();
  }

  /// 加载监控设置
  Future<void> loadSettings() async {
    try {
      await _loadOptionsAndSettings();
      notifyListeners();
    } catch (e, stackTrace) {
      _recordError('加载监控设置', e, stackTrace: stackTrace);
    }
  }

  /// 更新监控设置
  Future<bool> updateSettings({
    int? interval,
    int? retention,
    bool? enabled,
    String? defaultIO,
    String? defaultNetwork,
  }) async {
    try {
      final success = await _service.updateSetting(
        interval: interval,
        retention: retention,
        enabled: enabled,
        defaultIO: defaultIO,
        defaultNetwork: defaultNetwork,
      );
      if (success) {
        await _loadOptionsAndSettings();
        await load(silent: true);
      }
      return success;
    } catch (e, stackTrace) {
      _recordError('更新监控设置', e, stackTrace: stackTrace);
      return false;
    }
  }

  /// 清理监控数据
  Future<bool> cleanData() async {
    try {
      final success = await _service.cleanData();
      if (success) {
        await load(silent: true);
      }
      return success;
    } catch (e, stackTrace) {
      _recordError('清理监控数据', e, stackTrace: stackTrace);
      return false;
    }
  }
}
