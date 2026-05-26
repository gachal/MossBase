# MossBase

<img src="logo.png" width="200" />

> 苔藓（Moss）+ 知识库（Base），小而有机。

企业级 Wiki 知识库系统。

## 功能特性

- **Markdown 编辑器** — 基于 md-editor-v3 的纯 Markdown 编写体验
- **知识空间** — 多空间隔离，支持空间级成员权限（管理员/成员/查看者）
- **页面树** — 无限层级树形结构，拖拽排序
- **版本管理** — 页面历史版本查看与对比
- **RAG 语义搜索** — 基于 Qdrant 向量数据库的智能语义检索（可选）
- **MCP 集成** — 通过 MCP 协议让 AI 工具直接操作知识库（可选）
- **安装向导** — 首次启动时自动进入 Web 安装引导

## 技术栈

| 层 | 技术 |
|------|------|
| 后端 | Go 1.22+ (Gin + GORM) + MySQL 8.0 + JWT |
| 前端 | Vue 3 + TypeScript + Pinia + Element Plus + md-editor-v3 + Vite |
| RAG（可选） | Qdrant + Redis + OpenAI 兼容嵌入 API |
| MCP（可选） | Go MCP SDK (stdio / HTTP) |
| 部署 | Docker Compose / 二进制 + Nginx |

## 快速开始

### 方式一：Docker Compose 部署（推荐）

#### 1. 克隆项目

```bash
git clone https://github.com/gachal/MossBase.git
cd MossBase
```

#### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env`，填写以下必填项：

```bash
# 必填：MySQL root 密码
MYSQL_ROOT_PASSWORD=your-mysql-password

# 必填：JWT 签名密钥（建议 32 位以上随机字符串）
JWT_SECRET=your-jwt-secret

# 必填：RAG 服务 API Key（即使不启用 RAG 也需要设置）
RAG_API_KEY=your-rag-api-key

# 必填：MCP 服务 API Key
MCP_API_KEY=your-mcp-api-key
```

#### 3. 启动所有服务

```bash
docker-compose up -d
```

#### 4. 访问系统

浏览器打开 `http://localhost`，首次访问将自动进入安装向导。

安装向导将引导你完成：

1. 配置数据库连接（Docker 部署时已预配置）
2. 创建管理员账号

> Docker 模式下后端运行在 8033 端口，前端通过 Nginx 反向代理访问。无需单独访问后端端口。

### 方式二：二进制部署

适用于不使用 Docker 的环境（如宝塔面板、传统服务器）。

#### 前置条件

- Go 1.22+（仅编译时需要）
- Node.js 18+（仅前端构建时需要）
- MySQL 8.0
- Nginx（用于反向代理）

#### 1. 准备 MySQL 数据库

```sql
CREATE DATABASE mossbase CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER 'mossbase'@'%' IDENTIFIED BY 'your-password';
GRANT ALL PRIVILEGES ON mossbase.* TO 'mossbase'@'%';
FLUSH PRIVILEGES;
```

#### 2. 编译后端

```bash
cd backend
go mod tidy

# 编译主服务
CGO_ENABLED=0 go build -o mossbase-server ./cmd/server

# 编译 MCP 服务（可选）
CGO_ENABLED=0 go build -o mossbase-mcp-server ./cmd/mcp-server
```

#### 3. 编译前端

```bash
cd frontend
npm ci
npm run build
```

构建产物在 `frontend/dist/` 目录。

#### 4. 配置后端

编辑 `backend/configs/config.yaml`：

```yaml
server:
  port: 8033
  mode: release

database:
  host: 127.0.0.1
  port: 3306
  username: mossbase
  password: "your-password"
  dbname: mossbase
```

也可通过环境变量覆盖配置，前缀为 `MOSS_`，例如 `MOSS_DATABASE_PASSWORD=xxx`。

#### 5. 启动服务

```bash
# 首次运行会进入安装向导模式
./mossbase-server

# 安装向导完成后服务会自动退出，重新启动即可进入正常模式
./mossbase-server
```

#### 6. 配置 Nginx

