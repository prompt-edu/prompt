---
sidebar_position: 8
title: Glossary
description: Technical reference for PROMPT's domain terminology with pointers into the codebase.
keywords:
  - glossary
  - terminology
  - domain model
  - course phase
  - microfrontend
  - module federation
  - keycloak
  - rbac
---

# Glossary

This page is the developer-side counterpart to the [User Guide Glossary](/user/glossary). It uses the same two-section split (Core Concepts, then Course Phases) and adds pointers into the codebase so you can jump from a term to where it lives.

Pointers reference directories or files only, never line numbers, so they keep working through refactors. The most useful anchor files are:

- Domain model: `servers/core/db/sqlc/models.go`
- Role definitions: `servers/core/permissionValidation/roles.go`
- Phase registry: `servers/core/db/query/course_phase_type.sql`

---

## Core Concepts

**Course**
A top-level container with a name, type (Lecture, Seminar, Practical Course), date range, and semester tag. Owns a set of course phases and a set of course participations.
See: `servers/core/course/`, `servers/core/db/sqlc/models.go`

**Semester Tag**
A short, immutable identifier for a course instance (`ss25`, `ws24`, `ios25`). Used to distinguish course iterations.
See: `servers/core/course/`

**Course Phase**
A modular step in a course. Each phase belongs to a Course Phase Type and is rendered by a microfrontend that talks to its own service.
See: `servers/core/coursePhase/`

**Course Phase Type**
The "kind" of a phase (Application, Interview, Assessment, etc.). Registered centrally and referenced by every concrete phase. Defines the meta-data contract a phase exposes.
See: `servers/core/db/query/course_phase_type.sql`

**Course Phase Graph**
The directed, linear ordering students follow through a course's phases. Forward-only, no branches or loops.
See: `servers/core/course/`, `servers/core/db/sqlc/models.go`

**Course Configurator**
The instructor-facing UI for building the phase graph and configuring data flow between phases.
See: `clients/core/`

**Course Participation**
A student's enrollment in a course. Links a Student to a Course at the course level.
See: `servers/core/course/courseParticipation/`

**Course Phase Participation**
A student's enrollment record inside one specific phase. Carries the pass status for that phase plus the per-phase metadata payloads.
See: `servers/core/coursePhase/coursePhaseParticipation/`

**Pass Status**
Per-phase outcome enum: `passed`, `failed`, `not_assessed`. Drives student progression to the next phase.
See: `servers/core/coursePhase/coursePhaseParticipation/`

**Restricted Data**
JSON metadata writable by Lecturers and readable only by Lecturers/Editors. Used for confidential per-phase or per-participation state.
See: `servers/core/coursePhase/coursePhaseDTO/`

**Student-Readable Data**
JSON metadata writable by Lecturers and readable by Students, Lecturers, and Editors. Used for phase outputs students should see (e.g., interview slot assignment).
See: `servers/core/coursePhase/coursePhaseDTO/`

**Participation Data Dependency Graph**
Defines which per-student outputs of one phase become inputs to a later phase (e.g., application scores feeding the matching phase).
See: `servers/core/course/`, `servers/core/db/sqlc/models.go`

**Phase Data Dependency Graph**
Same idea as above, but at the phase (aggregate) level rather than per-student.
See: `servers/core/course/`, `servers/core/db/sqlc/models.go`

**Template**
A course flagged as `template = true`. Used as a blueprint to clone a fresh course instance with the same phase structure, graphs, and application form.
See: `docs/user/templates.md`, `servers/core/course/`

**Roles**
RBAC roles enforced by the permission validation layer:
- `PROMPT_Admin` - global admin
- `PROMPT_Lecturer` - global course-creation right
- `Course Lecturer` (`<courseId>-Lecturer`) - course-scoped owner
- `Course Editor` (`<courseId>-Editor`) - course-scoped editor
- `Course Student` (`<courseId>-Student`) - course-scoped participant

See: `servers/core/permissionValidation/roles.go`

**Keycloak**
The identity provider PROMPT integrates with for authentication and group-based role assignment.
See: `docs/contributor/keycloak-dev.md`

**Microfrontend / Module Federation**
Each course phase ships a client bundle (a Webpack Module Federation remote) that the core client loads dynamically at runtime. The shell hands the remote a context object and embeds the phase UI inline.
See: `docs/contributor/new_microfrontend.md`, `clients/core/`

**Remote Entry Service**
The mechanism that resolves a phase type to the URL of its microfrontend bundle, allowing the shell to discover and load phase clients without redeploying.
See: `servers/core/`, `clients/core/`

---

## Course Phases

Each phase below has a paragraph definition, a short list of key terms, and a "Code:" line pointing to the client (microfrontend) and server (Go service) modules.

### Application

Collects student applications via a configurable form. Outcomes feed the rest of the course.

- Application Form, Application Question, Application Answer, Auto Accept
- Code: client lives in `clients/core/` (application module shipped from core), server in `servers/core/` (application is part of core)

### Interview

Schedules interview slots, captures interviewer answers and scores per applicant.

- Interview Slot, Interview Question, Interview Assignment, Interview Score
- Code: client `clients/interview_component/`, server `servers/interview/`

### Matching

Assigns accepted students to projects or courses based on submitted preferences. Owns no dedicated server - handled by core.

- Preference, Ranking, Allocation, Data Import/Export
- Code: client `clients/matching_component/`, server: handled by `servers/core/`

### Team Allocation

Algorithmic team formation from skill surveys and team-mate preferences.

- Skill, Survey, Team Preference, Allocation, Tutor, Timeframe
- Code: client `clients/team_allocation_component/`, server `servers/team_allocation/`

### Self Team Allocation

Student-driven team formation inside a configurable survey window.

- Team Preference, Survey Timeframe, Skill Response
- Code: client `clients/self_team_allocation_component/`, server `servers/self_team_allocation/`

### Assessment

Structured per-student evaluation using a configurable schema.

- Competency, Category, Assessment Schema, Score Level (`VeryBad`, `Bad`, `Ok`, `Good`, `VeryGood`), Self Evaluation, Peer Evaluation, Tutor Evaluation
- Code: client `clients/assessment_component/`, server `servers/assessment/`

### Certificate

Generates and distributes course completion certificates.

- Certificate Template, Release Date, Download Status
- Code: client `clients/certificate_component/`, server `servers/certificate/`

### Presentation

Schedules individual or team presentations, stores slot-scoped materials, and collects categorized multi-instructor feedback in independent or shared editing mode.

- Presentation Slot, Presenter Assignment, Material, Feedback Category, Feedback Session, Feedback Release
- Code: client `clients/presentation_component/`, server `servers/presentation/`
