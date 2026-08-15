---
paths:
  - "servers/**/*.go"
---

# prompt-sdk API Catalog

Go SDK shared across all microservices (source: `github.com/prompt-edu/prompt-sdk`). Always prefer
these over custom implementations.

```go
import promptSDK "github.com/prompt-edu/prompt-sdk"
import "github.com/prompt-edu/prompt-sdk/promptTypes"
import sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
```

## Middleware & Auth

- `InitPhaseKeycloak() error` — **phase services use this**; reads `KEYCLOAK_HOST`,
  `KEYCLOAK_REALM_NAME`, and `SERVER_CORE_HOST` and calls `InitAuthenticationMiddleware`. Never
  hand-roll an `initKeycloak()`.
- `InitAuthenticationMiddleware(KeycloakURL, Realm, CoreURL string) error` — call once in `main.go`.
- `AuthenticationMiddleware(allowedRoles ...string) gin.HandlerFunc` — route protection factory.
- `CORSMiddleware(clientHost string) gin.HandlerFunc`.
- `GetTokenUser(c *gin.Context) (TokenUser, bool)` — `ID`, `Email`, `Roles`, `IsStudentOfCourse`,
  `CourseParticipationID`, …
- Role constants: `PromptAdmin`, `PromptLecturer`, `CourseLecturer`, `CourseEditor`, `CourseStudent`.

## Data fetching (inter-service)

- `FetchJSON(url, authHeader string) ([]byte, error)`
- `sdkUtils.SendCoreRequest(ctx, method, authHeader string, body io.Reader, url string)` — raw
  request against core when `FetchJSON` is not enough; pair with `sdkUtils.GetCoreUrl()`
- `FetchAndMergeParticipationsWithResolutions(coreURL, authHeader, coursePhaseID)`
- `FetchAndMergeCourseParticipationWithResolution(coreURL, authHeader, coursePhaseID, courseParticipationID)`
- `FetchAndMergeCoursePhaseWithResolution(coreURL, authHeader, coursePhaseID)`
- `ResolveParticipation`, `ResolveCoursePhaseData`, `ResolveAllParticipations`

## Utilities

- `GetEnv(key, defaultValue string) string`
- `DeferDBRollback(tx pgx.Tx, ctx context.Context)`
- `sdkUtils.GetCoreUrl() string` — core base URL from `SERVER_CORE_HOST`
- `sdkUtils.RunMigrations(databaseURL, migrationPath string) error` — startup migrations
- `sdkUtils.InitSentry(dsn string) error`

## Test helpers (`testutils` subpackage)

Import as `sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"`. Never fork these into a
service-local `testutils` package.

- `SetupTestDB(ctx, sqlDumpPath, queryFactory)` — testcontainers Postgres seeded from a dump:
  `sdkTestUtils.SetupTestDB(ctx, "../database_dumps/x.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })`
- `MockPermissionMiddleware(authRoles ...string) gin.HandlerFunc`, `MockAuthMiddleware(...)`

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
