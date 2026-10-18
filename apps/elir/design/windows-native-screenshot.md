# Windows 原生截图与选区设计

状态：按用户实现请求进入开发，自动验证与实机验收分别记录。日期：2026-10-18。

基线：`codex/windows-support`，`d24d9a2003a9ebc0ae3d0b96a08210af29402592`，MyGo `v0.5.5`。实现 agent 开始前须重新核对 HEAD、工作区改动及依赖；本文描述的“当前实现”仅对应此基线。

## 1. 决策与范围

保留现有截图业务流程，用进程内 Go/Win32 替换 PowerShell/WinForms：GDI 冻结虚拟桌面，原生无边框窗口完成选区，标准库编码 PNG，再走原有图库、剪贴板和缩略图流程。

本次完成四件事：

1. 截图路径不再启动 PowerShell，不再加载 .NET 或通过 `Add-Type` 动态编译。
2. 在截图线程内建立明确的 per-monitor DPI 上下文，捕获、鼠标、窗口边界、裁剪统一使用物理像素。
3. 窗口截图优先使用 DWM 可见边界，失败时受控回退到 `GetWindowRect`。
4. 明确取消、错误、文件提交和原生资源释放契约，并验证多屏交互。

不增加截图依赖、CGO、独立 helper、常驻捕获服务、后端选择设置或 PowerShell 回退。不做 WGC/DXGI、HDR 色调映射、滚动截图、标注、放大镜、鼠标指针合成、窗口阴影或遮挡窗口内容恢复。GDI 输出是桌面当时可见像素；受保护内容可能为空白，不能据此承诺完整捕获。

“逐步迁移”指分步开发和验证，最终只保留一个 Windows 实现。不能以“已完成 GDI 抓图但选区仍运行 PowerShell”作为本任务完成条件。

## 2. 当前调用链与必须保留的契约

```text
platform.EventTap → screenshot.shortcut.KeyDown
  → Screenshot.take（busy CAS，另起 goroutine）
  → Screenshot.capture
      → Library.NewPath
      → capture(context.Background(), path)
      → ReadFile → copyPNG（MyGo Clipboard / transfer.PNG）
      → Stat → thumbnail
  → Settings.Update → 更新图库列表、显示弹出缩略图
```

源码入口见 `../internal/screenshot/screenshot.go`、`shortcut.go`、`capture_windows.go`、`capture_windows.ps1`。macOS 由 `capture.go` 调用系统 `screencapture -i`，本次不动。

当前 Windows helper 先 `CopyFromScreen`，再打开覆盖虚拟桌面的 WinForms 窗口；拖动并松开左键确认，Esc 取消，Space 隐藏覆盖层后立即选择鼠标下的顶层窗口。没有“先空格切换模式，再点击确认”的交互。

保留包内签名：

```go
func capture(ctx context.Context, path string) (bool, error)
```

| 结果 | 返回值 | 文件及上层行为 |
| --- | --- | --- |
| 用户确认且 PNG 已提交 | `true, nil` | `path` 是完整、可解码 PNG；上层继续复制与展示 |
| Esc、无效选区、用户关闭或切走选区 | `false, nil` | 不新增截图，不触碰剪贴板 |
| context 取消或超时 | `false, ctx.Err()` | 不提交本次文件；释放全部会话资源 |
| Win32、显示配置、编码或文件错误 | `false, error` | 带阶段信息；不留下本次不完整 PNG |

现有调用者传入 `context.Background()`，并未接入应用退出取消。底层仍必须支持可取消 context，便于测试和以后接入；本次不扩展应用生命周期。重复快捷键仍由 `Screenshot.busy` 拦截。

当前剪贴板写入失败会返回错误，但磁盘中已保存的 PNG 保留，稍后图库刷新可发现它；本次保留这一行为，不在捕获后处理失败时删除用户截图。截图是用户显式复制，不改变现有剪贴板历史策略。

## 3. 代码边界与文件清单

所有 Win32 截图细节放入已有的 `internal/platform`，复用该包的 `systemDLL`、`user32` 和原生调用习惯。无需导出 DLL 加载器或新建通用 Windows 框架。

Windows 专用入口建议为：

```go
// Windows only. Called from a worker goroutine, never MyGo's UI thread.
func CaptureScreenshot(ctx context.Context, path string) (bool, error)
```

