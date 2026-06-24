---
paths:
  - "servers/**/*.go"
---

# Auth & Routing (Go services)

Authentication is provided by `prompt-sdk` (Keycloak-backed). Extends `../common/security.md`.

## Required route shape

Routes that serve course data MUST be mounted under `<server>/api/course_phase/:coursePhaseID` —
the SDK auth middleware depends on this path structure.

## Wiring

```go
import promptSDK "github.com/ls1intum/prompt-sdk"

// once in main.go
promptSDK.InitAuthenticationMiddleware(keycloakURL, realm, coreURL)

// per route — middleware factory takes allowed roles
router.GET("/endpoint", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), handler)
```

## Roles

`PromptAdmin`, `PromptLecturer`, `CourseLecturer`, `CourseEditor`, `CourseStudent`. Course-specific
roles are created dynamically with a naming convention that includes semester and course name.

## Helpers

- `GetTokenUser(c)` → authenticated user (`ID`, `Email`, `Roles`, `IsStudentOfCourse`,
  `CourseParticipationID`, …).
- `CORSMiddleware(clientHost)` — configure with `CORE_HOST`.
- Prefer SDK fetch helpers (`FetchJSON`, `FetchAndMerge*`) over hand-rolled inter-service HTTP.
