BEGIN;

DROP VIEW IF EXISTS completed_score_levels;
DROP VIEW IF EXISTS weighted_participant_scores;

-- Restore the views as they were after 0009
CREATE VIEW weighted_participant_scores AS WITH score_values AS (
        SELECT a.course_phase_id,
            a.course_participation_id,
            CASE
                a.score_level
                WHEN 'very_bad' THEN 5
                WHEN 'bad' THEN 4
                WHEN 'ok' THEN 3
                WHEN 'good' THEN 2
                WHEN 'very_good' THEN 1
            END AS score_numeric,
            comp.weight AS competency_weight,
            cat.weight AS category_weight
        FROM assessment a
            JOIN competency comp ON a.competency_id = comp.id
            JOIN category cat ON comp.category_id = cat.id
    ),
    weighted_scores AS (
        SELECT course_phase_id,
            course_participation_id,
            SUM(
                score_numeric * competency_weight * category_weight
            )::NUMERIC AS weighted_score_sum,
            SUM(competency_weight * category_weight)::NUMERIC AS total_weight
        FROM score_values
        GROUP BY course_phase_id,
            course_participation_id
    )
SELECT course_phase_id,
    course_participation_id,
    ROUND(
        (weighted_score_sum / NULLIF(total_weight, 0))::numeric,
        2
    ) AS score_numeric,
    CASE
        WHEN ROUND(
            (weighted_score_sum / NULLIF(total_weight, 0))::numeric,
            2
        ) <= 1.5 THEN 'very_good'
        WHEN ROUND(
            (weighted_score_sum / NULLIF(total_weight, 0))::numeric,
            2
        ) <= 2.5 THEN 'good'
        WHEN ROUND(
            (weighted_score_sum / NULLIF(total_weight, 0))::numeric,
            2
        ) <= 3.5 THEN 'ok'
        WHEN ROUND(
            (weighted_score_sum / NULLIF(total_weight, 0))::numeric,
            2
        ) <= 4.5 THEN 'bad'
        ELSE 'very_bad'
    END AS score_level
FROM weighted_scores;

CREATE VIEW completed_score_levels AS
SELECT slc.course_phase_id,
    slc.course_participation_id,
    slc.score_level
FROM weighted_participant_scores slc
WHERE EXISTS (
        SELECT 1
        FROM assessment_completion ac
        WHERE ac.course_participation_id = slc.course_participation_id
            AND ac.course_phase_id = slc.course_phase_id
    );

COMMIT;
