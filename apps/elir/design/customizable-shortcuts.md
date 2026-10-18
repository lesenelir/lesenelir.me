# Elir 全部快捷键自定义方案

日期：2026-10-18。状态：实现和自动化验证完成；macOS 真机及物理快捷键交互尚未验收。

实现先在独立 worktree 完成，从 `e102493` 起步并合入截图坐标修复 `fcb23a7`。按评审边界现已整理到独立分支 `codex/customizable-shortcuts`，基线为截图分支的 `fcb23a7`；PR 只展示快捷键差异，截图分支及其上游 PR 不包含快捷键提交。临时 worktree、编译／预览产物已清理。

当前实现记录（2026-10-18）：

- 已完成 10 个动作的默认值、左右修饰、冲突校验、录入候选、保存／恢复默认和统一路由；原硬编码检测器已迁移。
- 已完成 Store 成功落盘后发布内存、失败不泄漏到后续保存的修复及测试。
- Windows 已使用原生 GDI 选择层，方案第 6 节的 PowerShell helper 方案被当前代码替代；本实现只接入可配置局部按键。
- macOS 选择层通过固定系统 screencapture 的非交互快照加 MyGo 遮罩实现，窗口几何来自 CGWindowListCopyWindowInfo；没有模拟按键。双平台运行细节及最终检查结果在完成时补记。
- 已执行：`go test ./...`（Windows，全部通过）；`GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ./...`（通过）。
- 已完成录入 UI、全部动作改绑、截图局部键与负坐标／混合缩放裁剪回归、MyGo vet、适用模块 race、Windows 可执行文件构建、macOS 双架构及受影响测试编译，并已整理为独立 PR 分支。macOS 真机不可用，不能把交叉编译记为真机通过。

## 1. 目标和范围

用户要求项目内所有按键操作可自定义，不限截图中听写页面的两个键。覆盖 Elir 定义的全局触发键、任务取消键、任务中的切换键，以及截图选择器内部按键；Windows 与 macOS 都要完成。

保留各功能原有页面，键帽变成录入按钮；同一配置驱动事件匹配、设置页、浮层、说明文字。保存后生效，重启保留，支持单项恢复默认。

当前找到 10 个可配置动作：

| 动作 ID（建议） | 用途 | Windows 默认 | macOS 默认 | 生效范围 |
| --- | --- | --- | --- | --- |
| `voice.dictate` | 开始听写／结束当前录音 | F8 | Fn／Globe | 全局，沿用控制器状态规则 |
| `voice.translate` | 开始翻译／结束当前录音 | 左 Shift+F8 | 左 Shift+Fn | 全局，独立配置，不从听写键派生 |
| `voice.cancel` | 取消录音、转写或整理 | Esc | Esc | 语音任务可取消时 |
| `voice.cycleTarget` | 切换目标语言 | Tab | Tab | 翻译录音中 |
| `rewrite.translate` | 翻译选中文字 | Alt+T | Option+T | 全局 |
| `rewrite.polish` | 润色选中文字 | Alt+R | Option+R | 全局 |
| `rewrite.cancel` | 取消文字处理 | Esc | Esc | 文字处理进行中 |
| `screenshot.capture` | 开始截图 | Ctrl+Shift+2 | Command+Shift+2 | 全局 |
| `screenshot.window` | 选择／截取指针下窗口 | Space | Space | 截图选择器内，保留各平台既有交互 |
| `screenshot.cancel` | 取消截图 | Esc | Esc | 截图选择器内 |

三个取消动作分别设置，允许默认都用 Esc，不能为了省实现强制绑定成一项。

范围辨别：截图页的 Ctrl/Cmd+V、终端粘贴组合是“目标应用如何粘贴”的说明，不是 Elir 注册的动作；Elir 内部模拟复制／粘贴是与目标应用交互的协议，不能跟随用户配置任意替换。系统／MyGo 文本控件的选中、输入、Tab 焦点移动也不是业务快捷键。保留这些行为，并把外部粘贴说明与可编辑的 Elir 快捷键清楚区分。此次检索未发现托盘菜单自定义 accelerator，不新增退出、打开设置等未有的快捷键。

## 2. 当前实现与调用边界

所有路径相对于 `apps/elir`。

