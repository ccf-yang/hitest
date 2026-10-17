import 'package:flutter/foundation.dart';
import '../models/monitoring_runtime_models.dart';
import '../../core/services/logger/logger_service.dart';

const String _monitorRepoPackage = 'data.repositories.monitor_repository';

/// 磁盘数据
class DiskData {
  final String path;
  final String type;
  final String device;
  final int total;
  final int used;
  final int free;
  final double usedPercent;

  const DiskData({
    required this.path,
    required this.type,
    required this.device,
    required this.total,
    required this.used,
    required this.free,
    required this.usedPercent,
  });

  factory DiskData.fromJson(Map<String, dynamic> json) {
    return DiskData(
      path: json['path'] as String? ?? '/',
      type: json['type'] as String? ?? '',
      device: json['device'] as String? ?? '',
      total: json['total'] as int? ?? 0,
      used: json['used'] as int? ?? 0,
      free: json['free'] as int? ?? 0,
      usedPercent: (json['usedPercent'] as num?)?.toDouble() ?? 0,
    );
  }
}

/// 监控指标快照
///
/// 从API响应中提取的最新监控数据
class MonitorMetricsSnapshot {
  final double? cpuPercent;
  final double? memoryPercent;
  final double? diskPercent;
  final double? load1;
  final double? load5;
  final double? load15;
  final int? memoryUsed;
  final int? memoryTotal;
  final List<DiskData> diskData;
  final int? uptime;
  final DateTime? timestamp;

  const MonitorMetricsSnapshot({
    this.cpuPercent,
    this.memoryPercent,
    this.diskPercent,
    this.load1,
    this.load5,
    this.load15,
    this.memoryUsed,
    this.memoryTotal,
    this.diskData = const [],
    this.uptime,
    this.timestamp,
  });

  bool get hasData =>
      cpuPercent != null ||
      memoryPercent != null ||
      diskPercent != null ||
      load1 != null;

  factory MonitorMetricsSnapshot.empty() => const MonitorMetricsSnapshot();
}

/// 监控数据点
class MonitorDataPoint {
  final DateTime time;
  final double value;

  const MonitorDataPoint({
    required this.time,
    required this.value,
  });
}

/// 监控时间序列
class MonitorTimeSeries {
  final String name;
  final List<MonitorDataPoint> data;
  final double? min;
  final double? max;
  final double? avg;

  const MonitorTimeSeries({
    required this.name,
    required this.data,
    this.min,
    this.max,
    this.avg,
  });

  factory MonitorTimeSeries.empty(String name) =>
      MonitorTimeSeries(name: name, data: []);
}

/// 监控数据包
class MonitorDataPackage {
  final MonitorMetricsSnapshot current;
  final Map<String, MonitorTimeSeries> timeSeries;

  const MonitorDataPackage({
    required this.current,
    required this.timeSeries,
  });
}

/// 解析参数
class MonitorDataParseArgs {
  final dynamic data;
  final DateTime timestamp;

  MonitorDataParseArgs(this.data, this.timestamp);
}

/// 顶级解析函数 - 解析完整监控数据包 (用于 compute)
MonitorDataPackage parseMonitorDataPackage(MonitorDataParseArgs args) {
  final data = args.data;
  final now = args.timestamp;

  // 解析当前指标
  final current = parseMetricsResponse(data, now);

  // 解析各时间序列指标
  final timeSeries = {
    'cpu': parseTimeSeriesResponse(data, 'base', 'cpu'),
    'memory': parseTimeSeriesResponse(data, 'base', 'memory'),
    'load': parseTimeSeriesResponse(data, 'base', 'cpuLoad1'),
    'io': parseTimeSeriesResponse(data, 'io', 'ioThroughput'),
    'network': parseTimeSeriesResponse(data, 'network', 'networkThroughput'),
  };

  return MonitorDataPackage(current: current, timeSeries: timeSeries);
}

