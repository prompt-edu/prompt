# Presentation service

The Presentation phase owns presentation scheduling, material metadata, and structured instructor
feedback. Material objects are stored through the configured S3-compatible storage service.

All course data routes are below:

```text
/presentation/api/course_phase/:coursePhaseID
```

The service supports individual and team targets, independent and shared feedback modes, phase
copy, privacy export/deletion, optimistic feedback revisions, and server-sent events for shared
editing.
