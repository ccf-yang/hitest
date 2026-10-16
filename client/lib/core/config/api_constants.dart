/// hitest API 常量配置
class ApiConstants {
  // API 版本前缀 - hitest 服务端挂在 /api/v1
  static const String apiPrefix = '/api/v1';

  // 默认服务器配置
  static const String defaultBaseUrl = 'http://localhost:10086';

  // 调试模式
  static const bool isDebugMode = true;

  // 网络配置
  static const int connectTimeout = 8; // 秒
  static const int receiveTimeout = 8; // 秒
  static const int sendTimeout = 8; // 秒

  // 重试配置
  static const int maxRetryAttempts = 3;
  static const Duration retryDelay = Duration(seconds: 1);

  // 应用信息
  static const String appName = 'hitest Flutter App';
  static const String appVersion = '1.0.0';
  static const String userAgent = 'hitest-Flutter-App/1.0.0';

  // API路径构建
  static String buildApiPath(String endpoint) {
    return '$apiPrefix$endpoint';
  }

  // 认证相关常量
  static const String authHeaderToken = 'hitest-Token';
  static const String authHeaderTimestamp = 'hitest-Timestamp';
  static const String authPrefix = 'hitest'; // MD5计算前缀

  // 服务端没有登录接口, 探活复用 monitor 基础信息端点
  static const String authLogin = '/api/v1/monitor/base/os';
}