- `internal/app/app.go` 创建唯一 `platform.EventTap`，由 `app.Shortcuts` 持有配置及录入 UI，委托 `internal/shortcut.Router` 处理事件。各功能通过 `Host.Shortcuts.Register` 注册动作和有效状态；旧广播器和分散检测器已删除。
- `internal/shortcut` 统一 10 项默认绑定、稳定物理键 ID、精确修饰匹配、通用／左右限定冲突校验、释放触发手势、输入所有权和配置切换后的按键释放。Windows F8 是实际 F8；macOS Fn／Globe 归一化为 Fn。
- Windows 通过 VK／scan／extended 标志区分物理键位、主键盘与数字键盘，忽略 injected events；macOS 保持 EventTap 与 syntheticTag 过滤，并在监听重启／停用时取消未完成手势。
- 语音动作仍交给原 `controller.Controller.Handle`；Capture 回调只发布原子可取消／可切换状态。文字处理仍使用原 start／stop；截图仍经原图片库、剪贴板和缩略图流程。
- `Store.SetRaw` 基于候选副本写盘，成功后才发布 sections；`Section.Update` 同样成功才发布 value，Raw 返回独立副本。失败路径测试证明其他 section 的后续保存不会带上失败值。
- 三处功能页面保留原布局，键帽变为录入／保存／取消／恢复默认控件，并可选择是否区分左右修饰。错误、长组合键、页面切换和监听重启有回归验证。浮层与说明读取当前绑定；外部应用粘贴协议不改。
- Windows 截图保持现有 GDI 原生选择层；macOS 使用非交互系统快照加 MyGo 自有选择层和窗口几何快照。局部按键均取任务开始时绑定快照，按键释放由原输入所有者继续消费。

第 3～8 节保留设计契约与验收要求；已执行检查和未实现的系统选择器细节见第 9 节。交叉编译不等于 macOS 真机验收。

## 3. 推荐结构：一份配置、一套匹配、原有业务动作

新增一个小型 `internal/shortcut` 包，集中放稳定动作 ID、绑定模型、默认值、校验、格式化、事件匹配与按键录入状态。不要引入完整插件注册框架、宏引擎或另一个全局键盘监听库。

事件路径：

```text
Windows hook / macOS EventTap
    → 归一化 KeyEvent
    → shortcut 路由（录入优先，其次截图会话，再到正常动作）
    → voice.Controller.Handle / rewrite.start、stop / screenshot.take
```

截图独立 helper 使用同一绑定快照和匹配规则的等价实现；全局路由在 helper 拥有输入期间不抢走截图内部键。一个物理事件只交给一个输入所有者。

由 app 持有 manager，经 Host 提供给功能和编辑控件。功能提供短小动作回调和有效状态；不要让快捷键包导入 voice/rewrite/screenshot，避免环形依赖。语音仍由已有控制器决定开始、结束、取消，不复制录音状态机。

现有 voice 的 Action 类型可继续保留，manager 的动作通过适配回调映射；Fn 手势已迁到公共匹配器并迁移测试，原 Detector 已删除。`Host.HandleKeys` 和旧广播器若不再有调用者则删除，不能同时保留两条触发路径。

### 3.1 绑定和按键模型

- 使用稳定键 ID，例如 `KeyT`、`Digit2`、`F8`、`Escape`、`Tab`、`Space`、`Fn`；不保存显示字符串或混合平台原始数字。
- 普通字母键按物理键位识别，保持现有快捷键习惯；界面明确键位语义。区分主键盘／数字键盘，覆盖字母、数字、标点、方向、导航、编辑键、功能键。
- 一个绑定为一个主键和零到多个修饰键；修饰键支持通用和左／右限定。通用 Ctrl 与左 Ctrl 的匹配集合重叠，冲突判断不能仅比较 JSON 是否相等。
- Fn／Globe 在 macOS 归为同一个可识别按键；Windows 的 F8 是真实 F8，不再伪装 Fn。硬件未向操作系统报告的 Fn 无法录入；不把它静默转成另一个键。
- 保留 macOS Fn 单键及 Fn+左 Shift；本次不增加其他“仅 Ctrl/Alt/Shift 单独轻点”的手势、连续多段快捷键、双击、长按或宏。
- 触发时机是动作策略，不作为额外用户选项：语音启动／结束保留释放主键触发，其他动作保留首次按下触发。长按自动重复不重复执行业务。

建议存储示例（只展示部分动作，实际必须覆盖全部 10 项）：

```json
{
  "shortcuts": {
    "version": 1,
    "windows": {
      "voice.dictate": {"key": "F8", "modifiers": []},
      "voice.translate": {"key": "F8", "modifiers": ["LeftShift"]},
      "voice.cancel": {"key": "Escape", "modifiers": []}
    },
    "darwin": {
      "voice.dictate": {"key": "Fn", "modifiers": []}
    }
  }
}
```

缺失动作使用当前平台默认值；另一平台配置原样保留。默认值显式按平台定义，不靠 UI 字符替换。不加导入导出、云同步、配置方案切换；本次不提供清空／禁用快捷键，避免额外引入取消任务无出口的问题。

