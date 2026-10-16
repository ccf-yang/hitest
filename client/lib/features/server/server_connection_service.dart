import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../../core/network/hitest_auth_headers.dart';
import '../../core/network/machine_path.dart';

class ServerConnectionResult {
  const ServerConnectionResult({
    required this.success,
    this.errorMessage,
    this.osInfo,
    this.responseTime,
  });

  final bool success;
  final String? errorMessage;
  final Map<String, dynamic>? osInfo;
  final Duration? responseTime;
}

class ServerConnectionService {
  Future<ServerConnectionResult> testConnection({
    required String serverUrl,
    required String apiKey,
    bool allowInsecureTls = false,
    String machine = 'local',
  }) async {
    final stopwatch = Stopwatch()..start();

    try {
      final dio = Dio(BaseOptions(
        baseUrl: serverUrl,
        connectTimeout: const Duration(seconds: 10),
        receiveTimeout: const Duration(seconds: 10),
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      ));

      if (allowInsecureTls && !kIsWeb) {
        try {
          (dio.httpClientAdapter as dynamic).onHttpClientCreate =
              (dynamic httpClient) {
            httpClient.badCertificateCallback =
                (dynamic cert, String host, int port) => true;
            return httpClient;
          };
        } catch (_) {
          // Ignore adapter-level failures and continue with default validation.
        }
      }

      // 这里的 Dio 是独立于 DioClient 的（连接测试要在保存配置前跑）,
      // 所以机器前缀和信封头得自己补一遍, 否则会打到 hub 上不存在的无前缀路由。
      final headers = <String, String>{
        'X-Hitest-Envelope': '1',
      };
      if (apiKey.isNotEmpty) headers.addAll(OnePanelAuthHeaders.build(apiKey));

      final response = await dio.get(
        MachinePath.add('/api/v1/monitor/base/os', machine: machine),
        options: Options(
          headers: headers,
        ),
      );

      stopwatch.stop();

      if (response.statusCode == 200 && response.data != null) {
        final data = response.data as Map<String, dynamic>;
        if (data['data'] != null) {
          return ServerConnectionResult(
            success: true,
            osInfo: data['data'] as Map<String, dynamic>?,
            responseTime: stopwatch.elapsed,
          );
        }
      }

      return ServerConnectionResult(
        success: false,
        errorMessage: 'Invalid response from server',
        responseTime: stopwatch.elapsed,
      );
    } on DioException catch (e) {
      stopwatch.stop();
      String errorMessage;

      switch (e.type) {
        case DioExceptionType.connectionTimeout:
        case DioExceptionType.sendTimeout:
        case DioExceptionType.receiveTimeout:
          errorMessage = 'Connection timeout';
          break;
        case DioExceptionType.connectionError:
          errorMessage = 'Cannot connect to server';
          break;
        case DioExceptionType.badResponse:
          if (e.response?.statusCode == 401) {
            errorMessage = 'Authentication failed: Invalid API key';
          } else {
            errorMessage =
                'Server error: ${e.response?.statusCode ?? 'Unknown'}';
          }
          break;
        default:
          errorMessage = 'Connection failed: ${e.message}';
      }

      return ServerConnectionResult(
        success: false,
        errorMessage: errorMessage,
        responseTime: stopwatch.elapsed,
      );
    } catch (e) {
      stopwatch.stop();
      return ServerConnectionResult(
        success: false,
        errorMessage: 'Unexpected error: $e',
        responseTime: stopwatch.elapsed,
      );
    }
  }
}