参见下方的 [Nginx 反向代理部署](#nginx-反向代理部署) 章节。

## 宝塔面板部署

宝塔面板是国内常用的服务器管理面板，以下介绍如何在宝塔面板中部署 MossBase。

### 1. 安装基础环境

在宝塔面板「软件商店」中安装：

- **Nginx** 1.24+
- **MySQL** 8.0

如宝塔不提供 Go 环境，通过命令行安装：

```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. 创建数据库

在宝塔面板「数据库」页面：

1. 点击「添加数据库」
2. 数据库名：`mossbase`
3. 用户名：`mossbase`
4. 密码：自动生成或自定义
5. 访问权限：本地服务器
6. 记录下数据库名、用户名、密码

### 3. 编译项目

```bash
cd /www/wwwroot
git clone https://github.com/gachal/MossBase.git
cd MossBase

# 编译后端
cd backend
go mod tidy
CGO_ENABLED=0 go build -o mossbase-server ./cmd/server

# 编译前端
cd ../frontend
npm ci
npm run build
```

### 4. 配置后端

编辑 `/www/wwwroot/MossBase/backend/configs/config.yaml`：

```yaml
server:
  port: 8033
  mode: release

database:
  host: 127.0.0.1
  port: 3306
  username: mossbase
  password: "宝塔面板中设置的数据库密码"
  dbname: mossbase
```

### 5. 创建网站

在宝塔面板「网站」页面：

1. 点击「添加站点」
2. 域名：填写你的域名（如 `wiki.example.com`）
3. 根目录：`/www/wwwroot/MossBase/frontend/dist`
4. PHP 版本：纯静态
5. 提交

### 6. 配置反向代理

在网站设置中点击「反向代理」选项卡，添加反向代理：

- 代理名称：`mossbase-api`
- 目标 URL：`http://127.0.0.1:8033`
- 发送域名：`$host`
- 代理目录：`/api`

保存后宝塔会自动生成 Nginx 配置。如需手动调整，点击「配置文件」选项卡，确保包含：

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8033;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

location / {
    try_files $uri $uri/ /index.html;
}
```

### 7. 配置进程守护

建议使用进程守护保持后端持续运行。

**方式一：宝塔进程守护管理器**

在宝塔「软件商店」搜索并安装「进程守护管理器」，添加守护进程：

- 启动命令：`/www/wwwroot/MossBase/backend/mossbase-server`
- 运行目录：`/www/wwwroot/MossBase/backend`
- 进程名：`mossbase`

**方式二：systemd**

创建 `/etc/systemd/system/mossbase.service`：

```ini
[Unit]
Description=MossBase Server
After=network.target mysql.service

[Service]
Type=simple
WorkingDirectory=/www/wwwroot/MossBase/backend
ExecStart=/www/wwwroot/MossBase/backend/mossbase-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable mossbase
systemctl start mossbase
```

### 8. 完成安装

1. 浏览器访问你的域名
2. 首次访问进入安装向导，按提示完成安装
3. 安装向导完成后服务会自动退出，进程守护会自动重启
4. 重新访问域名，开始使用

## Nginx 反向代理部署

如果你使用独立 Nginx（非宝塔面板），以下是一个完整的配置示例。

### 1. 部署前端静态文件

```bash
cd frontend
npm ci
npm run build

sudo mkdir -p /usr/share/nginx/html/mossbase
sudo cp -r dist/* /usr/share/nginx/html/mossbase/
```

### 2. Nginx 配置

创建 `/etc/nginx/conf.d/mossbase.conf`：

```nginx
server {
    listen 80;
    server_name wiki.example.com;

    root /usr/share/nginx/html/mossbase;
    index index.html;

    # API 反向代理
    location /api/ {
        proxy_pass http://127.0.0.1:8033;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # SPA 路由回退
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

### 3. 启用 HTTPS（推荐）

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d wiki.example.com
```

### 4. 重载 Nginx

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## MCP 集成

MossBase 内置 MCP (Model Context Protocol) 服务，支持 AI 工具（Claude Desktop、Cursor 等）直接操作知识库。

### 功能概览

- **11 个工具** — 页面 CRUD、空间管理、关键词搜索、语义搜索
- **资源模板** — `mossbase://spaces/{spaceID}/pages/{pageID}`
- **4 个提示词** — 总结页面、搜索回答、解释内容、大纲扩展

### 传输模式

| 模式 | 用途 | 说明 |
|------|------|------|
| `stdio` | 本地 AI 工具 | Claude Desktop、Cursor 等 |
| `http` | 远程部署 | 通过 HTTP Streamable 传输 |
| `both` | 同时支持 | 两种传输同时运行 |

### 快速启用

编辑 `backend/configs/config.yaml`：

```yaml
mcp:
  enabled: true
  transport: "http"
  http_port: 8095
  api_keys:
    - "your-secure-api-key"
```

### 配置 Claude Desktop

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`（macOS）或对应路径：

```json
{
  "mcpServers": {
    "mossbase": {
      "command": "/path/to/mossbase-mcp-server",
      "env": {
        "MOSS_MCP_ENABLED": "true",
        "MOSS_MCP_TRANSPORT": "stdio",
        "MOSS_DATABASE_HOST": "127.0.0.1",
        "MOSS_DATABASE_PORT": "3306",
        "MOSS_DATABASE_USERNAME": "root",
        "MOSS_DATABASE_PASSWORD": "your-password",
        "MOSS_DATABASE_DBNAME": "mossbase"
      }
    }
  }
}
```

> 详细文档：[MCP 工具参考](docs/mcp/tools.md) | [MCP 部署指南](docs/mcp/deployment.md) | [MCP 资源参考](docs/mcp/resources.md) | [MCP 提示词参考](docs/mcp/prompts.md)

## RAG 语义搜索

MossBase RAG（检索增强生成）服务提供基于向量数据库的语义搜索能力，支持知识库内容的智能检索。

### 架构概览

```
用户 → Frontend → Backend → RAG Service → Qdrant (向量存储)
                                        → Redis (异步任务)
                                        → Embedding Provider (文本向量化)
```

### 组件

| 组件 | 用途 |
|------|------|
| RAG Service | 独立微服务，文档索引与语义搜索 |
| Qdrant | 高性能向量数据库 |
| Redis | 异步任务队列 (Asynq) |
| Embedding Provider | OpenAI 兼容 API（支持 OpenAI / 中转服务 / Ollama / LocalAI） |

### 快速启用

1. 在 `.env` 中配置嵌入服务：

```bash
OPENAI_API_KEY=sk-your-api-key
# 国内用户可使用中转服务
# OPENAI_BASE_URL=https://api.api2d.com/v1
```

2. 启动 Docker Compose 时确保包含 `rag`、`qdrant`、`redis` 服务。

3. 在后端配置中启用：

```yaml
rag:
  enabled: true
  base_url: "http://rag:8090"    # Docker 网络
  api_key: "${RAG_API_KEY}"
```

> RAG 为可选组件。未启用时，MossBase 核心功能不受影响，搜索将降级为数据库关键词搜索。
>
> 详细文档：[RAG 架构文档](docs/rag/architecture.md) | [RAG 部署指南](docs/rag/deployment.md) | [RAG API 参考](docs/rag/api.md) | [RAG 集成指南](docs/rag/integration-guide.md)

## 环境变量

所有配置均可通过环境变量覆盖，前缀为 `MOSS_`。完整列表见 [.env.example](.env.example)。

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `MOSS_SERVER_PORT` | 否 | `8033` | 后端监听端口 |
| `MOSS_SERVER_MODE` | 否 | `debug` | 运行模式：debug / release |
| `MOSS_DATABASE_HOST` | 否 | `127.0.0.1` | MySQL 主机 |
| `MOSS_DATABASE_PORT` | 否 | `3306` | MySQL 端口 |
| `MOSS_DATABASE_USERNAME` | 否 | `root` | MySQL 用户名 |
| `MOSS_DATABASE_PASSWORD` | 是 | | MySQL 密码 |
| `MOSS_DATABASE_DBNAME` | 否 | `mossbase` | 数据库名 |
| `MOSS_JWT_SECRET` | 是 | | JWT 签名密钥 |
| `MOSS_RAG_ENABLED` | 否 | `false` | 是否启用 RAG |
| `MOSS_RAG_BASE_URL` | 否 | `http://127.0.0.1:8090` | RAG 服务地址 |
| `MOSS_MCP_ENABLED` | 否 | `false` | 是否启用 MCP |
| `MOSS_MCP_TRANSPORT` | 否 | `stdio` | MCP 传输模式 |

## 开发

```bash
# 后端
cd backend
go mod tidy
go run cmd/server/main.go        # 启动后端 (:8033)

# 前端
cd frontend
npm install
npm run dev                      # 启动前端 (:5173，自动代理 /api 到 :8033)
```

## 项目结构

```
MossBase/
├── backend/                    # Go 后端
│   ├── cmd/server/             # 主服务入口
│   ├── cmd/mcp-server/         # MCP 服务入口
│   ├── configs/config.yaml     # 配置文件
│   ├── migrations/             # 数据库迁移
│   └── internal/               # 业务代码（DDD 分层）
├── frontend/                   # Vue 3 前端
│   ├── src/views/              # 页面视图
│   ├── src/components/         # 组件
│   ├── src/stores/             # Pinia 状态管理
│   └── nginx.conf              # 生产环境 Nginx 配置
├── docs/                       # 文档
│   ├── mcp/                    # MCP 详细文档
│   └── rag/                    # RAG 详细文档
├── docker-compose.yml          # Docker Compose 编排
└── .env.example                # 环境变量模板
```

## License

[MIT](LICENSE) Copyright 2026, gachal
