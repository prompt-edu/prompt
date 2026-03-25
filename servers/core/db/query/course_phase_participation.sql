-- name: GetCoursePhaseParticipation :one
SELECT * FROM course_phase_participation
WHERE course_phase_id = $1 
  AND course_participation_id = $2
LIMIT 1;

-- name: GetAllCoursePhaseParticipationsForCoursePhase :many
SELECT
    cpp.pass_status,
    cpp.restricted_data,
    cpp.student_readable_data,
    s.id AS student_id,
    s.first_name,
    s.last_name,
    s.email,
    s.matriculation_number,
    s.university_login,
    s.has_university_account,
    s.gender
FROM
    course_phase_participation cpp
JOIN
    course_participation cp ON cpp.course_participation_id = cp.id
JOIN
    student s ON cp.student_id = s.id
WHERE
    cpp.course_phase_id = $1;

-- name: GetAllCoursePhaseParticipationsForCourseParticipation :many
SELECT * FROM course_phase_participation
WHERE course_participation_id = $1;

--- We need to ensure that the course_participation_id and course_phase_id 
--- belong to the same course.
-- name: CreateOrUpdateCoursePhaseParticipation :one
INSERT INTO course_phase_participation
    (course_phase_id, course_participation_id, pass_status, restricted_data, student_readable_data)
SELECT
    $1 AS course_phase_id,
    $2 AS course_participation_id,
    COALESCE($3, 'not_assessed'::pass_status) AS pass_status,
    $4 AS restricted_data,
    $5 AS student_readable_data
FROM course_participation cp
JOIN course_phase cph ON cp.course_id = cph.course_id
WHERE cp.id = $2
  AND cph.id = $1
ON CONFLICT (course_phase_id, course_participation_id)
DO UPDATE SET
  pass_status = COALESCE($3, course_phase_participation.pass_status),
  restricted_data = course_phase_participation.restricted_data || $4,
  student_readable_data = course_phase_participation.student_readable_data || $5
RETURNING *;

-- name: UpdateCoursePhaseParticipation :one
UPDATE course_phase_participation
SET 
    pass_status = COALESCE($3, pass_status),   
    restricted_data = restricted_data || $4, 
    student_readable_data = student_readable_data || $5
WHERE course_phase_id = $1
  AND course_participation_id = $2
RETURNING course_phase_id, course_participation_id; -- important to trigger a no rows in result set error if ids mismatch

-- name: GetCoursePhaseParticipationByCourseParticipationAndCoursePhase :one
SELECT * FROM course_phase_participation
WHERE course_participation_id = $1 AND course_phase_id = $2 LIMIT 1;

-- name: UpdateCoursePhasePassStatus :many
UPDATE course_phase_participation
SET pass_status = sqlc.arg(pass_status)::pass_status
WHERE 
  course_participation_id = ANY(sqlc.arg(course_participation_id)::uuid[])
  AND course_phase_id = sqlc.arg(course_phase_id)::uuid
  AND pass_status != sqlc.arg(pass_status)::pass_status
RETURNING course_participation_id;

-- name: GetCoursePhaseParticipationStatusCounts :many
SELECT pass_status, COUNT(*) AS count
FROM course_phase_participation
WHERE course_phase_id = $1
GROUP BY pass_status;

