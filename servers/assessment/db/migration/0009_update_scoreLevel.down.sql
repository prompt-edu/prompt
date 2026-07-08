BEGIN;

-- Drop views that depend on the assessment score column / enum
DROP VIEW IF EXISTS completed_score_levels;
DROP VIEW IF EXISTS weighted_participant_scores;

-- 1. Recreate the old enum type under a temporary name
CREATE TYPE score_level_old AS ENUM ('novice', 'intermediate', 'advanced', 'expert');

-- 2. Add back the old score column and map data from the new enum
ALTER TABLE assessment
ADD COLUMN score score_level_old;

-- ponytail: 'very_bad' has no old-enum counterpart (up mapped novice->bad, nothing to very_bad); collapse to 'novice'
UPDATE assessment
SET score = CASE score_level
        WHEN 'bad' THEN 'novice'::score_level_old
        WHEN 'ok' THEN 'intermediate'::score_level_old
        WHEN 'good' THEN 'advanced'::score_level_old
        WHEN 'very_good' THEN 'expert'::score_level_old
        WHEN 'very_bad' THEN 'novice'::score_level_old
    END;

ALTER TABLE assessment
ALTER COLUMN score SET NOT NULL;

ALTER TABLE assessment DROP COLUMN score_level;

-- 3. Swap the enum types back
DROP TYPE score_level;
ALTER TYPE score_level_old RENAME TO score_level;

-- 4. Restore the old competency description columns
ALTER TABLE competency
ADD COLUMN novice TEXT NOT NULL DEFAULT '',
    ADD COLUMN intermediate TEXT NOT NULL DEFAULT '',
    ADD COLUMN advanced TEXT NOT NULL DEFAULT '',
    ADD COLUMN expert TEXT NOT NULL DEFAULT '';

UPDATE competency
SET novice = description_bad,
    intermediate = description_ok,
    advanced = description_good,
    expert = description_very_good;

-- ponytail: description_very_bad has no old counterpart; dropped (data loss on rollback)
ALTER TABLE competency DROP COLUMN description_very_bad,
    DROP COLUMN description_bad,
    DROP COLUMN description_ok,
    DROP COLUMN description_good,
    DROP COLUMN description_very_good;

-- 5. Recreate the views as they were after 0007
CREATE OR REPLACE VIEW weighted_participant_scores AS WITH score_values AS (
    SELECT
        a.course_phase_id,
        a.course_participation_id,
        CASE
            a.score
            WHEN 'novice' THEN 4
            WHEN 'intermediate' THEN 3
            WHEN 'advanced' THEN 2
            WHEN 'expert' THEN 1
        END AS score_numeric,
        comp.weight AS competency_weight,
        cat.weight AS category_weight
    FROM
        assessment a
        JOIN competency comp ON a.competency_id = comp.id
        JOIN category cat ON comp.category_id = cat.id
),
weighted_scores AS (
    SELECT
        course_phase_id,
        course_participation_id,
        SUM(
            score_numeric * competency_weight * category_weight
        ) :: NUMERIC AS weighted_score_sum,
        SUM(competency_weight * category_weight) :: NUMERIC AS total_weight
    FROM
        score_values
    GROUP BY
        course_phase_id,
        course_participation_id
)
SELECT
    course_phase_id,
    course_participation_id,
    ROUND(
        (weighted_score_sum / NULLIF(total_weight, 0)) :: numeric,
        2
    ) AS score_numeric,
    CASE
        WHEN ROUND(
            (weighted_score_sum / NULLIF(total_weight, 0)) :: numeric,
            2
        ) < 1.75 THEN 'expert'
        WHEN ROUND(
            (weighted_score_sum / NULLIF(total_weight, 0)) :: numeric,
            2
        ) < 2.5 THEN 'advanced'
        WHEN ROUND(
            (weighted_score_sum / NULLIF(total_weight, 0)) :: numeric,
            2
        ) < 3.25 THEN 'intermediate'
        ELSE 'novice'
    END AS score_level
FROM
    weighted_scores;

CREATE OR REPLACE VIEW completed_score_levels AS
SELECT
    slc.course_phase_id,
    slc.course_participation_id,
    slc.score_level
FROM
    weighted_participant_scores slc
WHERE
    EXISTS (
        SELECT 1
        FROM assessment_completion ac
        WHERE ac.course_participation_id = slc.course_participation_id
            AND ac.course_phase_id = slc.course_phase_id
    );

COMMIT;
