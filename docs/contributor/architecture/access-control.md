---
sidebar_position: 3
---

# 🔐 Access Control

PROMPT 2 uses **role-based access control (RBAC)** to manage user permissions throughout the system. This is implemented through **Keycloak**, which issues JSON Web Tokens (JWTs) and manages both system-wide and course-specific roles.

The platform distinguishes four main user types:

* Administrator
* Lecturer
* Teaching Personnel
* Student

Based on these, the system defines **five roles**, each with specific permissions:

## 🎓 Access Roles

| Role                     | Scope           | Description                                                                                                                                                              |
| ------------------------ | --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Prompt Administrator** | System-wide     | Full access to all courses, users, and settings. Handles configuration and maintenance.                                                                                  |
| **Prompt Lecturer**      | System-wide     | Can create new courses and view student history across all past courses.                                                                                                 |
| **Course Lecturer**      | Course-specific | Assigned when creating a course. Can configure course structure, manage phases, assess students, and assign final grades.                                                |
| **Course Editor**        | Course-specific | Supports course execution. Can view participants and progress but cannot write data or assess students. May gain additional permissions per phase (e.g., give feedback). |
| **Course Student**       | Course-specific | Can apply to courses, participate in phases, and view their progress and grades.                                                                                         |

---

## 🧩 Role Resolution and Enforcement

The PROMPT 2 system distinguishes **system-wide roles** from **course-specific roles**.

### 🔐 System-Wide Roles: `Prompt Administrator` and `Prompt Lecturer`

These roles are managed directly in **Keycloak** and are encoded in each user's **JWT token**. Any component (core or course phase service) can validate these roles by inspecting the token.

### 📘 Course-Specific Roles: `Course Lecturer` and `Course Editor`

These roles are dynamically created and linked to specific courses. The role names are generated with a naming convention that includes the **semester** and **course name** (e.g., `ios25-iPraktikum-Lecturer`). Because services typically operate using internal identifiers like `courseID` or `coursePhaseID`, they must map these IDs to role names.

To support this, the **Core Server provides an endpoint** that returns the correct role name for a given course or phase ID. Services can then:

1. Request the role name using the ID
2. Check whether the user’s JWT contains the required role
3. Cache the result, since mappings rarely change

### 🧪 Example: Course Lecturer Access Flow

import LecturerAccess from './img/prompt_2_access_control_lecturer.png';

<img src={LecturerAccess} alt="PROMPT 2 Access Control Flow for Course Lecturers" />

In this flow:

* A lecturer logs in through the Core Client, which authenticates via Keycloak.
* Keycloak issues a JWT token, which the client uses to retrieve associated courses from the Core Server.
* When the lecturer accesses a course phase (e.g., the Intro Course), the microfrontend sends a request to the respective course phase service.
* This request includes the `jwtToken` and `coursePhaseID`.
* The course phase service contacts the Core Server to resolve the `coursePhaseID` into a role name.
* It checks the JWT for the corresponding role before allowing access.

Note: This role mapping is public and not user-specific. Therefore, it does **not** require authentication and can be **cached safely**.

---

## 👩‍🎓 Student Role and Phase Participation

The `Course Student` role is handled differently. Because student access is **phase-dependent** and students progress from phase to phase dynamically, their membership **cannot** be captured in static Keycloak roles.

Instead, the **Core Server maintains authoritative information** about which students are allowed in which phases. Course phase services must verify student membership using the **dedicated Core endpoint**.

### 🧪 Example: Course Student Access Flow

import StudentAccess from './img/prompt_2_access_control_student.png';

<img src={StudentAccess} alt="PROMPT 2 Access Control Flow for Course Students" />

In this flow:

* A student logs in and receives a JWT token from Keycloak.
* The Core Client only shows courses and phases the student is currently enrolled in.
* When accessing a course phase (e.g., a survey in the Intro Course), the phase service:

  1. Sends the `jwtToken` and `coursePhaseID` to the Core Server
  2. Core verifies that the student is part of the phase
  3. Returns the `courseParticipationID`
  4. The phase service uses this ID to record progress (e.g., survey response)

Because phase participation may change (e.g., after passing a phase), the membership check result **should not be cached long-term**.

---

## 🎯 Custom Roles for Fine-Grained Access

PROMPT 2 also supports the creation of **custom Keycloak roles** for advanced access scenarios. Instructors can define roles for specific course phases and assign them to users as needed.

