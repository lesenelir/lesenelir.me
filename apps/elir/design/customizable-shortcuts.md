# 全部快捷键与透明浮窗设计（PR #10）

对应 [PR #10](https://github.com/lesenelir/lesenelir.me/pull/10)，分支 `codex/customizable-shortcuts`。更新：2026-10-18。依赖 [PR #9](windows-native-screenshot.md)；三条 PR 均以 master 为目标，前置 PR 未合并时后续 diff 会含前置改动。快捷键以 `fcb23a7` 为基线，功能提交 `c72cdfb`；透明浮窗是当前分支后续修复，不另建第四篇文档。

## 范围与调用边界

同一套配置、匹配和录入覆盖 Windows/macOS 十项业务快捷键；截图内部的窗口键和取消键只在 Windows 可改，macOS 截图由系统 `screencapture -i` 处理这两个键（`shortcut.Fixed`）。保留功能页、语音控制器、文字处理与图库；不引入热键库、宏、多段键、额外手势或配置方案系统。

```text
Windows hook / macOS EventTap → 物理键归一化
  → app.Shortcuts / shortcut.Router（录入、截图所有权、普通动作）
  → voice.Controller.Handle / rewrite.start、stop / Screenshot.take
```

app 创建唯一 EventTap，经 Host 提供 Shortcuts。internal/shortcut 只管键位、配置、匹配／校验和事件归属，不导入功能。旧检测器／广播器已删，业务沿原控制器，跨线程可用状态使用已有原子标志。

UI 状态只在主线程读写，经 Window.Update 发布工作结果；使用稳定控件身份和普通状态／Handle，不存过期 Element，不在视图构建时写盘。

## 动作与配置

| 动作 ID | Windows 默认 | macOS 默认 | 范围 |
| --- | --- | --- | --- |
| voice.dictate | F8 | Fn／Globe | 听写开始／结束，释放触发 |
| voice.translate | 左 Shift+F8 | 左 Shift+Fn | 翻译开始／结束，保留语音手势 |
| voice.cancel | Esc | Esc | 语音可取消时 |
| voice.cycleTarget | Tab | Tab | 翻译录音中 |
| rewrite.translate | Alt+T | Option+T | 选中文本翻译 |
| rewrite.polish | Alt+R | Option+R | 选中文本润色 |
| rewrite.cancel | Esc | Esc | 改写进行中 |
| screenshot.capture | Ctrl+Shift+2 | Command+Shift+2 | 启动截图 |
| screenshot.window | Space | Space（固定） | Windows 立即选窗；macOS 由系统截图切换窗口模式 |
| screenshot.cancel | Esc | Esc（固定） | 截图选择中 |

存稳定物理键 ID（KeyT、Digit2、F8、Fn），不存显示字符串／混合平台数字。修饰键支持通用 Ctrl／Alt／Shift／Meta 和左右限定，主键盘与数字键盘分开。Windows F8 为真实 F8，macOS Fn／Globe 合并；硬件未上报的 Fn 无法录入。

复用 settings.json 的 shortcuts section，version=1，windows／darwin 各存 Binding map；Binding 为 key 与 modifiers。缺失动作补当前平台默认，另一平台保留。JSON／版本／键位／冲突非法时报告并整套回退当前平台默认，不覆写原文件；不逐条回退制造新冲突。

模拟复制／粘贴协议和外部 Ctrl/Cmd+V 提示不属于业务键，系统文本控件快捷键也不在范围内。不为取消动作增加禁用／清空选项。

## 匹配、冲突与按键归属

- 精确匹配修饰集合，Ctrl+Alt+T 不算 Alt+T。裸 F8 与左 Shift+F8 可共存；通用 Shift+F8 和左 Shift+F8 匹配重叠，拒绝。
- 语音保留释放触发、左右 Shift 和 dirty 手势；其他动作首次按下触发，长按不重复执行业务。
- 同时可用动作整表校验，不限页面。语音／改写可并发，仅 cancel 可共享键，取消所有匹配且可取消的任务。
- 录入／截图临时拥有输入，不触发其他业务；截图取消仅取消截图，退出恢复路由。
- 已消费 KeyDown 的 KeyUp 必须由原所有者消费。改配置、任务结束、关闭录入、监听重启不丢释放归属；未消费按下不能因配置变化吞掉释放。
- 切配置取消未完成手势，等旧按键释放再开始。hook 停止／恢复时同步可获知的修饰状态，不把恢复当按键。
- Windows 忽略 injected events，macOS 忽略 syntheticTag；事件回调不写盘、起进程或等网络。
- 应用内冲突和平台保留／不支持组合明确拒绝，不保证完整探测其他应用占用。全局裸输入字符在 UI 提示会占用其他应用中的该键。

## 录入与成功后发布

点击原页面键帽 → 录入完整组合 → 候选／校验 → 保存或取消，可选左右修饰。Esc／Tab／Enter 都可录入，提供鼠标取消；失焦、隐藏／关闭设置、切页结束录入，丢弃未保存候选。

恢复默认只替换当前候选，仍需整表校验与显式保存，不覆盖其他动作。保存期间禁止重复提交：

```text
验证候选整表 → 后台 Store.SetRaw 写盘
  → Window.Update → 成功才发布配置、刷新绑定／界面
```

Store.SetRaw 用候选副本，持久化成功才发布 sections；Section.Update 成功才发布 value，Raw 返回独立副本，map/slice 深拷贝。失败保留旧配置并显示错误，后续保存其他 section 不得带上失败值。页面／浮层／中英文说明从有效绑定格式化，长键名不截断到不可辨认。

## 截图局部输入

Windows 保留 [PR #9 原生 GDI](windows-native-screenshot.md)。WithScreenshotShortcuts 经 context 传不可变快照；窗口按同一规则处理窗口／取消动作，由 Router.Claim 认领释放。全局路由在截图时透传本地输入，不启动语音／改写；DPI、文件、安全加载和清理契约不变，没有 helper 配置／环境通道。

macOS 继续运行系统的 `/usr/sbin/screencapture -i`：拖选、Space 切换窗口模式（带阴影、可截被遮挡的窗口）、拖动中移动选区和 Esc 取消都由系统提供，按下快捷键即出现，无需先抓屏解码。因此 macOS 上只有触发截图的快捷键可改；`Config.Bindings` 忽略 darwin 下保存的窗口键和取消键，设置页对它们显示固定键帽。截图进行中 Router 透传按键，交给系统截图。

曾尝试过用自绘选择层替换系统截图，以支持在 macOS 改这两个键；代价是遮罩出现前要抓取并解码整张 Retina 快照（数百毫秒以上），且失去窗口阴影、遮挡内容和移动选区，已撤回。

## Windows 透明浮窗后续修复

语音条、翻译／润色条、截图缩略图原先使用 Transparent Panel。MyGo v0.5.5 Windows 仅系统材质／Flyout 路径启用 noRedirect；普通 Panel 将胶囊外透明区显示成黑底，离屏 Tester alpha 检查无法发现。

共享 `widget.NewOverlay` 统一三处：Windows 使用托盘关联、初始隐藏、不聚焦、无系统阴影的 Flyout，走 DirectComposition；macOS 保留透明 Panel。app 在功能启动前建立 Tray，经 Host 提供。调用者仍控制 SetPosition、SetSize、ShowInactive、Hide；托盘关联避免受设置窗口移动／最小化影响，不依赖 Windows 11 系统材质。

内容自己画圆角／阴影，点击沿原视图。托盘创建失败仍沿日志错误路径使用旧 Panel，该降级仍有黑底限制。依赖、录音流程、设置格式不变；待 MyGo 普通透明 Panel 修复可移除此有注释兼容路径。

## 实际验证与交付

快捷键合入 fcb23a7 后已执行 Windows 全量测试、MyGo vet、普通构建，以及公共路由／Store／录入 UI／语音控制器适用的 race 测试。macOS arm64/amd64 正式代码与 app、voice、rewrite、screenshot、platform/darwin 测试二进制编译通过；ui.NewTester 渲染检查中文录入和长键名。分支整理仅核对源码一致，不当作重跑所有测试。

| 回归 | 覆盖 |
| --- | --- |
| 路由 | 十项改绑、旧键失效、左右修饰／数字键盘、按序／污染／重复、三类取消和输入归属 |
| 持久化 | 缺省／损坏／版本、跨平台保留、副本、写盘失败及后续保存不泄漏 |
| 录入 UI | Esc／Tab／Enter、失焦／切页、默认冲突、错误／长名、多帧生命周期 |
| 截图 | 局部快照／释放认领、负坐标／混合缩放合成裁剪，不替代硬件证据 |

本次浮窗修复已执行：

- Windows `CGO_ENABLED=0 go test ./...`、`go tool mygo vet .` 通过；正常权限重跑解除沙箱 socket／rename 限制，未放宽断言。
- `ELIR_OVERLAY_DESKTOP_TEST=1 go test ./internal/widget -run '^TestOverlayDesktop$' -count=1 -v -timeout=45s` 通过；加 `MYGO_GPU=0` 同样通过，覆盖默认 GPU 和软件绘制呈现。
- Windows 11（内核 10.0 build 26300）桌面取色：640×76 → 232×172 → 640×76 三次，角落透出后方 #2a699a，中央为 #141416；检查透明合成／不激活／工具窗样式、置顶、无系统阴影、初始隐藏、位置尺寸、隐藏重显和前台保持。
- macOS arm64/amd64 `CGO_ENABLED=0 go build ./...` 通过，未运行 macOS 测试。
- MyGo Windows amd64 打包通过。最新试用包为 `build/transparent-overlays-trial-20261018/windows-amd64/` 的 Elir.exe 和 Elir Setup 0.3.0 amd64.exe，版本仍为 0.3.0，未签名／发布。用户反馈可用并授权提交；运行前先从托盘退出旧实例。

## 未验收与下一步

无 macOS 真机：Fn／Globe、非美式键盘、权限恢复待验。Windows 真实终端／高权限目标物理快捷键、完整录音上传、缩略图拖放、Windows 10 真机透明合成未全面验收。用户试用反馈不替代逐项验收。

后续按三篇维护：平台与安全归 PR #8，Windows 捕获／文件归 PR #9，按键及本次浮窗归本文；更新对应正文和验证记录，不再为小修复增文档。
