# API_CONTRACT_GUIDE.md

关键接口契约说明。

这份文档优先记录项目中最关键、最容易被误解、且已经形成前后端事实约定的接口字段语义。

---

## 1. `/api/runs/active`

### 1.1 作用
用于提供运行中任务的实时展示数据。

当前约定：
- 这是运行中 UI 的主要事实来源
- 任务卡片、运行中提示小窗、运行中 ETA 等主展示应优先围绕该接口组织

### 1.2 关键字段
典型返回项中应包含：
- `runRecord`
- `progress`
- `progressLine`
- `progressSource`
- `progressMismatch`
- `progressCheck`

### 1.3 `runRecord`
用于承载运行记录基本信息，例如：
- 运行 id
- taskId
- status
- startedAt
- finishedAt

### 1.4 `progress`
语义：
- 运行中的实时进度主字段
- 当前运行中展示应优先使用它

典型字段包括：
- `bytes`
- `totalBytes`
- `speed`
- `eta`
- `percentage`
- `completedFiles`
- `totalCount`
- `phase`
- `lastUpdatedAt`

### 1.5 `progressLine`
语义：
- 最后成功解析到的原始进度日志行
- 主要用于调试和核对链路

### 1.6 `progressMismatch` / `progressCheck`
语义：
- 后端对当前 progress 自洽性的辅助自检结果
- 主要用于调试，不是主业务展示字段

---

## 2. 运行中字段语义说明

### 2.1 `bytes`
语义：
- 当前已传输字节数

### 2.2 `totalBytes`
语义：
- 当前运行中总量
- 前端运行中总量展示应优先使用该字段

注意：
- 早期 live progress 若给出偏小总量，不应覆盖更可信的预检总量

### 2.3 `speed`
语义：
- 当前实时速度

### 2.4 `eta`
语义：
- 当前基于 live progress 计算出的预计剩余时间

当前约定：
- 不再回退 `preflight.totalBytes` 来计算运行中 ETA

### 2.5 `percentage`
语义：
- 当前运行中百分比

当前约定：
- 若 `preflight` 接管了 `totalBytes`，则百分比应基于 `bytes / totalBytes` 重新计算，以保证字段自洽

### 2.6 `completedFiles`
语义：
- 已完成文件数

### 2.7 `totalCount`
语义：
- 当前运行中文件总数

注意：
- 不应长期出现“已完成文件数有值，但总数为 0”这种自相矛盾状态

---

## 3. `progress / finalSummary / preflight` 的区别

### 3.1 `progress`
- 运行中实时进度
- 当前运行中展示主数据源

### 3.2 `finalSummary`
- 历史详情 / 最终总结字段
- 用于完成态详情、统计概览、文件明细等展示

### 3.3 `preflight`
- 预估总量
- 主要来自运行前预检
- 仅用于预估展示或后端明确说明的兼容兜底
- 不应直接被前端当作运行中主展示字段使用

---

## 4. WebSocket 相关约定

### 4.1 `run_progress`
当前约定：
- 前端更新 active run 时，必须按 `runRecord.id` 匹配
- 不应错误按顶层 `r.id` 等字段匹配

### 4.2 `run_status`
当前约定：
- 主要用于状态切换与补充刷新触发
- 不应用它替代运行中进度主链

---

## 5. 任务卡片完成态规则

当前约定：
- 运行中：读取 `/api/runs/active.progress`
- 完成瞬间：前端冻结成单份完成帧 `completedFreezeByTask`
- 完成后短窗口：继续沿用同一份冻结帧
- 超过短窗口：清掉冻结帧，不再继续显示进度条

注意：
- 任务卡片不再在 finished 阶段切换到第二份完成态摘要
- 不要再重新引入新的 handoff 逻辑，避免完成后再抖一下

---

## 6. Active Runs 接口规则（必须遵守）

`/api/runs/active` 当前是运行中 UI 的唯一事实来源。

接口约定：
- `progress`：运行中实时进度主字段
- `progressLine`：最后成功解析到的原始进度日志行
- `progressSource`：当前进度来源
- `progressMismatch` / `progressCheck`：后端一致性自检结果

后续开发不得：
- 删除 `progress` 并要求前端把结束态过渡重新塞回运行中链
- 让 `finalSummary` 再次承担运行中主展示职责
