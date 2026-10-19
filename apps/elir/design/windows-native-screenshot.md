# Windows 原生截图设计与验证（PR #9）

对应 [PR #9](https://github.com/lesenelir/lesenelir.me/pull/9)，分支 `codex/windows-native-screenshot`，依赖 [PR #8](windows-support.md)。更新：2026-10-18。原生实现为 `e102493`，队列鼠标坐标修复为 `fcb23a7`。

## 决策与责任边界

进程内 Go/Win32 替换 PowerShell/WinForms：GDI 冻结可见虚拟桌面，原生窗口选区，标准库编码 PNG，再进入原图库、剪贴板、缩略图流程。保持 MyGo v0.5.5，不新增依赖、CGO、helper、后台服务或备用后端。

不实现 HDR 映射、滚动截图、标注、光标合成、窗口阴影或被遮挡内容恢复。GDI 返回冻结时可见桌面，受保护内容可能为空白。

当前调用链：

```text
platform.EventTap → shortcut.Router → Screenshot.take（busy CAS，工作 goroutine）
  → Screenshot.capture → Library.NewPath
  → capture(ctx, path, bindings, router) → platform.CaptureScreenshot(ctx, path)
  → 读取 PNG、复制剪贴板、生成缩略图
  → Settings.Update → 图库与缩略图
```

PR #9 的平台入口仍为 `CaptureScreenshot(context.Context, string) (bool, error)`。PR #10 在截图适配层增加绑定快照／按键归属回调，经 context 传入，不改变捕获和保存契约。

| 文件（相对 apps/elir） | 当前职责 |
| --- | --- |
| internal/screenshot/capture_windows.go | 转发至平台，当前同时接入 PR #10 快捷键快照 |
| internal/platform/screenshot_windows.go | 线程 DPI、显示器、GDI、裁剪、PNG 提交 |
| internal/platform/screenshot_overlay_windows.go | 原生窗口、消息循环、拖动、窗口命中、取消与清理 |
| internal/platform/screenshot_windows_test.go | 几何／像素、文件、DPI、原生资源回归 |
| 图库／剪贴板／缩略图 | 保持业务规则；剪贴板失败不删除已保存 PNG |

Windows 脚本、嵌入和 helper 测试已删除。macOS 选择器由 [PR #10](customizable-shortcuts.md#截图局部输入) 改造。

## 线程、DPI 与像素契约

会话在锁定的 OS 工作线程完成 DPI、GDI、HWND 和消息循环，不占用 MyGo 主线程。进入前设置并验证 per-monitor DPI：优先 PMv2，再尝试 per-monitor；API 缺失时仅在能证明继承状态满足要求时继续，否则返回截图功能错误，不改变全进程 DPI。

坐标统一为带符号的物理屏幕像素和半开矩形：

- 显示器并集为 V，快照大小 V.Dx×V.Dy；布局空洞清零填黑。
- 拖动使用投递时的 32 位 MSG.pt，避免 GetCursorPos 延迟读取已到终点，或 lParam 16 位截断。
- R 与 V 求交后以 R.Sub(V.Min) 裁剪，不再乘 DPI；完全空洞、零尺寸或任一边不超过一像素按无效取消。
- Space 在同一 DPI 上下文实时取鼠标坐标。
- 退出先释放 HWND/GDI，再恢复原 DPI；恢复失败不把带修改状态的线程放回池，也不提交 PNG。

覆盖层显示前建立一份 top-down、32-bit BI_RGB DIB，清零后按显示器 BitBlt（SRCCOPY|CAPTUREBLT）。任何屏捕获失败则整次失败。绘制／裁剪共享快照，拖动中不重复抓屏，也不为编码复制完整桌面。

读取前 GdiFlush，BGRA 转 RGB，PNG alpha 强制为 255。先检查尺寸／stride／总字节溢出和 Win32 范围，DIB 固定预算 512 MiB；这不是含选区图像、PNG、上层缓存的总峰值。多屏不是同一时刻原子采样。

## 交互与窗口命中

| 输入／事件 | 行为 |
| --- | --- |
| 左键按下／拖动／松开 | 捕获鼠标，最终坐标重算；有效矩形确认，否则取消 |
| 窗口键，默认 Space | 非重复时隐藏自身，WindowFromPoint → GA_ROOT，立即选窗 |
| 取消键，默认 Esc；右键；关闭；激活后失焦／Alt+Tab | 取消，不残留覆盖层或 busy |
| 非自身收尾的鼠标捕获丢失 | 取消 |
| 显示布局／分辨率／DPI 变化 | 返回配置变化错误，要求重试 |
| context 取消／超时 | 会话 timer 检查后退出，不保证中断阻塞原生 API |

结果只完成一次，自身隐藏、正常 ReleaseCapture 和迟到通知不能改写已确认结果。激活前先发送一个零位移的鼠标输入（`SendInput`，不移动指针），使 Elir 成为最后输入的接收者，再 `SetForegroundWindow`：快捷键按下时 Elir 在后台，否则前台锁会拒绝激活。仍未拿到前台时不退出：置顶覆盖层照常接收鼠标，第一次点击即激活；在此之前 `WM_TIMER` 轮询所配置的取消键（`GetAsyncKeyState`），右键也取消，覆盖层不会无法关闭。WM_PAINT 配对 BeginPaint/EndPaint，框线不烧进 PNG。

边界优先 DWM 扩展可见边界，失败回退 GetWindowRect，后者可能含不可见边框。圆角外保留桌面背景。冻结后外部窗口移动／关闭／换层叠时，实时命中可能与快照不符；当前没有冻结窗口树。

## 取消、文件与资源所有权

| 结果 | 返回及文件规则 |
| --- | --- |
| 确认且提交成功 | true, nil；最终路径是完整可解码 PNG |
| 用户取消／无效／失焦 | false, nil；不增截图，不碰剪贴板 |
| context 取消 | false, ctx.Err()；提交前不留本次最终文件 |
| 捕获／清理／编码／文件错误 | false, 带阶段信息的 error；不交付不完整 PNG |

先关闭覆盖层、转换选区、释放资源／恢复 DPI，再编码到目标目录隐藏临时文件；检查 Encode、Write、Close 和提交前 context。MoveFileExW 不使用 REPLACE_EXISTING，已有目标失败。成功移动是提交点，之后晚到取消不改成失败；失败只清理会话拥有的临时文件，不删已有目标或已提交 PNG。

这是进程内完整文件可见性保证，不是断电协议；强杀可能留下图库不识别的隐藏临时文件。上层仍用 Background context，未扩展应用退出取消协议。

桌面 DC 用 ReleaseDC，内存 DC 用 DeleteDC；恢复旧选入对象后再 DeleteObject。会话线程销毁 HWND、timer、鼠标捕获并移除注册关系。类／回调只注册一次，不把临时 Go 指针长期交给系统。DLL 遵守真实 System32 约束；错误不记录像素、key、剪贴板，清理错误不覆盖原错误。

## 已执行验证

以下整理 2026-10-18 已有验收；本次归并未重跑压力和性能测量。原始日志曾保留于忽略目录 build/.screenshot-validation/，目前不在此工作区，以下保留当时记录的结果与条件。

| 类别 | 证据 |
| --- | --- |
| 默认自动测试 | 负原点、四向／跨屏／越界、空洞、溢出、512 MiB；合成 DIB 方向、颜色、stride、alpha、PNG 解码 |
| 文件／状态 | 中文／空格／单引号路径、已有目标、取消、Encode／Write／Close／Move 失败、提交后取消；自身收尾不改写接受状态 |
| 原生资源 | 100 次 DIB 创建释放；逐 HWND 验证销毁、注册表清空，五轮各 100 隐藏会话 GDI 不增长。USER 预热波动仅记录，不代表每个 API 均故障注入 |
| 桌面取消 | 显式 TestScreenshotDesktopCancellation：三次真实抓取和 context 取消，无文件／注册残留、不写剪贴板 |
| 构建 | Windows 全量测试、MyGo vet、普通构建／NSIS，macOS arm64/amd64 编译通过 |
| 单屏集成 | Windows 11（内核 10.0 build 26300）amd64，3840×2160、200%；正反拖动均产出 600×240 相同 PNG |
| 窗口／剪贴板 | Space 输出 1498×1069，画图实际粘贴同尺寸；注册 PNG 原始 148,295 字节与磁盘一致。画图二次保存有像素差异，不作逐像素等价证明 |
| 连续操作 | 100 次确认 + 100 次 Esc 取消，最终 100 张；图库关闭重开可见。取消阶段 GDI 50，USER 99–102，工作集约 238.6→229.0 MiB，未持续线性增长 |

集成使用隔离配置／锁的埋点副本，受控调用原 Screenshot.take 后注入窗口事件；全局 hook 过滤注入事件，因此不是实体快捷键证据。交付不含探针，也不能将其等同未修改安装包全部交互。默认测试不抓桌面、不操作剪贴板，真实桌面检查需显式启用。

### 同机性能

新旧各新启动一次，记录首次和后续 20 次，不清 OS 缓存。首帧含绘制完成／DwmFlush，确认计时不含用户等待。

| 阶段（ms） | 旧首次 | 原生首次 | 旧后续中位数 / P95 | 原生后续中位数 / P95 |
| --- | ---: | ---: | ---: | ---: |
| 触发到首帧 | 1855.73 | 95.26 | 883.88 / 932.71 | 88.12 / 106.88 |
| 确认到 PNG 提交 | 74.69 | 50.82 | 75.42 / 78.39 | 51.41 / 56.08 |
| 确认到剪贴板／缩略图就绪 | 132.70 | 71.78 | 129.13 / 133.38 | 71.73 / 75.86 |

旧 helper 临时 Timer 调用原 Space，新版注入 Space。旧 GetWindowRect 输出 1520×1080，新 DWM 输出 1498×1069，面积差约 2.5%，不是同像素矩形的 PNG 编码微基准；误命中其他窗口的旧样本已作废。上述条件下后续首帧中位数约降 90%，上层就绪约降 44%，不外推设备。

原生首帧工作集 156.8–224.3 MiB；旧 helper 189.4–190.6 MiB，另有非同步主进程 93.8–167.7 MiB。原生 Go 堆驻留并非全面更低，不能推算精确峰值节省；图库缓存不属于 DIB 预算。

## 剩余边界与下一步

用户当前只有单屏。真实混合 DPI、多屏负坐标／空洞／竖屏／主屏切换、100%／150%、跨屏／部分离屏／最大化窗口、真实显示变化、实体快捷键和未修改安装包完整交互仍待验收。HDR、受保护内容、RDP、多 GPU、早期 Windows 10、禁用 PowerShell 的企业策略环境未测。

合成跨屏几何不替代硬件，带 darwin 标签的图库 UI 测试也不在 Windows 全量测试中执行。下一步在对应设备执行这些矩阵，保留 DWM 回退和冻结后实时命中的限制，不回退 PowerShell 掩盖错误。