/// 顶级解析函数 - 解析监控指标响应
MonitorMetricsSnapshot parseMetricsResponse(dynamic data, DateTime timestamp) {
  if (data == null || data is! Map) {
    return const MonitorMetricsSnapshot();
  }

  final responseData = data as Map<String, dynamic>;
  final dataList = responseData['data'] as List?;

  if (dataList == null) {
    return const MonitorMetricsSnapshot();
  }

  double? cpuPercent;
  double? memoryPercent;
  double? diskPercent;
  double? load1;
  double? load5;
  double? load15;
  int? memoryUsed;
  int? memoryTotal;
  List<DiskData> diskDataList = [];
  int? uptime;

  for (final item in dataList) {
    if (item is! Map<String, dynamic>) continue;

    final param = item['param'] as String?;
    final values = item['value'] as List?;

    if (values == null || values.isEmpty) continue;

    final lastValue = values.last;
    if (lastValue is! Map<String, dynamic>) continue;

    // Keep parsing silent in isolate to avoid side effects and output noise.

    switch (param) {
      case 'base':
        // base 包含所有指标
        cpuPercent = (lastValue['cpu'] as num?)?.toDouble();
        memoryPercent = (lastValue['memory'] as num?)?.toDouble();
        load1 = (lastValue['cpuLoad1'] as num?)?.toDouble();
        load5 = (lastValue['cpuLoad5'] as num?)?.toDouble();
        load15 = (lastValue['cpuLoad15'] as num?)?.toDouble();
        memoryUsed = lastValue['memoryUsed'] as int?;
        memoryTotal = lastValue['memoryTotal'] as int?;
        uptime = lastValue['uptime'] as int?;

        // 解析磁盘数据
        final diskDataRaw = lastValue['diskData'] as List?;
        if (diskDataRaw != null) {
          diskDataList = diskDataRaw
              .map((d) => DiskData.fromJson(d as Map<String, dynamic>))
              .toList();
          if (diskDataList.isNotEmpty) {
            diskPercent = diskDataList.first.usedPercent;
          }
        }
        break;
      case 'cpu':
        cpuPercent ??= (lastValue['cpu'] as num?)?.toDouble();
        break;
      case 'memory':
        memoryPercent ??= (lastValue['memory'] as num?)?.toDouble();
        break;
      case 'load':
        load1 ??= (lastValue['cpuLoad1'] as num?)?.toDouble();
        load5 ??= (lastValue['cpuLoad5'] as num?)?.toDouble();
        load15 ??= (lastValue['cpuLoad15'] as num?)?.toDouble();
        break;
      case 'io':
        diskPercent ??= (lastValue['disk'] as num?)?.toDouble();
        break;
    }
  }

  return MonitorMetricsSnapshot(
    cpuPercent: cpuPercent,
    memoryPercent: memoryPercent,
    diskPercent: diskPercent,
    load1: load1,
    load5: load5,
    load15: load15,
    memoryUsed: memoryUsed,
    memoryTotal: memoryTotal,
    diskData: diskDataList,
    uptime: uptime,
    timestamp: timestamp,
  );
}

