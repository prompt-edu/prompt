---
paths:
  - "servers/**/*.go"
---

# Go Coding Style

Stack: Go 1.26, Gin, PostgreSQL (pgx), sqlc, golang-migrate. Extends `../common/coding-style.md`.

## Naming

- **PascalCase:** exported types/functions (e.g. `CourseService`, `CreateCourse`).
- **camelCase:** unexported functions, local variables.
- **snake_case:** database columns and tables.

## Module structure

```text
module/
  moduleDTO/        # request/response DTOs
  router.go         # Gin routes, auth middleware
  service.go        # business logic
  service_test.go   # tests
  validation.go     # input validation
```

## Error handling

```go
func handler(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // ... business logic
}
```

## Logging

```go
import log "github.com/sirupsen/logrus"
log.Debug(...) ; log.Info(...) ; log.Warn(...) ; log.Error("...: ", err)
```

## Environment variables

```go
dbHost := utils.GetEnv("DB_HOST", "localhost")
```

Critical: `DB_*` (per-service), `KEYCLOAK_*`, `CORE_HOST` (CORS), `SERVER_CORE_HOST` (inter-service), `DEBUG`.
