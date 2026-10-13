import 'package:dio/dio.dart';

import '../../core/config/api_constants.dart';
import '../../core/network/dio_client.dart';
import 'api_response_parser.dart';

/// 预置快捷命令 API（GET /commands/preset、POST /commands/preset/run）
class QuickCommandV2Api {
  final DioClient _client;

  QuickCommandV2Api(this._client);

  Future<Response<List<Map<String, dynamic>>>> getPresetCommands() async {
    final response = await _client.get<Map<String, dynamic>>(
      ApiConstants.buildApiPath('/commands/preset'),
    );
    final list = ApiResponseParser.extractListData(response)
        .whereType<Map<String, dynamic>>()
        .toList(growable: false);
    return Response(
      data: list,
      statusCode: response.statusCode,
      statusMessage: response.statusMessage,
      requestOptions: response.requestOptions,
    );
  }

  Future<Response<String>> runPresetCommand(String name) async {
    final response = await _client.post<Map<String, dynamic>>(
      ApiConstants.buildApiPath('/commands/preset/run'),
      data: {'name': name},
    );
    final output = ApiResponseParser.extractMapData(response)['data'];
    return Response(
      data: output is String ? output : '',
      statusCode: response.statusCode,
      statusMessage: response.statusMessage,
      requestOptions: response.requestOptions,
    );
  }
}
