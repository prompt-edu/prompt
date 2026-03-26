# Certificate Service

A microservice for generating and managing course completion certificates in the PROMPT platform using [Typst](https://typst.app/) for PDF generation.

## Features

- Upload and manage certificate templates (Typst format)
- Preview certificates with mock data before releasing to students
- Generate certificates on-demand for individual students
- Track student certificate downloads
- Configurable release date for student access
- Compilation error reporting for template debugging

## Architecture

- **Server**: Go service (Gin framework) with sqlc for type-safe database operations
- **Client**: React micro-frontend (Webpack Module Federation) with TypeScript and shadcn/ui
- **Database**: PostgreSQL for template storage, configuration, and download tracking
- **Template Engine**: Typst compiler for PDF generation (included in Docker image)

## Development Setup

### Prerequisites

- Go 1.26+
- Node.js 18+
- Docker and Docker Compose
- PostgreSQL
- Typst compiler (`brew install typst` on macOS)
- golang-migrate CLI tool

### Environment Variables

```bash
# Database
DB_HOST_CERTIFICATE=localhost
DB_PORT_CERTIFICATE=5439
DB_USER=prompt-postgres
DB_PASSWORD=prompt-postgres
DB_NAME=prompt_certificate

# Service
SERVER_ADDRESS=localhost:8088
CORE_HOST=http://localhost:3000
SERVER_CORE_HOST=http://localhost:8080

# Keycloak
KEYCLOAK_HOST=http://localhost:8282
KEYCLOAK_REALM_NAME=prompt
```

### Running Locally

**Server:**

```bash
cd servers/certificate
go run main.go
```

**Client (as part of all micro-frontends):**

```bash
cd clients && yarn install && yarn run dev
```

**Supporting Services:**

```bash
docker-compose up db-certificate keycloak
```

### Database Setup

Migrations run automatically on server startup. To run manually:

```bash
migrate -path ./db/migration -database "postgres://prompt-postgres:prompt-postgres@localhost:5439/prompt_certificate?sslmode=disable" up
```

Regenerate sqlc code after modifying queries:

```bash
cd servers/certificate
sqlc generate
```

### Testing

```bash
cd servers/certificate
go test ./... -v
```

Tests use `testcontainers-go` for database isolation — Docker must be running.

## API Endpoints

All endpoints are under `/certificate/api/course_phase/:coursePhaseID/`.

### Configuration (Instructor)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/config` | Get certificate configuration |
| `PUT` | `/config` | Update certificate template |
| `PUT` | `/config/release-date` | Set/clear release date |
| `GET` | `/config/template` | Download raw template file |

### Certificate Generation

| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| `GET` | `/certificate/download` | All | Student downloads own certificate |
| `GET` | `/certificate/download/:studentID` | Admin, Lecturer, Editor | Download certificate for a student |
| `GET` | `/certificate/preview` | Admin, Lecturer | Preview with mock data |
| `GET` | `/certificate/status` | All | Check availability and download status |

### Participants

| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| `GET` | `/participants` | Admin, Lecturer, Editor | List participants with download status |

## Template Format

Certificate templates use the [Typst](https://typst.app/) markup language. Templates receive data via a JSON file.

### Data Access

Load certificate data in your template:

```typst
#let data = json("data.json")
```

### Available Fields

| Field | Type | Description |
|-------|------|-------------|
| `studentName` | string | Student's full name (first + last) |
| `courseName` | string | Course name from PROMPT |
| `teamName` | string | Team name (empty if not allocated) |
| `date` | string | Generation date (e.g., "January 2, 2006") |

### Example Template

```typst
#let data = json("data.json")

#set page(paper: "a4")
#set text(font: "Linux Libertine")

#align(center + horizon)[
  #text(size: 28pt, weight: "bold")[Certificate of Completion]

  #v(2em)

  #text(size: 18pt)[This certifies that]

  #v(1em)

  #text(size: 24pt, weight: "bold")[#data.studentName]

  #v(1em)

  #text(size: 18pt)[has successfully completed the course]

  #v(1em)

  #text(size: 22pt, style: "italic")[#data.courseName]

  #if data.teamName != "" [
    #v(1em)
    #text(size: 16pt)[as a member of team #data.teamName]
  ]

  #v(3em)

  #text(size: 14pt)[#data.date]
]
```

### Embedding Images

Since templates are stored as a single file, images (logos, signatures, etc.) must be embedded inline. Typst supports this via the `image()` function with `bytes()`.

**SVG graphics** (recommended for logos, icons, decorative elements):

```typst
#let logo = bytes("<svg xmlns='http://www.w3.org/2000/svg' width='200' height='100'>...</svg>")
#image(logo, width: 4cm)
```

**Raster images (PNG/JPG)** — wrap the base64-encoded image in an SVG data URI:

```typst
#let signature = bytes("<svg xmlns='http://www.w3.org/2000/svg' width='400' height='100'><image href='data:image/png;base64,iVBORw0KGgo...' width='400' height='100'/></svg>")
#image(signature, width: 5cm)
```

To get the base64 string from a PNG file:

```bash
base64 -i image.png | tr -d '\n'
```

Set the SVG `width` and `height` attributes to match the original image dimensions for correct aspect ratio.

### Template Requirements

- Must be a valid Typst (`.typ`) file
- Use `json("data.json")` to access certificate data (a `vars.json` symlink is also created for compatibility)
- Should target A4 page format
- Fonts must be system fonts or bundled with the template
- Images must be embedded inline (see above) — external file references will not work
- Use the **Test Certificate** button in the settings page to verify your template compiles correctly

## Database Schema

### `course_phase_config`

Stores the certificate template and configuration per course phase.

```sql
CREATE TABLE course_phase_config (
    course_phase_id     uuid PRIMARY KEY,
    template_content    text,
    created_at          timestamp with time zone NOT NULL DEFAULT NOW(),
    updated_at          timestamp with time zone NOT NULL DEFAULT NOW(),
    updated_by          text,
    release_date        timestamp with time zone
);
```

### `certificate_download`

Tracks student certificate downloads (only self-downloads are recorded, not instructor downloads).

```sql
CREATE TABLE certificate_download (
    id                  SERIAL PRIMARY KEY,
    student_id          uuid NOT NULL,
    course_phase_id     uuid NOT NULL,
    first_download      timestamp with time zone NOT NULL DEFAULT NOW(),
    last_download       timestamp with time zone NOT NULL DEFAULT NOW(),
    download_count      integer NOT NULL DEFAULT 1,
    CONSTRAINT idx_certificate_download_student_phase UNIQUE (student_id, course_phase_id)
);
```

## Docker

The certificate service uses a custom Dockerfile (unlike other services which share `servers/Dockerfile`) because it needs the Typst compiler bundled in the image.

```bash
# Build
docker build -t prompt-server-certificate ./servers/certificate

# Run
docker run -p 8088:8080 \
  -e DB_HOST_CERTIFICATE=host.docker.internal \
  -e DB_PORT_CERTIFICATE=5439 \
  -e DB_USER=prompt-postgres \
  -e DB_PASSWORD=prompt-postgres \
  -e DB_NAME=prompt_certificate \
  prompt-server-certificate
```