### 3.2 匹配和消费契约

1. 先更新按下键／修饰键状态，再匹配当前可用动作；CapsLock 等无关标志不参与快捷键比较。
2. 只匹配指定组合，不把 Ctrl+Alt+T 当成 Alt+T。Fn 手势继承现有 dirty 规则；F8 与左 Shift+F8 保留先后按序兼容，单次手势只产出一个动作。
3. 记录本次按键序列的所有权。即使任务结束、切换配置，已经吞掉 KeyDown 的 KeyUp 仍由原所有者消费；未消费的 KeyDown 不得仅因配置变化而吞掉 KeyUp。
4. 保存配置不立刻抹掉 held/swallowed 状态：取消未完成手势的业务触发，等待旧按键释放，再使用新配置匹配下一次完整按键。
5. 输入回调不得写盘、启动进程或等待网络；业务回调沿用异步处理。跨 goroutine 的状态使用原有原子标志或明确同步的快照。
6. Windows 继续忽略 injected events，macOS 继续忽略本应用 syntheticTag，保证模拟复制／粘贴不会递归触发新快捷键。
7. hook 停止、系统禁用监听、睡眠恢复时清理未完成手势；恢复后同步仍按住的修饰键，不把恢复本身当一次按键。无法获知的按键状态等待新一轮按下确认。

## 4. 冲突与作用域

作用域最少保留：正常全局、语音可取消、翻译录音、改写进行中、截图会话、快捷键录入。

- 同时可能生效的两个普通动作不能绑定同一匹配集合；全局动作与局部动作也要校验，不能仅检查同页重复。
- 裸 F8 与左 Shift+F8 通过精确修饰条件区分，允许共存；通用 Shift+F8 与左 Shift+F8 重叠，需要报冲突。
- 语音与改写可以并发，不能假定二者互斥。为保留默认 Esc 行为，仅允许多个 cancel 动作共享键：正常模式下取消所有匹配且可取消的任务；不同取消键只影响对应功能。
- 截图选择期间由截图输入所有者独占快捷键，不触发语音／改写；截图取消仅取消截图。其他任务不被取消或重置，选择结束后恢复匹配。
- 录入期间暂停业务匹配，但继续跟踪键状态；退出录入后等本次按键释放，防止刚录入的键立即触发业务。
- 应用内冲突可确定校验；系统保留组合／平台不支持的键明确拒绝并解释原因；与其他应用的全局快捷键冲突不保证全部探测到，不显示“绝对无冲突”。
- 全局绑定裸字母、数字等可输入字符时，保存前明确提示该键会在其他应用中被 Elir 占用；局部 Space/Tab/Esc 在相应状态外必须透传。

## 5. 设置交互和持久化

点击现有键帽 → 显示“请按快捷键” → 完整录入一次按键组合 → 显示候选值和校验结果 → 点击保存或取消。保留明确的鼠标取消按钮；录入态的 Esc、Tab、Enter 都可以作为候选，不硬编码成退出／跳焦／保存键。录入前后的控件仍支持正常键盘访问和可访问名称。

失焦、关闭／隐藏窗口或切换页面结束录入并丢弃未保存候选。单项“恢复默认”同样经过整表冲突校验；默认键被占用时指出冲突动作，不覆盖其他设置。保存期间禁用重复提交，使用版本号或串行写入避免旧结果覆盖新候选。

保存过程：校验候选整表 → 后台写入 settings.json → 主线程发布有效配置和刷新界面。失败显示错误并保留旧的有效配置，不提示成功。

复用 Store，而不是另建配置文件。需要小范围修正 `Store.SetRaw` 与 `Section.Update` 的提交顺序：基于候选副本序列化、持久化成功后才发布内存值；失败保持原值，保留其他 section。map/slice 必须深拷贝。新增失败路径测试，证明后续保存其他 section 不会意外带上失败的快捷键配置。

旧配置缺少 shortcuts section 自动采用默认值，不做侵入式迁移。非法键／错误版本／冲突配置必须在加载时校验，向用户报告并使用安全的整套当前平台默认值；不自动覆写原文件，方便修复。仅“逐条回退默认”可能制造新的交叉冲突，不能直接使用。

所有界面通过绑定模型格式化；`Keycaps` 的通用绘制可复用，但取消针对业务键的 `fn→F8` 等替换。浮层长组合键不能截断到无法识别，必要时调整宽度。中英文说明中的 Tab／Esc 改为参数或“当前切换／取消快捷键”。外部应用粘贴提示单独按平台输出。

