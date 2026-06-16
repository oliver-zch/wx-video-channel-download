# wx-video-channel-download

视频号视频下载工具 —— 前后端分离的 Web 服务，支持通过分享链接解析视频信息、在线播放和下载。

## 功能

- 🔗 粘贴视频号分享链接，自动解析视频元数据（标题、封面、作者、互动数据等）
- ▶️ 在线播放视频（流式解密代理）
- ⬇️ 下载视频到本地
- 🔑 运行时更新 SPH Cookie，无需重启服务
- 📦 单二进制部署，前端内嵌，无外部依赖

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go + Gin |
| 前端 | 原生 HTML/CSS/JS（无构建步骤） |
| 解密 | ISAAC64 流式 XOR 解密 |
| 部署 | 单二进制 |

## 项目结构

```
wx_channels_web/
├── main.go                          # 入口，嵌入前端页面
├── config.yaml                      # 配置文件
├── go.mod / go.sum                  # Go 依赖
├── internal/
│   ├── api/
│   │   ├── server.go                # Gin 引擎、CORS、路由注册
│   │   ├── handler.go               # API 处理函数
│   │   └── types.go                 # 请求/响应类型定义
│   ├── service/
│   │   ├── sph.go                   # SPH 分享链接解析（yuanbao → channels API）
│   │   ├── decryptor.go             # 视频流式解密代理
│   │   └── reader.go                # ISAAC64 流式解密 Reader
│   └── config/
│       └── config.go                # 配置加载与运行时更新
├── pkg/
│   └── decrypt/
│       └── decrypt.go               # ISAAC64 密码算法实现
└── web/
    └── index.html                   # 前端单页面（通过 go:embed 内嵌）
```

## 快速开始

### 1. 获取代码

```bash
git clone <your-repo-url>
cd wx_channels_web
```

### 2. 配置

编辑 `config.yaml`：

```yaml
api:
  hostname: "0.0.0.0"
  port: 2022

# yuanbao.tencent.com 的 session cookie
# 获取方式见下方 "Cookie 获取" 章节
sph_cookie: ""
```

### 3. 运行

```bash
# 安装依赖
go mod tidy

# 直接运行
go run .

# 或构建后运行
go build -o wx-video-channel-download .
./wx-video-channel-download
```

访问 `http://localhost:2022` 即可使用。

## Cookie 获取

SPH 解析依赖 [yuanbao.tencent.com](https://yuanbao.tencent.com) 的 session cookie：

1. 浏览器打开 [yuanbao.tencent.com](https://yuanbao.tencent.com)，用微信登录
2. 按 `F12` 打开开发者工具
3. 在地址栏输入 `copy(document.cookie)` 并回车，cookie 已复制到剪贴板
4. 粘贴到网页底部的 Cookie 设置框，点击"更新 Cookie"

> **注意**：Cookie 会过期（通常 7-30 天）。解析失败时请重新获取。

## API 文档

### 健康检查

```
GET /api/status
```

响应：
```json
{ "code": 0, "msg": "ok", "data": { "version": "1.0.0" } }
```

### 解析视频

```
POST /api/parse
Content-Type: application/json

{ "url": "https://weixin.qq.com/sph/xxxxx" }
```

响应：
```json
{
  "code": 0,
  "msg": "成功",
  "data": {
    "title": "视频描述",
    "author": "作者昵称",
    "author_avatar": "https://...",
    "cover_url": "https://...",
    "video_url": "https://...",
    "decrypt_key": "1234567890",
    "media_type": 1,
    "like_count": "1.2万",
    "fav_count": "3000",
    "comment_count": "500",
    "forward_count": "100",
    "create_time": 1700000000,
    "h264_url": "https://...",
    "h265_url": "https://...",
    "original_url": "https://..."
  }
}
```

### 代理播放 / 下载

```
GET /api/proxy?url=<视频URL>&key=<解密密钥>
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `url` | 是 | 视频 URL（来自 parse 接口） |
| `key` | 否 | 解密密钥，为空则直接代理不解密 |
| `filename` | 否 | 设置后触发文件下载 |

### 更新 Cookie

```
POST /api/config/cookie
Content-Type: application/json

{ "cookie": "hy_source=...; hy_user=...; hy_token=..." }
```

## 部署

### 直接部署

```bash
# 交叉编译 Linux 版本
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o wx_channels_web .

# 上传到服务器
scp wx_channels_web config.yaml user@server:/opt/wx_channels_web/

# 运行
ssh user@server "cd /opt/wx_channels_web && ./wx_channels_web"
```


## 技术原理

### 视频号加密机制

微信视频号的视频采用 ISAAC64 流密码加密，仅加密前 128KB（131072 字节），其余部分为明文。

### 解析流程

```
用户粘贴分享链接
    ↓
yuanbao.tencent.com 解析 → exportId + generalToken
    ↓
channels.weixin.qq.com 获取视频信息 → 视频 URL + 元数据
    ↓
前端展示 / 代理播放 / 下载
```

### 代理解密

视频代理端点实时流式解密：

1. 接收客户端请求，转发 Range header 到上游
2. 创建 ISAAC64 解密上下文，根据偏移量跳过 keystream 块
3. 流式 XOR 解密前 128KB，超过部分直接透传
4. 支持 HTTP Range 请求，实现视频拖拽 seek

## 致谢

- [wx_channels_download](https://github.com/ltaoo/wx_channels_download) - 原始项目
- [WechatSphDecrypt](https://github.com/Hanson/WechatSphDecrypt) - ISAAC64 解密算法

## License

MIT
