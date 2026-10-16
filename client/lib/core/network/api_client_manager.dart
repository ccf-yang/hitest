import '../config/api_config.dart';
import '../services/logger/logger_service.dart';
import 'dio_client.dart';

class ApiClientManager {
  ApiClientManager._internal();

  static final ApiClientManager _instance = ApiClientManager._internal();

  factory ApiClientManager() => _instance;

  static ApiClientManager get instance => _instance;

  final Map<String, DioClient> _clients = <String, DioClient>{};
  final Map<String, _ClientConfigMeta> _clientMeta =
      <String, _ClientConfigMeta>{};

  Future<ApiConfig> _getCurrentConfig() async {
    final config = await ApiConfigManager.getCurrentConfig();
    if (config == null) {
      throw StateError('No API config available');
    }
    return config;
  }

  DioClient getClient(
    String serverId,
    String serverUrl,
    String apiKey, {
    bool allowInsecureTls = false,
    String machine = 'local',
  }) {
    final nextMeta = _ClientConfigMeta(
      url: serverUrl,
      apiKey: apiKey,
      allowInsecureTls: allowInsecureTls,
      machine: machine,
    );
    final currentMeta = _clientMeta[serverId];

    if (currentMeta == nextMeta && _clients.containsKey(serverId)) {
      return _clients[serverId]!;
    }

    final client = DioClient(
      baseUrl: serverUrl,
      apiKey: apiKey,
      allowInsecureTls: allowInsecureTls,
      machine: machine,
    );
    _clients[serverId] = client;
    _clientMeta[serverId] = nextMeta;
    return client;
  }

  Future<DioClient> getCurrentClient() async {
    final config = await _getCurrentConfig();
    appLogger.dWithPackage(
      'core.network.api_client_manager',
      'getCurrentClient: serverId=${config.id}, url=${config.url}, machine=${config.machine}, apiKeyPresent=${config.apiKey.isNotEmpty}, allowInsecureTls=${config.allowInsecureTls}',
    );
    return getClient(
      config.id,
      config.url,
      config.apiKey,
      allowInsecureTls: config.allowInsecureTls,
      machine: config.machine,
    );
  }

  void removeClient(String serverId) {
    _clients.remove(serverId);
    _clientMeta.remove(serverId);
  }

  void clearAllClients() {
    _clients.clear();
    _clientMeta.clear();
  }
}

class _ClientConfigMeta {
  const _ClientConfigMeta({
    required this.url,
    required this.apiKey,
    required this.allowInsecureTls,
    required this.machine,
  });

  final String url;
  final String apiKey;
  final bool allowInsecureTls;
  final String machine;

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    return other is _ClientConfigMeta &&
        other.url == url &&
        other.apiKey == apiKey &&
        other.allowInsecureTls == allowInsecureTls &&
        other.machine == machine;
  }

  @override
  int get hashCode =>
      Object.hash(url, apiKey, allowInsecureTls, machine);
}