MyGo 仍为当前固定的 v0.5.5；实现 UI 前按 `.agents/skills/mygo-maintenance/SKILL.md` 核实本地 API。控件使用稳定身份，状态存普通数据／Handle，不保存过期 Element，不在构建视图时执行持久化。

## 6. 两个平台的关键改造

### Windows

保留 `WH_KEYBOARD_LL`，去掉 F8→Fn 特判。VK／scan／extended 标志归一化物理键位和数字键盘，保存左右修饰身份；F8 改绑后不再占用。

当前截图后端已是原生 Win32/GDI，不再有 PowerShell helper 或配置数据跨进程通道。Go context 传入不可变绑定快照；选择窗口首次 KeyDown 按同一规则匹配窗口／取消动作，旧默认键失效。全局路由在截图活动时透传本地输入，选择窗口认领已消费的键以配对全局 KeyUp，避免关闭后漏到原应用。保留原生句柄清理、可信 DLL、DPI 检查、显示器变化取消和 PNG 写入契约。

### macOS

沿用 EventTap，平台层只做事件归一化。Fn 的特殊手势保留在公共匹配器；不通过新全局热键库替换现有监听，以免丢失 Fn 和吞键控制。

截图内部 Space/Esc 是本次最大的工程点。实现采用 Elir 自有最小截图选择层，以控制所有局部快捷键；下列为原设计要求，未验收差异见第 9 节：

1. 在显示遮罩前获取显示器图像与窗口几何快照；原生能力通过 platform 边界提供，先核实固定 MyGo 版本已有能力，不凭空调用未确认 API。
2. 用现有 MyGo 窗口／绘制能力做区域拖选和窗口选择；仅接入本次两个局部动作，不扩展标注、录屏等功能。
3. macOS 保持现有窗口选择习惯：窗口键切换窗口模式，再点击确认；窗口命中排除 Elir 选择层，按显示次序取最前可选窗口。
4. 输出仍为原 `capture(ctx, path)` 等价契约：成功产出 PNG、取消不留文件、失败返回错误；后续图片库、剪贴板和缩略图流程不改。增加 bindings 参数及必要的平台封装即可。
5. 支持多显示器、负坐标、Retina／不同缩放映射，裁剪不包含自身遮罩；取消、权限失败和退出都关闭遮罩、释放位图并恢复快捷键作用域。

这一步先做独立可行性验证，确认现有 MyGo 加平台原生能力能完成区域／窗口选择，再接正式流程。旧系统选择器隐含的移动选区、窗口阴影等行为要做一次交互对照并记录差异，不能把丢失行为当成快捷键改造的默认结果。

不把“给系统 screencapture 注入 Space/Esc”作为默认方案：输入可能落到错误前台、受物理修饰键影响，且原始 Space/Esc 仍可能生效；需要额外处理生命周期、成对事件和默认键屏蔽，不能凭代码静态检查就认定完整支持。若实现 agent 提出该更小方案，应先给出 macOS 真机证据证明以上条件；未通过时执行自有选择层方案。只让 Windows 截图内部按键可配不能算本需求完成。

## 7. 改动清单与实施顺序

| 步骤 | 文件／范围 | 完成条件 |
| --- | --- | --- |
| 1 | 新 `internal/shortcut`、相关测试 | 10 个动作、平台默认、配置模型、匹配／冲突／Fn 测试成立 |
| 2 | `internal/platform/keyboard_windows.go`、`darwin/eventtap.go`、`platform_darwin.go`，新键码映射文件 | 两端向公共匹配器提供一致事件；模拟复制／粘贴不回流 |
| 3 | `internal/app/{app,feature,keys}.go`，voice/rewrite/screenshot 的 Start 和旧 shortcut 适配 | 统一事件所有者，业务行为沿用旧控制器，无双重触发 |
| 4 | `internal/store/store.go` 和失败路径测试 | 保存成功才发布，保存失败不污染内存或后续落盘 |
| 5 | 新录入控件；三处 page、voice overlay、rewrite hud、`widget.go`、语言文件 | 每个键可编辑，提示与生效值一致，重启可恢复 |
| 6 | `screenshot/capture_windows.{go,ps1}`、`capture.go`、截图会话管理和 macOS 选择层 | 截图内按键也生效，退出恢复，不损坏原截图结果流程 |
| 7 | 旧快捷键／页面测试、截图测试、README／CLAUDE 中快捷键描述、生成 locales | 完成回归与双平台真机验收，记录实际证据 |

