---
sidebar_position: 7
---

# Templating System (Developer Guide)

This document explains how the templating system works end-to-end for developers: data model, server endpoints, client flows, and expected copy semantics.

---

## Concept

A template is a first-class course record with `template = true`. It acts as a blueprint to create new courses. Creating from a template performs a deep copy of the original course’s configuration.

---

## Data Model and Constraints

- `course.template` (boolean) marks a course as a template.
- Migrations support template courses: e.g., `0015_check_dates_template.up.sql` allows `end_date` to be null for template courses.
- Name/SemesterTag validation on the client prevents `-` in either field.
- ECTS is fixed for some course types (Seminar = 5, Practical = 10) via UI rules; Lecture requires manual input.

---

## Server (Core) Endpoints

All routes below are in `servers/core` and use Gin with permission middleware.

- GET `/api/courses/template` → List template courses
  - Implementation: `course/router.go#getTemplateCourses` → `course/service.go#GetTemplateCourses`
  - Admins get all template courses; Lecturers see only templates they have roles for (restricted query).
- GET `/api/courses/:uuid/template` → Check if a course is a template
  - Implementation: `course/router.go#checkCourseTemplateStatus` → `service.CheckCourseTemplateStatus`
- PUT `/api/courses/:uuid/template` → Mark/unmark a course as a template
  - Implementation: `course/router.go#updateCourseTemplateStatus` → `service.UpdateCourseTemplateStatus`
  - DB: `MarkCourseAsTemplate` / `UnmarkCourseAsTemplate` in `db/query/template.sql`
- POST `/api/courses/:uuid/copy` → Copy a course (optionally creating a template)
  - Implementation: `course/copy/router.go#copyCourse` → `copy.copyCourseInternal`
  - See Copy Semantics below.
- GET `/api/courses/:uuid/copyable` → Verify all phases expose `/copy` endpoint and are reachable.

### Copy Semantics (servers/core/course/copy/copy_course.go)

When copying with `Template = false`:

- Creates a new course with given Name, SemesterTag, StartDate, EndDate
- Copies:
  - Course phases and their order (phase graph)
  - Phase/data/participation graphs
  - DTO mappings and meta graphs
  - Application form if both source and target course have an application phase
  - Student-readable and restricted metadata
  - CourseType and ECTS
- Creates Keycloak roles/groups for the new course

When copying with `Template = true`:

- Same as above, but:
  - StartDate and EndDate are omitted (NULL dates)
  - The new course is marked as template (`MarkCourseAsTemplate`)

Error handling: Runs within a DB transaction; partial failures roll back. Some operations (e.g., phase configuration copy) are executed after commit and may still return errors up the call chain.

---

## Server (Assessment) — Assessment Templates

Separate from course templates, the assessment service exposes assessment template management:

- Router: `servers/assessment/assessmentTemplates/router.go`
- Service: `servers/assessment/assessmentTemplates/service.go`
- DTOs: `assessmentTemplateDTO/*`

Endpoints under `/assessment-template` (Prompt Admin for mutating routes):

- GET `/assessment-template` → List
- GET `/assessment-template/:templateID` → Get by ID
- POST `/assessment-template` → Create
- PUT `/assessment-template/:templateID` → Update
- DELETE `/assessment-template/:templateID` → Delete

Use `GetCoursePhasesByAssessmentTemplate` to find course phases using a given assessment template.

---

## Client (Core) Flow

Key files under `clients/core`:

- Course creation choices UI: `managementConsole/courseOverview/AddingCourse/components/CourseCreationChoiceDialog.tsx`
  - Offers: New Course, Create New Template, Use Template
- Create Template from scratch: `managementConsole/courseOverview/AddTemplateDialog.tsx`
  - Properties form: `AddTemplateProperties.tsx` with `templateFormSchema`
  - Creates course via POST `/api/courses/` with `template: true`
- Make Template from existing course: `managementConsole/courseSettings/components/CourseTemplateToggle.tsx`
  - Opens `CopyCourseDialog` with `createTemplate=true`
  - `useTemplateForm` → validates name/semester tag
  - Backend call: POST `/api/courses/:id/copy` with `Template = true`
- Use Template to create course: `TemplateSelectionDialog.tsx`
  - Queries templates with `coreApi.courses.listTemplates` (GET `/api/courses/template`)
  - On select → opens `CopyCourseDialog` (`useTemplateCopy = true`, `createTemplate = false`)
  - Backend call: POST `/api/courses/:templateId/copy` with `Template = false`

Networking helpers:

- `src/network/api/courses.ts` — `listTemplates`, `templateStatus`, `setTemplateStatus`, `copy`
- `src/network/hooks/useCopyCourse.ts`
- `src/network/hooks/useTemplateForm.ts`

Validations:

- `src/validations/template.ts` and `makeTemplateCourse.ts` enforce name and semester tag rules and ECTS input constraints.

---

## Permissions

- Listing templates: Prompt Admin, Course Lecturer
- Creating templates (from scratch or from existing course): Prompt Admin, Course Lecturer
- Using templates to create a course: Prompt Admin, Course Lecturer

The server additionally filters template visibility for non-admins based on the user’s course roles.

---

## Edge Cases and Notes

- Name/SemesterTag uniqueness and format are enforced downstream; UI prevents hyphens and empty values.
- If the source course lacks an application phase, the application form is not copied (no-op).
- Copyability check: `/copyable` verifies all phases implement copy; UI shows a confirmation step.
- After creating a course (or template), the client refreshes Keycloak token and refetches courses to ensure new roles are loaded for navigation.

---

## Extending the System

- To add fields that should be copied, update `copy_course.go` and the DTO/metadata conversion utilities.
- To expose new template kinds (e.g., per-phase templates), add appropriate endpoints in the respective service and mirror the client flows.