`screenshot/capture_windows.go` 仅转发此入口。此函数不读图库、不访问剪贴板、不调用 MyGo UI；输出目录由调用者准备。无需为 Darwin 增加同名门面。

| 文件 | 预期改动 |
| --- | --- |
| `internal/screenshot/capture_windows.go` | 删除脚本嵌入与进程构造，保留原签名并转发 |
| `internal/screenshot/capture_windows.ps1` | 原生链路验收后删除 |
| `internal/screenshot/capture_windows_test.go` | 删除已失效的 helper 路径/环境/C# 编译测试；不保留无意义的转发测试 |
| `internal/platform/screenshot_windows.go`（新增） | 会话编排、DPI、GDI 捕获、裁剪和 PNG 提交 |
| `internal/platform/screenshot_overlay_windows.go`（新增） | 窗口类、消息循环、鼠标/键盘、窗口命中及边界 |
| `internal/platform/screenshot_windows_test.go`（新增） | 几何、像素、失败清理、取消与文件契约测试 |
| `internal/platform/system_windows.go` / 测试 | 原则上复用；仅在确有必要时最小调整，保持安全加载约束 |
| `internal/screenshot/screenshot.go`、图库、弹窗、快捷键及现有测试 | 调用兼容性复查；原则上不改业务逻辑 |
| `apps/elir/README.md` | 实现完成后删除 PowerShell/WinForms 依赖说明，准确描述 OS 能力边界 |
| `design/windows-support.md`、`design/windows-security.md` | 实现完成后追加迁移状态，保留原历史验证记录 |
| 本文 | 随实现更新进度、实际证据、剩余问题及下一步 |

表中 `internal/...` 路径相对 `apps/elir`。没有新增配置、导出到应用外部的 API、数据迁移、语言文本或测试 fixture 的需求；若实现必须增加可见错误提示，使用现有错误展示路径。

## 4. 线程、DPI 与坐标

### 4.1 专属会话线程

在现有截图工作 goroutine 内 `runtime.LockOSThread()`。DPI 设置、创建窗口、消息循环、GDI 操作及 HWND 销毁均在这条线程完成。不得把交互消息循环放在 MyGo 主线程，也不在窗口回调里同步等待 `mygo.RunOnMain`。

MyGo `v0.5.5` 的 `internal/windows/app.go` 在初始化中尝试设置 Per-Monitor V2；`cmd/mygo/winres.go` 的打包 manifest 也声明 PMv2，并带 `true/pm`。这不等于截图线程的 DPI 状态已经经过运行时验证，普通 `go build` 与打包产物也应分别检查。

截图线程在取得任何坐标/DC/窗口前调用 `SetThreadDpiAwarenessContext(PER_MONITOR_AWARE_V2)`，保存旧上下文；不支持 PMv2 时可尝试 `PER_MONITOR_AWARE`。成功后用查询 API 验证实际处于 per-monitor 模式。销毁窗口与释放 GDI 后恢复旧上下文，再解锁线程。恢复失败时不得把带有修改状态的线程重新投入线程池，须返回错误并让仍锁定的工作 goroutine 结束。

