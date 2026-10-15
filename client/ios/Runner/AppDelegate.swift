import UIKit
import Flutter

// 标准 Flutter iOS 入口。
// 说明: 旧版本曾在此文件内实现 SwiftUI 混合生命周期(@main RunnerApp +
// 原生 TabView 界面), 因 Dart 侧 getUIRenderMode handler 已随功能裁剪删除,
// SwiftUI 启动画面(Color.white)永远等不到回调, 造成启动白屏;
// 且原生界面(Server/Files/Containers/Apps/Websites)对应的
// 服务端功能均已删除。整体移除, 恢复标准 FlutterAppDelegate。
@main
@objc class AppDelegate: FlutterAppDelegate {
  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    GeneratedPluginRegistrant.register(with: self)
    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }
}
