---
title: Audit Log
sidebar_position: 90
---

# Audit Log

The audit log records important actions taken across PROMPT so instructors and administrators can
see **who did what, and when**.

## Where to find it

- **Per course** (course lecturers and admins): open a course and select **Audit Log** in the course
  sidebar. It shows the actions that happened in that course.
- **Platform-wide** (admins only): **Admin → Audit Log** in the main sidebar shows actions across all
  courses.

## What is recorded

Each entry captures:

- **Time** the action happened.
- **Actor** — the person who performed it (name and email) and the **role** they used.
- **Action** — a human-readable description, e.g. "Created slot" or "Published grades".
- **Outcome** — whether the action **succeeded** or was **denied** (a blocked, unauthorized attempt).
- **Entity** — the specific thing acted on, when available (e.g. a team or student).
- **Source** — which part of the system reported it (core or a course phase).

Read-only actions (just viewing data) are not recorded.

## Searching and filtering

The toolbar above the table lets you narrow the log:

- **Search** by actor, action, or entity.
- **Outcome** — show only successful actions or only denied attempts.
- **Role** — filter to actions by a specific role (e.g. only lecturers, or only students).
- **Source** — restrict to a specific service.
- **Date range** — limit to a time window.

Use **Load more** at the bottom to page back through older entries.
