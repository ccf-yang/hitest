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
  static const List<Locale> supportedLocales = <Locale>[
    Locale('zh')
  ];

  /// No description provided for @appName.
  ///
  /// In en, this message translates to:
  /// **'hitest Client'**
  String get appName;

  /// No description provided for @commonLoading.
  ///
  /// In en, this message translates to:
  /// **'Loading...'**
  String get commonLoading;

  /// No description provided for @commonRetry.
  ///
  /// In en, this message translates to:
  /// **'Retry'**
  String get commonRetry;

  /// No description provided for @commonCancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get commonCancel;

  /// No description provided for @commonConfirm.
  ///
  /// In en, this message translates to:
  /// **'Confirm'**
  String get commonConfirm;

  /// No description provided for @commonMore.
  ///
  /// In en, this message translates to:
  /// **'More'**
  String get commonMore;

  /// No description provided for @commonSave.
  ///
  /// In en, this message translates to:
  /// **'Save'**
  String get commonSave;

  /// No description provided for @commonDelete.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get commonDelete;

  /// No description provided for @commonRefresh.
  ///
  /// In en, this message translates to:
  /// **'Refresh'**
  String get commonRefresh;

  /// No description provided for @commonComingSoon.
  ///
  /// In en, this message translates to:
  /// **'Coming soon'**
  String get commonComingSoon;

  /// No description provided for @commonEmpty.
  ///
  /// In en, this message translates to:
  /// **'No data'**
  String get commonEmpty;

  /// No description provided for @commonLoadFailedTitle.
  ///
  /// In en, this message translates to:
  /// **'Failed to load'**
  String get commonLoadFailedTitle;

  /// No description provided for @monitorNetworkLabel.
  ///
  /// In en, this message translates to:
  /// **'Network'**
  String get monitorNetworkLabel;

  /// No description provided for @monitorMetricCurrent.
  ///
  /// In en, this message translates to:
  /// **'Current'**
  String get monitorMetricCurrent;

  /// No description provided for @monitorMetricMin.
  ///
  /// In en, this message translates to:
  /// **'Min'**
  String get monitorMetricMin;

  /// No description provided for @monitorMetricAvg.
  ///
  /// In en, this message translates to:
  /// **'Avg'**
  String get monitorMetricAvg;

  /// No description provided for @monitorMetricMax.
  ///
  /// In en, this message translates to:
  /// **'Max'**
  String get monitorMetricMax;

  /// No description provided for @navServer.
  ///
  /// In en, this message translates to:
  /// **'Servers'**
  String get navServer;

  /// No description provided for @noServerSelectedTitle.
  ///
  /// In en, this message translates to:
  /// **'{module} needs a server first'**
  String noServerSelectedTitle(String module);

  /// No description provided for @noServerSelectedDescription.
  ///
  /// In en, this message translates to:
  /// **'Choose a current server before opening this module.'**
  String get noServerSelectedDescription;

  /// No description provided for @serverPageTitle.
  ///
  /// In en, this message translates to:
  /// **'Servers'**
  String get serverPageTitle;

  /// No description provided for @serverSearchHint.
  ///
  /// In en, this message translates to:
  /// **'Search server name or IP'**
  String get serverSearchHint;

  /// No description provided for @serverAdd.
  ///
  /// In en, this message translates to:
  /// **'Add'**
  String get serverAdd;

  /// No description provided for @serverListEmptyTitle.
  ///
  /// In en, this message translates to:
  /// **'No servers yet'**
  String get serverListEmptyTitle;

  /// No description provided for @serverListEmptyDesc.
  ///
  /// In en, this message translates to:
  /// **'Add your first hitest server to get started.'**
  String get serverListEmptyDesc;

  /// No description provided for @serverCurrent.
  ///
  /// In en, this message translates to:
  /// **'Current'**
  String get serverCurrent;

  /// No description provided for @serverIpLabel.
  ///
  /// In en, this message translates to:
  /// **'IP'**
  String get serverIpLabel;

  /// No description provided for @serverCpuLabel.
  ///
  /// In en, this message translates to:
  /// **'CPU'**
  String get serverCpuLabel;

  /// No description provided for @serverMemoryLabel.
  ///
  /// In en, this message translates to:
  /// **'Memory'**
  String get serverMemoryLabel;

  /// No description provided for @serverLoadLabel.
  ///
  /// In en, this message translates to:
  /// **'Load'**
  String get serverLoadLabel;

  /// No description provided for @serverDiskLabel.
  ///
  /// In en, this message translates to:
  /// **'Disk'**
  String get serverDiskLabel;

  /// No description provided for @serverMetricsUnavailable.
  ///
  /// In en, this message translates to:
  /// **'Metrics unavailable'**
  String get serverMetricsUnavailable;

  /// No description provided for @serverOpenDetail.
  ///
  /// In en, this message translates to:
  /// **'Open details'**
  String get serverOpenDetail;

  /// No description provided for @serverDetailTitle.
  ///
  /// In en, this message translates to:
  /// **'Server Details'**
  String get serverDetailTitle;

  /// No description provided for @serverModuleMonitoring.
  ///
  /// In en, this message translates to:
  /// **'Monitoring'**
  String get serverModuleMonitoring;

  /// No description provided for @serverActionRefresh.
  ///
  /// In en, this message translates to:
  /// **'Refresh'**
  String get serverActionRefresh;

  /// No description provided for @serverActionSwitch.
  ///
  /// In en, this message translates to:
  /// **'Switch server'**
  String get serverActionSwitch;

  /// No description provided for @serverFormTitle.
  ///
  /// In en, this message translates to:
  /// **'Add Server'**
  String get serverFormTitle;

  /// No description provided for @serverFormName.
  ///
  /// In en, this message translates to:
  /// **'Server name'**
  String get serverFormName;

  /// No description provided for @serverFormNameHint.
  ///
  /// In en, this message translates to:
  /// **'e.g. Production'**
  String get serverFormNameHint;

  /// No description provided for @serverFormUrl.
  ///
  /// In en, this message translates to:
  /// **'Server URL'**
  String get serverFormUrl;

  /// No description provided for @serverFormUrlHint.
  ///
  /// In en, this message translates to:
  /// **'e.g. https://panel.example.com'**
  String get serverFormUrlHint;

  /// No description provided for @serverFormApiKey.
  ///
  /// In en, this message translates to:
  /// **'API key'**
  String get serverFormApiKey;

  /// No description provided for @serverFormApiKeyHint.
  ///
  /// In en, this message translates to:
  /// **'Enter API key'**
  String get serverFormApiKeyHint;

  /// No description provided for @serverFormSaveConnect.
  ///
  /// In en, this message translates to:
  /// **'Save and continue'**
  String get serverFormSaveConnect;

  /// No description provided for @serverFormTest.
  ///
  /// In en, this message translates to:
  /// **'Test connection'**
  String get serverFormTest;

  /// No description provided for @serverFormRequired.
  ///
  /// In en, this message translates to:
  /// **'This field is required'**
  String get serverFormRequired;

  /// No description provided for @serverFormSaveSuccess.
  ///
  /// In en, this message translates to:
  /// **'Server saved'**
  String get serverFormSaveSuccess;

  /// No description provided for @serverFormSaveFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to save server: {error}'**
  String serverFormSaveFailed(String error);

  /// No description provided for @serverDeleteConfirmTitle.
  ///
  /// In en, this message translates to:
  /// **'Delete server'**
  String get serverDeleteConfirmTitle;

  /// No description provided for @serverDeleteConfirmMessage.
  ///
  /// In en, this message translates to:
  /// **'Delete server {name}?'**
  String serverDeleteConfirmMessage(String name);

  /// No description provided for @serverDeleteSuccess.
  ///
  /// In en, this message translates to:
  /// **'Deleted server {name}'**
  String serverDeleteSuccess(String name);

  /// No description provided for @serverDeleteFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to delete server: {error}'**
  String serverDeleteFailed(String error);

  /// No description provided for @serverTestSuccess.
  ///
  /// In en, this message translates to:
  /// **'Connection successful'**
  String get serverTestSuccess;

  /// No description provided for @serverTestFailed.
  ///
  /// In en, this message translates to:
  /// **'Connection failed'**
  String get serverTestFailed;

  /// No description provided for @serverTestTesting.
  ///
  /// In en, this message translates to:
  /// **'Testing connection...'**
  String get serverTestTesting;

  /// No description provided for @serverMetricsAvailable.
  ///
  /// In en, this message translates to:
  /// **'Metrics loaded'**
  String get serverMetricsAvailable;

  /// No description provided for @serverTokenValidity.
  ///
  /// In en, this message translates to:
  /// **'Token validity (minutes)'**
  String get serverTokenValidity;

  /// No description provided for @serverTokenValidityHint.
  ///
  /// In en, this message translates to:
  /// **'Set to 0 to skip timestamp validation'**
  String get serverTokenValidityHint;

  /// No description provided for @serverFormAllowInsecureTls.
  ///
  /// In en, this message translates to:
  /// **'Ignore TLS certificate validation'**
  String get serverFormAllowInsecureTls;

  /// No description provided for @serverFormAllowInsecureTlsHint.
  ///
  /// In en, this message translates to:
  /// **'Only enable for trusted self-signed or internal endpoints.'**
  String get serverFormAllowInsecureTlsHint;

  /// No description provided for @serverFormMinutes.
  ///
  /// In en, this message translates to:
  /// **'minutes'**
  String get serverFormMinutes;

  /// No description provided for @settingsPageTitle.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settingsPageTitle;

  /// No description provided for @settingsGeneral.
  ///
  /// In en, this message translates to:
  /// **'General'**
  String get settingsGeneral;

  /// No description provided for @settingsAppSectionTitle.
  ///
  /// In en, this message translates to:
  /// **'App'**
  String get settingsAppSectionTitle;

  /// No description provided for @settingsTheme.
  ///
  /// In en, this message translates to:
  /// **'Theme'**
  /// No description provided for @quickCommandsTitle.
  String get quickCommandsTitle;

  /// No description provided for @quickCommandsEmptyDesc.
  String get quickCommandsEmptyDesc;

  /// No description provided for @quickCommandsNoOutput.
  String get quickCommandsNoOutput;

  String get settingsTheme;

  /// No description provided for @settingsServerManagement.
  ///
  /// In en, this message translates to:
  /// **'Server management'**
  String get settingsServerManagement;

  /// No description provided for @settingsServerManagementSubtitle.
  ///
  /// In en, this message translates to:
  /// **'Manage multiple connected servers. Server system settings are only available from each server detail page.'**
  String get settingsServerManagementSubtitle;

  /// No description provided for @themeSystem.
  ///
  /// In en, this message translates to:
  /// **'Follow system'**
  String get themeSystem;

  /// No description provided for @themeLight.
  ///
  /// In en, this message translates to:
  /// **'Light'**
  String get themeLight;

  /// No description provided for @themeDark.
  ///
  /// In en, this message translates to:
  /// **'Dark'**
  String get themeDark;

  /// No description provided for @authLogin.
  ///
  /// In en, this message translates to:
  /// **'Login'**
  String get authLogin;

  /// No description provided for @notFoundTitle.
  ///
  /// In en, this message translates to:
  /// **'Page not found'**
  String get notFoundTitle;

  /// No description provided for @notFoundDesc.
  ///
  /// In en, this message translates to:
  /// **'The requested page does not exist.'**
  String get notFoundDesc;

  /// No description provided for @monitorSettings.
  ///
  /// In en, this message translates to:
  /// **'Monitor Settings'**
  String get monitorSettings;

  /// No description provided for @monitorInterval.
  ///
  /// In en, this message translates to:
  /// **'Monitor Interval'**
  String get monitorInterval;

  /// No description provided for @monitorIntervalUnit.
  ///
  /// In en, this message translates to:
  /// **'seconds'**
  String get monitorIntervalUnit;

  /// No description provided for @monitorRetentionUnit.
  ///
  /// In en, this message translates to:
  /// **'days'**
  String get monitorRetentionUnit;

  /// No description provided for @monitorCleanData.
  ///
  /// In en, this message translates to:
  /// **'Clean Monitor Data'**
  String get monitorCleanData;

  /// No description provided for @monitorCleanConfirm.
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to clean all monitor data? This action cannot be undone.'**
  String get monitorCleanConfirm;

  /// No description provided for @monitorCleanSuccess.
  ///
  /// In en, this message translates to:
  /// **'Monitor data cleaned successfully'**
  String get monitorCleanSuccess;

  /// No description provided for @monitorCleanFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to clean monitor data'**
  String get monitorCleanFailed;

  /// No description provided for @monitorSettingsSaved.
  ///
  /// In en, this message translates to:
  /// **'Settings saved successfully'**
  String get monitorSettingsSaved;

  /// No description provided for @monitorSettingsFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to save settings'**
  String get monitorSettingsFailed;

  /// No description provided for @monitorSettingsTitle.
  ///
  /// In en, this message translates to:
  /// **'Monitor Settings'**
  String get monitorSettingsTitle;

  /// No description provided for @monitorSettingsInterval.
  ///
  /// In en, this message translates to:
  /// **'Monitor Interval'**
  String get monitorSettingsInterval;

  /// No description provided for @monitorSettingsStoreDays.
  ///
  /// In en, this message translates to:
  /// **'Data Retention Days'**
  String get monitorSettingsStoreDays;

  /// No description provided for @monitorSettingsEnable.
  ///
  /// In en, this message translates to:
  /// **'Enable Monitoring'**
  String get monitorSettingsEnable;

  /// No description provided for @commonError.
  ///
  /// In en, this message translates to:
  /// **'Error occurred'**
  String get commonError;

  /// No description provided for @env.
  ///
  /// In en, this message translates to:
  /// **'Environment'**
  String get env;


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
      <String>['en', 'zh'].contains(locale.languageCode);

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
