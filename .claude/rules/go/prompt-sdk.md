---
paths:
  - "servers/**/*.go"
---

# prompt-sdk API Catalog

Go SDK shared across all microservices (source: `../Prompt-SDK/`). Always prefer these over custom
implementations.

```go
import promptSDK "github.com/ls1intum/prompt-sdk"
import "github.com/ls1intum/prompt-sdk/promptTypes"
```

## Middleware & Auth

- `InitAuthenticationMiddleware(KeycloakURL, Realm, CoreURL string) error` — call once in `main.go`.
- `AuthenticationMiddleware(allowedRoles ...string) gin.HandlerFunc` — route protection factory.
- `CORSMiddleware(clientHost string) gin.HandlerFunc`.
- `GetTokenUser(c *gin.Context) (TokenUser, bool)` — `ID`, `Email`, `Roles`, `IsStudentOfCourse`,
  `CourseParticipationID`, …
- Role constants: `PromptAdmin`, `PromptLecturer`, `CourseLecturer`, `CourseEditor`, `CourseStudent`.

## Data fetching (inter-service)

- `FetchJSON(url, authHeader string) ([]byte, error)`
- `FetchAndMergeParticipationsWithResolutions(coreURL, authHeader, coursePhaseID)`
- `FetchAndMergeCourseParticipationWithResolution(coreURL, authHeader, coursePhaseID, courseParticipationID)`
- `FetchAndMergeCoursePhaseWithResolution(coreURL, authHeader, coursePhaseID)`
- `ResolveParticipation`, `ResolveCoursePhaseData`, `ResolveAllParticipations`

## Utilities

- `GetEnv(key, defaultValue string) string`
- `DeferDBRollback(tx pgx.Tx, ctx context.Context)`

## Validation (`utils` subpackage)

- `RegisterValidation(tag, fn)`, `ValidateStruct(s)`
- Built-in validators: `matriculationNumber` (8 digits starting with `0`), `universityLogin`
  (TUM ID format `aa00aaa`).

## Types (`promptTypes` subpackage)

- Domain: `Person`, `Student`, `Team`, `MetaData`, `CoursePhaseParticipationWithStudent`
- Application answers: `AnswersText`, `AnswersMultiSelect`, `ReadApplicationAnswersFromMetaData(data)`
- Phase management: `PhaseCopyRequest`, `PhaseCopyHandler`, `PhaseConfigHandler`
- Endpoint registration: `RegisterCopyEndpoint(...)`, `RegisterConfigEndpoint(...)`
- Enums: `Gender` (`male`|`female`|`diverse`|`prefer_not_to_say`), `StudyDegree` (`bachelor`|`master`)
