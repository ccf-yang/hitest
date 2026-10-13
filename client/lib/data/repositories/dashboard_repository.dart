import 'package:hitest/api/v2/dashboard_v2.dart';
import 'package:hitest/core/network/api_client_manager.dart';
import 'package:hitest/core/network/network_exceptions.dart';
import 'package:hitest/core/services/logger/logger_service.dart';

import '../../features/server/server_models.dart';

/// 服务器指标仓库：仅负责从 /dashboard/base 读取当前指标快照。
class DashboardRepository {
  DashboardRepository();

  Future<DashboardV2Api> _ensureApi() async {
    final client = await ApiClientManager().getCurrentClient();
    return DashboardV2Api(client);
  }

  Future<Map<String, dynamic>> getDashboardBase() async {
    final api = await _ensureApi();
    return (await api.getDashboardBase()).data ?? const <String, dynamic>{};
  }

  /// 返回当前服务器指标快照（CPU/内存/磁盘/负载）。
  ///
  /// 如果 API 调用失败或数据格式异常，返回空的 [ServerMetricsSnapshot]。
  Future<ServerMetricsSnapshot> getCurrentServerMetrics() async {
    try {
      appLogger.dWithPackage(
        'data.repositories.dashboard_repository',
        'getCurrentServerMetrics',
      );

      final baseData = await getDashboardBase();

      final currentInfo = baseData['currentInfo'] as Map<String, dynamic>?;

      if (currentInfo == null) {
        appLogger.wWithPackage(
          'data.repositories.dashboard_repository',
          'currentInfo is null in Dashboard API response',
        );
        return const ServerMetricsSnapshot();
      }

      final cpuPercent = (currentInfo['cpuUsedPercent'] as num?)?.toDouble();
      final memoryPercent =
          (currentInfo['memoryUsedPercent'] as num?)?.toDouble();

      double? diskPercent;
      final diskDataList = currentInfo['diskData'] as List?;
      if (diskDataList != null && diskDataList.isNotEmpty) {
        final mainDisk = diskDataList.first as Map<String, dynamic>;
        diskPercent = (mainDisk['usedPercent'] as num?)?.toDouble();
      }

      final load = (currentInfo['load1'] as num?)?.toDouble();

      appLogger.dWithPackage(
        'data.repositories.dashboard_repository',
        'Extracted metrics: cpu=$cpuPercent%, memory=$memoryPercent%, disk=$diskPercent%, load=$load',
      );

      return ServerMetricsSnapshot(
        cpuPercent: cpuPercent,
        memoryPercent: memoryPercent,
        diskPercent: diskPercent,
        load: load,
      );
    } catch (e, stack) {
      if (e is NetworkException) {
        appLogger.wWithPackage(
          'data.repositories.dashboard_repository',
          'Network error getting current server metrics: ${e.message}',
        );
      } else {
        appLogger.eWithPackage(
          'data.repositories.dashboard_repository',
          'Error getting current server metrics',
          error: e,
          stackTrace: stack,
        );
      }
      return const ServerMetricsSnapshot();
    }
  }
}
