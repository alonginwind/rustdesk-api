# 被控端策略配置文档

## 概述

被控端策略通过心跳接口 (`/api/heartbeat`) 下发给被控端。被控端收到策略后，会将 `config_options` 中的配置应用到本地。

### 策略优先级

- **设备策略** > **默认策略**
- 如果设备有单独配置的策略，使用设备策略
- 如果没有设备策略，回退到全局默认策略

### 下发机制

1. 被控端每 15 秒发送一次心跳
2. 心跳请求中包含 `modified_at`（客户端本地的策略时间戳）
3. 服务端比较 `modified_at`，如果不一致则下发策略
4. 被控端收到策略后应用到本地配置，并更新时间戳

---

## API 接口

### 管理端接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/admin/peer_strategy/list` | GET | 策略列表（分页） |
| `/admin/peer_strategy/detail/:id` | GET | 策略详情 |
| `/admin/peer_strategy/create` | POST | 创建设备策略 |
| `/admin/peer_strategy/update` | POST | 更新设备策略 |
| `/admin/peer_strategy/delete` | POST | 删除设备策略 |
| `/admin/peer_strategy/default` | GET | 获取默认策略 |
| `/admin/peer_strategy/default/update` | POST | 设置/更新默认策略 |

### 请求示例

**创建设备策略**

```json
POST /admin/peer_strategy/create
{
  "peer_id": "123456789",
  "config_options": {
    "access-mode": "view",
    "enable-clipboard": "N",
    "enable-file-transfer": "N"
  }
}
```

**设置默认策略**

```json
POST /admin/peer_strategy/default/update
{
  "config_options": {
    "enable-record-session": "Y",
    "allow-auto-record-incoming": "Y"
  }
}
```

---

## 支持的策略配置项

### 权限控制

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `access-mode` | 访问模式 | `custom`, `full`, `view` |
| `enable-keyboard` | 允许键盘输入 | `Y`, `N` |
| `enable-clipboard` | 允许剪贴板 | `Y`, `N` |
| `enable-file-transfer` | 允许文件传输 | `Y`, `N` |
| `enable-camera` | 允许摄像头 | `Y`, `N` |
| `enable-terminal` | 允许终端 | `Y`, `N` |
| `enable-remote-printer` | 允许远程打印 | `Y`, `N` |
| `enable-audio` | 允许音频 | `Y`, `N` |
| `enable-tunnel` | 允许隧道 | `Y`, `N` |
| `enable-remote-restart` | 允许远程重启 | `Y`, `N` |
| `enable-record-session` | 允许录制会话 | `Y`, `N` |
| `enable-block-input` | 允许阻断输入 | `Y`, `N` |
| `enable-privacy-mode` | 允许隐私模式 | `Y`, `N` |
| `allow-remote-config-modification` | 允许远程修改配置 | `Y`, `N` |
| `allow-remote-cm-modification` | 允许控制端点击接受窗口（连接管理窗口）来接受连接、修改权限等 | `Y`, `N` |
| `enable-perm-change-in-accept-window` | 允许用户在接受窗口（连接管理窗口）中接受传入会话前更改权限 | `Y`, `N` |
| `one-way-clipboard-redirection` | 禁用从被控端到控制端的剪贴板同步（被控端，>= 1.3.1） | `Y`, `N` |
| `one-way-file-transfer` | 禁用从被控端到控制端的文件传输（被控端，>= 1.3.1） | `Y`, `N` |
| `sync-init-clipboard` | 建立连接时同步初始剪贴板（仅从控制端到被控端，>= 1.3.1） | `Y`, `N` |
| `file-transfer-max-files` | 单次文件传输最大文件数 | 正整数，`0` 为内置默认值 |

