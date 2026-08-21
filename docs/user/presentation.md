---
sidebar_position: 8
---

# 🎤 Presentation Course Phase

The **Presentation Phase** keeps presentation scheduling, supporting materials, and instructor feedback in one course phase. It supports presentations by individual students or by teams.

## Overview

Instructors can:

- configure individual or team presentations
- choose which materials presenters have to hand in
- create time slots, as a single slot or a whole series, and assign presenters
- define written feedback categories
- choose independent or shared feedback
- review uploaded materials and release named feedback to students

Students can see their assigned slot, hand in the requested materials, and read feedback after it has been released.

The main page of the phase is the presenter's own page. Instructors and administrators are not presenters, so they see that page filled with sample data and disabled, together with a note explaining it. Everything instructors manage lives on **Schedule** and **Settings**.

## Configure the Phase

Course lecturers and PROMPT administrators configure the phase from **Settings**:

1. Select **Individual** or **Team** as the presentation target.
2. For team presentations, connect the optional team and team-allocation inputs in the Course Configurator.
3. Select a feedback mode:
   - **Independent** keeps each instructor's draft separate. Submitted evaluations are visible to other instructors.
   - **Shared** gives instructors one collaborative form. Updates are synchronized while they edit.
4. Select the uploads presenters have to hand in, see [Request Materials](#request-materials).
5. Add and order the written feedback categories, such as *Content*, *Delivery*, and *Q&A*.

Changing a locked target, feedback mode, or category structure can invalidate existing presentation data. PROMPT shows the affected data and requires an explicit reset before applying such a change. Changing the requested uploads is not locked: it never deletes files that were already handed in.

## Request Materials

**Settings** lists the uploads a presentation can ask for:

| Upload | Accepted formats |
| --- | --- |
| Presentation slides | PDF, PPT, PPTX, ODP |
| Summary or report | PDF, DOC, DOCX, ODT |
| Handout | PDF |
| Poster | PDF, PNG, JPG |
| Source code archive | ZIP |
| Video recording | MP4 |

Every selected upload is mandatory and gets its own slot on the presenter's page, which names the accepted formats and shows whether the file is still missing. Files that do not fit the slot are rejected. Slide decks are requested by default; deselecting everything means presenters cannot attach files at all.

Removing an upload from the list leaves already uploaded files in place. They stay visible and downloadable, marked as no longer requested.

## Schedule Presentations

From **Schedule**, lecturers and administrators create slots with a start time, end time, and optional location, then assign a student or team. Slots may overlap, which allows parallel presentation rooms or tracks.

To lay out a whole session, tick **Create multiple slots** in the create dialog and give a slot duration and an optional break between slots. PROMPT divides the time range accordingly and shows how many slots it will create, up to 100 per batch. A series is created in one step, so a rejected slot never leaves half a session behind.

The slot table is also the instructor's overview: it shows the assigned presenter, how many materials arrived, whether evaluations were submitted or released, and it links into the feedback workspace.

Students cannot select their own slots. Their presentation page shows only the slot assigned by the teaching team.

## Upload Materials

Presenters can attach files to the requested upload slots until the presentation starts. In a team phase, every current team member can manage the team's materials. Staff can upload, download, replace, or remove materials at any time.

Uploads use the platform's configured file-size limit, 50 MB per file by default, which is worth keeping in mind for video recordings.

## Give Feedback

Course editors, lecturers, and administrators can evaluate a presentation using the configured categories.

- In **Independent** mode, each instructor owns a private draft and submits it independently.
- In **Shared** mode, instructors edit the same category responses. If two people change the same category at once, PROMPT detects the revision conflict instead of silently overwriting work.

Drafts are never released. When all intended evaluations have been submitted, a lecturer or administrator can release a named feedback snapshot. A release can be withdrawn, revised, and released again when corrections are needed.

## View Feedback as a Student

Released feedback appears on the student's or team's presentation page with the instructor's name. Feedback that is still a draft or has only been submitted internally remains hidden from students.
