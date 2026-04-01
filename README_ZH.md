# RcloneFlow

基于 Web 的 Rclone 管理界面，用于多存储复制/同步/移动任务管理。

[English](README.md)

## 功能特点

- **多存储管理** - 添加、编辑和管理多个 Rclone 存储
- **文件浏览器** - 浏览和导航远程存储文件
- **任务管理** - 创建和管理存储间的复制/同步任务
- **定时任务** - 使用 cron 风格的调度设置自动化同步
- **运行记录** - 跟踪任务执行历史和状态
- **现代化界面** - 简洁、响应式的 Web 界面

## 系统要求

- Go 1.22+
- Rclone (需要开启 RC 模式)
- Git

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/ray5378/RcloneFlow.git
cd RcloneFlow
```

### 2. 配置 Rclone

确保已安装 Rclone 并配置好存储。配置文件通常在 `~/.config/rclone/rclone.conf`，或通过 `RCLONE_CONFIG` 环境变量指定。

### 3. 启动 Rclone RC 服务器

```bash
rclone rcd --rc-user=your_user --rc-pass=your_pass --rc-addr=localhost:5572
```

或使用环境变量：
```bash
export RCLONE_RC_URL=http://localhost:5572
export RCLONE_RC_USER=your_user
export RCLONE_RC_PASS=your_pass
```

### 4. 构建并运行

```bash
# 构建
go build -o server ./cmd/server

# 运行
./server
```

服务器默认在 17870 端口启动，访问 http://localhost:17870

### 5. 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `APP_ADDR` | 服务器地址 | `:17870` |
| `APP_DATA_DIR` | 数据目录 | `./data` |
| `RCLONE_RC_URL` | Rclone RC 地址 | `http://127.0.0.1:5572` |
| `RCLONE_RC_USER` | Rclone RC 用户名 | - |
| `RCLONE_RC_PASS` | Rclone RC 密码 | - |
| `RCLONE_RC_TIMEOUT` | RC 超时时间 | `120s` |

## Docker 部署

### 构建镜像

```bash
docker build -t rcloneflow .
```

### 运行容器

```bash
docker run -d \
  --name rcloneflow \
  -p 17870:17870 \
  -e RCLONE_RC_URL=http://host.docker.internal:5572 \
  -e RCLONE_RC_USER=your_user \
  -e RCLONE_RC_PASS=your_pass \
  -v /path/to/rclone/config:/root/.config/rclone \
  rcloneflow
```

### Docker Compose

```yaml
version: '3.8'
services:
  rcloneflow:
    build: .
    ports:
      - "17870:17870"
    environment:
      - RCLONE_RC_URL=http://rclone:5572
      - RCLONE_RC_USER=your_user
      - RCLONE_RC_PASS=your_pass
    volumes:
      - ./data:/app/data
      - /path/to/rclone/config:/root/.config/rclone
    depends_on:
      - rclone

  rclone:
    image: rclone/rclone
    container_name: rclone
    volumes:
      - /path/to/rclone/config:/config/rclone
      - /path/to/your/data:/data
    command: rcd --rc-user=your_user --rc-pass=your_pass --rc-addr=0.0.0.0:5572
```

## 项目结构

```
RcloneFlow/
├── cmd/
│   └── server/          # 主应用程序入口
├── internal/
│   ├── app/             # HTTP 服务器和 API 处理器
│   ├── rclone/          # Rclone RC 客户端封装
│   ├── scheduler/       # 任务调度逻辑
│   └── store/           # 数据持久化 (SQLite)
├── web/
│   ├── index.html       # 前端单页应用 (Vue.js)
│   └── vendor/          # Vue.js CDN 包
├── data/                # 应用数据目录
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## API 接口

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/remotes` | 列出所有存储 |
| POST | `/api/remotes` | 创建新存储 |
| PUT | `/api/remotes` | 更新存储配置 |
| GET | `/api/remotes/config/{name}` | 获取存储配置 |
| POST | `/api/remotes/test` | 测试存储连接 |
| GET | `/api/providers` | 获取支持的存储类型 |
| GET | `/api/browser/list` | 列出目录内容 |
| GET/POST | `/api/tasks` | 列出/创建任务 |
| POST | `/api/tasks/{id}/run` | 运行任务 |
| GET/POST | `/api/schedules` | 列出/创建定时任务 |
| GET | `/api/runs` | 列出运行历史 |

## 开发

### 构建前端（可选）

前端作为单 HTML 文件打包，内嵌 Vue.js。开发时：

```bash
# 前端直接从 web/index.html 提供服务
# 基础开发无需构建步骤
```

### 运行测试

```bash
go test ./...
```

## 贡献

欢迎提交 Pull Request！

## 开源协议

MIT License - 详见 LICENSE 文件。

## 致谢

- [Rclone](https://rclone.org/) - 强大的云存储同步工具
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