不得再调用 `SetProcessDPIAware`，也不得在截图时修改全进程 DPI。线程上下文 API 从 Windows 10 1607 提供，PMv2 从 1703 提供。[线程 DPI API](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setthreaddpiawarenesscontext)、[DPI 上下文版本](https://learn.microsoft.com/en-us/windows/win32/hidpi/dpi-awareness-context)。

旧版本兼容策略：API 不存在时，仅在查询证实进程已为 per-monitor aware、当前线程继承该状态时继续；不能证明则只让截图返回明确的不支持错误，应用仍可启动，不退回 system-aware 混合坐标。实现者需在 README 记录实际验证的最低版本，不得把“Windows 10/11”宣传为所有早期版本均已验证。动态查询可选入口，避免 `MustFindProc` 导致整个应用在旧系统启动失败。

### 4.2 唯一坐标约定

统一使用带符号的物理屏幕像素坐标及半开矩形 `[left, top, right, bottom)`：

- 枚举当前显示器矩形，计算并集 `V`；支持主屏左侧、上方的负坐标。
- 快照缓冲区原点是 `(0,0)`，尺寸为 `V.Dx() × V.Dy()`。
- 拖动使用 `GetMessage` 返回的 `MSG.pt`：它是消息投递时的 32 位屏幕坐标，避免 `lParam` 的 16 位截断，也避免窗口线程延后处理消息时 `GetCursorPos` 已到终点而丢失起点。Space 实时命中仍使用同一 DPI 上下文的 `GetCursorPos`。[MSG 坐标契约](https://learn.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-msg)。
- 屏幕矩形 `R` 先与 `V` 求交，裁剪坐标是 `R.Sub(V.Min)`；不得再次乘屏幕缩放率。
- 跨屏拖动按同一物理坐标计算，不经过 MyGo 逻辑点、DIP 或全局缩放比。
- 显示器布局中的空洞固定填黑、不读取未初始化内存；选区完全落在空洞中视为无效，跨空洞的矩形保留黑色部分。

单个覆盖窗口跨越 `V`，按原始像素 1:1 绘制冻结图；无需为每块屏幕创建交互窗口。若实测证明单窗口路径存在无法修正的 DPI 问题，再在本文记录证据并修订设计，不能用未经验证的二次缩放掩盖偏差。

## 5. GDI 捕获与内存

在覆盖窗口显示前完成捕获：取得桌面 DC，创建兼容内存 DC 和 top-down、32-bit、`BI_RGB` 的 DIB section，选入内存 DC。先清零，然后按显示器矩形把同一桌面 DC 中的可见像素复制到 `rect.Min - V.Min`，使用 `SRCCOPY | CAPTUREBLT`。任何显示器捕获失败则整次失败，不返回缺屏的成功结果。

仍然只有一份虚拟桌面 DIB；按显示器复制用于定义空洞，不引入多个独立后端。不同显示器之间不保证同一时刻原子采样，这是本期 GDI 路线的限制。[BitBlt 语义](https://learn.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-bitblt)。

读取 DIB 内存前完成 `GdiFlush` 同步。用同一份快照绘制选区背景，确认后仅把选区转换为 Go 图像；BGR 通道转为 RGB，输出 alpha 强制为 255，避免未定义的第四字节使 PNG 透明。不为 PNG 再复制一次完整桌面，也不在用户拖动期间重新抓屏。[DIB 访问同步](https://learn.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-createdibsection)。

宽高、stride、总字节数在分配前使用宽整数检查，拒绝零/负值、乘法溢出、超出 Win32/Go 可表示范围。建议本期采用内部 512 MiB 快照上限，非用户配置；超出返回可理解的尺寸错误，不能等原生越界或 Go OOM。该值是待实现验证的工程预算，并非测量结论；总峰值还包括选区图像及已有上层处理。

每个资源只有一个所有者；部分初始化失败也必须清理：

| 资源 | 释放方式 |
| --- | --- |
| `GetDC(0)` 桌面 DC | `ReleaseDC(0, dc)` |
| `CreateCompatibleDC` 内存 DC | `DeleteDC` |
| DIB、非 stock 的画笔等 GDI 对象 | 先恢复此前选入对象，再 `DeleteObject` |
| HWND、timer、鼠标捕获 | owner 线程 `KillTimer` / `ReleaseCapture` / `DestroyWindow` |
| Go 会话引用 | HWND 销毁后移除句柄到会话的注册关系 |

窗口类与 `syscall.NewCallback` 回调只注册一次，不能每次截图都创建永久回调。WndProc 的状态通过 Go 端受保护的注册表/会话标识访问；不要把临时 Go 对象指针长期存入 Windows。窗口类从创建期间到 `WM_NCDESTROY` 都需有合法处理路径；未知消息交给 `DefWindowProc`。

所有新增系统 DLL 仍经 `systemDLL` 使用真实 System32 绝对路径及 `LOAD_LIBRARY_SEARCH_SYSTEM32`，不引入工作目录、PATH 或环境变量搜索。可选 API 缺失返回功能错误，禁止为绕过加载失败回退到不受限加载。

## 6. 选区窗口与窗口截图

### 6.1 选区状态

创建无边框、置顶、不进任务栏且可接收键盘焦点的窗口，使用十字光标。首帧直接显示快照及选区边线，不新增蒙层效果或文字工具条。

状态只需要 `ready → dragging → accepted/cancelled/failed`；完成结果只能写入一次。进入窗口时主动检查激活/焦点是否成功；无法接收 Esc 的覆盖层不能留在屏幕上。

| 输入/事件 | 行为 |
| --- | --- |
| 左键按下 | 记录物理屏幕起点，捕获鼠标 |
| 拖动 | 归一化反向拖动矩形，重绘选区线；不修改快照 |
| 左键松开 | 用最终鼠标坐标重算；宽高均大于 1 像素且与显示器相交则确认，否则取消 |
| Space（非自动重复） | 立即按鼠标位置选择窗口，兼容现有 Windows 行为 |
| Esc / 关闭 | 取消 |
| 已激活后失焦、Alt+Tab、拖动中意外失去鼠标捕获 | 取消，不能留下置顶窗口或卡住 busy |
| 显示器布局/分辨率/DPI 变化 | 终止并返回“显示配置已变化，请重试”，不用旧快照配新坐标 |

主动隐藏/销毁、正常释放鼠标捕获引发的通知不能把已接受结果改写为取消；会话状态要区分自身收尾和意外事件。处理 `WM_PAINT` 必须配对 Begin/EndPaint；边线只画在窗口 DC，不烧进保存的图像。

消息循环检查 `GetMessage` 的错误和退出值。context 取消可用一个仅在交互期间存在的约 50 ms `WM_TIMER` 检查并唤醒收尾；不需要跨 goroutine 销毁 HWND 或轮询全桌面。检查 timer 创建失败，避免无法取消的会话。该间隔不是阻塞 Win32 调用期间的硬实时保证。

### 6.2 Space 的命中与边界

采用兼容现有行为的最小实现：记录指针屏幕坐标，标记正在完成窗口选择，隐藏覆盖窗口，再调用 `WindowFromPoint` → `GetAncestor(GA_ROOT)`。不重新抓屏，也不采用遍历所有窗口/缓存 Z 序的额外系统。无法得到有效窗口或与显示器无交集时按无效选区取消。

目标边界优先取 `DwmGetWindowAttribute(DWMWA_EXTENDED_FRAME_BOUNDS)`，检查 HRESULT、尺寸及交集。DWM 查询不可用或失败时，在相同 per-monitor 上下文中回退 `GetWindowRect`；该回退可能保留不可见边框，应在测试结果中标识，不能宣称等价于 DWM。微软明确说明 GetWindowRect 有 DPI 虚拟化且可能含不可见边框，而 DWM 扩展边界不做 DPI 调整。[边界文档](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getwindowrect)。

选取的仍是冻结桌面内相应矩形，包括当时的遮挡物，不是 `PrintWindow` 或独立窗口内容。圆角矩形外部保留桌面像素，不生成透明抠图。

已知兼容限制：从冻结到按 Space 之间，外部窗口若移动、关闭或改变层叠顺序，实时命中的边界可能与冻结图不一致；现有实现也有这一问题。本期保留该边界，测试与交付必须注明，不承诺“冻结窗口树”。若以后要求消除它，再单独设计与快照同步的窗口清单和命中模型。对应实现处用一行 `shortcut:` 注明此限制及升级条件。

## 7. 取消、文件提交与错误

交互结束后先销毁覆盖窗口、恢复桌面交互，再把选区转换为 Go 图像。DIB 在转换完成前保持有效；随后释放原生资源、恢复 DPI 并解锁线程，最后编码和提交 PNG。原生清理或 DPI 恢复失败时不进入文件提交，避免已经保存成功却返回捕获失败。所有提前返回路径也要执行对应的清理。

只对已确认选区写文件：在目标目录 `CreateTemp` 一个图库不会识别的临时文件，`png.Encode` 后检查写入和 Close 错误；提交前再次检查 context。Windows 使用不覆盖已有目标的同目录 rename/move 语义提交，必须明确验证目标已存在时失败，不能依赖跨平台 `os.Rename` 的“可能替换”行为。

成功移动为最终 `path` 是提交点：提交前取消则删临时文件并返回 context 错误；提交后不再把晚到的取消转换为失败。碰到已有目标返回冲突，不能删除或覆盖该文件。无论成功、取消、编码失败、关闭失败或提交失败，都只清理本会话实际拥有的临时文件。

这是进程内失败下的完整文件可见性保证，不扩展为断电持久性协议。进程强制终止可能留下隐藏临时文件，本期不新增后台清理服务；临时文件不能被 `Library.List` 当作截图。

Win32 错误按各 API 返回契约判断，不能把成功调用后的陈旧 `GetLastError` 当失败；DWM 使用 HRESULT。错误带阶段名称，如 `capture desktop`、`create selection window`、`write png`，不记录像素、API key 或剪贴板内容。资源清理错误不能掩盖原始错误。

## 8. 实施顺序与停止条件

1. **核对与基线**：确认源码版本及范围，补充当前机器的显示布局、缩放、Windows 版本；先采集旧 helper 性能基线。没有设备条件时明确留下待测项。
2. **原生捕获核心**：实现物理坐标、DIB、裁剪、文件提交，完成不依赖真实桌面的测试和必要的隔离 Win32 自检。此阶段不切换生产入口。
3. **选区会话**：完成消息循环、拖动、Space、Esc、失焦、context 和资源清理；在真实桌面单独验收。不能把打开真实屏幕/修改剪贴板的测试放进默认测试命令。
4. **接入与删除**：一次切换 `capture_windows.go`，删除脚本及过期测试，更新 README 和历史设计文档的当前状态；运行回归和打包检查。
5. **交付**：完成实机矩阵、资源与延迟对比；填写实际结果、不能验证的边界。若关键混合 DPI/取消测试失败，修复后再交付，不回退 PowerShell 隐藏问题。

实现者可以将 2、3 合并提交，避免为了分阶段创造临时公共接口。不要顺手修复截图页现有的非本任务文案、重构图库或更改语音/粘贴逻辑。

## 9. 验证与验收

### 9.1 默认自动测试

测试实际逻辑，使用最少的函数级注入点，不为模拟 Win32 建立庞大接口层：

- 几何表驱动：负原点、四向拖动、跨屏、1 像素/零尺寸、越界、窗口部分离屏、完全空洞、溢出和内存上限。
- 小型合成 DIB：红蓝通道、上下方向、stride、alpha、裁剪边界；解码最终 PNG 核验像素和尺寸。
- 文件契约：成功、预取消、提交前取消、写入/Close/移动失败、目标已存在、包含空格/中文/单引号的路径；旧文件不能被覆盖，临时文件清理正确。
- 状态与生命周期：终止幂等、Space 隐藏时的失焦通知、正常 ReleaseCapture 通知、失败到任一步都释放已获资源、下一次截图可重新开始。
- DPI：原生自检验证设置与恢复；API 不可用分支返回错误；加载 user32/gdi32/dwmapi 继续遵守系统路径约束。

注意 `internal/screenshot/screenshot_test.go` 当前有 `darwin` build tag，Windows 的 `go test ./...` 不会执行其中图库 UI 测试；不能误报为完整 UI 覆盖。保留现有纯图库/快捷键测试，必要的新契约测试独立放到 Windows 可执行文件。

### 9.2 构建检查

从 `apps/elir` 执行，环境变量按 PowerShell 语法设置并在结束后恢复：

```text
CGO_ENABLED=0 go test ./...
go tool mygo vet .
CGO_ENABLED=0 go build ./...
go tool mygo build -platform windows/amd64
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ./...
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build ./...
git diff --check
```

上面是命令/环境意图，不是可直接粘贴的 PowerShell 赋值语法。macOS 这里只是交叉编译，不是运行验证。检查 MyGo 打包造成的生成文件变更，避免夹带无关 diff。最终截图路径静态检查不得再含 PowerShell 启动、脚本嵌入或 C# 编译依赖。

### 9.3 必需实机矩阵

| 场景 | 验收内容 |
| --- | --- |
| 单屏 100%、150%、200% | 点位、框线与 PNG 像素位置一致；无透明、反色、翻转 |
| 双屏 100% + 150% / 200% | 各屏内选区及双向跨屏选区，无整体缩放和偏移 |
| 副屏在左/上，主屏切换，竖屏及布局空洞 | 负坐标、矩形交集及黑色空洞正确 |
| 普通/最大化/部分离屏/跨屏窗口 | Space 边界正确，DWM 无额外不可见边框；记录回退情况 |
| 冻结后后台窗口变化 | 记录第 6.2 节限制，不能作为已修复能力汇报 |
| Esc、无效拖动、Alt+Tab、context 取消、显示配置变化 | 无残留覆盖层/文件，无 busy 卡死，下一次可截图 |
| 连续触发快捷键、连续确认/取消各 100 次 | 单会话、资源不持续增长、主 UI 仍响应 |
| PNG 复制、粘贴到支持图片的应用、弹出缩略图、图库重开 | 现有用户链路完整；剪贴板失败时 PNG 保留 |
| 普通 go build 与 MyGo 安装包 | 两条启动路径均确认 DPI 和选区；不启动 PowerShell |
| 禁用 PowerShell 的测试环境 | 截图可工作；不把此项扩大解释为所有企业策略都兼容 |

HDR、受保护窗口、远程桌面可做观察项，记录黑屏、颜色或会话差异；未测不能宣称支持。多 GPU、多显示器的 GDI 兼容性须由实测确认。

### 9.4 性能与资源

同一台机器、同一显示布局，对旧/新实现分别测首次截图和至少 20 次后续截图。分开记录“触发到选区首帧完成”“确认到 PNG 提交”“确认到剪贴板及缩略图可用”，排除用户拖动时间；旧实现首帧可用外部录屏计时，新实现用临时埋点辅助，不把进程启动耗时当完整交互延迟。

记录中位数、P95、工作集、GDI/USER 对象数。重复截图结束并稳定后，资源不应随次数持续增加。新实现必须证实无 helper 子进程；预期启动工作减少，但不在测量前许诺具体毫秒数或倍数。若新实现延迟/内存退化，先定位绘制、转换、编码或已有上层重读，不能为通过对比删除安全与清理逻辑。

## 10. 交接状态

- [x] 核对当前截图入口、脚本、上层调用链、安全加载器及固定 MyGo DPI 实现。
- [x] 查询微软 DPI、GDI、DIB 与 DWM 边界文档。
- [x] 前期形成设计文档。
- [x] 用户授权按本文实现；保持既定交互与兼容边界。
- [x] 原生捕获、选区、文件提交与自动测试实现。
- [x] 生产入口切换、旧脚本移除、README 更新。
- [x] Windows 测试/打包与 macOS 交叉编译。
- [x] 单屏 200% 真实拖动/Space、图片粘贴、连续确认/取消、图库重开。
- [x] 同机新旧各首次 + 20 次性能对比及资源采样。
- [ ] 混合 DPI、多屏与其余硬件矩阵实测（用户确认目前只有单屏）。

实现基线核对：从 `codex/customizable-shortcuts` 的 `d24d9a2003a9ebc0ae3d0b96a08210af29402592` 新建独立分支 **`codex/windows-native-screenshot`**，MyGo `v0.5.5`。开始时仅本文未跟踪；随后出现的 `customizable-shortcuts.md` 属于其他开发，未修改。无 CodeGraph 索引或 llmdoc binding。

实际实现：`platform/screenshot_windows.go` 负责线程 DPI、显示器枚举、限额 DIB、每屏 BitBlt、BGRA 裁剪及非覆盖 PNG 提交；`screenshot_overlay_windows.go` 负责单次注册的类/回调、物理坐标拖动、Space/DWM、Esc/失焦/显示变化/context 取消与资源释放。截图包只转发；旧脚本及 helper 测试删除。可选旧系统 shcore 查询所需的 loader 仅拆出返回 error 的变体，保持真实 System32 和依赖搜索限制，不因可选 DLL 缺失 panic。图库、剪贴板、快捷键和 macOS 业务代码未改。

### 2026-10-18 自动验证证据

- `CGO_ENABLED=0 go test ./... -count=1`：通过，包含本次拖动修复。先在沙箱内因既有 HTTP/WebSocket 测试无法连接回环地址失败，随后在正常本地权限下完整重跑通过。文件提交测试使用仓库内 `build/.native-screenshot-test-tmp` 作为 TEMP/TMP；系统临时目录的移动操作在沙箱内被拒绝。执行时保存并恢复环境变量。
- 几何测试覆盖负原点、四向拖动、跨屏/越界/部分离屏、完全空洞、1 像素/零尺寸、坐标溢出及 512 MiB 边界。合成 top-down DIB 测试行步幅、颜色、alpha、裁剪，并解码最终 PNG 核验像素。
- 文件契约覆盖中文/空格/单引号路径、成功、预取消/提交前取消、编码/写入/Close/移动失败、已有目标不覆盖、提交后晚到取消；临时文件不会进入图库。
- 默认测试仅使用合成像素及隐藏窗口，不捕获桌面、不操作剪贴板。真实原生 DIB 创建/释放 100 次通过；隐藏 HWND 经真实消息循环执行 100 次关闭/Esc/context/显示/DPI 消息终止，GDI 对象 **2 → 2**、USER 对象 **0 → 0**。这不是全桌面/用户链路的资源基准，也不代表每一个 Win32 调用均做了故障注入。
- 补充回归时进程级 USER 计数偶尔增加 1，首次窗口预热仍不稳定；未把原因归结为已证明的 IME 行为。测试改为用真实 `IsWindow` 逐个确认本会话 HWND 已销毁，并继续断言注册表清空、清理返回成功、GDI 不增长；USER 总量仅记录。连续 5 轮各 100 个隐藏会话通过（GDI 均 2 → 2），随后完整测试通过。完整桌面资源趋势另见下文。
- DPI 原生自检证明设置与恢复成功；查询 API 缺失分支拒绝未证明的 DPI 状态；新增 DLL 的 System32 路径检查通过。状态测试确认已接受结果不被自身隐藏、ReleaseCapture 或迟到通知改写。
- 显式 `ELIR_SCREENSHOT_DESKTOP_TEST=1 go test ./internal/platform -run '^TestScreenshotDesktopCancellation$' -count=1 -v`：正常桌面权限下通过。Windows **10.0 build 26300 amd64**，枚举到单屏 **(0,0)-(3840,2160)**，系统 DPI **192（200%）**。连续 3 次真实 GDI 抓取和覆盖层 context 超时取消，无输出文件、无会话注册残留，不写剪贴板。首次沙箱内执行因前台焦点不可用而按设计失败并退出，未将其当成功或放宽焦点检查。
- `go tool mygo vet .`、`CGO_ENABLED=0 go build ./...`、Darwin arm64/amd64 `CGO_ENABLED=0 go build ./...`：通过。macOS 仅交叉编译，未运行。最终并行复查中 vet 曾扫到正在清理的仓库内 Go 临时目录而失败；后续 TEMP/TMP 改为 Go 会跳过的 `build/.native-screenshot-test-tmp`，清除已空的旧目录，串行重跑 vet 通过。
- 修复后再次通过 Windows 全量测试、MyGo vet、普通构建及 Darwin arm64/amd64 交叉编译。`go tool mygo build -platform windows/amd64 -o build/native-screenshot-verified`：通过，生成该目录 `windows-amd64/Elir.exe` 与 `Elir Setup 0.3.0 amd64.exe`，这是当前修复版本。前一版位于 `build/native-screenshot`；独立输出避免覆盖仍在运行的旧 exe，未终止其他开发进程。安装包无签名/更新签名，没有安装或发布。
- MyGo 生成的 `locales/catalog_gen.go` 仅换行差异，确认无语义变化后恢复。截图 Windows 路径静态检查无 PowerShell 启动、脚本嵌入、C# 或 helper 依赖。
- `git diff --check`：通过。

### 单屏交互与用户链路实测

2026-10-18，Windows build 26300、单屏 3840×2160、200%。在独立配置/图库和独立单实例锁目录下，从 `e102493`（同步本次 MSG.pt 修复）与旧基线 `d24d9a2` 各准备一份忽略目录中的测试副本，仅添加计时、受控触发及 UI 自动化识别探针，正式程序不加入测试开关。全局快捷键主动过滤注入事件，所以 Computer Use 无法把注入的 Ctrl+Shift+2 当作实体键盘测试；副本直接调用原 `Screenshot.take`，随后用真实窗口的鼠标/键盘事件验证交互。

- 真实快速反向拖动暴露了 `GetCursorPos` 读取晚于事件的问题：按下与松开可能均读到终点而取消。已改为消息投递时的 `MSG.pt`，保留 32 位物理坐标；Space 继续实时查询。新增负坐标、超过 16 位范围及正/反方向消息坐标回归通过。修复后正、反拖动同一区域均输出 **600×240** PNG，SHA-256 均为 `2C8B94F0839F4AF4C884D614ABEE7D61FB6DDAE328C54EB107C9C3CB28CA7AA3`。
- Space 命中测试窗口，输出 **1498×1069**；通过 Windows 画图 Ctrl+V 真正粘贴，选区尺寸同为 1498×1069，已保存本地证据。原 PNG 全部 alpha 为 255。画图二次保存有 28,623 个像素差异，因此不使用它证明逐像素等价。另一次受控截图后，立即读取本应用写入的注册格式 `PNG`，去除 GlobalAlloc 尾部填充后与保存文件 **148,295 字节完全相同**；这是剪贴板原始 PNG 的直接证据，未由此归因画图二次保存差异。
- 原生连续 **100 次确认 + 100 次 Esc 取消** 均完成，最终磁盘 **100 张**，取消不增文件。图库实际显示“共 100 张 · 13.8 MB”，缩略图正确绘制；正常关闭设置再经第二实例激活后仍能看到同一图库。
- 100 次取消期间 GDI 始终 **50**，USER **99–102**，最后为 101；确认阶段 USER 预热后稳定约 101。工作集不是持续线性增长：取消阶段约 **238.6 → 229.0 MiB**；保留了 100 张图片的上层缩略图缓存，此数据不可解释为单次捕获峰值预算。
- 探针在首帧完成后临时暴露 tool window 供 Computer Use 选择，在输入事件开始时恢复正式样式，再执行原有处理。首帧计时点在暴露之前；上述 GUI 证据属于带探针副本的集成验证，不等同于未修改安装包的全部用户链路。

### 同机性能与资源结果

新旧各记录首次及后续 20 次；首次指新启动应用的第一次截图，没有清理 OS/磁盘缓存。首帧在绘制完成并 `DwmFlush` 后记录；确认耗时从选择事件开始计时，不包含人等待时间。旧 WinForms 窗口无法由 Computer Use 选择，故用临时 Timer 调用原 Space 处理；新版本使用真实注入 Space。两者均执行原截图、保存及上层剪贴板/缩略图流程，埋点副本未进入交付产物。

| 阶段（ms） | 旧版首次 | 原生首次 | 旧版后续中位数 / P95 | 原生后续中位数 / P95 |
| --- | ---: | ---: | ---: | ---: |
| 触发到首帧 | 1855.73 | 95.26 | 883.88 / 932.71 | 88.12 / 106.88 |
| 确认到 PNG 提交 | 74.69 | 50.82 | 75.42 / 78.39 | 51.41 / 56.08 |
| 确认到剪贴板与缩略图就绪 | 132.70 | 71.78 | 129.13 / 133.38 | 71.73 / 75.86 |

同一目标窗口，旧版 GetWindowRect 输出 **1520×1080**，原生 DWM 输出 **1498×1069**，面积相差约 2.5%；这是实际行为对比，非严格相同像素矩形的编码微基准。旧版首次试跑在首张后命中了其他窗口，已按尺寸发现并作废，表中只使用重新激活同一目标后的 21 个有效样本。该条件下原生后续首帧中位数约降低 90%，确认到上层就绪约降低 44%，不外推到所有设备。

首帧工作集采样：原生单进程 **156.8–224.3 MiB**；旧 helper **189.4–190.6 MiB**，另有主进程触发前 **93.8–167.7 MiB**。这些是阶段采样，旧主进程与 helper 非同步峰值；不能据此给出精确峰值节省比例。就绪后主进程原生 **139.6–192.8 MiB**、旧版 **100.9–162.6 MiB**，原生 Go 堆驻留并未全面更低。旧 helper 首帧 GDI 30 / USER 30–31；原生首帧包括自身覆盖层，完整压力趋势见上文。原生截图生产路径静态不含进程启动或 helper 依赖；普通 console 构建可能有 conhost，不能把它误报为截图 helper。

临时源码、计时与资源日志、截图和画图证据均保留在忽略目录 `build/.screenshot-validation/`，不提交包含桌面像素的产物。有效性能文件为 `native-latency.log`、`native-metrics.json`、`old-latency.log`、`old-metrics.json`；压力文件为 `native-stress.log`、`native-cancel-100.log`。原始注册 PNG 验证副本为 `clipboard-proof.png`。

### 尚未验证的边界

用户确认目前只有单屏，因此混合 DPI/多屏、负坐标真实显示布局、竖屏、主屏切换及布局空洞保留为硬件待验，合成几何测试不能替代实机。100%/150%、最大化/跨屏/部分离屏窗口、真实显示配置变更、实体快捷键、未修改 exe 与安装版的完整交互、禁用 PowerShell 的企业策略环境也未验证；HDR、受保护内容、RDP、多 GPU 未测。新安装包已构建，未安装或发布，不宣称完整硬件兼容。

下一步在具备对应硬件时执行第 9.3 节剩余矩阵；遇到混合 DPI 或取消失败先修复，不回退 PowerShell。Space 仍是实时命中冻结像素，窗口移动限制与 DWM 回退边框限制保留。
