# RcloneFlow

基于 Web 的 Rclone 管理界面，用于多存储复制/同步/移动任务管理。

[English](README_EN.md)

## 功能特点

- **多存储管理** - 添加、编辑和管理多个 Rclone 存储，支持 SMB/CIFS 等协议
- **文件浏览器** - 浏览和导航远程存储文件，支持右键菜单操作（复制/移动/删除/重命名）
- **任务管理** - 创建和管理存储间的复制/同步/移动任务
- **定时任务** - 使用 cron 风格的调度设置自动化同步
- **运行记录** - 跟踪任务执行历史和实时状态
- **实时状态同步** - 后台自动同步 rclone job 状态
- **现代化界面** - 简洁、响应式的 Web 界面，支持深色/浅色模式
- **统一错误处理** - Toast 通知，友好的错误提示
- **单元测试** - 前后端完整的单元测试覆盖

## 系统要求

- Go 1.22+
- Rclone (需要开启 RC 模式)
- Git
- Node.js 18+ (前端开发)

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/ray5378/RcloneFlow.git
cd RcloneFlow
```

### 2. 配置 Rclone

确保已安装 Rclone 并配置好存储。配置文件通常在 `~/.config/rclone/rclone.conf`。

### 3. 启动 Rclone RC 服务器

```bash
rclone rcd --rc-user=your_user --rc-pass=your_pass --rc-addr=localhost:5572
```

### 4. 构建并运行

```bash
# 构建后端
go build -o server ./cmd/server

# 运行
./server
```

服务器默认在 17870 端口启动，访问 http://localhost:17870

### 5. 配置

配置文件 `config.yaml`：
```yaml
server:
  addr: ":17870"
  static_dir: "./web"

rclone:
  rc_url: "http://127.0.0.1:5572"
  rc_user: ""
  rc_pass: ""
  timeout: "120s"

storage:
  data_dir: "./data"

log:
  level: "info"
  output: "stdout"

sync:
  pool_interval: 5      # 任务状态同步间隔（秒）
  schedule_interval: 1   # 定时任务检查间隔（分钟）
```

### 6. 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `APP_ADDR` | 服务器地址 | `:17870` |
| `APP_DATA_DIR` | 数据目录 | `./data` |
| `RCLONE_RC_URL` | Rclone RC 地址 | `http://127.0.0.1:5572` |
| `RCLONE_RC_USER` | Rclone RC 用户名 | - |
| `RCLONE_RC_PASS` | Rclone RC 密码 | - |
| `RCLONE_RC_TIMEOUT` | RC 超时时间 | `120s` |

## 项目结构

```
RcloneFlow/
├── cmd/
│   └── server/              # 主应用程序入口
├── internal/
│   ├── adapter/            # Rclone API 封装层
│   ├── controller/         # HTTP 控制器
│   ├── dao/                # 数据访问层
│   ├── service/            # 业务逻辑层
│   ├── scheduler/           # 任务调度器
│   ├── router/             # 路由定义
│   ├── store/              # 数据库封装 (SQLite)
│   └── config/             # 配置管理
├── frontend/               # 前端源码 (Vue 3 + TypeScript)
│   └── src/
│       ├── api/            # API 调用层 (统一封装)
│       ├── components/      # Vue 组件
│       └── views/           # 页面视图
├── migrations/              # 数据库迁移文件 (goose)
├── web/                    # 编译后的前端文件
├── config.yaml             # 配置文件
├── Dockerfile
└── docker-compose.yml
```

## 技术架构

### 后端 (Go)

- **Router** - HTTP 路由层，处理请求分发
- **Controller** - 控制器层，参数校验，调用 Service
- **Service** - 业务逻辑层，核心业务处理
- **DAO** - 数据访问层，数据库操作封装
- **Adapter** - Rclone API 适配器封装

### 前端 (Vue 3 + TypeScript)

- **API 层** - 统一封装的 API 调用
  - `api/client.ts` - HTTP 客户端，拦截器
  - `api/errors.ts` - 统一错误处理，Toast 通知
  - `api/task.ts` - 任务相关 API
  - `api/run.ts` - 运行记录 API
  - `api/remote.ts` - 远程存储 API
  - `api/browser.ts` - 文件浏览器 API
- **组件** - Toast、Modal 等 UI 组件
- **视图** - TaskView、BrowserView 等页面

## API 接口

### 任务管理
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/tasks` | 列出所有任务 |
| POST | `/api/tasks` | 创建任务 |
| PUT | `/api/tasks` | 更新任务 |
| DELETE | `/api/tasks/{id}` | 删除任务 |
| POST | `/api/tasks/{id}/run` | 运行任务 |

### 定时任务
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/schedules` | 列出所有定时任务 |
| POST | `/api/schedules` | 创建定时任务 |
| DELETE | `/api/schedules/{id}` | 删除定时任务 |

### 运行记录
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/runs` | 列出运行历史 |
| GET | `/api/runs/{id}` | 获取运行详情 |
| DELETE | `/api/runs/{id}` | 清除运行记录 |
| GET | `/api/runs/active` | 获取运行中的任务及实时状态 |

### 远程存储
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/remotes` | 列出所有存储 |
| POST | `/api/remotes` | 创建存储 |
| PUT | `/api/remotes` | 更新存储 |
| GET | `/api/remotes/config/{name}` | 获取存储配置 |
| DELETE | `/api/config/{name}` | 删除存储 |
| POST | `/api/remotes/test` | 测试存储连接 |
| GET | `/api/providers` | 获取支持的存储类型 |

### 文件浏览器
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/browser/list` | 列出目录内容 |
| POST | `/api/fs/copy` | 复制文件 |
| POST | `/api/fs/move` | 移动文件 |
| POST | `/api/fs/copyDir` | 复制目录 |
| POST | `/api/fs/moveDir` | 移动目录 |
| POST | `/api/fs/delete` | 删除文件 |
| POST | `/api/fs/purge` | 删除目录 |
| POST | `/api/fs/mkdir` | 创建目录 |

## 数据库迁移

使用 goose 进行数据库版本化管理：

```bash
# 查看迁移状态
goose status

# 运行迁移
goose up

# 回滚一个版本
goose down

# 创建新迁移
goose create add_new_field
```

迁移文件位于 `migrations/` 目录。

## 开发

### 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 开发模式 (热重载)
npm run dev

# 运行测试
npm test

# 测试覆盖率
npm run test:coverage

# 生产构建
npm run build
```

### 后端开发

```bash
# 运行测试
go test ./...

# 测试覆盖率
go test -cover ./...

# 构建
go build -o server ./cmd/server
```

### 测试覆盖率

| 模块 | 覆盖率 |
|------|--------|
| adapter | ~80% |
| dao | ~90% |
| service | ~60% |
| 前端 API | ~70% |

## Docker 部署

### 构建镜像

```bash
docker build -t rcloneflow .
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
    volumes:
      - /path/to/rclone/config:/config/rclone
      - /path/to/your/data:/data
    command: rcd --rc-user=your_user --rc-pass=your_pass --rc-addr=0.0.0.0:5572
```

## 贡献

欢迎提交 Pull Request！

## 开源协议

MIT License - 详见 LICENSE 文件。

## 致谢

- [Rclone](https://rclone.org/) - 强大的云存储同步工具
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Vitest](https://vitest.dev/) - 快速的前端测试框架