/// 顶级解析函数 - 解析时间序列响应
MonitorTimeSeries parseTimeSeriesResponse(
  dynamic data,
  String param,
  String valueKey,
) {
  if (data == null || data is! Map) {
    return MonitorTimeSeries.empty(param);
  }

  final responseData = data as Map<String, dynamic>;
  final dataList = responseData['data'] as List?;

  if (dataList == null || dataList.isEmpty) {
    return MonitorTimeSeries.empty(param);
  }

  // 查找匹配的数据项
  Map<String, dynamic>? item;
  try {
    item = dataList.firstWhere(
      (e) => (e as Map)['param'] == param,
    ) as Map<String, dynamic>?;
  } catch (_) {
    item = null;
  }

  // 如果没有找到精确匹配，尝试使用 'base' 数据（base包含cpu/memory/load）
  if (item == null && param != 'base') {
    try {
      item = dataList.firstWhere(
        (e) => (e as Map)['param'] == 'base',
      ) as Map<String, dynamic>?;
    } catch (_) {
      item = null;
    }
  }

  if (item == null) {
    return MonitorTimeSeries.empty(param);
  }

  final itemMap = item;
  final dates = (itemMap['date'] as List?)?.map((e) => e.toString()).toList();
  final values = itemMap['value'] as List?;

  if (values == null || values.isEmpty) {
    return MonitorTimeSeries.empty(param);
  }

  final dataPoints = <MonitorDataPoint>[];
  double? min;
  double? max;
  double sum = 0;
  int count = 0;
  final now = DateTime.now();

  for (var i = 0; i < values.length; i++) {
    final valueMap = values[i] as Map<String, dynamic>?;
    if (valueMap == null) continue;

    final value = _extractMetricValue(valueMap, valueKey);
    if (value == null) {
      continue;
    }

    // 后端发的时间戳数字部分是 UTC（盒子时钟即 UTC），却硬贴了 +08:00 偏移标签，
    // 自相矛盾。这里取数字当作 UTC 构造真实瞬间，交给设备本地化（东八区即 +8）。
    final createdAt = valueMap['createdAt'] as String?;

    // 优先用 date[] 数组（与 value 同序），其次 valueMap 里的 createdAt。
    String? rawTime;
    if (dates != null && i < dates.length && dates[i].isNotEmpty) {
      rawTime = dates[i];
    } else {
      rawTime = createdAt;
    }

    final time = _parseTimestampAsUtc(rawTime) ?? now;

    dataPoints.add(MonitorDataPoint(time: time, value: value));

    if (min == null || value < min) min = value;
    if (max == null || value > max) max = value;
    sum += value;
    count++;
  }

  return MonitorTimeSeries(
    name: param,
    data: dataPoints,
    min: min,
    max: max,
    avg: count > 0 ? sum / count : null,
  );
}

/// 取 ISO 时间戳的数字部分当作 UTC 瞬间（忽略其自带的偏移标签）。
/// 形如 2026-10-17T03:13:54+08:00 → DateTime.utc(2026-10-17 03:13:54)。
/// 盒子写入的挂钟数字实为 UTC，+08:00 标签是错的，故丢弃标签、按 UTC 还原真实时刻，
/// 之后 .toLocal() 才能在东八区设备上正确显示为本地时间。
DateTime? _parseTimestampAsUtc(String? raw) {
  if (raw == null || raw.isEmpty) return null;
  final m = RegExp(
    r'^(\d{4})-(\d{2})-(\d{2})[T ](\d{2}):(\d{2}):(\d{2})',
  ).firstMatch(raw);
  if (m == null) return null;
  return DateTime.utc(
    int.parse(m.group(1)!),
    int.parse(m.group(2)!),
    int.parse(m.group(3)!),
    int.parse(m.group(4)!),
    int.parse(m.group(5)!),
    int.parse(m.group(6)!),
  );
}

double? _extractMetricValue(Map<String, dynamic> valueMap, String valueKey) {
  final directValue = (valueMap[valueKey] as num?)?.toDouble();
  if (directValue != null) {
    return directValue;
  }

  switch (valueKey) {
    case 'ioThroughput':
      final read = (valueMap['read'] as num?)?.toDouble();
      final write = (valueMap['write'] as num?)?.toDouble();
      if (read != null || write != null) {
        return ((read ?? 0) + (write ?? 0)) / 1024;
      }

      final fallback = (valueMap['diskRead'] as num?)?.toDouble() ??
          (valueMap['diskWrite'] as num?)?.toDouble() ??
          (valueMap['ioRead'] as num?)?.toDouble() ??
          (valueMap['ioWrite'] as num?)?.toDouble();
      return fallback != null ? fallback / 1024 : null;
    case 'networkThroughput':
      final up = (valueMap['up'] as num?)?.toDouble();
      final down = (valueMap['down'] as num?)?.toDouble();
      if (up != null || down != null) {
        return (up ?? 0) + (down ?? 0);
      }

      return (valueMap['networkIn'] as num?)?.toDouble() ??
          (valueMap['networkOut'] as num?)?.toDouble();
    case 'disk':
      return (valueMap['diskRead'] as num?)?.toDouble() ??
          (valueMap['diskWrite'] as num?)?.toDouble() ??
          (valueMap['ioRead'] as num?)?.toDouble() ??
          (valueMap['ioWrite'] as num?)?.toDouble();
    case 'networkIn':
      return (valueMap['up'] as num?)?.toDouble() ??
          (valueMap['down'] as num?)?.toDouble() ??
          (valueMap['networkOut'] as num?)?.toDouble();
    default:
      return null;
  }
}

