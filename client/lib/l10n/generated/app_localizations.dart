import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_zh.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'generated/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
      : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
    delegate,
    GlobalMaterialLocalizations.delegate,
    GlobalCupertinoLocalizations.delegate,
    GlobalWidgetsLocalizations.delegate,
  ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[Locale('zh')];

  /// No description provided for @appName.
  ///
  /// In zh, this message translates to:
  /// **'hitest Client'**
  String get appName;

  /// No description provided for @commonLoading.
  ///
  /// In zh, this message translates to:
  /// **'加载中...'**
  String get commonLoading;

  /// No description provided for @commonRetry.
  ///
  /// In zh, this message translates to:
  /// **'重试'**
  String get commonRetry;

  /// No description provided for @commonCancel.
  ///
  /// In zh, this message translates to:
  /// **'取消'**
  String get commonCancel;

  /// No description provided for @commonConfirm.
  ///
  /// In zh, this message translates to:
  /// **'确认'**
  String get commonConfirm;

  /// No description provided for @commonMore.
  ///
  /// In zh, this message translates to:
  /// **'更多'**
  String get commonMore;

  /// No description provided for @commonSave.
  ///
  /// In zh, this message translates to:
  /// **'保存'**
  String get commonSave;

  /// No description provided for @commonDelete.
  ///
  /// In zh, this message translates to:
  /// **'删除'**
  String get commonDelete;

  /// No description provided for @commonClose.
  ///
  /// In zh, this message translates to:
  /// **'关闭'**
  String get commonClose;

  /// No description provided for @commonRefresh.
  ///
  /// In zh, this message translates to:
  /// **'刷新'**
  String get commonRefresh;

  /// No description provided for @commonComingSoon.
  ///
  /// In zh, this message translates to:
  /// **'即将支持'**
  String get commonComingSoon;

  /// No description provided for @commonEmpty.
  ///
  /// In zh, this message translates to:
  /// **'暂无数据'**
  String get commonEmpty;

  /// No description provided for @commonLoadFailedTitle.
  ///
  /// In zh, this message translates to:
  /// **'加载失败'**
  String get commonLoadFailedTitle;

  /// No description provided for @monitorNetworkLabel.
  ///
  /// In zh, this message translates to:
  /// **'网络'**
  String get monitorNetworkLabel;

  /// No description provided for @monitorMetricCurrent.
  ///
  /// In zh, this message translates to:
  /// **'当前'**
  String get monitorMetricCurrent;

  /// No description provided for @monitorMetricMin.
  ///
  /// In zh, this message translates to:
  /// **'最小'**
  String get monitorMetricMin;

  /// No description provided for @monitorMetricAvg.
  ///
  /// In zh, this message translates to:
  /// **'平均'**
  String get monitorMetricAvg;

  /// No description provided for @monitorMetricMax.
  ///
  /// In zh, this message translates to:
  /// **'最大'**
  String get monitorMetricMax;

  /// No description provided for @navServer.
  ///
  /// In zh, this message translates to:
  /// **'服务器'**
  String get navServer;

  /// No description provided for @noServerSelectedTitle.
  ///
  /// In zh, this message translates to:
  /// **'{module} 需要先选择服务器'**
  String noServerSelectedTitle(String module);

  /// No description provided for @noServerSelectedDescription.
  ///
  /// In zh, this message translates to:
  /// **'请先选择当前服务器，再进入该模块。'**
  String get noServerSelectedDescription;

  /// No description provided for @serverPageTitle.
  ///
  /// In zh, this message translates to:
  /// **'服务器'**
  String get serverPageTitle;

  /// No description provided for @serverSearchHint.
  ///
  /// In zh, this message translates to:
  /// **'请输入服务器名称或 IP'**
  String get serverSearchHint;

  /// No description provided for @serverAdd.
  ///
  /// In zh, this message translates to:
  /// **'添加'**
  String get serverAdd;

  /// No description provided for @serverListEmptyTitle.
  ///
  /// In zh, this message translates to:
  /// **'暂无服务器'**
  String get serverListEmptyTitle;

  /// No description provided for @serverListEmptyDesc.
  ///
  /// In zh, this message translates to:
  /// **'先添加一个 hitest 服务器开始使用。'**
  String get serverListEmptyDesc;

  /// No description provided for @serverCurrent.
  ///
  /// In zh, this message translates to:
  /// **'当前'**
  String get serverCurrent;

  /// No description provided for @serverIpLabel.
  ///
  /// In zh, this message translates to:
  /// **'IP'**
  String get serverIpLabel;

  /// No description provided for @serverCpuLabel.
  ///
  /// In zh, this message translates to:
  /// **'CPU'**
  String get serverCpuLabel;

  /// No description provided for @serverMemoryLabel.
  ///
  /// In zh, this message translates to:
  /// **'内存'**
  String get serverMemoryLabel;

  /// No description provided for @serverLoadLabel.
  ///
  /// In zh, this message translates to:
  /// **'负载'**
  String get serverLoadLabel;

  /// No description provided for @serverDiskLabel.
  ///
  /// In zh, this message translates to:
  /// **'磁盘'**
  String get serverDiskLabel;

  /// No description provided for @serverMetricsUnavailable.
  ///
  /// In zh, this message translates to:
  /// **'监控数据暂不可用'**
  String get serverMetricsUnavailable;

  /// No description provided for @serverOpenDetail.
  ///
  /// In zh, this message translates to:
  /// **'查看详情'**
  String get serverOpenDetail;

  /// No description provided for @serverDetailTitle.
  ///
  /// In zh, this message translates to:
  /// **'服务器详情'**
  String get serverDetailTitle;

  /// No description provided for @serverModuleMonitoring.
  ///
  /// In zh, this message translates to:
  /// **'监控'**
  String get serverModuleMonitoring;

  /// No description provided for @serverActionRefresh.
  ///
  /// In zh, this message translates to:
  /// **'刷新'**
  String get serverActionRefresh;

  /// No description provided for @serverActionSwitch.
  ///
  /// In zh, this message translates to:
  /// **'切换服务器'**
  String get serverActionSwitch;

  /// No description provided for @serverFormTitle.
  ///
  /// In zh, this message translates to:
  /// **'添加服务器'**
  String get serverFormTitle;

  /// No description provided for @serverFormName.
  ///
  /// In zh, this message translates to:
  /// **'服务器名称'**
  String get serverFormName;

  /// No description provided for @serverFormNameHint.
  ///
  /// In zh, this message translates to:
  /// **'例如：生产环境'**
  String get serverFormNameHint;

  /// No description provided for @serverFormUrl.
  ///
  /// In zh, this message translates to:
  /// **'服务器地址'**
  String get serverFormUrl;

  /// No description provided for @serverFormUrlHint.
  ///
  /// In zh, this message translates to:
  /// **'例如：https://panel.example.com'**
  String get serverFormUrlHint;

  /// No description provided for @serverFormApiKey.
  ///
  /// In zh, this message translates to:
  /// **'API 密钥'**
  String get serverFormApiKey;

  /// No description provided for @serverFormApiKeyHint.
  ///
  /// In zh, this message translates to:
  /// **'服务器未开启鉴权时可留空'**
  String get serverFormApiKeyHint;

  /// No description provided for @serverFormSaveConnect.
  ///
  /// In zh, this message translates to:
  /// **'保存并继续'**
  String get serverFormSaveConnect;

  /// No description provided for @serverFormTest.
  ///
  /// In zh, this message translates to:
  /// **'测试连接'**
  String get serverFormTest;

  /// No description provided for @serverFormRequired.
  ///
  /// In zh, this message translates to:
  /// **'该字段不能为空'**
  String get serverFormRequired;

  /// No description provided for @serverFormSaveSuccess.
  ///
  /// In zh, this message translates to:
  /// **'服务器已保存'**
  String get serverFormSaveSuccess;

  /// No description provided for @serverFormSaveFailed.
  ///
  /// In zh, this message translates to:
  /// **'保存服务器失败：{error}'**
  String serverFormSaveFailed(String error);

  /// No description provided for @serverDeleteConfirmTitle.
  ///
  /// In zh, this message translates to:
  /// **'删除服务器'**
  String get serverDeleteConfirmTitle;

  /// No description provided for @serverDeleteConfirmMessage.
  ///
  /// In zh, this message translates to:
  /// **'确定删除服务器 {name} 吗？'**
  String serverDeleteConfirmMessage(String name);

  /// No description provided for @serverDeleteSuccess.
  ///
  /// In zh, this message translates to:
  /// **'已删除服务器 {name}'**
  String serverDeleteSuccess(String name);

  /// No description provided for @serverDeleteFailed.
  ///
  /// In zh, this message translates to:
  /// **'删除服务器失败：{error}'**
  String serverDeleteFailed(String error);

  /// No description provided for @serverTestSuccess.
  ///
  /// In zh, this message translates to:
  /// **'连接成功'**
  String get serverTestSuccess;

  /// No description provided for @serverTestFailed.
  ///
  /// In zh, this message translates to:
  /// **'连接失败'**
  String get serverTestFailed;

  /// No description provided for @serverTestTesting.
  ///
  /// In zh, this message translates to:
  /// **'正在测试连接...'**
  String get serverTestTesting;

  /// No description provided for @serverMetricsAvailable.
  ///
  /// In zh, this message translates to:
  /// **'监控数据已加载'**
  String get serverMetricsAvailable;

  /// No description provided for @serverTokenValidity.
  ///
  /// In zh, this message translates to:
  /// **'接口密钥有效期'**
  String get serverTokenValidity;

  /// No description provided for @serverTokenValidityHint.
  ///
  /// In zh, this message translates to:
  /// **'设置为0时不校验时间戳'**
  String get serverTokenValidityHint;

  /// No description provided for @serverFormAllowInsecureTls.
  ///
  /// In zh, this message translates to:
  /// **'忽略 TLS 证书校验'**
  String get serverFormAllowInsecureTls;

  /// No description provided for @serverFormAllowInsecureTlsHint.
  ///
  /// In zh, this message translates to:
  /// **'仅在可信的自签名或内网环境中启用。'**
  String get serverFormAllowInsecureTlsHint;

  /// No description provided for @serverFormMinutes.
  ///
  /// In zh, this message translates to:
  /// **'分钟'**
  String get serverFormMinutes;

  /// No description provided for @settingsPageTitle.
  ///
  /// In zh, this message translates to:
  /// **'设置'**
  String get settingsPageTitle;

  /// No description provided for @settingsGeneral.
  ///
  /// In zh, this message translates to:
  /// **'通用'**
  String get settingsGeneral;

  /// No description provided for @settingsAppSectionTitle.
  ///
  /// In zh, this message translates to:
  /// **'应用'**
  String get settingsAppSectionTitle;

  /// No description provided for @settingsTheme.
  ///
  /// In zh, this message translates to:
  /// **'主题'**
  String get settingsTheme;

  /// No description provided for @settingsServerManagement.
  ///
  /// In zh, this message translates to:
  /// **'服务器管理'**
  String get settingsServerManagement;

  /// No description provided for @settingsServerManagementSubtitle.
  ///
  /// In zh, this message translates to:
  /// **'管理多台已连接服务器；服务端系统设置仅在对应服务器详情页中提供。'**
  String get settingsServerManagementSubtitle;

  /// No description provided for @themeSystem.
  ///
  /// In zh, this message translates to:
  /// **'跟随系统'**
  String get themeSystem;

  /// No description provided for @themeLight.
  ///
  /// In zh, this message translates to:
  /// **'浅色'**
  String get themeLight;

  /// No description provided for @themeDark.
  ///
  /// In zh, this message translates to:
  /// **'深色'**
  String get themeDark;

  /// No description provided for @authLogin.
  ///
  /// In zh, this message translates to:
  /// **'登录'**
  String get authLogin;

  /// No description provided for @notFoundTitle.
  ///
  /// In zh, this message translates to:
  /// **'页面不存在'**
  String get notFoundTitle;

  /// No description provided for @notFoundDesc.
  ///
  /// In zh, this message translates to:
  /// **'请求的页面不存在或已迁移。'**
  String get notFoundDesc;

  /// No description provided for @monitorSettings.
  ///
  /// In zh, this message translates to:
  /// **'监控设置'**
  String get monitorSettings;

  /// No description provided for @monitorInterval.
  ///
  /// In zh, this message translates to:
  /// **'监控间隔'**
  String get monitorInterval;

  /// No description provided for @monitorIntervalUnit.
  ///
  /// In zh, this message translates to:
  /// **'秒'**
  String get monitorIntervalUnit;

  /// No description provided for @monitorRetentionUnit.
  ///
  /// In zh, this message translates to:
  /// **'天'**
  String get monitorRetentionUnit;

  /// No description provided for @monitorCleanData.
  ///
  /// In zh, this message translates to:
  /// **'清理监控数据'**
  String get monitorCleanData;

  /// No description provided for @monitorCleanConfirm.
  ///
  /// In zh, this message translates to:
  /// **'确定要清理所有监控数据吗？此操作无法撤销。'**
  String get monitorCleanConfirm;

  /// No description provided for @monitorCleanSuccess.
  ///
  /// In zh, this message translates to:
  /// **'监控数据清理成功'**
  String get monitorCleanSuccess;

  /// No description provided for @monitorCleanFailed.
  ///
  /// In zh, this message translates to:
  /// **'清理监控数据失败'**
  String get monitorCleanFailed;

  /// No description provided for @monitorSettingsSaved.
  ///
  /// In zh, this message translates to:
  /// **'设置保存成功'**
  String get monitorSettingsSaved;

  /// No description provided for @monitorSettingsFailed.
  ///
  /// In zh, this message translates to:
  /// **'保存设置失败'**
  String get monitorSettingsFailed;

  /// No description provided for @monitorSettingsTitle.
  ///
  /// In zh, this message translates to:
  /// **'监控设置'**
  String get monitorSettingsTitle;

  /// No description provided for @monitorSettingsInterval.
  ///
  /// In zh, this message translates to:
  /// **'监控间隔'**
  String get monitorSettingsInterval;

  /// No description provided for @monitorSettingsStoreDays.
  ///
  /// In zh, this message translates to:
  /// **'数据保留天数'**
  String get monitorSettingsStoreDays;

  /// No description provided for @monitorSettingsEnable.
  ///
  /// In zh, this message translates to:
  /// **'启用监控'**
  String get monitorSettingsEnable;

  /// No description provided for @commonError.
  ///
  /// In zh, this message translates to:
  /// **'发生错误'**
  String get commonError;

  /// No description provided for @env.
  ///
  /// In zh, this message translates to:
  /// **'环境变量'**
  String get env;

  /// No description provided for @quickCommandsTitle.
  ///
  /// In zh, this message translates to:
  /// **'快捷命令'**
  String get quickCommandsTitle;

  /// No description provided for @quickCommandsEmptyDesc.
  ///
  /// In zh, this message translates to:
  /// **'暂无预置命令。请在服务端 hitest/conf/commands.yaml 中按 name/command 配置。'**
  String get quickCommandsEmptyDesc;

  /// No description provided for @quickCommandsNoOutput.
  ///
  /// In zh, this message translates to:
  /// **'（无输出）'**
  String get quickCommandsNoOutput;

  /// No description provided for @serverFormMachine.
  ///
  /// In zh, this message translates to:
  /// **'机器 ID'**
  String get serverFormMachine;

  /// No description provided for @serverFormMachineHint.
  ///
  /// In zh, this message translates to:
  /// **'local 或留空 = 中心服务所在机器'**
  String get serverFormMachineHint;

  /// No description provided for @serverFormMachineHelp.
  ///
  /// In zh, this message translates to:
  /// **'hub 下挂了远程节点时，填注册上来的 agent id（例如 box）才能操作那台机器；留空只看 hub 本机。'**
  String get serverFormMachineHelp;

  /// No description provided for @navFiles.
  ///
  /// In zh, this message translates to:
  /// **'文件'**
  String get navFiles;

  /// No description provided for @navTerminal.
  ///
  /// In zh, this message translates to:
  /// **'终端'**
  String get navTerminal;

  /// No description provided for @filesTitle.
  ///
  /// In zh, this message translates to:
  /// **'文件'**
  String get filesTitle;

  /// No description provided for @filesEmpty.
  ///
  /// In zh, this message translates to:
  /// **'空目录'**
  String get filesEmpty;

  /// No description provided for @filesLoadFailed.
  ///
  /// In zh, this message translates to:
  /// **'读取目录失败：{error}'**
  String filesLoadFailed(String error);

  /// No description provided for @filesUp.
  ///
  /// In zh, this message translates to:
  /// **'返回上级'**
  String get filesUp;

  /// No description provided for @filesHome.
  ///
  /// In zh, this message translates to:
  /// **'根目录'**
  String get filesHome;

  /// No description provided for @filesNewFolder.
  ///
  /// In zh, this message translates to:
  /// **'新建目录'**
  String get filesNewFolder;

  /// No description provided for @filesNewFolderHint.
  ///
  /// In zh, this message translates to:
  /// **'目录名'**
  String get filesNewFolderHint;

  /// No description provided for @filesRename.
  ///
  /// In zh, this message translates to:
  /// **'重命名'**
  String get filesRename;

  /// No description provided for @filesRenameHint.
  ///
  /// In zh, this message translates to:
  /// **'新名称'**
  String get filesRenameHint;

  /// No description provided for @filesDeleteConfirm.
  ///
  /// In zh, this message translates to:
  /// **'删除「{name}」？目录会连同内容一并删除。'**
  String filesDeleteConfirm(String name);

  /// No description provided for @filesOpSuccess.
  ///
  /// In zh, this message translates to:
  /// **'操作完成'**
  String get filesOpSuccess;

  /// No description provided for @filesOpFailed.
  ///
  /// In zh, this message translates to:
  /// **'操作失败：{error}'**
  String filesOpFailed(String error);

  /// No description provided for @filesDownloaded.
  ///
  /// In zh, this message translates to:
  /// **'已保存到本机'**
  String get filesDownloaded;

  /// No description provided for @filesPickToUpload.
  ///
  /// In zh, this message translates to:
  /// **'选择要上传的文件'**
  String get filesPickToUpload;

  /// No description provided for @filesUploading.
  ///
  /// In zh, this message translates to:
  /// **'上传中'**
  String get filesUploading;

  /// No description provided for @filesPreview.
  ///
  /// In zh, this message translates to:
  /// **'查看内容'**
  String get filesPreview;

  /// No description provided for @filesPreviewTooLarge.
  ///
  /// In zh, this message translates to:
  /// **'文件较大，不在此预览，请下载后查看。'**
  String get filesPreviewTooLarge;

  /// No description provided for @filesFileSizeUnknown.
  ///
  /// In zh, this message translates to:
  /// **'未知大小'**
  String get filesFileSizeUnknown;

  /// No description provided for @terminalTitle.
  ///
  /// In zh, this message translates to:
  /// **'终端'**
  String get terminalTitle;

  /// No description provided for @terminalConnect.
  ///
  /// In zh, this message translates to:
  /// **'连接'**
  String get terminalConnect;

  /// No description provided for @terminalDisconnect.
  ///
  /// In zh, this message translates to:
  /// **'断开'**
  String get terminalDisconnect;

  /// No description provided for @terminalReconnect.
  ///
  /// In zh, this message translates to:
  /// **'重新连接'**
  String get terminalReconnect;

  /// No description provided for @terminalNotConnected.
  ///
  /// In zh, this message translates to:
  /// **'未连接'**
  String get terminalNotConnected;

  /// No description provided for @terminalConnecting.
  ///
  /// In zh, this message translates to:
  /// **'连接中…'**
  String get terminalConnecting;

  /// No description provided for @terminalClosed.
  ///
  /// In zh, this message translates to:
  /// **'会话已结束'**
  String get terminalClosed;

  /// No description provided for @terminalFailed.
  ///
  /// In zh, this message translates to:
  /// **'连接失败：{error}'**
  String terminalFailed(String error);

  /// No description provided for @terminalShell.
  ///
  /// In zh, this message translates to:
  /// **'Shell'**
  String get terminalShell;

  /// No description provided for @terminalShellHint.
  ///
  /// In zh, this message translates to:
  /// **'/bin/bash -l'**
  String get terminalShellHint;

  /// No description provided for @terminalSend.
  ///
  /// In zh, this message translates to:
  /// **'发送'**
  String get terminalSend;

  /// No description provided for @terminalClear.
  ///
  /// In zh, this message translates to:
  /// **'清屏'**
  String get terminalClear;

  /// No description provided for @terminalCtrlC.
  ///
  /// In zh, this message translates to:
  /// **'Ctrl-C'**
  String get terminalCtrlC;

  /// No description provided for @terminalKeyboard.
  ///
  /// In zh, this message translates to:
  /// **'键盘'**
  String get terminalKeyboard;

  /// No description provided for @terminalHideKeyboard.
  ///
  /// In zh, this message translates to:
  /// **'收起键盘'**
  String get terminalHideKeyboard;

  /// No description provided for @terminalUp.
  ///
  /// In zh, this message translates to:
  /// **'上方向键'**
  String get terminalUp;

  /// No description provided for @terminalDown.
  ///
  /// In zh, this message translates to:
  /// **'下方向键'**
  String get terminalDown;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['zh'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'zh':
      return AppLocalizationsZh();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}
