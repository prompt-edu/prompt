BEGIN;

DROP INDEX IF EXISTS idx_assessment_completion_participation_phase;

DROP VIEW IF EXISTS completed_score_levels;
DROP VIEW IF EXISTS weighted_participant_scores;

CREATE VIEW weighted_participant_scores AS
WITH score_values AS (
    SELECT
        a.course_phase_id,
        a.course_participation_id,
        CASE a.score
            WHEN 'novice' THEN 1
            WHEN 'intermediate' THEN 2
            WHEN 'advanced' THEN 3
            WHEN 'expert' THEN 4
        END AS score_numeric,
        comp.weight AS competency_weight,
        cat.weight AS category_weight
    FROM assessment a
    JOIN competency comp ON a.competency_id = comp.id
    JOIN category cat ON comp.category_id = cat.id
),
weighted_scores AS (
    SELECT
        course_phase_id,
        course_participation_id,
        SUM(score_numeric * competency_weight * category_weight)::FLOAT AS weighted_score_sum,
        SUM(competency_weight * category_weight)::FLOAT AS total_weight
    FROM score_values
    GROUP BY course_phase_id, course_participation_id
)
SELECT
    course_phase_id,
    course_participation_id,
    ROUND((weighted_score_sum / total_weight)::numeric, 2) AS score_numeric
FROM weighted_scores;

CREATE OR REPLACE VIEW score_level_categories AS
SELECT
    course_phase_id,
    course_participation_id,
    score_numeric,
    CASE
        WHEN score_numeric < 1.5 THEN 'novice'
        WHEN score_numeric < 2.5 THEN 'intermediate'
        WHEN score_numeric < 3.5 THEN 'advanced'
        ELSE 'expert'
    END AS score_level
FROM weighted_participant_scores;

CREATE VIEW completed_score_levels AS
SELECT
    slc.course_phase_id,
    slc.course_participation_id,
    slc.score_level
FROM
    score_level_categories slc
WHERE
    EXISTS (
        SELECT 1
        FROM assessment_completion ac
        WHERE ac.course_participation_id = slc.course_participation_id
            AND ac.course_phase_id = slc.course_phase_id
    );

COMMIT;
