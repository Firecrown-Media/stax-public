# Task Lists

This directory contains generated task lists from implementation plans.

## Task Format

Tasks are generated from plans using `/speckit.tasks` and follow this format:

```markdown
# Tasks for [Feature Name]

## Phase 1: [Phase Name]

- [ ] Task 1 description
  - Dependencies: none
  - Files: path/to/file.go
- [ ] Task 2 description
  - Dependencies: Task 1
  - Files: path/to/other.go
```

## Naming Convention

`YYYY-MM-DD-feature-name-tasks.md`

## Notes

- Files matching `*.local.md` are gitignored for personal task tracking
- Update task status as work progresses
- Reference the parent plan for context

## Current Task Lists

_(Add task lists as they are generated)_
