BEGIN;

CREATE TABLE category_assessment (
    id                      uuid        NOT NULL PRIMARY KEY,
    category_id             uuid        NOT NULL,
    course_phase_id         uuid        NOT NULL,
    course_participation_id uuid        NOT NULL,
    comment                 text        NOT NULL DEFAULT '',
    author                  text        NOT NULL DEFAULT '',
    author_id               text        NOT NULL DEFAULT '',
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    UNIQUE (category_id, course_phase_id, course_participation_id),
    FOREIGN KEY (category_id) REFERENCES category (id) ON DELETE CASCADE
);

ALTER TABLE assessment
    ADD COLUMN author_id text NOT NULL DEFAULT '';

INSERT INTO category_assessment (id, category_id, course_phase_id, course_participation_id,
                                 comment, author, author_id, created_at, updated_at)
SELECT gen_random_uuid(),
       c.category_id,
       a.course_phase_id,
       a.course_participation_id,
       string_agg(
           '## ' || c.name || ' (by ' || COALESCE(NULLIF(a.author, ''), 'Unknown') || ')'
           || E'\nExamples: ' || COALESCE(NULLIF(a.examples, ''), '-')
           || E'\nComment: '  || COALESCE(NULLIF(a.comment, ''),  '-'),
           E'\n\n'
           ORDER BY c.name, c.id
       ),
       'migration',
       'migration',
       MAX(a.assessed_at),
       MAX(a.assessed_at)
FROM   assessment a
JOIN   competency c ON c.id = a.competency_id
WHERE  COALESCE(NULLIF(a.examples, ''), NULLIF(a.comment, '')) IS NOT NULL
GROUP  BY c.category_id, a.course_phase_id, a.course_participation_id;

ALTER TABLE assessment
    DROP COLUMN comment,
    DROP COLUMN examples;

COMMIT;
