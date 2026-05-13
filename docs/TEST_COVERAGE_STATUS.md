# Test Coverage Status

Last updated: 2026-05-14 (legacy sort_index schema repair batch)

## Purpose

This document tracks:

- what test coverage work has already been added
- what major test gaps still remain
- what the next highest-yield targets are
- the working rules for how coverage/test changes should be landed

It is intentionally pragmatic rather than exhaustive.

## Working rules

1. Every time one coverage/test item is completed, first run the relevant tests and coverage measurement, then update this document once with both the newly added coverage work and the refreshed coverage figures.
2. Any future code commit must first pass the relevant tests before commit.
3. Any future code commit for this work must include the updated `docs/TEST_COVERAGE_STATUS.md` in the same commit.
4. Validation should be recorded here in a concise form whenever a coverage batch lands.
5. New or changed code should not be treated as merge-ready unless its coverage reaches the required gate for new code; test work should target the changed logic and risk surface, not just inflate overall coverage.
6. Default gate recommendation for this repo: `new code line coverage >= 80%`, `new code branch coverage >= 70%`; if the current toolchain only reports line coverage reliably, at minimum enforce `new code line coverage >= 80%`.

---

## Commit/validation policy

For this repo, coverage/test work should follow this rule:

- do the work
- run the relevant tests
- re-measure the relevant coverage numbers
- confirm the new/changed code coverage reaches the merge gate when such a gate exists
- use the current diff/changed-lines scope as the default definition of “new code”; if only file/function-level evidence is available, state that approximation explicitly
- update this document once as part of the same work batch, recording both what was added and the refreshed coverage figures
- only commit after those tests pass
- include this updated document in the same commit as the code changes

This applies to future coverage pushes unless the user explicitly says otherwise.

## Latest validated state

Validated for this batch:

- `go test ./internal/store ./internal/service ./internal/controller` ✅
- `go test ./internal/store -coverprofile=/tmp/rcloneflow-store.out` ✅
- `go test ./... -coverprofile=/tmp/rcloneflow-all.out` ✅

Latest pushed commit for the previous coverage pass:

- `08068d1`
- `test: expand frontend and service coverage`

Approximate coverage checkpoints during this pass:

### Go

- baseline around `40.9%`
- later around `42.1%`
- earlier service-focused checkpoint around `42.6%`
- latest freshly re-measured overall checkpoint after recent controller additions: `41.3%`
- latest freshly re-measured overall checkpoint after task reorder persistence batch: `48.5%`
- latest freshly re-measured overall checkpoint after task `sort_index` self-heal batch: `49.0%`
- latest freshly re-measured overall checkpoint after legacy `sort_index` schema repair batch: `49.1%`

### Controller package

- earlier recorded `internal/controller` checkpoint around `29.2%`
- later checkpoint after `task_controller_test.go` + first `run_status_log_test.go` batch: `45.9%`
- later freshly re-measured checkpoint after adding `HandleRuns` + `HandleRunsByTask` coverage: `42.1%`
- later freshly re-measured checkpoint after `HandleRunFiles` coverage batch: `43.0%`
- later freshly re-measured checkpoint after kill-related coverage batch: `46.6%`
- later freshly re-measured checkpoint after `HandleGlobalStats` coverage batch: `47.9%`
- latest freshly re-measured `internal/controller` checkpoint after `settings.go` coverage batch: `51.7%`
- latest freshly re-measured `internal/controller` checkpoint after task reorder persistence batch: `51.7%`

### Store package

- latest freshly re-measured `internal/store` checkpoint after task `sort_index` self-heal batch: `47.1%`
- latest freshly re-measured `internal/store` checkpoint after legacy `sort_index` schema repair batch: `49.2%`

### Frontend

Coverage rose across several rounds from roughly:

- `14.77%`
- `19.72%`
- `25.73%`
- `29.38%`
- `39.00%`
- `48.57%`
- `52.14%`
- `52.43%`
- `55.24%`

---

## What was added

## Frontend test additions

Large coverage expansion was added across:

### API tests

- `frontend/src/api/activeTransfer.test.ts`
- `frontend/src/api/browser.test.ts`
- `frontend/src/api/remote.test.ts`
- `frontend/src/api/schedule.test.ts`
- `frontend/src/api/settings.test.ts`