第 6 步 macOS 可行性验证应在第 1 步前或早期开展，尽早暴露主要风险。其余按依赖顺序推进，不需要同时重构语音流水线、密钥存储或截图历史管理。生成的 `locales/catalog_gen.go` 使用 `go tool mygo generate`，不手改。

## 8. 验收要求

自动测试必须包含：

- 无新配置时恢复全部默认行为；10 个动作逐项改绑，旧键失效，新键仅在对应上下文生效；重启仍有效。
- F8／左 Shift+F8、Fn／左 Shift+Fn 的两种按下顺序；右 Shift 不误当左 Shift；长按一次触发、额外键污染、修饰先松开。
- 左右修饰限定、通用修饰重叠、裸键与组合键、字母／数字键盘映射；不匹配的普通输入透传。
- 语音与改写并发时取消键共享／不同的行为；录入和截图会话不触发其他业务。
- 改配置、关闭录入、任务提前结束、监听重启时按键仍按住，KeyDown/KeyUp 的消费一致，无残留修饰状态。
- 保存失败、配置损坏、未知版本、加载后冲突、另一平台字段保留、含 map/slice 的副本隔离。
- UI 录入 Esc/Tab/Enter、失焦退出、恢复默认冲突、保存错误、长键名布局；额外帧和控件消失／重现不产生 MyGo 生命周期问题。
- Windows helper 数据白名单、继承环境隔离、模拟事件不递归；两端截图取消不入库，成功仍复制并生成缩略图。

实现后的检查：在 `apps/elir` 跑受影响测试、`go tool mygo vet .`、`CGO_ENABLED=0 go test ./...`（按宿主 shell 设置环境）；跨平台编译只证明编译，macOS build tags 下的测试应在 macOS 执行。发生共享状态改动时补适用的 race 检查。

真机至少覆盖 Windows 普通应用／终端／不同权限窗口，以及 macOS Fn／Globe、非美式键盘、辅助功能权限丢失恢复、截图多屏和不同缩放。系统无法拦截的保留键不可在 UI 中宣称可用。没有 macOS 设备时，应明确列为未验收，不能用 Windows 成功替代。

## 9. 开源参考与当前进度

已查阅以下项目，参考交互和责任拆分，不作为直接引入依赖的决定：

- [KeyboardShortcuts](https://github.com/sindresorhus/KeyboardShortcuts)：动作名称、快捷键录入控件、持久化和系统／菜单冲突提示；它是 Swift/macOS 方案，不能直接覆盖现有 Go 双平台路径。
- [Flameshot 快捷键设置](https://github.com/flameshot-org/flameshot/blob/master/src/config/shortcutswidget.cpp) 与 [截图选择器](https://github.com/flameshot-org/flameshot/blob/master/src/widgets/capture/capturewidget.cpp)：参考截图内部动作的可配置设计；不迁入其整个 Qt 架构。

实施结果：10 项动作、键位归一化、录入及左右限定选项、冲突校验、配置发布、原有页面和浮层、两平台截图局部键均已实现。旧检测器／广播器已移除，回归案例迁入公共路由测试。

实际验证：Windows `CGO_ENABLED=0 go test ./...`、`go tool mygo vet .`、普通可执行文件构建已通过；公共路由／Store／录入 UI／语音控制器 `go test -race` 已通过；macOS arm64/amd64 正式代码与 app、voice、rewrite、screenshot、platform/darwin 测试编译已通过。中文录入 UI 与长组合键已用 ui.NewTester 渲染查看。上述检查均在合入 `fcb23a7` 后执行；拆分为独立分支时核对源码与已验证版本完全一致，仅调整提交组织和本记录，没有重跑测试。未另做安装包或发布。

未验收：没有 macOS 真机；Fn/Globe 硬件、macOS 截图多显示器／Retina、权限丢失恢复，以及 Windows 终端／不同权限窗口的真人快捷键操作未验证。Windows 物理键位、数字键盘、左右修饰、输入所有权、三类取消、保存失败、所有动作重绑和多屏裁剪均有自动化回归，不以它们冒充真机证据。

macOS 与旧系统选择器的明确差异：自有选择层裁剪冻结时可见桌面，不会捕获被其他窗口遮挡的目标内容，不额外生成窗口阴影；暂不支持拖选时移动选区。窗口键仍切换窗口模式后点击确认。上述差异仅由实现对照得出，未作真机交互对照。临时快照保存在随机目录，成功／取消／失败均清理；无按键注入。窗口几何使用 [CGWindowListCopyWindowInfo](https://developer.apple.com/documentation/coregraphics/cgwindowlistcopywindowinfo(_:_:))，坐标契约见 [kCGWindowBounds](https://developer.apple.com/documentation/coregraphics/kcgwindowbounds)。
