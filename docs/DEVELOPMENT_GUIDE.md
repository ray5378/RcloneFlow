# RcloneFlow 开发指南

本文档是 RcloneFlow 项目的完整开发指南，用于指导后续开发者遵循代码规范、了解项目结构、掌握开发流程。

---

## 目录

1. [项目概述](#1-项目概述)
2. [技术栈](#2-技术栈)
3. [项目结构](#3-项目结构)
4. [开发环境配置](#4-开发环境配置)
5. [开发流程规范](#5-开发流程规范)
6. [前端开发规范](#6-前端开发规范)
7. [后端开发规范](#7-后端开发规范)
8. [测试规范](#8-测试规范)
9. [构建与部署](#9-构建与部署)
10. [问题排查](#10-问题排查)

---

## 1. 项目概述

### 1.1 项目简介
RcloneFlow 是一个基于 rclone 的可视化文件传输与管理工具，提供：
- 任务配置与调度
- 实时进度展示
- 远程存储管理
- 文件浏览与操作
- 历史记录查看

### 1.2 核心特性
- 稳态进度展示（Stable Progress）
- 历史冻结（Summary Freeze）
- JWT 鉴权
- 单例模式（Singleton Mode）
- Webhook 通知
- 定时调度（Cron）

### 1.3 关键概念

#### 稳态进度 (Stable Progress)
运行中的任务使用"稳态进度"展示，包括阶段信息：
- `preparing` - 准备阶段
- `transferring` - 传输中
- `between_files` - 文件间隔
- `finalizing` - 完成中

#### 历史冻结 (Summary Freeze)
任务结束时，将最终总结一次性写入数据库（`summary.finalSummary`），历史页面只渲染这份冻结数据，不再更新。

#### 单例模式 (Singleton Mode)
开启单例模式的任务在触发时会检查是否有其他任务正在运行：
- 如果有其他任务在运行，跳过本次执行
- 如果没有其他任务在运行，启动当前任务
- 使用数据库事务确保原子性

---

## 2. 技术栈

### 2.1 前端技术栈
- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript 5.3+
- **构建工具**: Vite 5.0+
- **测试框架**: Vitest 4.1+
- **类型检查**: vue-tsc

### 2.2 后端技术栈
- **语言**: Go 1.25+
- **Web 框架**: Gin
- **数据库**: SQLite (modernc.org/sqlite)
- **WebSocket**: gorilla/websocket
- **任务调度**: robfig/cron/v3
- **日志**: go.uber.org/zap
- **JWT**: github.com/golang-jwt/jwt/v5
- **测试**: testify

### 2.3 依赖工具
- **rclone**: 文件传输引擎（CLI + RC 双模式）
- **Docker**: 容器化部署

---

## 3. 项目结构

```
rcloneflow/
├── cmd/server/              # 后端入口
├── internal/
│   ├── app/                 # 应用初始化
│   ├── auth/                # 鉴权
│   ├── config/              # 配置
│   ├── controller/          # 控制器（HTTP 接口）
│   ├── dao/                 # 数据访问对象
│   ├── logger/              # 日志
│   ├── router/              # 路由配置
│   ├── service/             # 业务逻辑层
│   ├── store/               # 数据存储层
│   ├── websocket/           # WebSocket 处理
│   ├── scheduler/           # 调度器
│   ├── runnercli/           # rclone 执行器（当前主链）
│   ├── rclone/              # rclone 适配
│   └── adapter/             # 适配器
├── frontend/                # 前端源码
│   ├── src/
│   │   ├── views/           # 页面组件
│   │   ├── components/      # 可复用组件
│   │   ├── composables/     # 组合式函数
│   │   ├── api/             # API 请求封装
│   │   └── main.ts          # 前端入口
│   ├── package.json
│   └── vite.config.ts
├── web/                     # 前端构建输出（可重建）
├── docs/                    # 项目文档（非常重要！）
│   ├── ENGINEERING_RULES.md    # 工程总规范
│   ├── FRONTEND_RULES.md       # 前端实现规则
│   ├── BACKEND_RULES.md        # 后端实现规则
│   ├── DEVELOPMENT_CHECKLIST.md # 开发检查清单
│   ├── ARCHITECTURE_OVERVIEW.md # 架构总览
│   ├── TECH_DEBT.md            # 技术债与拆分进度
│   └── ...                     # 更多文档
├── migrations/              # 数据库迁移
├── scripts/                 # 工具脚本
├── data/                    # 数据目录（运行时）
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── config.yaml
```

### 3.1 关键文件说明

| 文件/目录 | 说明 |
|----------|------|
| `docs/` | **最重要的目录**！包含所有开发规范、指南、检查清单 |
| `internal/runnercli/runner.go` | rclone 执行器主链，负责解析进度 |
| `internal/controller/run.go` | 运行记录控制器，`/api/runs/active` 接口 |
| `frontend/src/views/TaskView.vue` | 任务主页面（高风险区） |
| `frontend/src/views/BrowserView.vue` | 文件浏览页面（已优化） |
| `frontend/src/composables/useBrowser.ts` | 统一的浏览器组合式 |

---

## 4. 开发环境配置

### 4.1 前置要求
- Node.js 18+
- Go 1.25+
- Docker（可选，用于容器化）
- Git

### 4.2 克隆项目
```bash
git clone <repository-url>
cd rcloneflow
```

### 4.3 前端环境
```bash
cd frontend
npm install

# 开发模式
npm run dev

# 构建
npm run build

# 测试
npm run test
```

### 4.4 后端环境
```bash
# 安装依赖
go mod download

# 运行服务
go run ./cmd/server

# 或编译后运行
go build -o server ./cmd/server
./server
```

### 4.5 配置文件
复制示例配置：
```bash
cp .env.example .env
cp config.yaml config.local.yaml
```

### 4.6 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| APP_ADDR | :17870 | 服务地址 |
| APP_DATA_DIR | ./data | 数据目录 |
| LOG_LEVEL | info | 日志级别 (debug/info/warn/error) |
| PRECHECK_MODE | none | 预检模式 |
| PROGRESS_FLUSH_INTERVAL | 5s | 进度刷新间隔 |
| FINISH_WAIT_INTERVAL | 5s | 完成后等待间隔 |
| FINAL_SUMMARY_RETENTION_DAYS | 7 | 历史保留天数 |
| CLEANUP_INTERVAL_HOURS | 24 | 清理扫描间隔 |

---

## 5. 开发流程规范

### 5.1 开始开发前必读

在开始任何开发工作前，**必须先阅读以下文档**：

1. `docs/README.md` - 文档索引，了解该看哪些文档
2. `docs/ENGINEERING_RULES.md` - 工程总规范
3. `docs/DEVELOPMENT_CHECKLIST.md` - 开发检查清单
4. `docs/ARCHITECTURE_OVERVIEW.md` - 架构总览

根据你的角色，还需要阅读：
- **前端开发**: `docs/FRONTEND_RULES.md`
- **后端开发**: `docs/BACKEND_RULES.md`

### 5.2 Git 分支策略

- `master` - 生产环境分支，稳定版本
- `dev` - 开发分支，所有功能开发先合并到这里
- `feature/*` - 功能分支，从 dev 分出，合并回 dev
- `bugfix/*` - Bug 修复分支

### 5.3 开发流程

1. **从 dev 创建分支**
   ```bash
   git checkout dev
   git pull origin dev
   git checkout -b feature/your-feature-name
   ```

2. **开发过程**
   - 小步提交，每步停测
   - 使用 `docs/DEVELOPMENT_CHECKLIST.md` 做自检
   - 避免同时改多个高风险区域

3. **提交前检查**
   - 前端：运行 `npm run build` 确保能构建
   - 后端：运行 `go test` 确保测试通过
   - 检查没有引入回归

4. **合并回 dev**
   - 创建 PR/MR
   - 代码审阅（如果有团队）
   - 合并到 dev

5. **合并到 master（发布时）**
   - 遵循 `docs/RELEASE_GUIDE.md`

### 5.4 提交信息规范

格式：
```
<type>(<scope>): <subject>

<body>
```

Type 类型：
- `feat`: 新功能
- `fix`: Bug 修复
- `refactor`: 重构
- `docs`: 文档更新
- `test`: 测试相关
- `chore`: 构建/工具链

示例：
```
feat(browser): add useBrowser composable

- Create unified useBrowser composable
- Split BrowserView into child components
- Add complete type definitions
```

---

## 6. 前端开发规范

### 6.1 核心原则

完整规范请参考：`docs/FRONTEND_RULES.md`

关键原则：
1. **逻辑分层** - 区分 view、component、composable、helper
2. **职责分离** - 避免超大文件
3. **类型安全** - 完整的 TypeScript 类型定义
4. **可复用性** - 提取公共组件和 composable

### 6.2 目录规范

```
frontend/src/
├── views/          # 页面组件（路由级）
│   ├── TaskView.vue
│   ├── BrowserView.vue
│   ├── RunView.vue
│   ├── ScheduleView.vue
│   └── LoginView.vue
├── components/     # 可复用组件
│   ├── task/       # 任务相关组件
│   ├── StoragePanel.vue
│   ├── BrowserPanel.vue
│   └── ManageStoragePanel.vue
├── composables/    # 组合式函数
│   ├── useBrowser.ts
│   ├── useBrowserClipboard.ts
│   ├── useBrowserContextMenu.ts
│   ├── useBrowserFileOps.ts
│   ├── useBrowserRemoteManagement.ts
│   ├── useRunningHint.ts
│   ├── useActiveRunLookup.ts
│   └── useWebSocket.ts
├── api/            # API 请求封装
├── utils/          # 工具函数
└── types/          # 类型定义
```

### 6.3 Vue 组件规范

#### 6.3.1 组件命名
- 使用 PascalCase：`StoragePanel.vue`
- 组件名应反映其功能

#### 6.3.2 单文件组件结构
```vue
<script setup lang="ts">
// 1. 导入
import { ref, computed } from 'vue'

// 2. Props 定义
interface Props {
  title: string
}
const props = defineProps<Props>()

// 3. Emits 定义
interface Emits {
  (e: 'update'): void
}
const emit = defineEmits<Emits>()

// 4. 响应式状态
const count = ref(0)

// 5. 计算属性
const doubled = computed(() => count.value * 2)

// 6. 方法
const increment = () => {
  count.value++
}
</script>

<template>
  <div class="component-name">
    <!-- 模板内容 -->
  </div>
</template>

<style scoped>
.component-name {
  /* 样式 */
}
</style>
```

#### 6.3.3 模板规范
- 模板复杂度不超过 100 行
- 复杂逻辑提取到 computed 或方法
- 使用语义化 HTML 标签

#### 6.3.4 弹窗规范
**禁止使用原生 `alert()`、`confirm()`、`prompt()`**
必须使用自定义 Modal 组件：
```vue
<div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
  <div class="modal-content">
    <div class="modal-header">
      <h3>{{ title }}</h3>
      <button class="close-btn" @click="showModal = false">×</button>
    </div>
    <div class="modal-body">
      <!-- 内容 -->
    </div>
    <div class="modal-footer">
      <!-- 按钮 -->
    </div>
  </div>
</div>
```

样式类名：
- `modal-overlay` - 遮罩层
- `modal-content` - 弹窗内容区
- `modal-header` - 标题栏
- `modal-body` - 主体内容
- `modal-footer` - 底部按钮区
- `close-btn` - 关闭按钮

### 6.4 Composable 规范

#### 6.4.1 命名
- 以 `use` 开头：`useBrowser.ts`
- 导出完整的类型定义

#### 6.4.2 完整示例
```typescript
// frontend/src/composables/useSomething.ts

// 1. 导出返回类型
export interface UseSomethingReturn {
  state: Ref<State>
  someMethod: () => void
}

// 2. 实现 composable
export function useSomething(): UseSomethingReturn {
  const state = ref<State>(initialState)

  const someMethod = () => {
    // 实现
  }

  return {
    state,
    someMethod,
  }
}
```

#### 6.4.3 统一入口 Composable
对于相关功能，创建统一入口：
```typescript
// frontend/src/composables/useBrowser.ts
export function useBrowser() {
  const clipboard = useBrowserClipboard()
  const contextMenu = useBrowserContextMenu()
  const fileOps = useBrowserFileOps()
  const remoteMgmt = useBrowserRemoteManagement()

  return {
    // 整合所有功能
    ...clipboard,
    ...contextMenu,
    ...fileOps,
    ...remoteMgmt,
  }
}
```

### 6.5 TypeScript 规范

#### 6.5.1 类型定义
- 为所有 Props、Emits 定义接口
- 为 composable 返回值定义类型
- 避免使用 `any`

#### 6.5.2 类型导出
相关类型应导出，供子组件使用：
```typescript
export interface UseBrowserRemoteManagementReturn {
  remotes: Ref<Remote[]>
  selectedRemote: Ref<string | null>
  loadRemotes: () => Promise<void>
}
```

### 6.6 状态管理规范

- **运行中 UI 主数据源**: `/api/runs/active.progress`
- **任务卡片完成态**: 前端冻结帧 `completedFreezeByTask`
- **`preflight`**: 只保留预估语义，不直接驱动运行中主展示

---

## 7. 后端开发规范

### 7.1 核心原则

完整规范请参考：`docs/BACKEND_RULES.md`

关键原则：
1. **分层架构** - controller、service、store 职责清晰
2. **接口契约** - 明确字段语义、兼容字段、回退逻辑
3. **错误处理** - 统一的错误处理方式
4. **测试分层** - 单元测试、集成测试

### 7.2 目录规范

```
internal/
├── controller/    # HTTP 接口层
│   ├── run.go     # 运行记录接口
│   ├── task.go    # 任务接口
│   ├── browser.go # 文件浏览（RC 模式）
│   └── fs_cli.go  # 文件操作（CLI 模式）
├── service/       # 业务逻辑层
├── store/         # 数据存储层
├── dao/           # 数据访问对象
├── runnercli/     # rclone 执行器（当前主链）
└── ...
```

### 7.3 分层规范

#### 7.3.1 Controller 层
职责：
- HTTP 请求解析与响应
- 接口字段组装
- 参数校验
- 调用 service 层

不要：
- 直接写业务逻辑
- 直接操作数据库

#### 7.3.2 Service 层
职责：
- 业务规则
- 状态聚合
- 流程协调
- 调用 store/dao 层

#### 7.3.3 Store/Dao 层
职责：
- 数据读写
- 数据库操作
- 不包含业务逻辑

### 7.4 关键链路说明

#### 7.4.1 运行中进度链
固定排查顺序：
```
日志原文 -> summary.progress -> /api/runs/active.progress -> 前端展示
```

关键文件：
- `internal/runnercli/runner.go` - 解析 rclone 进度日志
- `internal/controller/run.go` - `/api/runs/active` 接口

#### 7.4.2 文件浏览与操作
**两条独立链路，不要混淆**：
- `internal/controller/browser.go` - 走 rclone RC 浏览
- `internal/controller/fs_cli.go` - 走 CLI 文件操作

### 7.5 API 接口规范

#### 7.5.1 基础信息
- 基础路径: `/api`
- 认证方式: JWT Bearer Token

#### 7.5.2 主要接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/tasks | 获取任务列表 |
| POST | /api/tasks | 创建任务 |
| PATCH | /api/tasks | 更新任务 |
| DELETE | /api/tasks/{id} | 删除任务 |
| POST | /api/tasks/{id}/run | 运行任务 |
| GET | /api/runs | 历史列表 |
| GET | /api/runs/active | 运行中的任务 |
| GET | /api/runs/{id} | 单条记录 |
| GET | /api/runs/{id}/files | 文件列表 |
| GET | /api/remotes | 存储列表 |
| POST | /api/remotes | 添加存储 |
| GET | /api/config | rclone 配置 |
| GET | /api/providers | 支持的存储类型 |

### 7.6 高风险区域

以下区域改动时需特别小心：
- `internal/controller/run.go`
- `internal/runnercli/runner.go`
- 运行中进度相关代码

原因：
- 多层状态链汇聚
- 用户可直接看到效果
- 回退逻辑较多

---

## 8. 测试规范

完整指南请参考：`docs/TESTING_GUIDE.md`

### 8.1 前端测试

```bash
cd frontend
npm run test
```

测试文件位置：`frontend/src/__tests__/` 或同目录 `.test.ts`

### 8.2 后端测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/controller -v

# 生成覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 8.3 测试策略

- **重构型改动**: 至少验证主链路功能正常
- **新功能**: 补充相应测试
- **Bug 修复**: 补充回归测试

---

## 9. 构建与部署

### 9.1 前端构建

```bash
cd frontend
npm install
npm run build
```

构建产物输出到 `../web` 目录。

### 9.2 后端构建

```bash
go build -o server ./cmd/server
```

### 9.3 Docker 构建

```bash
docker build --no-cache -t ray5378/rcloneflow:dev .
```

### 9.4 Docker Compose

```bash
docker-compose up -d
```

### 9.5 发布流程

完整指南请参考：`docs/RELEASE_GUIDE.md`

---

## 10. 问题排查

完整排查手册请参考：`docs/DEBUGGING_PLAYBOOK.md`

### 10.1 常见问题

#### 10.1.1 页面显示旧前端
**原因**: 挂载覆盖了 `/app/web`
**解决**: 不要挂载覆盖 /app/web 目录

#### 10.1.2 运行中进度不更新
**排查顺序**:
1. 检查 rclone 日志输出
2. 检查 `internal/runnercli/runner.go` 是否正确解析
3. 检查 `/api/runs/active` 接口返回
4. 检查前端 WebSocket 连接

#### 10.1.3 前端构建失败
- 检查 TypeScript 类型错误
- 检查 Vue 模板语法
- 运行 `npm run build` 查看详细错误

#### 10.1.4 rclone 操作失败
检查 rclone.conf 配置是否正确，确保远程存储名称与任务中一致

### 10.2 日志查看

```bash
# Docker 日志
docker logs rcloneflow

# 或直接运行时查看
./server
```

### 10.3 数据库位置
数据目录 `data/` 下可找到 SQLite 数据库文件。

---

## 11. 文档体系

项目有非常完善的文档体系，**开发前务必先看文档**！

### 11.1 文档索引

| 文档 | 用途 |
|------|------|
| `docs/README.md` | **必读** - 文档索引 |
| `docs/ENGINEERING_RULES.md` | **必读** - 工程总规范 |
| `docs/DEVELOPMENT_CHECKLIST.md` | **必读** - 开发检查清单 |
| `docs/ARCHITECTURE_OVERVIEW.md` | 架构总览 |
| `docs/TECH_DEBT.md` | 技术债与拆分进度 |
| `docs/FRONTEND_RULES.md` | 前端实现规则 |
| `docs/BACKEND_RULES.md` | 后端实现规则 |
| `docs/API_CONTRACT_GUIDE.md` | 接口契约指南 |
| `docs/DEBUGGING_PLAYBOOK.md` | 排查手册 |
| `docs/TESTING_GUIDE.md` | 测试指南 |
| `docs/RELEASE_GUIDE.md` | 发布指南 |
| `docs/REFACTORING_PLAYBOOK.md` | 重构手册 |
| `docs/ONBOARDING_GUIDE.md` | 新手指南 |
| `docs/GLOSSARY.md` | 术语表 |
| `docs/KNOWN_PITFALLS.md` | 已知坑点 |
| `docs/DECISION_LOG.md` | 决策记录 |

### 11.2 推荐阅读顺序

**新接手项目**:
1. `docs/README.md`
2. `docs/ENGINEERING_RULES.md`
3. `docs/DEVELOPMENT_CHECKLIST.md`
4. `docs/ARCHITECTURE_OVERVIEW.md`
5. `docs/TECH_DEBT.md`
6. `docs/FRONTEND_RULES.md` 或 `docs/BACKEND_RULES.md`

**前端开发**:
1. `docs/DEVELOPMENT_CHECKLIST.md`
2. `docs/FRONTEND_RULES.md`
3. `docs/ENGINEERING_RULES.md`
4. `docs/TECH_DEBT.md`

**后端开发**:
1. `docs/DEVELOPMENT_CHECKLIST.md`
2. `docs/BACKEND_RULES.md`
3. `docs/ENGINEERING_RULES.md`
4. `docs/TECH_DEBT.md`

---

## 12. 关键约束

开发时必须遵守以下约束（来自 `docs/ENGINEERING_RULES.md`）：

1. **`dev -> master` 是固定主线流程**
2. **运行中 UI 主真源是 `/api/runs/active.progress`**
3. **运行中主展示真源固定为 `progress`，任务卡片完成态由前端冻结帧 `completedFreezeByTask` 承接**
4. **`preflight` 保留预估语义，不直接驱动运行中主展示**
5. **高风险改动优先小步推进、每步停测**

---

## 附录

### A. 参考资源

- Vue 3 文档: https://vuejs.org/
- TypeScript 文档: https://www.typescriptlang.org/
- Go 文档: https://go.dev/doc/
- rclone 文档: https://rclone.org/

### B. 联系方式

如有问题，请查看 `docs/AGENT_COLLABORATION_GUIDE.md` 了解如何高效协作。

---

*最后更新: 2026-05-26*
