---
sidebar_position: 5
---

# Pull Request Structure

Every change reaches `main` through a pull request. A consistent PR structure keeps the review
fast and the release notes readable. This page describes how to title and describe a PR.

## Title

Use the format:

```text
`Tag`: Short description
```

- The **tag** is backticked and names the affected area.
- The **description** is a plain, sentence-case summary suitable for release notes (capitalize only
  the first word). Use `Fix` for bugfixes and another action verb (`Add`, `Allow`, `Show`,
  `Enable`, …) otherwise. Do not repeat the tag inside the description, and do not add emojis.

Examples:

```text
`Assessment`: Fix tutor evaluation without deadline blocking Unmark as Final
`Team Allocation`: Scope tutor access to assigned team
`Documentation`: Add pull request structure guide
```

### Available tags

`Core`, `Course Phases`, `Authentication`, `Module Federation`, `Interview`, `Team Allocation`,
`Matching`, `Assessment`, `Certificate`, `Intro Course`, `Self Team Allocation`, `Templating`,
`New Phase`, `Database`, `API`, `Testing`, `Documentation`, `Infrastructure`, `Deployment`,
`UI/UX`, `Performance`, `General`.

Use `General` when none of the others fit.

## Body

Fill in every section of the [PR template](https://github.com/prompt-edu/prompt/blob/main/.github/pull_request_template.md).
Keep all headings, helper text, and checklist items, and tick the boxes that apply.

| Section                        | What to write                                                              |
| ------------------------------ | -------------------------------------------------------------------------- |
| **What is the change?**        | A brief description of what was changed or added.                          |
| **Reason for the change**      | Why the change was made. Link the related issue (e.g. `Closes #123`).      |
| **How to Test**                | Numbered steps a reviewer can follow to verify the change.                 |
| **Screenshots**                | Before/after images for any visual change.                                 |
| **PR Checklist**               | Tick what is true: tested, clean code, tests updated, screenshots, docs.   |

## Before you open the PR

- `make lint` and `make test` pass.
- Backend: protected routes sit under `/api/course_phase/:coursePhaseID` with the correct roles,
  and `db/sqlc/` is regenerated and committed alongside any migration or query change.
- Frontend: no `any`, and shared libraries are reused over custom code.
- Tests are added or updated for the behavior you changed.
