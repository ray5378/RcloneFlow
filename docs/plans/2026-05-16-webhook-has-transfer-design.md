# Webhook 通知“有传输”状态过滤设计

## 背景
当前 rcloneflow 的 Webhook 通知支持两类状态过滤：`success` 和 `failed`。用户希望新增一个与这两项并列显示、可勾选并在保存后立即生效的过滤项：`有传输`。

## 需求结论
采用方案 A：在现有 `webhookNotifyStatus` 结构内新增 `hasTransfer` 字段。

### 用户可见行为
- Webhook 通知弹窗的“状态过滤”区域新增复选框：`有传输`
- 与“成功 / 失败”并列显示
- 勾选后保存即生效
- 旧任务未配置该字段时，默认关闭，不改变现有通知行为

## 判定规则
用户确认采用 C 口径：满足以下任一条件视为“有传输”
1. `summary.transferredBytes > 0`
2. 存在成功处理条目（优先使用 `finalSummary.counts.copied > 0`，兼容单文件/零字节等场景）

## 数据结构
当前：
```ts
webhookNotifyStatus: {
  success: boolean
  failed: boolean
}
```

扩展后：
```ts
webhookNotifyStatus: {
  success: boolean
  failed: boolean
  hasTransfer: boolean
}
```

## 前端改动
### 文件
- `frontend/src/composables/useTaskWebhookConfig.ts`
- `frontend/src/components/task/WebhookConfigModal.vue`
- `frontend/src/components/task/TaskListViewShell.vue`
- `frontend/src/i18n/zh.ts`
- `frontend/src/i18n/en.ts`

### 具体内容
- 表单默认值加入 `status.hasTransfer = false`
- 打开任务配置时从 `options.webhookNotifyStatus.hasTransfer` 回填
- 保存时把 `hasTransfer` 一并写入 `webhookNotifyStatus`
- 弹窗 props / emits 增加 `statusHasTransfer`
- 状态过滤 UI 中增加第三个勾选项“有传输”
- 中英文提示文案补充说明：勾选后，仅当本次任务实际发生传输或存在成功处理文件时才通知

## 后端改动
### 文件
- `internal/runnercli/webhook.go`

### 具体内容
- 读取 `webhookNotifyStatus.hasTransfer`
- 在原有 success / failed 过滤通过后，再按需执行“有传输”附加判定
- 附加判定逻辑：
  - `transferredBytes > 0` 返回 true
  - 否则若 `completedCount > 0` 返回 true
  - 否则返回 false
- 保持旧配置兼容：未配置 `hasTransfer` 时视为 false

## 测试策略
### 前端
- 为 `useTaskWebhookConfig` 新增单测：
  - 默认值包含 `hasTransfer: false`
  - 读取旧配置时默认 false
  - 读取新配置时正确回填 true
  - 保存时 payload 包含 `webhookNotifyStatus.hasTransfer`

### 后端
- 为 `internal/runnercli/webhook.go` 新增测试：
  - 勾选 `hasTransfer=true` 且 `transferredBytes>0` 时允许发送
  - 勾选 `hasTransfer=true` 且 `completedCount>0, transferredBytes=0` 时允许发送
  - 勾选 `hasTransfer=true` 且两者都不满足时不发送
  - 未配置 `hasTransfer` 时保持原 success/failed 逻辑

## 风险与兼容性
- 该字段新增在 options JSON 中，属于向后兼容扩展
- 旧任务不会因为本次改动额外发送通知
- “有传输”作为附加条件存在，语义虽挂在 status 下，但与本次用户预期一致，且实现成本最低
