# CLI Proxy API

[English](README_EN.md) | 中文

一个为 CLI 提供 OpenAI/Gemini/Claude/Codex 兼容 API 接口的代理服务器。

支持通过 OAuth 登录使用 OpenAI Codex (GPT 模型) 和 Claude Code。

可以使用本地或多账户 CLI 访问 OpenAI(包括 Responses)/Gemini/Claude 兼容的客户端和 SDK。

## 功能概览

- CLI 模型的 OpenAI/Gemini/Claude 兼容 API 端点
- 通过 OAuth 登录支持 OpenAI Codex (GPT 模型)
- 通过 OAuth 登录支持 Claude Code
- 通过 OAuth 登录支持 Qwen Code
- 通过 OAuth 登录支持 iFlow
- 支持 Amp CLI 和 IDE 扩展的提供商路由
- 流式和非流式响应
- 函数调用/工具支持
- 多模态输入支持（文本和图像）
- 多账户轮询负载均衡（Gemini、OpenAI、Claude、Qwen 和 iFlow）
- 简单的 CLI 认证流程
- Generative Language API Key 支持
- 多账户负载均衡
- 通过配置支持 OpenAI 兼容的上游提供商
- 可复用的 Go SDK 用于嵌入代理
- **捐赠站点** - Linux Do OAuth 登录与 New-API 集成的额度管理

## 部署教程

### 方式一：Docker Compose 部署（推荐）

1. 克隆仓库并进入目录：

```bash
git clone https://github.com/router-for-me/CLIProxyAPI.git
cd CLIProxyAPI
```

2. 复制配置文件：

```bash
cp config.example.yaml config.yaml
```

3. 编辑 `config.yaml`，配置必要参数：

```yaml
# 服务端口
port: 8317

# API 密钥（用于客户端认证）
api-keys:
  - "your-api-key-1"

# 认证文件目录
auth-dir: "~/.cli-proxy-api"
```

4. 启动服务：

```bash
docker-compose up -d
```

5. 查看日志：

```bash
docker-compose logs -f
```

### 方式二：二进制部署

1. 从 [Releases](https://github.com/router-for-me/CLIProxyAPI/releases) 下载对应平台的二进制文件

2. 创建配置文件 `config.yaml`（参考 `config.example.yaml`）

3. 运行：

```bash
./CLIProxyAPI
```

### 方式三：源码编译

1. 确保已安装 Go 1.24+

2. 克隆并编译：

```bash
git clone https://github.com/router-for-me/CLIProxyAPI.git
cd CLIProxyAPI
go build -o CLIProxyAPI ./cmd/server/
```

3. 运行：

```bash
./CLIProxyAPI
```

## 端口说明

| 端口 | 用途 |
|------|------|
| 8317 | 主 API 服务端口 |
| 8085 | 管理面板端口 |
| 1455 | Claude Code 端口 |
| 54545 | Gemini CLI 端口 |
| 51121 | Qwen Code 端口 |
| 11451 | iFlow 端口 |

## 配置说明

### 基础配置

```yaml
# 服务绑定地址，留空绑定所有接口
host: ""

# 服务端口
port: 8317

# TLS 配置（可选）
tls:
  enable: false
  cert: ""
  key: ""

# API 密钥列表
api-keys:
  - "your-api-key-1"
  - "your-api-key-2"

# 认证文件目录
auth-dir: "~/.cli-proxy-api"

# 调试模式
debug: false
```

### 管理 API 配置

```yaml
remote-management:
  # 是否允许远程访问管理 API
  allow-remote: false
  # 管理密钥（必填，否则管理 API 不可用）
  secret-key: "your-secret-key"
  # 禁用控制面板
  disable-control-panel: false
```

### 代理配置

```yaml
# 全局代理 URL（支持 socks5/http/https）
proxy-url: "socks5://user:pass@192.168.1.1:1080"

# 请求重试次数
request-retry: 3

# 最大重试等待时间（秒）
max-retry-interval: 30
```

## 捐赠站点

CLIProxyAPI 内置捐赠站点功能，允许用户：

- 通过 Linux Do Connect OAuth 登录
- 绑定 New-API 用户账户
- 确认捐赠后获得额度奖励

### 捐赠站点配置

在 `config.yaml` 中添加：

```yaml
linux-do-connect:
  client-id: "your-client-id"
  client-secret: "your-client-secret"
  redirect-uri: "https://your-domain.com/linuxdo/callback"

donation:
  quota-amount: 2000000  # 额度单位，2000000 = $20
  admin-linux-do-ids:
    - 12345  # 管理员的 Linux Do ID
```

设置环境变量（或在 `.env` 文件中）：

```bash
NEW_API_BASE_URL=https://your-newapi-instance.com
NEW_API_ADMIN_TOKEN=your-admin-token
```

### 捐赠站点 API 端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/` | GET | 捐赠站点主页 |
| `/linuxdo/login` | GET | 发起 Linux Do OAuth 登录 |
| `/linuxdo/callback` | GET | OAuth 回调处理 |
| `/status` | GET | 获取当前登录状态 |
| `/bind` | GET/POST | 查看/提交 New-API 账户绑定 |
| `/donate` | GET | 查看捐赠信息 |
| `/donate/confirm` | POST | 确认捐赠并获得额度 |
| `/logout` | POST | 退出登录并清除会话 |

### 访问控制

- 管理员用户（通过 `admin-linux-do-ids` 配置）拥有 auth 文件管理的完整权限
- 普通用户无法访问 auth 文件操作（返回 403 Forbidden）

## 环境变量

| 变量名 | 描述 |
|--------|------|
| `DEPLOY` | 部署标识 |
| `NEW_API_BASE_URL` | New-API 服务地址 |
| `NEW_API_ADMIN_TOKEN` | New-API 管理员令牌 |

## 目录结构

```
CLIProxyAPI/
├── config.yaml          # 配置文件
├── auths/               # OAuth 认证文件目录
├── logs/                # 日志目录
└── CLIProxyAPI          # 可执行文件
```

## SDK 文档

- 使用说明：[docs/sdk-usage_CN.md](docs/sdk-usage_CN.md)
- 高级用法：[docs/sdk-advanced_CN.md](docs/sdk-advanced_CN.md)
- 访问控制：[docs/sdk-access_CN.md](docs/sdk-access_CN.md)
- 监视器：[docs/sdk-watcher_CN.md](docs/sdk-watcher_CN.md)

## 贡献

欢迎贡献！请随时提交 Pull Request。

1. Fork 本仓库
2. 创建你的功能分支（`git checkout -b feature/amazing-feature`）
3. 提交你的更改（`git commit -m 'Add some amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。