These custom roles:

* Do **not** inherit permissions from system-defined roles
* Can be used by phase services to define access levels such as team membership or feedback permissions

Example: A course phase service may define team roles (`team-1`, `team-2`) and use them to differentiate student permissions within a team-based project.

---

## ✅ Summary

| Role Type            | Defined In  | Scope        | Validated Using       |
| -------------------- | ----------- | ------------ | --------------------- |
| Prompt Administrator | Keycloak    | System-wide  | JWT inspection        |
| Prompt Lecturer      | Keycloak    | System-wide  | JWT inspection        |
| Course Lecturer      | PROMPT Core | Per Course   | Mapped role + JWT     |
| Course Editor        | PROMPT Core | Per Course   | Mapped role + JWT     |
| Course Student       | PROMPT Core | Per Phase    | Membership endpoint   |
| Custom Role          | Keycloak    | Custom Logic | Service-defined usage |

---

## 🧑‍🤝‍🧑 Course Team Management API

Lecturers and Editors are managed via Keycloak groups (`/Prompt/{semesterTag}-{courseName}/Lecturer` and `.../Editor`). The PROMPT Core server exposes a small REST API on top of those groups so the UI does not have to talk to Keycloak directly. All routes live in `servers/core/keycloakRealmManager/`.

### Endpoints

| Method | Path | Authorization |
| :--- | :--- | :--- |
| `GET` | `/api/keycloak/:courseID/group/team` | `PromptAdmin`, `CourseLecturer` (per course) |
| `PUT` | `/api/keycloak/:courseID/group/:groupName/members/:userID` | `PromptAdmin`, `CourseLecturer` (per course) |
| `DELETE` | `/api/keycloak/:courseID/group/:groupName/members/:userID` | `PromptAdmin`, `CourseLecturer` (per course) |
| `GET` | `/api/keycloak/users/search?q=...&limit=...` | `PromptAdmin`, `CourseLecturer` (realm-wide) |
| `GET` | `/api/keycloak/status` | `PromptAdmin` only |

### Security Model

The defence against cross-course privilege escalation rests on **three** independent invariants:

1. **The Keycloak group path is derived server-side from the `:courseID` in the URL.** The client never sends a group name, path, or ID. `GetCourseSubgroup` (in `realmManagement.go`) is the only function that resolves the path - it concatenates `courseGroupName` (built from the DB row) with the sanitised role suffix.
2. **`:groupName` is validated against a hard-coded allow-list `{"Lecturer", "Editor"}` *before* any Keycloak call.** Anything else returns `400` immediately. The allow-list values are `permissionValidation.CourseLecturer` / `CourseEditor`, which are also passed to `GetGroupByPath` as `expectedName` so a Keycloak misdirect would still be rejected.
3. **Self-removal is rejected at the server with `400`.** The caller's Keycloak `sub` (read from `keycloakTokenVerifier.CtxUserID`) is compared against the `:userID` path parameter. This single rule prevents an empty Lecturer group via this UI - the final remover can never delete themselves - which makes a count check or advisory lock unnecessary.

Every mutating call also writes an INFO-level audit log line in the form:

```text
course-team audit: caller=<sub> action=<add|remove> target=<userID> group=<groupName> course=<courseID>
```

### Path Normalisation Gotcha

`gocloak.GetGroupByPath` joins URL segments with `/`, so a leading-slash `groupPath` produces `group-by-path//Prompt/...` which recent Keycloak versions reject under strict path normalisation (`missingNormalization: Request path not normalized`). PROMPT's `GetGroupByPath` wrapper strips the leading slash before delegating. Keycloak itself strips a leading slash internally (see `KeycloakModelUtils.findGroupByPath`), so behaviour is unchanged at the business-logic layer.

**If you add new code that calls `client.GetGroupByPath` directly, do not prefix the path with `/`.** Prefer routing through the `GetGroupByPath` wrapper or `GetCourseSubgroup`, both of which handle this for you.

### Service-Account Requirements

The Core server uses its `prompt-server` Keycloak client's service account for these calls. The required `realm-management` role mappings are documented in the [admin guide](../../admin/keycloak-prod#9-user-management-service-account). The `GET /api/keycloak/status` endpoint probes each of them and powers the **Keycloak** card on the admin **System Status** page.