-- name: GetAllCoursePhaseParticipationsForCoursePhaseIncludingPrevious :many
WITH 
  -----------------------------------------------------------------------
  -- A) Phases a student must have passed (per course_phase_graph)
  -----------------------------------------------------------------------
  direct_predecessor_for_pass AS (
      SELECT cpg.from_course_phase_id AS phase_id
      FROM course_phase_graph cpg
      WHERE cpg.to_course_phase_id = $1
  ),
  
  -----------------------------------------------------------------------
  -- B) Predecessor phases from which we pull meta data via participation_data_dependency_graph.
  -- Only include those whose course_phase_type has url = 'core'
  -- Copy the meta data from the predecessor to the current phase.
  -----------------------------------------------------------------------
  direct_predecessors_for_meta AS (
    SELECT 
      cpt.name                      AS from_course_phase_type_name,
      cpt.id                        AS from_course_phase_type_id,
      mdg.from_course_phase_id      AS from_course_phase_id,
      cp.restricted_data            AS course_phase_restricted_data,
      array_agg(po.dto_name)        AS from_dto_names -- all exported DTOs from this one phase from the core
    FROM participation_data_dependency_graph mdg
    JOIN course_phase cp 
        ON cp.id = mdg.from_course_phase_id
    JOIN course_phase_type cpt
        ON cpt.id = cp.course_phase_type_id
    JOIN course_phase_type_participation_provided_output_dto po 
      ON po.id = mdg.from_course_phase_DTO_id
    JOIN course_phase_type_participation_required_input_dto ri
      ON ri.id = mdg.to_course_phase_DTO_id
    WHERE mdg.to_course_phase_id = $1
      AND po.endpoint_path = 'core'
    GROUP BY 
      cpt.name,
      cpt.id,
      mdg.from_course_phase_id,
      cp.restricted_data
  ),
  
  -----------------------------------------------------------------------
  -- 1) Existing participants in the current phase
  -----------------------------------------------------------------------
  current_phase_participations AS (
      SELECT
          cpp.course_phase_id            AS course_phase_id,
          cpp.course_participation_id    AS course_participation_id,
          cpp.pass_status                AS pass_status,
          cpp.restricted_data            AS restricted_data,
          cpp.student_readable_data      AS student_readable_data,
          s.id                           AS student_id,
          s.first_name,
          s.last_name,
          s.email,
          s.matriculation_number,
          s.university_login,
          s.has_university_account,
          s.gender,
          s.nationality,
          s.study_degree,
          s.study_program,
          s.current_semester
      FROM course_phase_participation cpp
      JOIN course_participation cp 
        ON cpp.course_participation_id = cp.id
      JOIN student s 
        ON cp.student_id = s.id
      WHERE cpp.course_phase_id = $1
  ),
  
  -----------------------------------------------------------------------
  -- 2) Qualified non-participants:
  --    They do NOT yet have a participation in $1 and
  --    must have passed ALL direct_predecessors_for_pass.
  -----------------------------------------------------------------------
  qualified_non_participants AS (
      SELECT
          $1::uuid                     AS course_phase_id,
          cp.id                        AS course_participation_id,
          'not_assessed'::pass_status  AS pass_status,
          '{}'::jsonb                  AS restricted_data,
          '{}'::jsonb                  AS student_readable_data,
          s.id                         AS student_id,
          s.first_name,
          s.last_name,
          s.email,
          s.matriculation_number,
          s.university_login,
          s.has_university_account,
          s.gender,
          s.nationality,
          s.study_degree,
          s.study_program,
          s.current_semester
      FROM course_participation cp
      JOIN student s 
        ON cp.student_id = s.id
      WHERE 
          -- Exclude if they already have a participation in the current phase
          NOT EXISTS (
            SELECT 1
            FROM course_phase_participation new_cpp
            WHERE new_cpp.course_phase_id = $1
              AND new_cpp.course_participation_id = cp.id
          )
          -- And ensure they have 'passed' the direct predecessor (if any)
          AND EXISTS (
              SELECT 1
              FROM direct_predecessor_for_pass dpp
              JOIN course_phase_participation pcpp
                ON pcpp.course_phase_id = dpp.phase_id
                AND pcpp.course_participation_id = cp.id
              WHERE pcpp.pass_status = 'passed'
          )
  )
  
