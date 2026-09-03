# Seed data

Data-only SQL, one file per service database, loaded by `scripts/seed.sh`
(`make seed`) and by the e2e stack's one-shot `seed` service.

Full documentation: [`docs/contributor/guide/seeding.md`](../docs/contributor/guide/seeding.md).

## Rules

- **Data only.** The servers own the schema. Never add DDL, and never write
  `schema_migrations`.
- **Never insert a `course_phase_type` row.** Core creates every type at startup with
  a random UUID and *skips its provided/required DTO descriptors* when the type row
  already exists, so pinning a type quietly costs you the inter-phase data dependency
  metadata. Resolve types by name:
  `(SELECT id FROM course_phase_type WHERE name = 'Interview')`. The single exception
  is `example_component`: no service creates it, so `core.sql` inserts it (with a
  random id, like core does) purely so the rest of the file can resolve it by name.
  That leaves its `base_url` seed-owned and unreconciled - correct where core fronts
  the phase services, wrong for a host-run example server.
- **The seed is authoritative.** Each file deletes the rows it owns and inserts them
  again, so `make seed` reconciles instead of failing. Scope every `DELETE` to pinned
  ids - never truncate.
- **Nothing enforces the links between databases.** Run `make seed-check` after
  touching an id that crosses a service boundary.
- `course_phase_graph` is UNIQUE on both endpoints, so a course's phases are a strict
  chain. Append to the tail, or leave the phase off the graph.

## Files

| File | Database | Owns |
| --- | --- | --- |
| `core.sql` | core | courses, students, phases, the graph, participations, application form and answers, data dependency edges, mail campaign |
| `interview.sql` | interview | slots, assignments, reviews |
| `team_allocation.sql` | team_allocation | skills, teams, survey, responses, allocation, tutors |
| `self_team_allocation.sql` | self_team_allocation | teams, timeframe, assignments |
| `assessment.sql` | assessment | schemas, categories, competencies, config, assessments, evaluations |
| `certificate.sql` | certificate | phase config (Typst template), downloads |
| `presentation.sql` | presentation | phase config, feedback categories, slots, presentations |
| `example_server.sql` | example_server | the module's single row |

`infrastructure_setup` has no file: everything it stores is either encrypted
provider credentials or a record of resources provisioned in a real external
system, so there is nothing a seed can own. `core.sql` still creates the phase
the e2e suite drives through the service's own API.

## Pinned id ranges

Everything the seed creates has a fixed UUID so the files can reference each other
across databases. `scripts/seed-check.sh` knows the core-owned prefixes.

| Prefix | What |
| --- | --- |
| `c0000001…` / `c0000002…` / `c0000003…` | courses `iPraktikumFull` / `iPraktikumDemo` / `iPraktikumWelcome` |
| `d7307be2…`, `e12ffe63…`, `be780b32…` | the e2e fixture courses |
| `f00000NN…` | demo course phases (1 Application … 9 Example) |
| `cd0000NN…` | demo course participations |
| `e10000NN…` | demo students |
| `e0000005…` / `a5000007…` | the Keycloak `student` / `student2` rows |
| `e0000009…` / `e0000010…` | the privacy suite's deletion subjects |
| `fa…`, `fb…`, `fc…`, `fd…`, `fe…`, `ff…` | demo application questions, answers, assessments, mail campaign |
| `1a…`, `1b…` | interview slots, assignments |
| `2a…`, `2b…` | self team allocation teams, assignments |
| `3a…`, `3b…`, `3c…` | team allocation teams, skills, allocations |
| `4a…`, `4b…`, `4c…` | presentation categories, slots, presentations |
| `5a…` … `60…`, `6a…`, `6b…` | assessment schemas, categories, competencies, assessments, evaluations, feedback, action items |
| `a…`, `b3…`, `d…` | the e2e fixture phases and participations (see `e2e/src/data/constants.ts`) |

## The demo course

`iPraktikumDemo` (`ios2526`) is the full-course example. Its graph is

```text
Application → Interview → Matching → Team Allocation → Assessment → Presentation → Certificate
```

with **Self Team Allocation** and **Example** deliberately off the graph: the graph
cannot branch, Self Team Allocation is the alternative to Team Allocation rather than
a successor to it, and Example is a developer placeholder. Both still carry data.

Eight applicants, six of whom pass into the rest of the course. `cd000001` and
`cd000002` are the Keycloak `student` and `student2` users, so a local login lands on
a populated course.

Two things are intentionally absent:

- **File-upload answers.** `application_answer_file_upload` needs a `files` row backed
  by a real object in SeaweedFS, which SQL alone cannot create. The upload *question*
  is seeded.
- **Peer and tutor evaluations.** The config leaves those schema columns to their
  migration-set defaults; only the self-evaluation round is seeded.

## The demo timeline

Every date in the seed is absolute and belongs to one semester, `ios2526`
(`2026-04-01` to `2026-09-30`). The phases are dated in the order the course runs
them, so the demo reads as a semester that has been played through end to end:

| Phase | Dated | State |
| --- | --- | --- |
| Application | open 2026-01-01 to 2026-03-15 | eight applications answered |
| Interview | slots 2026-03-20 and 2026-03-21 | assigned and reviewed |
| Team Allocation | survey 2026-04-01 to 2026-04-06 | responses in, teams allocated |
| Self Team Allocation | timeframe 2026-04-07 to 2026-04-14 | teams filled |
| Assessment | 2026-07-01 to 2026-09-20, self-evaluation 2026-06-15 to 2026-06-30 | graded, results released |
| Presentation | slots 2026-09-24 | scheduled, feedback recorded |
| Certificate | template released 2026-09-30 | one download recorded |

The dates are pinned rather than derived from `NOW()`, for the same reason the ids
are: a reseed has to reproduce the same course, and a relative timeline would
disagree with the semester tag and the course's own start and end date. The trade is
that the demo ages. Once summer 2026 is behind us every window is closed, and the
phases that gate on one (the self-evaluation, both allocation timeframes, the
application) will refuse new input. To act against a live deadline, move that one
window in the phase's configuration screen after seeding, or edit the row in the
seed file before loading it.
