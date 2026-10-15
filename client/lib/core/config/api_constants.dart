/// hitest API常量配置
class ApiConstants {
  // API版本配置 - hitest 使用 /api/v1 路径前缀
  static const String apiVersion = 'v2';
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
}

/// hitest API路径常量
class ApiPaths {
  // 基础路径
  static String ai(String endpoint) =>
      ApiConstants.buildApiPath('/ai$endpoint');
  static String dashboard(String endpoint) =>
      ApiConstants.buildApiPath('/dashboard$endpoint');
  static String apps(String endpoint) =>
      ApiConstants.buildApiPath('/apps$endpoint');
  static String files(String endpoint) =>
      ApiConstants.buildApiPath('/files$endpoint');
  static String settings(String endpoint) =>
      ApiConstants.buildApiPath('/settings$endpoint');
  static String containers(String endpoint) =>
      ApiConstants.buildApiPath('/containers$endpoint');
  static String backups(String endpoint) =>
      ApiConstants.buildApiPath('/backups$endpoint');
  static String cronjob(String endpoint) =>
      ApiConstants.buildApiPath('/cronjob$endpoint');
  static String database(String endpoint) =>
      ApiConstants.buildApiPath('/database$endpoint');
  static String host(String endpoint) =>
      ApiConstants.buildApiPath('/host$endpoint');
  static String nginx(String endpoint) =>
      ApiConstants.buildApiPath('/nginx$endpoint');
  static String process(String endpoint) =>
      ApiConstants.buildApiPath('/process$endpoint');
  static String website(String endpoint) =>
      ApiConstants.buildApiPath('/website$endpoint');
  static String group(String endpoint) =>
      ApiConstants.buildApiPath('/group$endpoint');

  // 特殊路径（无需认证）
  static const String authLogin = '/api/v1/monitor/base/os'; // 新服务端无登录接口, 复用探活路径
  static const String healthCheck = '/api/v1/monitor/base/os';

  // AI 相关具体路径
  static const String aiOllamaModel = '/ai/ollama/model';
  static const String aiOllamaModelSearch = '/ai/ollama/model/search';
  static const String aiOllamaModelSync = '/ai/ollama/model/sync';
  static const String aiOllamaModelDelete = '/ai/ollama/model/del';
  static const String aiOllamaModelClose = '/ai/ollama/close';
  static const String aiGpuLoad = '/ai/gpu/load';
  static const String aiDomainBind = '/ai/domain/bind';
  static const String aiDomainGet = '/ai/domain/get';

  // Dashboard 相关具体路径
  static const String dashboardBaseOs = '/dashboard/base/os';
  static const String dashboardQuickOption = '/dashboard/quick/option';
  static const String dashboardAppLauncher = '/dashboard/app/launcher';
}
