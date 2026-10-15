import 'package:dio/dio.dart';

import '../../core/config/api_constants.dart';
import '../../core/network/dio_client.dart';
import 'api_response_parser.dart';

/// Dashboard API 客户端（仅保留监控所需的端点）
class DashboardV2Api {
  final DioClient _client;

  DashboardV2Api(this._client);

  /// 获取仪表盘基础信息（含 currentInfo 指标）
  ///
  /// GET /monitor/base/:ioOption/:netOption
  Future<Response<Map<String, dynamic>>> getDashboardBase({
    String? ioOption,
    String? netOption,
  }) async {
    final response = await _client.get<Map<String, dynamic>>(
      ApiConstants.buildApiPath(
          '/monitor/base/${ioOption ?? 'all'}/${netOption ?? 'all'}'),
    );
    return Response(
      data: ApiResponseParser.extractMapData(response),
      statusCode: response.statusCode,
      statusMessage: response.statusMessage,
      requestOptions: response.requestOptions,
    );
  }
}
