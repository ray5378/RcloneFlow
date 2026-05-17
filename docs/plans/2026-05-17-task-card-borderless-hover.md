# Task Card Borderless Hover Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make frontend task list cards borderless, rounded, spaced, and highlighted by whole-card hover/active backgrounds.

**Architecture:** This is a presentational CSS-only change. Keep the Vue template and task behavior unchanged, and adjust the component-scoped card styles plus the shared list-item base styles imported by task cards.

**Tech Stack:** Vue 3 single-file components, scoped CSS, Vite/Vitest frontend.

### Task 1: Add CSS guard test for borderless task cards

**Files:**
- Modify: `frontend/src/components/task/TaskCard.test.ts`
- Inspect: `frontend/src/components/task/TaskCard.vue`
- Inspect: `frontend/src/components/task/listItemBase.css`

**Step 1: Write the failing test**
Add a source-inspection test that asserts the approved style contract:
- `TaskCard.vue` should not define `.task-paths` with a `background` or `border-radius`.
- `TaskCard.vue` should include the dark/light whole-card hover colors.
- `listItemBase.css` should remove `border-bottom` and `border-left` from shared task/run list items.

**Step 2: Run test to verify it fails**
Run: `cd frontend && npm test -- TaskCard.test.ts`
Expected: FAIL before CSS implementation.

**Step 3: Commit test**
Run:
```bash
git add frontend/src/components/task/TaskCard.test.ts
git commit -m "test: cover borderless task card styling"
```

### Task 2: Implement borderless rounded task card styling

**Files:**
- Modify: `frontend/src/components/task/TaskCard.vue`
- Modify: `frontend/src/components/task/listItemBase.css`

**Step 1: Update shared base list item styling**
In `listItemBase.css`:
- Remove visible border-bottom and border-left defaults for `.task-card` / `.run-item`.
- Add rounded corners.
- Add bottom spacing between adjacent items.
- Keep pointer cursor and smooth background transition.

**Step 2: Update TaskCard scoped styling**
In `TaskCard.vue`:
- Remove hover left-line styling.
- Set dark hover background to `rgba(99, 102, 241, 0.10)`.
- Set light hover background to `rgba(25, 118, 210, 0.08)`.
- Set active/running card background to a subtle persistent whole-card tint.
- Remove `.task-paths` independent background and radius, leaving only light spacing/padding.

**Step 3: Run targeted test**
Run: `cd frontend && npm test -- TaskCard.test.ts`
Expected: PASS.

**Step 4: Run frontend verification**
Run the smallest project gate available, preferably:
```bash
cd frontend && npm test -- --run
cd frontend && npm run build
```
Expected: both pass.

**Step 5: Commit implementation**
Run:
```bash
git add frontend/src/components/task/TaskCard.vue frontend/src/components/task/listItemBase.css frontend/src/components/task/TaskCard.test.ts
git commit -m "style: make task cards borderless with hover highlight"
```