### UI/helper/component tests

- `frontend/src/components/task/progressText.test.ts`
- `frontend/src/components/task/runningHint.test.ts`
- `frontend/src/components/task/scheduleOptions.test.ts`
- `frontend/src/components/task/transferring/transferringLabels.test.ts`
- `frontend/src/composables/activeRunProgress.test.ts`
- `frontend/src/composables/useRunDisplayHelpers.test.ts`
- `frontend/src/composables/useRunningHint.test.ts`

### Run detail / transfer detail tests

- `frontend/src/composables/useActiveTransferDetail.test.ts`
- `frontend/src/composables/useRunDetailEntry.test.ts`
- `frontend/src/composables/useRunDetailFiles.test.ts`
- `frontend/src/composables/useRunDetailRuntime.test.ts`
- `frontend/src/composables/useRunDetailState.test.ts`

### Task form / command / path / view tests

- `frontend/src/composables/useTaskCommandParse.test.ts`
- `frontend/src/composables/useTaskFormEntry.test.ts`
- `frontend/src/composables/useTaskFormFlow.test.ts`
- `frontend/src/composables/useTaskFormNormalize.test.ts`
- `frontend/src/composables/useTaskFormOrchestrator.test.ts`
- `frontend/src/composables/useTaskFormPrepare.test.ts`
- `frontend/src/composables/useTaskFormRuntime.test.ts`
- `frontend/src/composables/useTaskFormState.test.ts`
- `frontend/src/composables/useTaskFormSubmit.test.ts`
- `frontend/src/composables/useTaskPathBrowse.test.ts`
- `frontend/src/composables/useTaskViewRuntimeState.test.ts`
- `frontend/src/composables/useTaskViewState.test.ts`

### Task history / task list / schedule tests

- `frontend/src/composables/useTaskHistoryActions.test.ts`
- `frontend/src/composables/useTaskHistoryComputed.test.ts`
- `frontend/src/composables/useTaskHistoryLoader.test.ts`
- `frontend/src/composables/useTaskHistoryPagingEntry.test.ts`
- `frontend/src/composables/useTaskHistoryRuntime.test.ts`
- `frontend/src/composables/useTaskListActions.test.ts`
- `frontend/src/composables/useTaskListRuntime.test.ts`
- `frontend/src/composables/useTaskRunActions.test.ts`
- `frontend/src/composables/useTaskScheduleDisplay.test.ts`
- `frontend/src/composables/useTaskScheduleLookup.test.ts`

### Frontend fixes / infra changes made during the pass

- `frontend/src/composables/useTaskCommandParse.ts`
  - fixed quoted remote-path parsing, e.g. `'src:/a:b'`
- `frontend/vitest.config.ts`
  - widened useful coverage scope
- `frontend/package.json`
  - coverage/test tooling updates
- `.github/workflows/docker.yml`
  - CI updated to actually run tests / coverage flow

---

## Go test additions

Added:

- `internal/service/cleanup_schedule_test.go`
- `internal/service/run_lifecycle_test.go`
- `internal/service/task_options_runtime_test.go`
- `internal/store/store_test.go`
  - added `TestReorderTasksPersistsOrder`
  - verifies `ReorderTasks(ids)` rewrites `sort_index` and `ListTasks()` returns the persisted order on reread
  - added `TestListTasks_NormalizesMissingSortIndexes`
  - verifies tasks with missing/invalid `sort_index` are automatically normalized on `ListTasks()` and the repaired sort values persist
  - added `TestListTasks_NormalizesDuplicateSortIndexes`
  - verifies duplicate `sort_index` values are normalized back to a stable continuous order on `ListTasks()`
  - tightened `TestOpenCreatesDir` to use a unique temp dir, avoiding cross-run SQLite path collisions during coverage runs
  - added `TestOpen_RepairsMissingSortIndexColumnWhenSchemaVersionAlreadyRecorded`
  - verifies `Open()` repairs legacy databases where `schema_migrations` already records v4 but `tasks.sort_index` is still missing
- `internal/controller/task_controller_test.go`
  - added `PATCH reorder persists order`
  - verifies `PATCH /api/tasks` with `{ order: [...] }` reaches persistence and the stored order is reflected by a fresh `ListTasks()` read