### 安全设置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `approve-mode` | 审批模式 | `password`, `click`, `password-click` |
| `verification-method` | 验证方式 | `use-temporary-password`, `use-permanent-password`, `use-both-passwords` |
| `temporary-password-length` | 临时密码长度 | `6`, `8`, `10` |
| `allow-numeric-one-time-password` | 允许纯数字一次性密码（>= 1.4.1） | `Y`, `N` |
| `allow-logon-screen-password` | 允许在登录屏幕使用密码输入（click 审批模式下，>= 1.4.7） | `Y`, `N` |
| `allow-scope-violation-close` | 允许违规关闭 | `Y`, `N` |
| `allow-scope-violation-alarm` | 允许违规告警 | `Y`, `N` |
| `disable-change-permanent-password` | 禁止修改永久密码（>= 1.4.5） | `Y`, `N` |
| `disable-change-id` | 禁止修改设备 ID（>= 1.4.5） | `Y`, `N` |
| `disable-unlock-pin` | 禁止使用 PIN 解锁设置（>= 1.4.5） | `Y`, `N` |
| `allow-command-line-settings-when-settings-disabled` | 禁用设置时仍允许命令行配置（>= 1.4.7） | `Y`, `N` |
| `allow-deep-link-password` | 允许通过 deep link 设置密码（`rustdesk://password/xxx`，仅 Android/iOS） | `Y`, `N` |
| `allow-deep-link-server-settings` | 允许通过 deep link 导入服务器配置（`rustdesk://config/xxx`，仅 Android/iOS） | `Y`, `N` |

### 网络设置

| 配置项 | 说明 | 示例值 |
|--------|------|--------|
| `custom-rendezvous-server` | 自定义中继服务器 | `rs.example.com` |
| `api-server` | API 服务器 | `api.example.com` |
| `key` | 连接密钥 | `your-secret-key` |
| `relay-server` | 中继服务器 | `relay.example.com` |
| `ice-servers` | ICE 服务器 | JSON 格式 |
| `direct-server` | 启用直接 IP 访问 | `Y`, `N` |
| `direct-access-port` | 直接 IP 访问端口 | `21118` |
| `allow-websocket` | 允许 WebSocket（>= 1.4.0） | `Y`, `N` |
| `disable-udp` | 仅使用 TCP | `Y`, `N` |
| `allow-insecure-tls-fallback` | 允许不安全 TLS 回退（>= 1.4.4） | `Y`, `N` |
| `allow-https-21114` | 允许 21114 作为 HTTPS 端口（>= 1.3.9） | `Y`, `N` |
| `enable-udp-punch` | 启用 UDP 打洞（>= 1.4.1） | `Y`, `N` |
| `enable-ipv6-punch` | 启用 IPv6 P2P 连接（>= 1.4.1） | `Y`, `N` |
| `allow-hostname-as-id` | 使用主机名作为 ID（>= 1.4.0） | `Y`, `N` |
| `enable-tcp-punch` | 启用 TCP 打洞 | `Y`, `N` |
| `enable-webrtc` | 启用 WebRTC | `Y`, `N` |
| `enable-port-forward-mux` | 启用端口转发复用 | `Y`, `N` |
| `allow-kcp-congestion-control` | 允许 KCP 拥塞控制 | `Y`, `N` |
| `allow-webrtc-congestion-control` | 允许 WebRTC 拥塞控制 | `Y`, `N` |
| `relay-fallback-delay` | 中继回退延迟（毫秒） | 正整数 |

### 连接控制

| 配置项 | 说明 | 可选值/示例 |
|--------|------|-------------|
| `enable-lan-discovery` | 启用局域网发现 | `Y`, `N` |
| `whitelist` | IP 白名单 | IP 列表，逗号分隔，支持 CIDR |
| `id-whitelist` | ID 白名单（>= 1.5.0） | ID 列表，支持 `*` 和 `?` 通配符 |
| `allow-auto-disconnect` | 允许自动断开 | `Y`, `N` |
| `auto-disconnect-timeout` | 自动断开超时（分钟） | `10` |
| `allow-only-conn-window-open` | 仅窗口打开时允许连接 | `Y`, `N` |
| `register-device` | 是否注册设备（Pro >= 1.6.0） | `Y`, `N` |

