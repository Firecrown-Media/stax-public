# Stax Spec-Kit Directory

This directory contains specification artifacts for the Stax project, following the [spec-kit](https://github.com/github/spec-kit) methodology for spec-driven development.

## Directory Structure

```
.speckit/
├── README.md           # This file
├── constitution.md     # Project principles, standards, and constraints
├── specs/              # Feature specifications
├── plans/              # Implementation plans
└── tasks/              # Generated task lists
```

## Files

### constitution.md

The project constitution defines:
- Core principles guiding development
- Technical standards and conventions
- Architectural decisions
- Quality gates and release criteria

### specs/

Feature specifications that define the "what" and "why":
- User-focused capability descriptions
- Acceptance criteria
- Non-functional requirements

### plans/

Implementation plans that define the "how":
- Technical approach and architecture
- Phase breakdown
- Success criteria (automated and manual)

### tasks/

Generated task lists from plans:
- Actionable work items
- Dependencies between tasks
- Progress tracking

## Workflow

1. **Specify**: Define what needs to be built (`/speckit.specify`)
2. **Plan**: Design how to build it (`/speckit.plan`)
3. **Tasks**: Break into actionable items (`/speckit.tasks`)
4. **Implement**: Execute the tasks (`/speckit.implement`)
5. **Analyze**: Validate consistency (`/speckit.analyze`)

## Committing

These files **should be committed** to the repository:
- They serve as living documentation
- They help onboard new contributors
- They provide context for AI assistants
- They track architectural decisions over time

Exception: Files matching `*.local.md` are gitignored for personal notes.