These covered major service-level behavior in:

### `internal/service/run.go`

Covered paths including:

- `ListRuns`
- `ListRunsByTask`
- `ListActiveRuns`
- `GetActiveRunByTaskID`
- `UpdateRunStatus`
- `deepMerge`
- `DeleteRun`
- `DeleteAllRuns`
- `DeleteRunsByTask`
- `CleanOldRuns`
- `cleanupRunLog`

### `internal/service/cleanup.go` and `schedule.go`

Covered key lifecycle behavior including:

- start / stop / replan paths
- cleanup execution behavior
- schedule list/create/update/delete flows via service layer

### `internal/service/task.go`

Covered important paths including:

- `UpdateTask`
- `UpdateTaskOptions`
- `RunTask` selected branches
- `GetTask`

Notable tested `RunTask` branches:

- task not found
- singleton success path
- singleton DB failure path
- non-singleton `AddRun` failure path
- settings load success
- settings load failure but run still starts
- task options parse failure with valid raw JSON fallback behavior
- raw option merge still honoring `enableStreaming=false`

Recorded function-level checkpoints during this pass:

- `UpdateTaskOptions` about `93.8%`
- `RunTask` about `54.0%`

---

## What still looks missing

This is the practical backlog, ordered by expected payoff.

## Highest-yield Go gaps

### 1) `internal/controller`

This is still one of the biggest remaining gaps, though one useful batch has now been added for `task.go`.

Why it matters:

- low coverage relative to importance
- user-facing API / orchestration behavior lives here
- usually good ROI because a few focused tests can cover a lot of branch logic

Completed in the latest controller batch:

- added `internal/controller/task_controller_test.go`
- covered `HandleTasks` core CRUD/PATCH branches:
  - GET list
  - POST create
  - POST duplicate-name conflict
  - PUT update
  - PATCH options merge
  - invalid JSON / missing id / not found / method-not-allowed branches
- covered `HandleBootstrap` core behavior:
  - successful bootstrap response with tasks + activeRuns
  - active run progress normalization/clamping
  - completed-files fallback from stderr log parsing
  - finalizing phase when `finishWait.enabled=true` and not done
  - duration fields on `runRecord`
  - method-not-allowed and active-run-error branches
- further covered `HandleTaskActions` branches:
  - DELETE success
  - DELETE invalid id
  - DELETE task-not-found -> 500
  - POST `/api/tasks/:id/run` when task is already running
  - POST `/api/tasks/:id/run` task-not-found -> 500
  - non-`/run` POST path -> 404
- added `internal/controller/run_status_log_test.go`
- covered `RunController` entry branches:
  - `HandleRunStatus` GET success / GET not found / GET list error
  - `HandleRunStatus` DELETE success / DELETE error
  - `HandleRunLog` serving `stderrFile` from summary
  - `HandleRunLog` fallback search under `/app/data/logs/<task-xxxx>/`
  - `HandleRunLog` list error / missing run 404 branches
  - `HandleRuns` GET paging/defaults/list error + DELETE success/error
  - `HandleRunsByTask` GET filtered list/error + DELETE success/error
  - `HandleRunFiles` raw mode
  - `HandleRunFiles` read-log error path
  - `HandleRunFiles` move-mode copy+delete merge behavior
  - `HandleRunFiles` pagination branches
  - `HandleRunFiles` CAS-noise filtering branches
  - `HandleRunKillCLI` method/list-error/not-found/success branches
  - `HandleTaskKill` method/invalid-id/error/no-runs/no-pid/success branches
  - `killRunBySummary` real pid kill path
  - `HandleGlobalStats` list-error branch
  - `HandleGlobalStats` mixed map/string progress aggregation
  - `HandleGlobalStats` zero-total percentage branch
- added `internal/controller/settings_test.go`
- covered `SettingsController` branches:
  - `HandleSettings` GET / PUT / method-not-allowed dispatch
  - `handleGet` default values / stored overrides / env precedence / meta fields
  - `handlePut` valid-key persistence / unknown-key ignore / reset / hook triggering
  - `atoiDefault` helper paths
