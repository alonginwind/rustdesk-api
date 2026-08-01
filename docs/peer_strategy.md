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
| `access-mode` | 访问模式 | `full`, `view` |
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

### 安全设置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `approve-mode` | 审批模式 | `password`, `click`, `both` |
| `verification-method` | 验证方式 | `use-temporary-password`, `use-permanent-password`, `use-both-passwords` |
| `temporary-password-length` | 临时密码长度 | `6`, `8`, `10` |
| `allow-numeric-one-time-password` | 允许数字一次性密码 | `Y`, `N` |
| `allow-scope-violation-close` | 允许违规关闭 | `Y`, `N` |
| `allow-scope-violation-alarm` | 允许违规告警 | `Y`, `N` |
| `allow-remote-config-modification` | 允许远程修改配置 | `Y`, `N` |

### 网络设置

| 配置项 | 说明 | 示例值 |
|--------|------|--------|
| `custom-rendezvous-server` | 自定义中继服务器 | `rs.example.com` |
| `api-server` | API 服务器 | `api.example.com` |
| `key` | 连接密钥 | `your-secret-key` |
| `relay-server` | 中继服务器 | `relay.example.com` |
| `ice-servers` | ICE 服务器 | JSON 格式 |
| `direct-server` | 直连服务器 | `direct.example.com` |
| `direct-access-port` | 直连访问端口 | `21118` |
| `allow-websocket` | 允许 WebSocket | `Y`, `N` |
| `disable-udp` | 禁用 UDP | `Y`, `N` |
| `allow-insecure-tls-fallback` | 允许不安全 TLS 回退 | `Y`, `N` |

### 连接控制

| 配置项 | 说明 | 可选值/示例 |
|--------|------|-------------|
| `enable-lan-discovery` | 启用局域网发现 | `Y`, `N` |
| `whitelist` | 白名单 | IP 列表，逗号分隔 |
| `allow-auto-disconnect` | 允许自动断开 | `Y`, `N` |
| `auto-disconnect-timeout` | 自动断开超时（秒） | `600` |
| `allow-only-conn-window-open` | 仅窗口打开时允许连接 | `Y`, `N` |

### 录制设置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `allow-auto-record-incoming` | 自动录制传入连接 | `Y`, `N` |
| `enable-abr` | 启用自适应比特率 | `Y`, `N` |

### 显示与性能

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `allow-remove-wallpaper` | 允许移除壁纸 | `Y`, `N` |
| `allow-always-software-render` | 允许始终软件渲染 | `Y`, `N` |
| `allow-linux-headless` | 允许 Linux 无头模式 | `Y`, `N` |
| `enable-hwcodec` | 启用硬件编解码 | `Y`, `N` |
| `enable-directx-capture` | 启用 DirectX 捕获 | `Y`, `N` |
| `enable-android-software-encoding-half-scale` | Android 软件编码半分辨率 | `Y`, `N` |

### 代理设置

| 配置项 | 说明 | 示例值 |
|--------|------|--------|
| `proxy-url` | 代理 URL | `http://proxy:8080` |
| `proxy-username` | 代理用户名 | `user` |
| `proxy-password` | 代理密码 | `pass` |

### 预设配置

| 配置项 | 说明 | 示例值 |
|--------|------|--------|
| `preset-address-book-name` | 预设地址簿名称 | `default` |
| `preset-address-book-tag` | 预设地址簿标签 | `work` |
| `preset-address-book-alias` | 预设地址簿别名 | `Office PC` |
| `preset-address-book-password` | 预设地址簿密码 | `password123` |
| `preset-address-book-note` | 预设地址簿备注 | `My note` |
| `preset-device-username` | 预设设备用户名 | `admin` |
| `preset-device-name` | 预设设备名称 | `Office-PC` |
| `preset-note` | 预设备注 | `Production server` |

### 其他设置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `enable-trusted-devices` | 启用信任设备 | `Y`, `N` |
| `allow-auto-update` | 允许自动更新 | `Y`, `N` |
| `keep-awake-during-incoming-sessions` | 传入会话时保持唤醒 | `Y`, `N` |

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