### 录制设置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `allow-auto-record-incoming` | 自动录制传入连接 | `Y`, `N` |
| `allow-auto-record-outgoing` | 自动录制传出连接（>= 1.3.2） | `Y`, `N` |
| `hide-recording-button` | 隐藏录制按钮（不禁止录制） | `Y`, `N` |
| `video-save-directory` | 录制视频保存目录 | 路径 |
| `windows-service-video-save-directory` | Windows 服务模式下录制保存目录（需安装） | Windows 绝对路径 |
| `enable-abr` | 启用自适应比特率 | `Y`, `N` |

### 显示与性能

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `allow-remove-wallpaper` | 传入会话时移除壁纸 | `Y`, `N` |
| `allow-always-software-render` | 始终软件渲染 | `Y`, `N` |
| `enable-hwcodec` | 启用硬件编解码 | `Y`, `N` |
| `enable-directx-capture` | 启用 DirectX 捕获（Windows） | `Y`, `N` |
| `enable-android-software-encoding-half-scale` | Android 软件编码半分辨率 | `Y`, `N` |
| `allow-d3d-render` | 允许 D3D 渲染（>= 1.3.9，Windows） | `Y`, `N` |
| `use-texture-render` | 使用纹理渲染 | `Y`, `N` |

### 代理设置

| 配置项 | 说明 | 示例值 |
|--------|------|--------|
| `proxy-url` | 代理 URL（支持 http/https/socks5） | `http://proxy:8080` |
| `proxy-username` | 代理用户名 | `user` |
| `proxy-password` | 代理密码 | `pass` |

### 显示默认选项

以下配置项会设置每个 peer 首次连接后的默认值，之后每个 peer 的单独设置会覆盖这些默认值。

> **注意**：以下部分配置项的 key 使用下划线（`_`）而非连字符（`-`），这是由 RustDesk 客户端内部实现决定的。策略下发时 key 不做转换，直接使用原始字符串匹配，因此必须使用下划线版本才能生效。

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `view_only` | 只读模式 | `Y`, `N` |
| `show_monitors_toolbar` | 在工具栏显示显示器 | `Y`, `N` |
| `collapse_toolbar` | 连接后折叠工具栏 | `Y`, `N` |
| `show_remote_cursor` | 显示远程光标 | `Y`, `N` |
| `follow_remote_cursor` | 跟随远程光标（>= 1.4.4） | `Y`, `N` |
| `follow_remote_window` | 跟随远程窗口焦点（>= 1.4.4） | `Y`, `N` |
| `zoom-cursor` | 缩放光标 | `Y`, `N` |
| `show_quality_monitor` | 显示质量监视器 | `Y`, `N` |
| `disable_audio` | 静音 | `Y`, `N` |
| `enable-file-copy-paste` | 启用文件复制粘贴（Windows） | `Y`, `N` |
| `disable_clipboard` | 禁用剪贴板 | `Y`, `N` |
| `lock_after_session_end` | 会话结束后锁定 | `Y`, `N` |
| `privacy_mode` | 隐私模式 | `Y`, `N` |
| `i444` | 真彩 (4:4:4) | `Y`, `N` |
| `reverse_mouse_wheel` | 反转鼠标滚轮 | `Y`, `N` |
| `swap-left-right-mouse` | 交换左右鼠标按钮 | `Y`, `N` |
| `displays_as_individual_windows` | 将显示器显示为独立窗口 | `Y`, `N` |
| `use_all_my_displays_for_the_remote_session` | 使用所有显示器进行远程会话 | `Y`, `N` |
| `terminal-persistent` | 断开连接时保留终端会话 | `Y`, `N` |
| `view_style` | 默认视图样式 | `original`, `adaptive` |
| `scroll_style` | 默认滚动样式（>= 1.4.4 支持 `scrolledge`） | `scrollauto`, `scrollbar`, `scrolledge` |
| `edge-scroll-edge-thickness` | 边缘滚动厚度（>= 1.4.4） | `20`-`150` |
| `image_quality` | 默认图像质量 | `best`, `balanced`, `low`, `custom` |
| `custom_image_quality` | 自定义图像质量 | `[10.0, 2000.0]` |
| `custom-fps` | 自定义 FPS | `[5, 120]` |
| `codec-preference` | 默认编解码器 | `auto`, `vp8`, `vp9`, `av1`, `h264`, `h265` |
| `trackpad-speed` | 默认触控板速度 | `[10, 1000]` |

