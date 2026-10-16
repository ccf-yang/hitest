import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:dio/dio.dart';

import '../../core/config/api_constants.dart';
import '../../core/network/dio_client.dart';
import 'api_response_parser.dart';

/// 文件路径拼接（服务端接受绝对路径，POSIX 语义）。
String joinPath(String dir, String name) {
  if (dir.endsWith('/')) return '$dir$name';
  return '$dir/$name';
}

String parentPath(String path) {
  final trimmed = path.endsWith('/') && path != '/'
      ? path.substring(0, path.length - 1)
      : path;
  final idx = trimmed.lastIndexOf('/');
  if (idx <= 0) return '/';
  return trimmed.substring(0, idx);
}

String formatSize(int bytes) {
  if (bytes < 1024) return '$bytes B';
  if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
  if (bytes < 1024 * 1024 * 1024) {
    return '${(bytes / 1024 / 1024).toStringAsFixed(1)} MB';
  }
  return '${(bytes / 1024 / 1024 / 1024).toStringAsFixed(2)} GB';
}

/// 服务端 `GET /files/list` 返回的一条目录项。
class FileEntry {
  const FileEntry({
    required this.name,
    required this.path,
    required this.size,
    required this.mode,
    required this.mtime,
    required this.isDir,
  });

  final String name;
  final String path;
  final int size;

  /// Go `os.FileMode.String()`，例如 `drwxr-xr-x`、`-rw-r--r--`。
  final String mode;
  final int mtime;
  final bool isDir;

  factory FileEntry.fromMap(Map<String, dynamic> map, String dir) {
    final name = map['name']?.toString() ?? '';
    return FileEntry(
      name: name,
      path: joinPath(dir, name),
      size: (map['size'] as num?)?.toInt() ?? 0,
      mode: map['mode']?.toString() ?? '',
      mtime: (map['mtime'] as num?)?.toInt() ?? 0,
      isDir: map['is_dir'] == true,
    );
  }

  DateTime get modified => DateTime.fromMillisecondsSinceEpoch(mtime * 1000);

  String get sizeLabel => formatSize(size);
}

/// 文件模块客户端。
///
/// 注意：服务端 envelope 中间件对 `/files/` 路径**不做包裹**（见
/// internal/server/envelope.go），所以这里按裸 JSON 解析；上传是 PUT 裸字节体，
/// 下载是 GET 裸字节体（支持 Range）。
class FilesV2Api {
  FilesV2Api(this._client);

  final DioClient _client;

  Future<List<FileEntry>> list(String path) async {
    final response = await _client.get<Map<String, dynamic>>(
      ApiConstants.buildApiPath('/files/list'),
      queryParameters: {'path': path},
    );
    final payload = ApiResponseParser.asMap(response.data,
        fallbackToRootMap: true);
    final entries = payload['entries'];
    if (entries is! List) return const [];
    return [
      for (final e in entries.whereType<Map>())
        FileEntry.fromMap(Map<String, dynamic>.from(e), path),
    ];
  }

  Future<void> mkdir(String path) => _post('/files/mkdir', {'path': path});

  Future<void> delete(String path) => _post('/files/delete', {'path': path});

  Future<void> rename(String path, String newPath) =>
      _post('/files/rename', {'path': path, 'new_path': newPath});

  Future<void> _post(String endpoint, Map<String, dynamic> body) async {
    await _client.post<Map<String, dynamic>>(
      ApiConstants.buildApiPath(endpoint),
      data: body,
    );
  }

  /// 上传：请求体即文件内容，服务端流式落盘，不分块、不驻内存。
  Future<void> uploadBytes(String path, List<int> bytes) async {
    await _client.put<Map<String, dynamic>>(
      ApiConstants.buildApiPath('/files/file'),
      queryParameters: {'path': path},
      data: bytes,
      options: Options(
        headers: {HttpHeaders.contentTypeHeader: 'application/octet-stream'},
        sendTimeout: const Duration(minutes: 5),
        receiveTimeout: const Duration(minutes: 5),
      ),
    );
  }

  /// 下载到本机文件（沙盒目录），大文件不整段进内存。
  Future<void> downloadToFile(String remotePath, String localPath) async {
    await _client.download(
      ApiConstants.buildApiPath('/files/file'),
      localPath,
      queryParameters: {'path': remotePath},
      options: Options(
        responseType: ResponseType.bytes,
        receiveTimeout: const Duration(minutes: 10),
      ),
    );
  }

  /// 取文件原始字节。注意 validateStatus 只放行 2xx, 所以失败一定走异常。
  Future<Uint8List> downloadToBytes(String remotePath) async {
    final response = await _client.get<Uint8List>(
      ApiConstants.buildApiPath('/files/file'),
      queryParameters: {'path': remotePath},
      options: Options(
        responseType: ResponseType.bytes,
        receiveTimeout: const Duration(minutes: 10),
      ),
    );
    final bytes = response.data;
    if (bytes == null) {
      throw StateError('empty file body: $remotePath');
    }
    return bytes;
  }

  static const int previewLimit = 256 * 1024;

  /// 读文本预览（至多 previewLimit 字节, 超过直接返回空串让 UI 提示下载）。
  Future<String> readPreview(String remotePath, int size) async {
    if (size > previewLimit) return '';
    final bytes = await downloadToBytes(remotePath);
    return decodeUtf8AllowErrors(bytes);
  }

  static bool isProbablyText(FileEntry entry) {
    return !entry.isDir && entry.size <= previewLimit;
  }
}

/// 二进制安全解码：非法字节替换为 U+FFFD 而不是抛异常。
String decodeUtf8AllowErrors(List<int> bytes) {
  return const Utf8Decoder(allowMalformed: true).convert(bytes);
}
