import 'dart:io';
import 'dart:typed_data';
import 'package:path_provider/path_provider.dart';
import 'package:file_picker/file_picker.dart';
import 'package:hitest/core/services/logger/logger_service.dart';

class FileSaveResult {
  final String? filePath;
  final bool success;
  final String? errorMessage;

  const FileSaveResult({
    this.filePath,
    required this.success,
    this.errorMessage,
  });
}

class FileSaveService {
  static final FileSaveService _instance = FileSaveService._internal();
  factory FileSaveService() => _instance;
  FileSaveService._internal();

  Future<FileSaveResult> saveFile({
    required String fileName,
    required Uint8List bytes,
    String? mimeType,
  }) async {
    appLogger.dWithPackage('file_save',
        'saveFile: fileName=$fileName, bytesLength=${bytes.length}');

    try {
      final result = await FilePicker.platform.saveFile(
        dialogTitle: '保存文件',
        fileName: fileName,
        bytes: bytes,
      );

      if (result != null) {
        appLogger.iWithPackage('file_save', 'saveFile: 文件已保存到 $result');
        return FileSaveResult(
          success: true,
          filePath: result,
        );
      } else {
        appLogger.iWithPackage('file_save', 'saveFile: 用户取消保存');
        return const FileSaveResult(
          success: false,
          errorMessage: '用户取消保存',
        );
      }
    } catch (e) {
      appLogger.wWithPackage(
          'file_save', 'saveFile: FilePicker保存失败，降级到默认下载目录: $e');

      final downloadDir = await _getFallbackDownloadDir();
      final safeFileName = _sanitizeFileName(fileName);
      final filePath = await _getUniqueFilePath(downloadDir.path, safeFileName);

      final file = await File(filePath).create(recursive: true);
      await file.writeAsBytes(bytes);

      appLogger.iWithPackage(
          'file_save', 'saveFile: 文件已保存到降级目录 $filePath');
      return FileSaveResult(
        success: true,
        filePath: filePath,
      );
    }
  }

  Future<Directory> _getFallbackDownloadDir() async {
    if (Platform.isAndroid || Platform.isIOS) {
      return await getApplicationDocumentsDirectory();
    } else {
      return await getDownloadsDirectory() ??
          await getApplicationDocumentsDirectory();
    }
  }

  String _sanitizeFileName(String fileName) {
    return fileName.replaceAll(RegExp(r'[<>:"/\\|?*]'), '_');
  }

  Future<String> _getUniqueFilePath(String directory, String fileName) async {
    final filePath = '$directory/$fileName';

    if (!await File(filePath).exists()) {
      return filePath;
    }

    final lastDot = fileName.lastIndexOf('.');
    String baseName;
    String extension;

    if (lastDot > 0) {
      baseName = fileName.substring(0, lastDot);
      extension = fileName.substring(lastDot);
    } else {
      baseName = fileName;
      extension = '';
    }

    int counter = 1;
    while (true) {
      final newPath = '$directory/${baseName}_$counter$extension';
      if (!await File(newPath).exists()) {
        return newPath;
      }
      counter++;
    }
  }

  Future<void> openFile(String filePath) async {
    appLogger.dWithPackage('file_save', 'openFile: filePath=$filePath');

    if (!await File(filePath).exists()) {
      throw Exception('文件不存在: $filePath');
    }

    throw UnimplementedError(
      '请在 pubspec.yaml 中添加 open_file 包依赖后使用。\n'
      '依赖: open_file: ^3.5.10\n'
      '导入: import \'package:open_file/open_file.dart\';',
    );
  }

  Future<void> openFileLocation(String filePath) async {
    appLogger.dWithPackage('file_save', 'openFileLocation: filePath=$filePath');

    final file = File(filePath);
    if (!await file.exists()) {
      throw Exception('文件不存在: $filePath');
    }

    final directory = file.parent;

    if (Platform.isMacOS) {
      await Process.run('open', [directory.path]);
    } else if (Platform.isWindows) {
      await Process.run('explorer', [directory.path]);
    } else if (Platform.isLinux) {
      await Process.run('xdg-open', [directory.path]);
    } else {
      throw UnsupportedError('当前平台不支持打开文件所在目录');
    }

    appLogger.iWithPackage(
        'file_save', 'openFileLocation: 已打开目录 ${directory.path}');
  }

  Future<String?> getDownloadDirectoryPath() async {
    try {
      if (Platform.isAndroid) {
        // 对于 Android，默认返回公共下载目录，以便 FlutterDownloader 使用
        final dir = Directory('/storage/emulated/0/Download');
        if (!await dir.exists()) {
          final extDir = await getExternalStorageDirectory();
          return extDir?.path ?? (await getApplicationDocumentsDirectory()).path;
        }
        return dir.path;
      } else if (Platform.isIOS) {
        final dir = await getApplicationDocumentsDirectory();
        return dir.path;
      } else {
        final dir = await getDownloadsDirectory() ??
            await getApplicationDocumentsDirectory();
        return dir.path;
      }
    } catch (e) {
      appLogger.eWithPackage('file_save', 'getDownloadDirectoryPath: 获取下载目录失败',
          error: e);
      return null;
    }
  }
}
