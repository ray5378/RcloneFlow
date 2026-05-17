# Task card borderless hover design

## Goal
Optimize the frontend task list card visual style so each task reads as a borderless rounded list item. The source, target, and progress details should blend into the task card instead of appearing inside a separate shaded block.

## Approved visual behavior
- Remove the independent background and rounded box styling from the source/target/progress area (`.task-paths`).
- Remove task-card borders, including the bottom divider and left accent line.
- Keep the task card default background transparent.
- Add spacing between adjacent task cards so separate tasks remain visually distinct.
- Give the whole task card rounded corners.
- Highlight the whole task card on hover:
  - dark theme: `rgba(99, 102, 241, 0.10)`
  - light theme: `rgba(25, 118, 210, 0.08)`
- Represent active/running state with a subtle whole-card background rather than a left line.

## Implementation scope
Update task-list card CSS in `frontend/src/components/task/TaskCard.vue` and shared list item base styling if needed. Keep existing structure and behavior intact: clicking a card still opens history, buttons still stop propagation, and progress text/bar behavior is unchanged.

## Verification
Run the frontend test/build gate available in the project after the style change. If there is no visual regression test, verify through static inspection of the CSS and the existing test suite/build.
