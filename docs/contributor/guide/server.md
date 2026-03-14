# Server Guidelines

## 0. General

**Project Context:**  
This project follows a microservices architecture using Go (Golang) as the primary backend language. Please adhere to Go best practices, the project's architectural patterns, and the conventions described in the Architecture Documentation.

**Shared Code:**  
Utilities shared between microservices should be implemented in the [Prompt SDK](https://github.com/prompt-edu/Prompt-SDK).

## 1. Naming Conventions

- **PascalCase** should be used for:

  - Exported types, structs, and interfaces (e.g., `CourseService`)
  - Exported functions and methods (e.g., `CreateCourse`)
  - Package-level constants (e.g., `MaxRetryAttempts`)

- **camelCase** should be used for:

  - Unexported (private) functions and methods (e.g., `setupRouter`)
  - Local variables and function parameters (e.g., `courseID`)
  - Struct fields that are not exported
  - Unexported constants (e.g., `maxRetryAttempts`)

- **snake_case** should be used for:

  - Database column names and table names

- Use descriptive names that clearly indicate the purpose. For example:
  - Service files should be suffixed with `Service` (e.g., `CourseService`)
  - DTO files should be clearly named for their purpose (e.g., `CreateCourseRequest`)

## 2. Folder Structure

Each microservice follows a consistent modular structure. For each module/component, use the following folder organization:

```text
servers/
├── assessment/                     -- Core server (main application)
│   ├── module/                     -- Course management module
│   │   ├── moduleDTO/              -- Contains Data Transfer Objects for API requests/responses
│   │   ├── submodule/              -- Sub-module (in this case for course participation)
│   │   ├── main.go                 -- Setup module and middleware, initialize service as singletons
│   │   ├── router.go               -- HTTP endpoint definitions and middleware setup
│   │   ├── router_test.go          -- Router tests
│   │   ├── service.go              -- Core business logic using standalone functions with singleton dependencies
│   │   ├── service_test.go         -- Service tests for standalone functions
│   │   ├── validation.go           -- Input validation logic
│   │   └── validation_test.go      -- Validation tests
│   ├── db/                         -- Database layer
│   │   ├── migration/              -- SQL migration files
│   │   ├── query/                  -- SQL query files (for SQLC)
│   │   └── sqlc/                   -- SQLC auto-generated database code
│   ├── utils/                      -- Shared utilities
│   ├── testutils/                  -- Test utilities
│   ├── docs/                       -- Swagger documentation
│   ├── main.go                     -- Server entry point
│   ├── go.mod                      -- Go module dependencies
│   └── sqlc.yaml                   -- SQLC configuration
└── [other microservices]
```

**Architectural Pattern Note:**  
Services in this project use standalone functions rather than methods on service structs. Dependencies (database connections, queries, etc.) are initialized as singletons in `main.go` and passed to functions as needed. This pattern promotes simplicity, testability, and consistency across the codebase.

## 3. Dependency Management

### 3.1 Common Dependencies

The project uses several standard dependencies across services:

- **Gin** ([github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)) - HTTP web framework
- **pgx** ([github.com/jackc/pgx/v5](https://github.com/jackc/pgx/v5)) - PostgreSQL driver
- **UUID** ([github.com/google/uuid](https://github.com/google/uuid)) - UUID generation
- **Logrus** ([github.com/sirupsen/logrus](https://github.com/sirupsen/logrus)) - Structured logging
- **Testify** ([github.com/stretchr/testify](https://github.com/stretchr/testify)) - Testing framework
- **SQLC** ([github.com/sqlc-dev/sqlc](https://github.com/sqlc-dev/sqlc)) - SQL code generation

### 3.2 Prompt SDK

- Use the internal `Prompt-SDK` ([github.com/prompt-edu/Prompt-SDK](https://github.com/prompt-edu/Prompt-SDK)) for:
  - Authentication middleware (Keycloak integration)
  - Authorization roles and permissions (RBAC)
  - CORS configuration
  - Common utilities

**Authentication & Authorization:**

- Use Keycloak for authentication via the `Prompt-SDK`
- Define role-based access control (RBAC) for all endpoints → more details: [Authorization](../architecture/access-control.md)
- Use permission validation middleware

**Example from assessment server:**

```go
// In assessment main.go
func main() {
    // ... database setup ...

    router := gin.Default()
    router.Use(promptSDK.CORSMiddleware(coreHost))
    api := router.Group("assessment/api/course_phase/:coursePhaseID")

    // Use SDK authentication middleware in router setup
    assessments.InitAssessmentModule(api, *query, conn)
}

// In assessment router.go
func setupAssessmentRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
    assessmentRouter := routerGroup.Group("/student-assessment")

    // Use Prompt-SDK roles for authorization
    assessmentRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), listAssessmentsByCoursePhase)
}
```

## 4. Database Management

### 4.1 Migration System

- Use `golang-migrate` for database schema changes
- Place migration files in `db/migration/` directory
- Follow naming convention: `####_description.up.sql` and `####_description.down.sql`
- Run migrations on service startup

### 4.2 Query Management with SQLC

- Write SQL queries in `db/query/` directory
- Use SQLC to generate type-safe Go code
- Generated code goes in `db/sqlc/` directory

**Example SQLC configuration:**

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/query/"
    schema: "db/migration/"
    gen:
      go:
        package: "db"
        sql_package: "pgx/v5"
        out: "db/sqlc"
        emit_json_tags: true
        overrides:
          - db_type: "uuid"
            go_type:
              import: "github.com/google/uuid"
              type: "UUID"
```

## 5. API Design and Documentation

### 5.1 RESTful Conventions

- Follow REST principles for endpoint design:
  - Use plural nouns for collections: `/courses`, `/assessments`
  - Use specific resource IDs for individual items: `/courses/{courseId}`
  - Use HTTP methods semantically: GET (read), POST (create), PUT (update), DELETE (remove)
- All servers must expose routes under `<server>/api/course_phase/:coursePhaseID` to use the Prompt-SDK authentication middleware (e.g., `assessment/api/course_phase/:coursePhaseID`).

### 5.2 Swagger Documentation

- Use `swaggo/swag` for API documentation
- Add Swagger comments to all public endpoints
- Include examples and proper status codes in documentation
- Generate and serve documentation automatically

### 5.3 Data Transfer Objects (DTOs)

- Use DTOs for request and response payloads
- Create separate DTOs for different operations if they differ from the main entity (Create, Update, Get)
- Use JSON tags for proper serialization
- Convert between DTOs and database models explicitly

**Example DTO structure:**

```go
type CreateCourse struct {
    Name                string        `json:"name" binding:"required"`
    StartDate           pgtype.Date   `json:"startDate" swaggertype:"string"`
    EndDate             pgtype.Date   `json:"endDate" swaggertype:"string"`
    RestrictedData      meta.MetaData `json:"restrictedData"`
    StudentReadableData meta.MetaData `json:"studentReadableData"`
}

// Convert dto to database model
func (c CreateCourse) GetDBModel() (db.CreateCourseParams, error) {/*...*/}
```

## 6. Error Handling and Logging

### 6.1 Error Handling Policy

- Use consistent error response format across all services
- Log errors with appropriate context using logrus
- Return structured `utils.ErrorResponse` for client-facing errors
- Never expose internal error details to clients; use generic user-friendly messages

### 6.2 Logging Levels

```go
log.Info("Assessment service started on port 8085")
log.WithField("coursePhaseID", coursePhaseID).Warn("No assessments found for course phase")
log.WithError(err).Error("Database connection failed")
```
