import 'package:dio/dio.dart';

import '../machine_path.dart';

/// 请求发出前补机器前缀: /api/v1/monitor/... → /api/v1/local/monitor/...
/// 必须排在其他拦截器之前, 让日志与重试看到的都是最终 URL。
class MachinePrefixInterceptor extends Interceptor {
  MachinePrefixInterceptor({this.machine = MachinePath.localMachine});

  final String machine;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    options.path = MachinePath.add(options.path, machine: machine);
    // 信封头是协议协商而非鉴权: AuthInterceptor 只在有 apiKey 时加头, 留空就会漏掉,
    // 而各仓库都按 data['data'] 解析, 所以这里无条件补上。
    options.headers['X-Hitest-Envelope'] = '1';
    handler.next(options);
  }
}
