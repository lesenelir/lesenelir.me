# Windows 桌面支持设计与验证（PR #8）

对应 [PR #8](https://github.com/lesenelir/lesenelir.me/pull/8)，分支 `codex/windows-support`。更新：2026-10-18。平台移植、安全加固、启动音量恢复和托盘交互统一在本文维护，不再为每个后续修复单独建文档。

## 范围与当前状态

保持 MyGo v0.5.5、Go 原生 UI、语音控制器、转写客户端、历史和图库。共享业务依赖 `internal/platform`，macOS 实现通过 Darwin 门面保留，Windows 提供原生实现；不新增依赖或 CGO。

本 PR 提交为平台支持 `4fe5f23`、安全加固 `f280e03`、音量恢复 `73dda92`、托盘交互 `d24d9a2`。后续 [PR #9 原生截图](windows-native-screenshot.md) 替换最初的 PowerShell/WinForms 路径；[PR #10 快捷键与浮窗](customizable-shortcuts.md) 管理配置按键和透明浮窗修复。当前 Windows 截图不再依赖 PowerShell/.NET，默认键也已可修改。

## 平台边界

| 范围 | Windows 实现及契约 |
| --- | --- |
| 入口与界面 | Windows/macOS 共享入口与设置；隐藏 Windows 不适用的 Dock、macOS 权限项 |
| 键盘 | WH_KEYBOARD_LL 运行在独立线程（自有消息循环），按顺序把按键交给主线程的 Router，最多等 200 ms，超时放行，避免主线程繁忙时拖慢全局输入或被系统摘除（`LowLevelHooksTimeout`）。Alt/Win 按住时被吞掉的按键、注入 Ctrl+C/V 前后释放和恢复修饰键时，都补发未分配键 VK 0xE8，前台应用不会把单独的 Alt 当作菜单键；模拟输入不回流业务快捷键 |
| 麦克风 | WinMM PCM；三个复用的 100 ms 缓冲区按队列顺序处理，reset 后收取不足整块的尾部 |
| 粘贴／选中复制 | 剪贴板与原生按键注入；剪贴板读写失败时停止，不继续注入 |
| 输出音量 | Core Audio endpoint API；保存设备 ID，在原设备恢复，不误改切换后的默认设备 |
| 密钥 | Windows 用户级 DPAPI；保留 macOS JSON 格式与旧数据读取逻辑 |
| 托盘 | Windows 左键经 showSettings 恢复同一窗口、保留页面并重读设置；右键菜单及 macOS 点击行为不变 |
| 发布 | Windows 手工安装；不修改 macOS 更新源和签名发布流程 |

Windows 默认 F8／左 Shift+F8 启动听写／翻译，Alt+T／Alt+R 处理选中文本，Ctrl+Shift+2 截图，Esc 取消，翻译中 Tab 切换语言。完整动作及配置契约由 [PR #10](customizable-shortcuts.md#动作与配置) 维护。

## 安全与数据保护

新增 DLL 从 OS 查询得到的真实 System32 目录加载，用 `LOAD_LIBRARY_SEARCH_SYSTEM32` 限制依赖搜索，不信任 PATH、当前目录或 SystemRoot 环境输入。可选 API/DLL 缺失返回相关功能错误，不回退不受限搜索。截图资源与文件提交由 [PR #9](windows-native-screenshot.md#取消文件与资源所有权) 维护。

DPAPI 保护当前用户持久化的 API key，不抵御以同一用户身份运行的进程。密钥文件读取失败时拒绝覆盖；能读出但解密或 JSON 解析失败（如密码被重置后 DPAPI 失效）时，原文件改名为 `secrets.json.unreadable-<时间>` 原样保留，以空存储继续，用户可重新填写 key。写盘成功后才发布新 key，失败保留原密文和内存值。

共享 settings.json 采用相同策略：读取失败时使用默认值，SetRaw 和 Section.Update 均拒绝写入、内存值也不发布，避免覆盖词典、快捷键等原配置，文件恢复可读后重启即可；能读出但整文件 JSON 解析失败时，原文件改名为 `settings.json.unreadable-<时间>` 原样保留，以默认值继续并可正常保存，用户不会被困在“无法保存”。

自动粘贴临时文本请求排除 Windows 剪贴板历史／云同步；显式复制保持现有行为。这不阻止本机其他程序读取当前剪贴板。普通权限程序向更高权限目标粘贴可能被系统阻止，沿用手动粘贴提示。

PR #8 曾对截图 helper 使用系统路径、系统模块范围并移除 API key 环境变量。这是历史实现的约束；PR #9 已删除脚本、C# 编译和子进程环境通道，当前不保留此旁路。

参考：[DLL 搜索安全](https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-security)、[剪贴板历史／同步排除格式](https://learn.microsoft.com/en-us/windows/win32/dataxchg/clipboard-formats#cloud-clipboard-and-clipboard-history-formats)。

## 音量恢复与退出

调用链为 `Voice.Start → quieter.recover → RestoreOutput → withOutput`。MyGo Windows 主线程已是 STA，录音回调常在未初始化或 MTA 工作线程执行。

1. withOutput 锁定 OS 线程，先尝试 MTA，仅在 RPC_E_CHANGED_MODE 时按已有 STA 初始化。
2. 每次成功初始化（包括 S_FALSE）配对一次 CoUninitialize，不改变或释放调用者持有的 COM apartment。
3. 恢复失败保留有效恢复记录与内存状态；再次录音调整音量前先重试，未恢复则不覆盖待恢复设备／状态。
4. 恢复成功才清空状态并删除记录；设备已不存在（`ErrOutputGone`：CoreAudio 设备不再 alive，Windows endpoint 找不到或不再 active）时同样丢弃，否则之后每次录音都无法静音。schema 不变，无效 JSON 沿既有逻辑丢弃。

录音开始时，静音或降低音量后必须检查恢复记录的序列化与写入结果。保存失败记录错误并立即调用 restore；恢复也失败则保留内存状态，在录音结束或下一次调整音量前重试。若保存与恢复持续失败且进程随后退出，仍没有可用的持久化恢复保证。

正常退出链为 `App.beforeQuit → Voice.Stop → Controller.Close → endSession → Recorder.Cancel → Recording(false)`。首次退出先 PreventDefault，在后台取消控制器并等待录音、音量恢复回调和流关闭完成，之后再次请求 MyGo Quit；等待中的重复退出也被拦住。等待最多 3 秒（`stopTimeout`），超时记录日志后照常退出，未完成的音量恢复由下次启动的恢复记录兜底，卡住的录音器不会让应用无法退出。清理期间保持 UI 循环运行，因为 Darwin Recorder.Cancel 通过 RunOnMain 释放录音资源。未启动的 Voice 和重复 Close 均可安全处理；强制终止仍由下次启动恢复记录兜底。

原生回归使用无效设备 ID，覆盖未初始化、已有 MTA／STA 线程，不改真实音量；通过后续 apartment 切换检查 S_FALSE 引用平衡。修复前 STA 复现 HRESULT 0x80010106，修复后通过。共享 quieter 测试用假 API 验证失败保留、重试和成功清理。

## 已有验证与证据边界

以下整理 2026-10-17～18 已执行记录，不将历史构建和短期 smoke 当作完整设备验收。

| 检查 | 实际结果 |
| --- | --- |
| Windows CGO_ENABLED=0 go test ./...、MyGo vet | 通过；覆盖输入／ABI、音频缓冲所有权、剪贴板失败、DPAPI、损坏密钥、写盘失败、COM 和音量恢复 |
| Windows 构建及 MyGo NSIS 打包 | 通过；生成 Elir.exe 与安装包，未签名／发布 |
| macOS arm64/amd64 编译及相关测试二进制编译 | 通过；未在 Windows 执行 macOS 测试 |
| 隔离 APPDATA 启动 smoke | 进程存活、hook 启动、无错误日志；单次三秒空闲约 88.6 MiB 私有内存、CPU 增量 0 ms，不构成性能基准 |
| 托盘原生消息 smoke | 隔离进程隐藏设置后，左键恢复同一窗口，右键进入菜单；不是实体任务栏点击验收 |
| 当时独立评审 | 修复并复查剪贴板失败、权限提示、DLL 依赖搜索与 COM 问题后通过 |

完整测试曾受沙箱回环 socket／临时文件 rename 限制，在正常本地权限下重跑通过；没有放松业务保护。

2026-10-18 三项审查修复已完成：退出等待清理、配置加载失败禁止覆盖、恢复记录写入失败立即回滚。新增回归覆盖阻塞恢复回调时等待退出、重复退出、配置读取／解析失败的文件及内存保护、修复后重新打开、恢复记录不可写后的立即恢复及失败重试。Windows `CGO_ENABLED=0 go test ./...`、相关四包 `go test -race` 和 `go tool mygo vet .` 通过；macOS arm64/amd64 的 `go build ./...` 及相关四包测试二进制编译通过。本轮测试使用假录音／音量 API，没有修改真实系统音量；下一步是真机录音中从托盘退出验收，未重新打包安装版。

## 构建与剩余验收

仓库根执行 `pnpm dev:elir:windows` 或 `pnpm build:elir:windows`；也可在 apps/elir 执行 `go tool mygo build -platform windows/amd64`，产物在 build/windows-amd64/，首次可能下载 NSIS。macOS 测试交叉编译应在打包结束、临时 Windows .syso 清理后执行。

仍需验收真实麦克风与断连、音量变更后的崩溃恢复、普通／高权限目标粘贴、实体托盘及快捷键、macOS 真机发布与交互。Windows arm64 未运行硬件测试，没有发起真实语音／LLM 请求。截图多屏／DPI 归 PR #9，macOS 新选择器和透明浮窗归 PR #10，避免重复记录。