- coverage measurement was later re-run explicitly (not just tests):
  - `go test ./internal/controller -coverprofile=...` => later checkpoint about `45.9%`
  - `go test ./... -coverprofile=...` => overall checkpoint about `41.3%`
  - later `go test ./internal/controller -coverprofile=...` after `HandleRuns` / `HandleRunsByTask` batch => about `42.1%`
  - later `go test ./internal/controller -coverprofile=...` after `HandleRunFiles` batch => about `43.0%`
  - later `go test ./internal/controller -coverprofile=...` after kill-related batch => about `46.6%`
  - later `go test ./internal/controller -coverprofile=...` after `HandleGlobalStats` batch => about `47.9%`
  - latest `go test ./internal/controller -coverprofile=...` after `settings.go` batch => about `51.7%`
  - latest `go test ./internal/controller -coverprofile=/tmp/rcloneflow-controller.out` after task reorder persistence batch => `51.7%`
  - latest `go test ./... -coverprofile=/tmp/rcloneflow-all.out` after task reorder persistence batch => `48.5%`
  - latest `go test ./internal/store -coverprofile=/tmp/rcloneflow-store.out` after task `sort_index` self-heal batch => `47.1%`
  - latest `go test ./... -coverprofile=/tmp/rcloneflow-all.out` after task `sort_index` self-heal batch => `49.0%`
- notable latest `run.go` function checkpoints from that re-measurement:
  - `HandleRuns` about `96.4%`
  - `HandleRunsByTask` about `100.0%`
  - `HandleRunStatus` about `91.3%`
  - `HandleRunKillCLI` about `94.1%`
  - `HandleTaskKill` about `96.6%`
  - `killRunBySummary` about `90.9%`
  - `HandleRunFiles` about `69.6%`
  - `HandleRunLog` about `82.8%`
  - `HandleGlobalStats` about `92.3%`
- notable latest `settings.go` function checkpoints from that re-measurement:
  - `HandleSettings` about `100.0%`
  - `handleGet` about `100.0%`
  - `handlePut` about `90.6%`
  - `atoiDefault` about `100.0%`

Recommended remaining controller targets:

- active transfer related controller paths not yet covered
- lower-covered controller files such as `remote.go`, `browser.go`, `schedule.go`, and `auth.go`
- error/HTTP status mapping branches in other controllers
- any still-uncovered `HandleTaskActions` edges only if they remain materially useful

### 2) remaining `internal/service/task.go` branches

`RunTask` improved, but still has uncovered logic.

Recommended remaining branches:

- more singleton blocking / existing-run edge cases
- more effective/default option merge edges
- startup bookkeeping branches not yet exercised
- branches around task lifecycle that do not currently fail fast in simple test setups

Also worth checking:

- `ListTasks` was previously at `0%`

### 3) `internal/config`

This package was previously very low and may still have easy gains.

Recommended targets:

- default loading behavior
- invalid config fallback / error paths
- env/config precedence
- serialization / persistence edge cases

---

## Frontend gaps still worth doing

### 1) controller/API integration-adjacent behavior in composables

Even with the big test increase, some runtime wiring branches are still likely thin around:

- API error propagation edge cases
- stale-state refresh behavior
- multi-step async orchestration branches
- websocket-driven state transitions beyond the already covered core paths

### 2) composables with Vue lifecycle warnings

These tests are passing, but some are still a little rough structurally.

Known note:

- `useTaskHistoryLoader.test.ts`
- `useActiveTransferDetail.test.ts`

Warnings seen during earlier rounds:

- `onMounted/onUnmounted is called when there is no active component instance...`

Suggested cleanup:

- run these composables inside a tiny host component `setup()` instead of directly invoking them in bare test context

### 3) remaining UI state edges

Good candidates when coverage pushes resume:

- more task history paging/filter corner cases
- more transfer-detail websocket/state degradation edges
- more command-mode parsing/error recovery combinations
- modal/view-shell event wiring smoke tests if needed

---

## Suggested next order of work

If continuing the coverage push, the recommended order is:

1. `internal/controller`
2. remaining high-value branches in `internal/service/task.go`
3. `internal/config`
4. frontend lifecycle-warning cleanup / residual runtime edge branches

---

## Notes

This document is a working tracker, not a generated report.
Update it whenever a notable coverage batch lands, especially when:

- a new package becomes the main focus
- a previously recommended gap is closed
- overall Go or frontend coverage moves materially
