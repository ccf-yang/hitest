// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Chinese (`zh`).
class AppLocalizationsZh extends AppLocalizations {
  AppLocalizationsZh([String locale = 'zh']) : super(locale);

  @override
  String get appName => 'hitest Client';

  @override
  String get commonLoading => '加载中...';

  @override
  String get commonRetry => '重试';

  @override
  String get commonCancel => '取消';

  @override
  String get commonConfirm => '确认';

  @override
  String get commonMore => '更多';

  @override
  String get commonSave => '保存';

  @override
  String get commonDelete => '删除';

  @override
  String get commonClose => '关闭';

  @override
  String get commonRefresh => '刷新';

  @override
  String get commonComingSoon => '即将支持';

  @override
  String get commonEmpty => '暂无数据';

  @override
  String get commonLoadFailedTitle => '加载失败';

  @override
  String get monitorNetworkLabel => '网络';

  @override
  String get monitorMetricCurrent => '当前';

  @override
  String get monitorMetricMin => '最小';

  @override
  String get monitorMetricAvg => '平均';

  @override
  String get monitorMetricMax => '最大';

  @override
  String get navServer => '服务器';

  @override
  String noServerSelectedTitle(String module) {
    return '$module 需要先选择服务器';
  }

  @override
  String get noServerSelectedDescription => '请先选择当前服务器，再进入该模块。';

  @override
  String get serverPageTitle => '服务器';

  @override
  String get serverSearchHint => '请输入服务器名称或 IP';

  @override
  String get serverAdd => '添加';

  @override
  String get serverListEmptyTitle => '暂无服务器';

  @override
  String get serverListEmptyDesc => '先添加一个 hitest 服务器开始使用。';

  @override
  String get serverCurrent => '当前';

  @override
  String get serverIpLabel => 'IP';

  @override
  String get serverCpuLabel => 'CPU';

  @override
  String get serverMemoryLabel => '内存';

  @override
  String get serverLoadLabel => '负载';

  @override
  String get serverDiskLabel => '磁盘';

  @override
  String get serverMetricsUnavailable => '监控数据暂不可用';

  @override
  String get serverOpenDetail => '查看详情';

  @override
  String get serverDetailTitle => '服务器详情';

  @override
  String get serverModuleMonitoring => '监控';

  @override
  String get serverActionRefresh => '刷新';

  @override
  String get serverActionSwitch => '切换服务器';

  @override
  String get serverFormTitle => '添加服务器';

  @override
  String get serverFormName => '服务器名称';

  @override
  String get serverFormNameHint => '例如：生产环境';

  @override
  String get serverFormUrl => '服务器地址';

  @override
  String get serverFormUrlHint => '例如：https://panel.example.com';

  @override
  String get serverFormApiKey => 'API 密钥';

  @override
  String get serverFormApiKeyHint => '服务器未开启鉴权时可留空';

  @override
  String get serverFormSaveConnect => '保存并继续';

  @override
  String get serverFormTest => '测试连接';

  @override
  String get serverFormRequired => '该字段不能为空';

  @override
  String get serverFormSaveSuccess => '服务器已保存';

  @override
  String serverFormSaveFailed(String error) {
    return '保存服务器失败：$error';
  }

  @override
  String get serverDeleteConfirmTitle => '删除服务器';

  @override
  String serverDeleteConfirmMessage(String name) {
    return '确定删除服务器 $name 吗？';
  }

  @override
  String serverDeleteSuccess(String name) {
    return '已删除服务器 $name';
  }

  @override
  String serverDeleteFailed(String error) {
    return '删除服务器失败：$error';
  }

  @override
  String get serverTestSuccess => '连接成功';

  @override
  String get serverTestFailed => '连接失败';

  @override
  String get serverTestTesting => '正在测试连接...';

  @override
  String get serverMetricsAvailable => '监控数据已加载';

  @override
  String get serverTokenValidity => '接口密钥有效期';

  @override
  String get serverTokenValidityHint => '设置为0时不校验时间戳';

  @override
  String get serverFormAllowInsecureTls => '忽略 TLS 证书校验';

  @override
  String get serverFormAllowInsecureTlsHint => '仅在可信的自签名或内网环境中启用。';

  @override
  String get serverFormMinutes => '分钟';

  @override
  String get settingsPageTitle => '设置';

  @override
  String get settingsGeneral => '通用';

  @override
  String get settingsAppSectionTitle => '应用';