### 预设配置

| 配置项 | 说明 | 示例值 |
|--------|------|--------|
| `preset-address-book-name` | 预设地址簿名称 | `default` |
| `preset-address-book-tag` | 预设地址簿标签 | `work` |
| `preset-address-book-alias` | 预设地址簿别名（>= 1.4.3） | `Office PC` |
| `preset-address-book-password` | 预设地址簿密码（>= 1.4.3） | `password123` |
| `preset-address-book-note` | 预设地址簿备注（>= 1.4.3） | `My note` |
| `preset-user-name` | 预设用户名称 | `admin` |
| `preset-strategy-name` | 预设策略名称 | `default` |
| `preset-device-group-name` | 预设设备组名称（>= 1.3.8） | `servers` |
| `preset-device-username` | 预设设备用户名（>= 1.4.3） | `admin` |
| `preset-device-name` | 预设设备名称（>= 1.4.3） | `Office-PC` |
| `preset-note` | 预设设备备注（>= 1.4.3） | `Production server` |
| `default-connect-password` | 默认连接密码（控制端连接远程设备的密码） | `abcd1234` |
| `display-name` | 显示名称（连接时弹窗显示的名称） | `My PC` |
| `avatar` | 用户头像 | 头像路径或标识 |

### 其他设置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `enable-trusted-devices` | 启用信任设备（跳过 2FA） | `Y`, `N` |
| `allow-auto-update` | 允许自动更新（>= 1.4.6，Windows） | `Y`, `N` |
| `keep-awake-during-incoming-sessions` | 传入会话时保持唤醒 | `Y`, `N` |
| `keep-awake-during-outgoing-sessions` | 传出会话时保持唤醒 | `Y`, `N` |
| `enable-confirm-closing-tabs` | 关闭多个远程标签前确认 | `Y`, `N` |
| `enable-open-new-connections-in-tabs` | 新连接在标签页中打开 | `Y`, `N` |
| `allow-ask-for-note` | 连接结束时提示输入备注（>= 1.4.4） | `Y`, `N` |
| `pre-elevate-service` | Windows 便携版自动提权运行 | `Y`, `N` |
| `remove-preset-password-warning` | 移除预设密码的安全警告 | `Y`, `N` |
| `enable-check-update` | 启用检查更新 | `Y`, `N` |

### UI/界面设置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `theme` | UI 主题 | `dark`, `light`, `system` |
| `lang` | 界面语言 | `default`, `en`, `zh-cn`, `zh-tw`, `ja`, `ko`, `fr`, `de` 等 |
| `peer-card-ui-type` | Peer 卡片视图 | `0`（大图标）, `1`（小图标）, `2`（列表） |
| `peer-sorting` | Peer 排序方式 | `Remote ID`, `Remote Host`, `Username` |
| `sync-ab-with-recent-sessions` | 地址簿与最近会话同步 | `Y`, `N` |
| `sync-ab-tags` | 地址簿标签排序 | `Y`, `N` |
| `filter-ab-by-intersection` | 按标签交集过滤地址簿 | `Y`, `N` |
| `disable-group-panel` | 禁用分组面板 | `Y`, `N` |
| `disable-discovery-panel` | 禁用发现面板 | `Y`, `N` |
| `hide-tray` | 隐藏系统托盘图标 | `Y`, `N` |
| `hide-stop-service` | 隐藏停止/切换服务控件 | `Y`, `N` |
| `hide-general-settings` | 隐藏“常规”设置选项卡（>= 1.5.0） | `Y`, `N` |
| `hide-security-settings` | 隐藏“安全”设置选项卡 | `Y`, `N` |
| `hide-network-settings` | 隐藏“网络”设置选项卡 | `Y`, `N` |
| `hide-server-settings` | 隐藏“服务器”设置选项卡 | `Y`, `N` |
| `hide-proxy-settings` | 隐藏“代理”设置选项卡 | `Y`, `N` |
| `hide-websocket-settings` | 隐藏“WebSocket”设置选项卡 | `Y`, `N` |
| `hide-remote-printer-settings` | 隐藏“远程打印机”设置选项卡 | `Y`, `N` |
| `hide-username-on-card` | 隐藏设备列表中的用户名 | `Y`, `N` |
| `hide-help-cards` | 隐藏 UAC/权限警告 | `Y`, `N` |
| `hide-powered-by-me` | 隐藏 "Powered by Me" 标识 | `Y`, `N` |
| `main-window-always-on-top` | 主窗口始终置顶（1.4.2） | `Y`, `N` |

