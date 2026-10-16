import 'dart:io';

import '../config/api_config.dart';
import '../config/api_constants.dart';
import 'machine_path.dart';
import 'hitest_auth_headers.dart';

class OnePanelWebSocketConnector {
  const OnePanelWebSocketConnector._();

  static Future<WebSocket> connect({
    required ApiConfig config,
    required String path,
    Map<String, String> queryParameters = const <String, String>{},
  }) async {
    final headers = <String, dynamic>{
      ...OnePanelAuthHeaders.build(config.apiKey),
      'User-Agent': ApiConstants.userAgent,
    };

    HttpClient? customClient;
    if (config.allowInsecureTls) {
      customClient = HttpClient()
        ..badCertificateCallback =
            (X509Certificate _, String __, int ___) => true;
    }

    return WebSocket.connect(
      _buildUri(
        baseUrl: config.url,
        path: path,
        queryParameters: queryParameters,
        machine: config.machine,
      ).toString(),
      headers: headers,
      customClient: customClient,
    );
  }

  static Uri buildUri({
    required ApiConfig config,
    required String path,
    Map<String, String> queryParameters = const <String, String>{},
  }) {
    return _buildUri(
      baseUrl: config.url,
      path: path,
      queryParameters: queryParameters,
      machine: config.machine,
    );
  }

  static Uri _buildUri({
    required String baseUrl,
    required String path,
    required Map<String, String> queryParameters,
    required String machine,
  }) {
    final baseUri = Uri.parse(baseUrl);
    final normalizedBasePath = baseUri.path.endsWith('/')
        ? baseUri.path.substring(0, baseUri.path.length - 1)
        : baseUri.path;
    // 服务端 WS 同样带机器前缀 (/api/v1/local/... 或 /api/v1/m/{id}/...)
    final normalizedPath = path.startsWith('/') ? path : '/$path';
    final machinePath = MachinePath.add(normalizedPath, machine: machine);

    return baseUri.replace(
      scheme: baseUri.scheme == 'https' ? 'wss' : 'ws',
      path: '$normalizedBasePath$machinePath',
      queryParameters:
          queryParameters.isEmpty ? null : Map<String, String>.from(queryParameters),
    );
  }
}