/// 统一监控数据仓库
///
/// 提供监控数据获取的统一入口，避免代码重复
class MonitorRepository {
  const MonitorRepository();

  /// 获取当前监控指标
  ///
  /// 返回最新的CPU、内存、磁盘、负载数据
  Future<MonitorMetricsSnapshot> getCurrentMetrics(dynamic client) async {
    try {
      final now = DateTime.now();
      final startTime = now.subtract(const Duration(hours: 1));

      appLogger.dWithPackage(
          _monitorRepoPackage, 'Calling monitor metrics API');
      final response = await client.post(
        '/api/v1/monitor/search',
        data: {
          'param': 'all',
          'startTime': startTime.toUtc().toIso8601String(),
          'endTime': now.toUtc().toIso8601String(),
        },
      );

      return parseMetricsResponse(response.data, now);
    } catch (e, stack) {
      appLogger.eWithPackage(
        _monitorRepoPackage,
        'getCurrentMetrics failed',
        error: e,
        stackTrace: stack,
      );
      return const MonitorMetricsSnapshot();
    }
  }

  /// 获取时间序列数据
  ///
  /// [client] API客户端
  /// [param] 参数类型: cpu, memory, load, io, network (注意: base单独使用会返回400)
  /// [valueKey] 值字段名
  /// [duration] 时间范围
  Future<MonitorTimeSeries> getTimeSeries(
    dynamic client,
    String param,
    String valueKey, {
    String? io,
    String? network,
    Duration duration = const Duration(hours: 1),
  }) async {
    try {
      final now = DateTime.now();
      final startTime = now.subtract(duration);

      // 使用 'all' 参数获取所有数据，然后从中提取需要的字段
      final response = await client.post(
        '/api/v1/monitor/search',
        data: {
          'param': 'all',
          if (io != null) 'io': io,
          if (network != null) 'network': network,
          'startTime': startTime.toUtc().toIso8601String(),
          'endTime': now.toUtc().toIso8601String(),
        },
      );

      return parseTimeSeriesResponse(response.data, param, valueKey);
    } catch (e, stack) {
      appLogger.eWithPackage(
        _monitorRepoPackage,
        'getTimeSeries failed: param=$param valueKey=$valueKey',
        error: e,
        stackTrace: stack,
      );
      return MonitorTimeSeries.empty(param);
    }
  }

  /// 批量获取所有监控数据（当前指标 + 时间序列）
  Future<MonitorDataPackage> getMonitorData(
    dynamic client, {
    String? io,
    String? network,
    Duration duration = const Duration(hours: 1),
    DateTime? startTime,
  }) async {
    try {
      final now = DateTime.now();
      final start = startTime ?? now.subtract(duration);

      // 使用 'all' 参数获取所有数据
      final response = await client.post(
        '/api/v1/monitor/search',
        data: {
          'param': 'all',
          if (io != null) 'io': io,
          if (network != null) 'network': network,
          'startTime': start.toUtc().toIso8601String(),
          'endTime': now.toUtc().toIso8601String(),
        },
      );

      // 使用 compute 在后台 Isolate 中解析数据，避免阻塞 UI 线程
      return await compute(
        parseMonitorDataPackage,
        MonitorDataParseArgs(response.data, now),
      );
    } catch (e, stack) {
      appLogger.eWithPackage(
        _monitorRepoPackage,
        'getMonitorData failed',
        error: e,
        stackTrace: stack,
      );
      return MonitorDataPackage(
        current: const MonitorMetricsSnapshot(),
        timeSeries: {},
      );
    }
  }

