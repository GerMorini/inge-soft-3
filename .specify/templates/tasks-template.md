---

description: "Task list template for feature implementation"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[###-feature-name]/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are REQUIRED for meaningful behavior introduced or changed by the feature. They
must cover the testable behaviors from spec.md and contribute to the project minimum of 8 useful
backend tests and 4 useful frontend tests. Documentation-only changes may state why tests do not apply.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Backend**: `backend/internal/<module>/{controller,service,repository,dto,dao}/`; `dto/` and
  `dao/` are optional and MUST appear only when the plan identifies concrete structures for them
- **Backend tests**: beside tested Go packages unless plan.md records an existing simpler layout
- **Frontend**: `frontend/src/` with tests in `frontend/tests/` or beside tested components
- **Frontend styles**: keep specific CSS beside its component or feature; reserve the global
  stylesheet for Tailwind/daisyUI integration, document-wide rules and shared palette tokens
- Paths MUST follow the concrete structure selected in plan.md

<!--
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.

  The /speckit-tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements from plan.md
  - Entities from data-model.md
  - Endpoints from contracts/

  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Tested independently
  - Delivered as an MVP increment

  DO NOT keep these sample tasks in the generated tasks.md file.
  ============================================================================
-->

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project structure per implementation plan
- [ ] T002 Initialize backend and React projects with only planned dependencies
- [ ] T003 Configure Tailwind CSS and daisyUI through the frontend build
- [ ] T004 [P] Configure linting and formatting tools

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

Examples of foundational tasks (include only when a current story requires them):

- [ ] T005 Setup database schema and migrations framework
- [ ] T006 [P] Implement only the authentication/authorization required by current stories
- [ ] T007 [P] Setup required API routes and middleware without generic frameworks
- [ ] T008 Create only models/entities shared by current stories
- [ ] T009 Configure required error handling and logging
- [ ] T010 Setup environment configuration without repository secrets

Create DTO or DAO tasks only when the plan identifies a concrete HTTP/use-case boundary or
persistence-specific structure. Do not generate directory, mapping, or duplicate-type tasks merely
to mirror the conceptual data flow.

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 1 *(required for changed behavior)* ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T011 [P] [US1] Backend test for [business behavior] in backend/internal/[module]/service/[name]_test.go
- [ ] T012 [P] [US1] Frontend test for [interface behavior] in frontend/tests/[name].test.tsx

### Implementation for User Story 1

- [ ] T013 [P] [US1] Add [Entity1] data type in backend/internal/[module]/[entity1].go
- [ ] T014 [US1] Add PostgreSQL access in backend/internal/[module]/repository/[name].go
- [ ] T015 [US1] Implement use case in backend/internal/[module]/service/[name].go
- [ ] T016 [US1] Implement HTTP mapping in backend/internal/[module]/controller/[name].go
- [ ] T017 [US1] Add validation and error handling
- [ ] T018 [US1] Add logging for user story 1 operations
- [ ] T019 [US1] Compose frontend with daisyUI, Tailwind utilities, the constitutional dark palette,
      and any justified custom CSS localized beside the affected component or feature

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 2 *(required for changed behavior)* ⚠️

- [ ] T020 [P] [US2] Backend test for [business behavior] in backend/internal/[module]/service/[name]_test.go
- [ ] T021 [P] [US2] Frontend test for [interface behavior] in frontend/tests/[name].test.tsx

### Implementation for User Story 2

- [ ] T022 [P] [US2] Add [Entity] data type in backend/internal/[module]/[entity].go
- [ ] T023 [US2] Implement use case in backend/internal/[module]/service/[name].go
- [ ] T024 [US2] Implement HTTP mapping in backend/internal/[module]/controller/[name].go
- [ ] T025 [US2] Integrate with User Story 1 components (if needed)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - [Title] (Priority: P3)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 3 *(required for changed behavior)* ⚠️

- [ ] T026 [P] [US3] Backend test for [business behavior] in backend/internal/[module]/service/[name]_test.go
- [ ] T027 [P] [US3] Frontend test for [interface behavior] in frontend/tests/[name].test.tsx

### Implementation for User Story 3

- [ ] T028 [P] [US3] Add [Entity] data type in backend/internal/[module]/[entity].go
- [ ] T029 [US3] Implement use case in backend/internal/[module]/service/[name].go
- [ ] T030 [US3] Implement HTTP mapping in backend/internal/[module]/controller/[name].go

**Checkpoint**: All user stories should now be independently functional

---

[Add more user story phases as needed, following the same pattern]

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] TXXX [P] Documentation updates in docs/ when required
- [ ] TXXX Verify daisyUI component reuse and remove unnecessary custom visual wrappers
- [ ] TXXX Verify feature CSS remains local, global CSS contains no component-specific rules, and
      semantic colors match the constitutional palette
- [ ] TXXX Code cleanup and refactoring
- [ ] TXXX Performance optimization across all stories
- [ ] TXXX [P] Additional behavior tests needed to preserve backend/frontend minimums
- [ ] TXXX Security hardening
- [ ] TXXX Run quickstart.md validation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Required behavior tests MUST be written and fail before implementation
- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all required behavior tests for User Story 1 together:
Task: "Backend test for [behavior] in backend/internal/[module]/service/[name]_test.go"
Task: "Frontend test for [behavior] in frontend/tests/[name].test.tsx"

# Launch all models for User Story 1 together:
Task: "Add [Entity1] data type in backend/internal/[module]/[entity1].go"
Task: "Add [Entity2] data type in backend/internal/[module]/[entity2].go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