### 打印机设置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `printer-incomming-job-action` | 远程打印传入任务的处理动作 | 动作值 |
| `printer-selected-name` | 选中的打印机名称 | 打印机名称 |
| `allow-printer-auto-print` | 允许打印机自动打印 | `Y`, `N` |

### Android 特定设置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `disable-floating-window` | 禁用浮动窗口 | `Y`, `N` |
| `floating-window-size` | 浮动窗口大小 | `[32, 320]`，默认 `120` |
| `floating-window-untouchable` | 浮动窗口不可触摸（点击穿透） | `Y`, `N` |
| `floating-window-transparency` | 浮动窗口透明度 | `[0, 10]`，默认 `10` |
| `floating-window-svg` | 浮动窗口图标 | SVG 文本（单行） |
| `keep-screen-on` | 保持屏幕常亮 | `never`, `during-controlled`, `service-on` |
| `touch-mode` | 触摸模式（>= 1.4.3 统一控制） | `Y`, `N` |
| `show-virtual-mouse` | 显示虚拟鼠标（>= 1.4.3） | `Y`, `N` |
| `show-virtual-joystick` | 显示虚拟摇杆（>= 1.4.3，需启用虚拟鼠标） | `Y`, `N` |

---

## 配置示例

### 示例 1：只读访问模式

```json
{
  "config_options": {
    "access-mode": "view",
    "enable-keyboard": "N",
    "enable-clipboard": "N",
    "enable-file-transfer": "N"
  }
}
```

### 示例 2：启用自动录制

```json
{
  "config_options": {
    "enable-record-session": "Y",
    "allow-auto-record-incoming": "Y"
  }
}
```

### 示例 3：限制特定 IP 访问

```json
{
  "config_options": {
    "whitelist": "192.168.1.100,192.168.1.101"
  }
}
```

### 示例 4：禁用不必要的功能

```json
{
  "config_options": {
    "enable-camera": "N",
    "enable-terminal": "N",
    "enable-remote-printer": "N",
    "enable-tunnel": "N",
    "enable-remote-restart": "N"
  }
}
```

### 示例 5：配置服务器地址

```json
{
  "config_options": {
    "custom-rendezvous-server": "rs.example.com",
    "api-server": "api.example.com",
    "key": "your-secret-key",
    "relay-server": "relay.example.com"
  }
}
```

---

## 注意事项

1. **配置生效时间**：策略下发后，被控端会在下次心跳时收到并应用，最长 15 秒
2. **配置覆盖**：策略配置会覆盖被控端本地配置
3. **空值处理**：如果策略中某个配置项的值为空字符串，且默认高级设置也为空，则会删除该配置项（回退到内置默认值）
4. **默认策略**：默认策略的 `peer_id` 为空，列表接口不会显示默认策略
5. **设备策略唯一性**：每个设备只能有一条策略，重复创建会返回错误

---

## 数据模型

### PeerStrategy

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 主键 |
| `peer_id` | string | 设备 ID（默认策略为空） |
| `config_options` | object | 配置项 JSON |
| `modified_at` | int64 | 策略版本时间戳 |
| `created_at` | timestamp | 创建时间 |
| `updated_at` | timestamp | 更新时间 |
| `peer_alias` | string | 关联的设备别名（仅列表接口返回） |