-----------------------------------------------------------------------
-- 3) Final SELECT: Merge the participants and meta data from all predecessors
-----------------------------------------------------------------------
SELECT
    main.*,
    (COALESCE(      
      (
            ----------------------------------------------------------------
            -- (I) Meta data from non-Application phases
            ----------------------------------------------------------------
            SELECT jsonb_object_agg(each.key, each.value)
            FROM direct_predecessors_for_meta dpm
            JOIN course_phase_participation pcpp
              ON pcpp.course_phase_id = dpm.from_course_phase_id
             AND pcpp.course_participation_id = main.course_participation_id
            CROSS JOIN LATERAL (
              SELECT key, value
              FROM (
                SELECT (jsonb_each(pcpp.student_readable_data)).key   AS key,
                       (jsonb_each(pcpp.student_readable_data)).value AS value
                UNION
                SELECT (jsonb_each(pcpp.restricted_data)).key   AS key,
                       (jsonb_each(pcpp.restricted_data)).value AS value
              ) sub
            ) AS each
            WHERE dpm.from_course_phase_type_name NOT IN ('Application', 'Interview')
              AND each.key = ANY(dpm.from_dto_names)
         ),
         '{}'::jsonb
      )::jsonb ||
      COALESCE(
          (
             -- (II) Meta data from the Interview phase (special handling)
            SELECT interviewdata.obj
            FROM direct_predecessors_for_meta dpm
            JOIN course_phase_participation pcpp
              ON pcpp.course_phase_id = dpm.from_course_phase_id
             AND pcpp.course_participation_id = main.course_participation_id
            JOIN course_phase_type cpt
              ON cpt.id = dpm.from_course_phase_type_id
             AND cpt.name = 'Interview'
            CROSS JOIN LATERAL (
               SELECT jsonb_strip_nulls(
                  jsonb_build_object(
                     'score', CASE
                         WHEN 'score' = ANY(dpm.from_dto_names) THEN pcpp.restricted_data -> 'score'
                         ELSE NULL
                     END,
                     'scoreLevel', CASE
                         WHEN 'scoreLevel' = ANY(dpm.from_dto_names) THEN
                           CASE
                             WHEN NOT (pcpp.restricted_data ? 'score') THEN NULL
                             WHEN jsonb_typeof(pcpp.restricted_data -> 'score') <> 'number' THEN NULL
                             WHEN (pcpp.restricted_data ->> 'score')::numeric < 1
                               OR (pcpp.restricted_data ->> 'score')::numeric > 5 THEN NULL
                             WHEN (pcpp.restricted_data ->> 'score')::numeric <= 1.5 THEN to_jsonb('veryGood'::text)
                             WHEN (pcpp.restricted_data ->> 'score')::numeric <= 2.5 THEN to_jsonb('good'::text)
                             WHEN (pcpp.restricted_data ->> 'score')::numeric <= 3.5 THEN to_jsonb('ok'::text)
                             WHEN (pcpp.restricted_data ->> 'score')::numeric <= 4.5 THEN to_jsonb('bad'::text)
                             ELSE to_jsonb('veryBad'::text)
                           END
                         ELSE NULL
                     END
                  )
               ) AS obj
            ) interviewdata
         ),
         '{}'::jsonb
      )::jsonb ||
      COALESCE(
          (
             -- (III) Meta data from the Application phase (special handling)
            SELECT appdata.obj
            FROM direct_predecessors_for_meta dpm
            JOIN course_phase_participation pcpp
              ON pcpp.course_phase_id = dpm.from_course_phase_id
             AND pcpp.course_participation_id = main.course_participation_id
            JOIN course_phase_type cpt
              ON cpt.id = dpm.from_course_phase_type_id
             AND cpt.name = 'Application'
            CROSS JOIN LATERAL (
               SELECT jsonb_strip_nulls(
                  jsonb_build_object(
                     'score', CASE 
                         WHEN 'score' = ANY(dpm.from_dto_names) THEN 
                           (SELECT to_jsonb(aasm.score)
                            FROM application_assessment aasm
                            WHERE aasm.course_phase_id = pcpp.course_phase_id AND aasm.course_participation_id = pcpp.course_participation_id)
                         ELSE NULL 
                     END,
                     'scoreLevel', CASE
                         WHEN 'scoreLevel' = ANY(dpm.from_dto_names) THEN
                           (SELECT CASE
                              WHEN aasm.score IS NULL OR aasm.score < 1 OR aasm.score > 5 THEN NULL
                              WHEN aasm.score <= 1 THEN to_jsonb('veryGood'::text)
                              WHEN aasm.score <= 2 THEN to_jsonb('good'::text)
                              WHEN aasm.score <= 3 THEN to_jsonb('ok'::text)
                              WHEN aasm.score <= 4 THEN to_jsonb('bad'::text)
                              ELSE to_jsonb('veryBad'::text)
                            END
                            FROM application_assessment aasm
                            WHERE aasm.course_phase_id = pcpp.course_phase_id AND aasm.course_participation_id = pcpp.course_participation_id)
                         ELSE NULL 
                     END,
                     'additionalScores', CASE 
                         WHEN 'additionalScores' = ANY(dpm.from_dto_names) THEN 
                           (SELECT jsonb_agg(
                                  jsonb_build_object(
                                    'key', question_config->>'key',
                                    'answer', pcpp.restricted_data -> (question_config->>'key')
                                  )
                           )
                            FROM jsonb_array_elements(dpm.course_phase_restricted_data->'additionalScores') question_config)
                         ELSE NULL 
                     END,
                     'applicationAnswers', CASE 
                         WHEN 'applicationAnswers' = ANY(dpm.from_dto_names) THEN 
                           (SELECT jsonb_agg(answer_obj)
                            FROM (
                               -- Text answers
                               SELECT jsonb_build_object(
                                  'key', qt.access_key,
                                  'answer', to_jsonb(aat.answer),
                                  'order_num', qt.order_num, 
                                  'type', 'text'
                               ) AS answer_obj
                               FROM application_question_text qt
                               JOIN application_answer_text aat
                                 ON aat.application_question_id = qt.id
                                AND aat.course_participation_id = pcpp.course_participation_id
                               WHERE qt.course_phase_id = dpm.from_course_phase_id
                                 AND qt.accessible_for_other_phases = true
                                 AND qt.access_key IS NOT NULL
                                 AND qt.access_key <> ''
                               UNION ALL
                               -- Multi-select answers
                               SELECT jsonb_build_object(
                                  'key', qm.access_key,
                                  'answer', to_jsonb(aams.answer),
                                  'order_num', qm.order_num, 
                                  'type', 'multiselect'
                               ) AS answer_obj
                               FROM application_question_multi_select qm
                               JOIN application_answer_multi_select aams
                                 ON aams.application_question_id = qm.id
                                AND aams.course_participation_id = pcpp.course_participation_id
                               WHERE qm.course_phase_id = dpm.from_course_phase_id
                                 AND qm.accessible_for_other_phases = true
                                 AND qm.access_key IS NOT NULL
                                 AND qm.access_key <> ''
                            ) answer_union)
                         ELSE NULL 
                     END
                  )
               ) AS obj
            ) appdata
         ),
         '{}'::jsonb
      )::jsonb)::jsonb AS prev_data
