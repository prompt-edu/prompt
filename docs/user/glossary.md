---
sidebar_position: 8
title: Glossary
description: Plain-language definitions of the most important terms in PROMPT.
keywords:
  - glossary
  - terminology
  - course
  - phase
  - participation
  - lecturer
  - editor
  - student
  - application
  - interview
  - matching
  - assessment
  - certificate
---

# Glossary

This page collects the words you will see most often in PROMPT and explains what they mean in plain language. It is split into two parts:

- **Core Concepts** apply across the whole platform.
- **Course Phases** lists each phase you can add to a course, with the few terms specific to that phase.

Developers can find the technical equivalent of this page in the [Contributor Guide Glossary](/contributor/glossary).

---

## Core Concepts

**Course**
A class you run for a group of students, like an iOS Practical or a Patterns Seminar. A course has a name, start and end dates, a semester label, and the phases students go through.

**Semester**
A short label that pins your course to a term, such as `ss25` or `ws24`. You set it once when you create the course and it stays the same afterward.

**Phase**
A single step in your course - for example, collecting applications, running interviews, or evaluating students at the end. You chain phases together to define the path students take from start to finish.

**Phase Type**
The kind of work a phase does, such as Application, Interview, or Assessment. When you add a new phase to your course, you pick its type from the list of available phase types.

**Phase Graph**
The order students move through the phases of your course. You draw it visually in the Course Configurator by connecting one phase to the next.

**Course Configurator**
The page where you set up your course. You add phases, connect them in the order students will move through them, and decide which information flows from one phase to the next.

**Participation**
A student's record of being in your course. It tracks who the student is and how far they have progressed through your phases.

**Pass Status**
The outcome of a phase for one student: passed, failed, or not yet judged. You set the pass status to decide who moves on to the next phase.

**Template**
A saved copy of a course that you can reuse next semester. Templates let you recreate a familiar course in seconds instead of building it from scratch every term.

**Lecturer**
A course owner. Lecturers have full control over their course: phases, participants, settings, and data.

**Editor**
A teaching assistant or co-organizer of a course. Editors run the day-to-day work inside phases but do not change the overall course structure.

**Student**
A participant enrolled in the course.

**PROMPT Lecturer**
Someone allowed to create new courses on the platform. Every Lecturer of a course is also a PROMPT Lecturer, but a PROMPT Lecturer may not yet own any courses.

**PROMPT Admin**
A platform-wide administrator with access to every course and every setting in PROMPT.

---

## Course Phases

### Application

Students fill out a form to apply for your course. You decide what to ask them and when applications open and close.

- **Application Form** - the set of questions students see when they apply.
- **Application Question** - a single question on the form, such as a text answer, a multi-select, or a checkbox.
- **Application Answer** - one student's response to a question.

### Interview

Schedule short interviews with applicants and capture how they did.

- **Interview Slot** - a time window when an interview takes place.
- **Interview Question** - a question you want every interviewer to ask.
- **Interview Score** - the rating an interviewer gives a student after an interview.

### Matching

Assign accepted students to specific courses or projects based on what they would like to do.

- **Preference** - a student's wish for a project or course, ranked from most to least preferred.
- **Ranking** - the ordered list of preferences a student submits.
- **Allocation** - the final assignment of a student to a project or course.

### Team Allocation

Form teams automatically based on student skills and what they would like to work on.

- **Skill** - something a student says they are good at, used to balance teams.
- **Team Preference** - a student's wish to work with specific peers or on specific projects.
- **Survey** - the questionnaire students fill out before allocation runs.
- **Tutor** - a teaching assistant who guides one or more teams.

### Self Team Allocation

Let students form their own teams within a time window you control.

- **Team Preference** - the teammates a student wants to work with.
- **Survey Timeframe** - the window during which students can fill out the team survey.
- **Skill Response** - a student's answer about which skills they bring to a team.

### Assessment

Evaluate students against a structured set of criteria at the end of the course.

- **Competency** - a single skill or behavior you are evaluating, such as "code quality" or "communication".
- **Category** - a group of related competencies that belong together.
- **Schema** - the full set of categories and competencies you grade against.
- **Score Level** - one of the levels a student can receive on a competency, from very bad to very good.
- **Self Evaluation** - the student's own grading of their performance.
- **Tutor Evaluation** - a student's grading of their tutor's performance.

### Certificate

Generate and hand out course completion certificates.

- **Certificate Template** - the layout and wording your certificates will use.
- **Release Date** - the day students can first download their certificate.
- **Download** - the action a student takes to get their certificate as a PDF.
