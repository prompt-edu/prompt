# Presentation service

The Presentation phase owns presentation scheduling, material metadata, and structured instructor
feedback. Material objects are stored through the configured S3-compatible storage service.

All course data routes are below:

```text
/presentation/api/course_phase/:coursePhaseID
```

The service supports individual and team targets, independent and shared feedback modes, phase
copy, phase deletion, privacy export/deletion, optimistic feedback revisions, and server-sent events
for shared editing.

Deleting a phase (`DELETE /presentation/api/course_phase/:coursePhaseID`, called by core) removes
its rows and every stored object under the phase's `presentations/<coursePhaseID>/` key prefix.
