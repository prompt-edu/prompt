# Interview Server

This microservice handles interview scheduling for the PROMPT platform.

## Features

- **Interview Slot Management**: Create, update, and delete interview time slots
- **Student Booking**: Students can book available interview slots
- **Capacity Management**: Track slot capacity and prevent overbooking
- **Location Support**: Optional location information for each slot

## Database Schema

### Tables

- `interview_slot`: Stores interview time slots with start/end times, location, and capacity
- `interview_assignment`: Links students (via course_participation_id) to their booked slots

## API Endpoints

### Interview Slots (Admin)

- `POST /interview/api/course_phase/:coursePhaseID/interview-slots` - Create a new slot
- `GET /interview/api/course_phase/:coursePhaseID/interview-slots` - List all slots
- `GET /interview/api/course_phase/:coursePhaseID/interview-slots/:slotId` - Get specific slot
- `PUT /interview/api/course_phase/:coursePhaseID/interview-slots/:slotId` - Update slot
- `DELETE /interview/api/course_phase/:coursePhaseID/interview-slots/:slotId` - Delete slot

### Interview Assignments (Students)

- `POST /interview/api/course_phase/:coursePhaseID/interview-assignments` - Book a slot
- `GET /interview/api/course_phase/:coursePhaseID/interview-assignments/my-assignment` - Get current booking
- `DELETE /interview/api/course_phase/:coursePhaseID/interview-assignments/:assignmentId` - Cancel booking

## Environment Variables

- `DB_HOST_INTERVIEW`: Database host (default: localhost)
- `DB_PORT_INTERVIEW`: Database port (default: 5438)
- `DB_NAME`: Database name (default: prompt)
- `DB_USER`: Database user (default: prompt-postgres)
- `DB_PASSWORD`: Database password (default: prompt-postgres)
- `SERVER_ADDRESS`: Server bind address (default: localhost:8087)
- `KEYCLOAK_HOST`: Keycloak authentication server URL
- `KEYCLOAK_REALM_NAME`: Keycloak realm (default: prompt)
- `CORE_HOST`: Frontend URL for CORS
- `SERVER_CORE_HOST`: Core server URL for inter-service communication

## Running Locally

1. Ensure PostgreSQL is running (or use Docker Compose)
2. Set environment variables (or use `.env` file)
3. Run migrations automatically on startup
4. Start the server:

```bash
go run main.go
```

## Docker Compose

The service is configured in the main `docker-compose.yml`:

```yaml
server-interview:
  container_name: prompt-server-interview
  ports:
    - "8087:8080"
  depends_on:
    - db-interview
```

## Development

### Adding New Queries

1. Add SQL queries to `db/query/interview_slot.sql`
2. Run `sqlc generate` to generate Go code
3. Use the generated methods in your service layer

### Running Tests

```bash
go test ./...
```