FROM
(
    SELECT * FROM current_phase_participations
    UNION
    SELECT * FROM qualified_non_participants
) AS main
ORDER BY main.last_name, main.first_name;



-- name: GetCoursePhaseParticipationByUniversityLoginAndCoursePhase :one
WITH 
-----------------------------------------------------------------------
-- A) Phases a student must have 'passed' (per course_phase_graph)
-- Identify the single previous phase (if any) required for PASS 
-----------------------------------------------------------------------
direct_predecessor_for_pass AS (
    SELECT cpg.from_course_phase_id AS phase_id
    FROM course_phase_graph cpg
    WHERE cpg.to_course_phase_id = $1
),

-----------------------------------------------------------------------
-- 1) Existing participants in the current phase
-----------------------------------------------------------------------
current_phase_participation AS (
    SELECT
        cpp.course_phase_id            AS course_phase_id,
        cpp.course_participation_id    AS course_participation_id,
        cpp.student_readable_data      AS student_readable_data,
        s.id                           AS student_id,
        s.first_name,
        s.last_name,
        s.email,
        s.matriculation_number,
        s.university_login,
        s.has_university_account,
        s.gender,
        s.nationality,
        s.study_degree,
        s.study_program,
        s.current_semester
    FROM course_phase_participation cpp
    INNER JOIN course_participation cp 
      ON cpp.course_participation_id = cp.id
    INNER JOIN student s 
      ON cp.student_id = s.id
    WHERE cpp.course_phase_id = $1
      AND s.university_login = $2
      AND s.matriculation_number = $3 
),