  @override
  String get settingsTheme => '主题';

  @override
  String get settingsServerManagement => '服务器管理';

  @override
  String get settingsServerManagementSubtitle =>
      '管理多台已连接服务器；服务端系统设置仅在对应服务器详情页中提供。';

  @override
  String get themeSystem => '跟随系统';

  @override
  String get themeLight => '浅色';

  @override
  String get themeDark => '深色';

  @override
  String get authLogin => '登录';

  @override
  String get notFoundTitle => '页面不存在';

  @override
  String get notFoundDesc => '请求的页面不存在或已迁移。';

  @override
  String get monitorSettings => '监控设置';

  @override
  String get monitorInterval => '监控间隔';

  @override
  String get monitorIntervalUnit => '秒';

  @override
  String get monitorRetentionUnit => '天';

  @override
  String get monitorCleanData => '清理监控数据';

  @override
  String get monitorCleanConfirm => '确定要清理所有监控数据吗？此操作无法撤销。';

  @override
  String get monitorCleanSuccess => '监控数据清理成功';

  @override
  String get monitorCleanFailed => '清理监控数据失败';

  @override
  String get monitorSettingsSaved => '设置保存成功';

  @override
  String get monitorSettingsFailed => '保存设置失败';

  @override
  String get monitorSettingsTitle => '监控设置';

  @override
  String get monitorSettingsInterval => '监控间隔';

  @override
  String get monitorSettingsStoreDays => '数据保留天数';

  @override
  String get monitorSettingsEnable => '启用监控';

  @override
  String get commonError => '发生错误';

  @override
  String get env => '环境变量';

  @override
  String get quickCommandsTitle => '快捷命令';

  @override
  String get quickCommandsEmptyDesc =>
      '暂无预置命令。请在服务端 hitest/conf/commands.yaml 中按 name/command 配置。';

  @override
  String get quickCommandsNoOutput => '（无输出）';

  @override
  String get serverFormMachine => '机器 ID';

  @override
  String get serverFormMachineHint => 'local 或留空 = 中心服务所在机器';

  @override
  String get serverFormMachineHelp =>
      'hub 下挂了远程节点时，填注册上来的 agent id（例如 box）才能操作那台机器；留空只看 hub 本机。';

  @override
  String get navFiles => '文件';

  @override
  String get navTerminal => '终端';

  @override
  String get filesTitle => '文件';

  @override
  String get filesEmpty => '空目录';

  @override
  String filesLoadFailed(String error) {
    return '读取目录失败：$error';
  }

  @override
  String get filesUp => '返回上级';

  @override
  String get filesHome => '根目录';

  @override
  String get filesNewFolder => '新建目录';

  @override
  String get filesNewFolderHint => '目录名';

  @override
  String get filesRename => '重命名';

  @override
  String get filesRenameHint => '新名称';

  @override
  String filesDeleteConfirm(String name) {
    return '删除「$name」？目录会连同内容一并删除。';
  }

  @override
  String get filesOpSuccess => '操作完成';

  @override
  String filesOpFailed(String error) {
    return '操作失败：$error';
  }

  @override
  String get filesDownloaded => '已保存到本机';

  @override
  String get filesPickToUpload => '选择要上传的文件';

  @override
  String get filesUploading => '上传中';

  @override
  String get filesPreview => '查看内容';

  @override
  String get filesPreviewTooLarge => '文件较大，不在此预览，请下载后查看。';

  @override
  String get filesFileSizeUnknown => '未知大小';

  @override
  String get terminalTitle => '终端';

  @override
  String get terminalConnect => '连接';

  @override
  String get terminalDisconnect => '断开';

  @override
  String get terminalReconnect => '重新连接';

  @override
  String get terminalNotConnected => '未连接';

  @override
  String get terminalConnecting => '连接中…';

  @override
  String get terminalClosed => '会话已结束';

  @override
  String terminalFailed(String error) {
    return '连接失败：$error';
  }

  @override
  String get terminalShell => 'Shell';

  @override
  String get terminalShellHint => '/bin/bash -l';

  @override
  String get terminalSend => '发送';

  @override
  String get terminalClear => '清屏';

  @override
  String get terminalCtrlC => 'Ctrl-C';

  @override
  String get terminalKeyboard => '键盘';

  @override
  String get terminalHideKeyboard => '收起键盘';

  @override
  String get terminalUp => '上方向键';

  @override
  String get terminalDown => '下方向键';
}
