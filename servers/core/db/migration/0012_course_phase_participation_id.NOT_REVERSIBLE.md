# 0012 has no down migration (intentional)

`0012_course_phase_participation_id.up.sql` renames `course_phase_participation.id` to
`old_id` and then **drops it permanently** after repointing `application_answer_text`,
`application_answer_multi_select`, and `application_assessment` onto
`course_participation_id`. The original surrogate-key UUID values are lost, so the migration
cannot be faithfully reversed.

Rather than ship a `.down.sql` that would either silently regenerate different surrogate keys
(breaking any external reference to the old ids) or leave the schema in a state where
`0012.up.sql` can no longer be re-applied, `0012` deliberately has **no down migration** — the
same convention the repo already uses for other irreversible steps (e.g.
`servers/team_allocation/db/migration/0008_skill_level_five_options`).

Practical effect: `migrate down` can roll a database back to **version 12**, not below it. Real
databases are all far past 12, so this floor is not a limitation in practice.
