# Webhook Has Transfer Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a new "has transfer" checkbox to webhook notification status filters so notifications are sent only when a run both matches success/failed status and has actual transferred data or successful processed items.

**Architecture:** Extend the existing `webhookNotifyStatus` object with a backward-compatible `hasTransfer` boolean. Keep the frontend shape aligned with current success/failed handling, and add a backend post-filter in `internal/runnercli/webhook.go` that evaluates transfer evidence from `finalSummary`.

**Tech Stack:** Vue 3, TypeScript, Vitest, Go, standard library HTTP/json, existing runner webhook pipeline.

### Task 1: Add failing frontend tests for webhook form state

**Files:**
- Create: `frontend/src/composables/useTaskWebhookConfig.test.ts`
- Modify: `frontend/src/composables/useTaskWebhookConfig.ts`

**Step 1: Write the failing test**
- Cover default `status.hasTransfer = false`
- Cover config hydration from `task.options`
- Cover `saveWebhook()` payload including `webhookNotifyStatus.hasTransfer`

**Step 2: Run test to verify it fails**
Run: `cd frontend && npm test -- src/composables/useTaskWebhookConfig.test.ts`
Expected: FAIL because `hasTransfer` is missing.

**Step 3: Write minimal implementation**
- Add `hasTransfer` to default form state
- Read `opts.webhookNotifyStatus.hasTransfer`
- Persist it in `saveWebhook()`

**Step 4: Run test to verify it passes**
Run: `cd frontend && npm test -- src/composables/useTaskWebhookConfig.test.ts`
Expected: PASS

**Step 5: Commit**
```bash
git add frontend/src/composables/useTaskWebhookConfig.ts frontend/src/composables/useTaskWebhookConfig.test.ts
git commit -m "feat: persist webhook has-transfer filter"
```

### Task 2: Add failing UI tests and wire modal props

**Files:**
- Modify: `frontend/src/components/task/WebhookConfigModal.vue`
- Modify: `frontend/src/components/task/TaskListViewShell.vue`
- Modify: `frontend/src/i18n/zh.ts`
- Modify: `frontend/src/i18n/en.ts`

**Step 1: Write the failing test**
- Add or extend component-level test to assert the new checkbox is rendered and bound
- Verify emitted update event for `statusHasTransfer`

**Step 2: Run test to verify it fails**
Run: `cd frontend && npm test -- src/components/...`
Expected: FAIL because modal does not expose the prop/event/checkbox.

**Step 3: Write minimal implementation**
- Add `statusHasTransfer` prop and update emit
- Add checkbox in the status filter row
- Pass state through `TaskListViewShell`
- Add zh/en copy for `hasTransfer` and updated hint

**Step 4: Run test to verify it passes**
Run: `cd frontend && npm test -- src/components/...`
Expected: PASS

**Step 5: Commit**
```bash
git add frontend/src/components/task/WebhookConfigModal.vue frontend/src/components/task/TaskListViewShell.vue frontend/src/i18n/zh.ts frontend/src/i18n/en.ts
git commit -m "feat: expose webhook has-transfer filter in ui"
```

### Task 3: Add failing backend tests for webhook send filtering

**Files:**
- Create: `internal/runnercli/webhook_test.go`
- Modify: `internal/runnercli/webhook.go`

**Step 1: Write the failing test**
- Case A: `hasTransfer=true`, `transferredBytes>0` => send
- Case B: `hasTransfer=true`, `completedCount>0`, bytes 0 => send
- Case C: `hasTransfer=true`, neither condition met => skip
- Case D: missing `hasTransfer` => old success/failed path still sends

**Step 2: Run test to verify it fails**
Run: `go test ./internal/runnercli -run TestWebhook`
Expected: FAIL because backend does not understand `hasTransfer`.

**Step 3: Write minimal implementation**
- Parse `webhookNotifyStatus.hasTransfer`
- Extract a small helper for transfer evidence
- Apply it after existing status matching

**Step 4: Run test to verify it passes**
Run: `go test ./internal/runnercli -run TestWebhook`
Expected: PASS

**Step 5: Commit**
```bash
git add internal/runnercli/webhook.go internal/runnercli/webhook_test.go
git commit -m "feat: filter webhook notifications by transfer activity"
```

### Task 4: Verify integrated behavior

**Files:**
- No code changes required unless fixes are needed

**Step 1: Run targeted frontend tests**
Run: `cd frontend && npm test -- src/composables/useTaskWebhookConfig.test.ts`
Expected: PASS

**Step 2: Run targeted backend tests**
Run: `go test ./internal/runnercli -run TestWebhook`
Expected: PASS

**Step 3: Run broader safety checks**
Run: `cd frontend && npm test`
Expected: PASS

Run: `go test ./...`
Expected: PASS

**Step 4: Commit if any final fixups were needed**
```bash
git add -A
git commit -m "test: verify webhook has-transfer support"
```