-----------------------------------------------------------------------
-- 2) Would-be participants: 
--    - They do NOT yet have a course_phase_participation for $1
--    - Must have passed ALL direct_predecessors_for_pass
-----------------------------------------------------------------------
qualified_non_participant AS (
    SELECT
        $1::uuid                     AS course_phase_id,
        cp.id                        AS course_participation_id,
        '{}'::jsonb                  AS student_readable_data,
        s.id                         AS student_id,
        s.first_name,
        s.last_name,
        s.email,
        s.matriculation_number,
        s.university_login,
        s.has_university_account,
        s.gender,
        s.nationality,
        s.study_degree,
        s.study_program,
        s.current_semester
    FROM course_participation cp
    JOIN student s 
      ON cp.student_id = s.id

    WHERE 
      s.university_login = $2
      AND s.matriculation_number = $3 
      -- Exclude if they already have a participation in the current phase
      AND NOT EXISTS (
        SELECT 1
        FROM course_phase_participation new_cpp
        WHERE new_cpp.course_phase_id = $1
          AND new_cpp.course_participation_id = cp.id
      )
    -- And ensure they have 'passed' in the previous phase 
    -- We filter just previous, not all since phase order might change or allow for non-linear courses at some point
    AND EXISTS (
        SELECT 1
        FROM direct_predecessor_for_pass dpp
        JOIN  course_phase_participation pcpp
          ON pcpp.course_phase_id = dpp.phase_id
          AND pcpp.course_participation_id = cp.id
        WHERE (pcpp.pass_status = 'passed')
    )
)
SELECT main.*
FROM
(
    SELECT * FROM current_phase_participation
    UNION
    SELECT * FROM qualified_non_participant
) AS main
LIMIT 1;


-- name: GetStudentsOfCoursePhase :many
WITH 
  -----------------------------------------------------------------------
  -- A) Phases a student must have passed (per course_phase_graph)
  -----------------------------------------------------------------------
  direct_predecessor_for_pass AS (
      SELECT cpg.from_course_phase_id AS phase_id
      FROM course_phase_graph cpg
      WHERE cpg.to_course_phase_id = $1
  ),
  
  -----------------------------------------------------------------------
  -- 1) Existing participants in the current phase
  -----------------------------------------------------------------------
  current_phase_students AS (
      SELECT
          s.*
      FROM course_phase_participation cpp
      JOIN course_participation cp 
        ON cpp.course_participation_id = cp.id
      JOIN student s 
        ON cp.student_id = s.id
      WHERE cpp.course_phase_id = $1
  ),
  
  -----------------------------------------------------------------------
  -- 2) Qualified non-participants:
  --    They do NOT yet have a participation in $1 and
  --    must have passed ALL direct_predecessors_for_pass.
  -----------------------------------------------------------------------
  qualified_students AS (
      SELECT
          s.*
      FROM course_participation cp
      JOIN student s 
        ON cp.student_id = s.id
      WHERE 
          -- Exclude if they already have a participation in the current phase
          NOT EXISTS (
            SELECT 1
            FROM course_phase_participation new_cpp
            WHERE new_cpp.course_phase_id = $1
              AND new_cpp.course_participation_id = cp.id
          )
          -- And ensure they have 'passed' the direct predecessor (if any)
          AND EXISTS (
              SELECT 1
              FROM direct_predecessor_for_pass dpp
              JOIN course_phase_participation pcpp
                ON pcpp.course_phase_id = dpp.phase_id
                AND pcpp.course_participation_id = cp.id
              WHERE pcpp.pass_status = 'passed'
          )
  )
  
-----------------------------------------------------------------------
-- 3) Final SELECT: Merge the participants and meta data from all predecessors
-----------------------------------------------------------------------
SELECT
    main.*
FROM
(
    SELECT * FROM current_phase_students
    UNION
    SELECT * FROM qualified_students
) AS main
ORDER BY main.last_name, main.first_name;