  /// 获取GPU信息

  /// 获取监控设置
  Future<MonitorSetting?> getSetting(dynamic client) async {
    try {
      final response = await client.get('/api/v1/monitor/setting');
      if (response.data != null && response.data is Map) {
        final body = response.data as Map<String, dynamic>;
        final payload = body['data'];
        if (payload is Map<String, dynamic>) {
          return MonitorSetting.fromJson(payload);
        }
        return MonitorSetting.fromJson(body);
      }
      return null;
    } catch (e, stack) {
      appLogger.eWithPackage(
        _monitorRepoPackage,
        'getSetting failed',
        error: e,
        stackTrace: stack,
      );
      return null;
    }
  }

  /// 更新监控设置
  Future<bool> updateSetting(
    dynamic client, {
    int? interval,
    int? retention,
    bool? enabled,
    String? defaultIO,
    String? defaultNetwork,
  }) async {
    try {
      final updates = <Map<String, dynamic>>[];
      if (interval != null) {
        updates.add({'key': 'MonitorInterval', 'value': interval.toString()});
      }
      if (retention != null) {
        updates.add({'key': 'MonitorStoreDays', 'value': retention.toString()});
      }
      if (enabled != null) {
        updates.add({
          'key': 'MonitorStatus',
          'value': enabled ? 'Enable' : 'Disable',
        });
      }
      if (defaultIO != null) {
        updates.add({'key': 'DefaultIO', 'value': defaultIO});
      }
      if (defaultNetwork != null) {
        updates.add({'key': 'DefaultNetwork', 'value': defaultNetwork});
      }

      for (final update in updates) {
        await client.post(
          '/api/v1/monitor/setting/update',
          data: update,
        );
      }
      return true;
    } catch (e, stack) {
      appLogger.eWithPackage(
        _monitorRepoPackage,
        'updateSetting failed',
        error: e,
        stackTrace: stack,
      );
      return false;
    }
  }

  /// 清理监控数据
  Future<bool> cleanData(dynamic client) async {
    try {
      await client.post('/api/v1/monitor/clean');
      return true;
    } catch (e, stack) {
      appLogger.eWithPackage(
        _monitorRepoPackage,
        'cleanData failed',
        error: e,
        stackTrace: stack,
      );
      return false;
    }
  }

  /// 获取网络接口列表
  Future<List<String>> getNetworkOptions(dynamic client) async {
    try {
      final response = await client.get('/api/v1/monitor/netoptions');
      if (response.data != null && response.data is Map) {
        final body = response.data as Map<String, dynamic>;
        final dataList = body['data'] as List?;
        if (dataList != null) {
          return dataList.map((e) => e.toString()).toList();
        }
      }
      return ['all'];
    } catch (e, stack) {
      appLogger.eWithPackage(
        _monitorRepoPackage,
        'getNetworkOptions failed',
        error: e,
        stackTrace: stack,
      );
      return ['all'];
    }
  }

  /// 获取IO设备列表
  Future<List<String>> getIOOptions(dynamic client) async {
    try {
      final response = await client.get('/api/v1/monitor/iooptions');
      if (response.data != null && response.data is Map) {
        final body = response.data as Map<String, dynamic>;
        final dataList = body['data'] as List?;
        if (dataList != null) {
          return dataList.map((e) => e.toString()).toList();
        }
      }
      return ['all'];
    } catch (e, stack) {
      appLogger.eWithPackage(
        _monitorRepoPackage,
        'getIOOptions failed',
        error: e,
        stackTrace: stack,
      );
      return ['all'];
    }
  }
}
