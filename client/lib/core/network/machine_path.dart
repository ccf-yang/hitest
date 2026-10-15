/// 服务端把「机器」做进了 URL 路径：
///   /api/v1/local/*   中心服务内置的本机 agent
///   /api/v1/m/{id}/*  隧道反代到远程 agent
/// App 沿用了 1Panel 时代的无前缀路径（/api/v1/monitor/...），逐仓库改路径改动面太大，
/// 这里统一补前缀：HTTP 走 MachinePrefixInterceptor，WebSocket 在拼 URI 时调用。
class MachinePath {
  static const String legacyPrefix = '/api/v1';
  static const String localMachine = 'local';

  /// 无前缀路径 → 带机器前缀；已带前缀、服务端顶层路由、非 API 路径原样返回。
  static String add(String path, {String machine = localMachine}) {
    if (!path.startsWith('$legacyPrefix/')) return path;
    final rest = path.substring(legacyPrefix.length);
    if (rest.startsWith('/local/') || rest.startsWith('/m/') || rest == '/agents') {
      return path;
    }
    if (machine == localMachine) return '$legacyPrefix/local$rest';
    return '$legacyPrefix/m/$machine$rest';
  }
}
