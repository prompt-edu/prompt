BEGIN;

UPDATE course_phase_type_participation_provided_output_dto po
SET endpoint_path = '/interview-review/score'
FROM course_phase_type cpt
WHERE po.course_phase_type_id = cpt.id
  AND cpt.name = 'Interview'
  AND po.dto_name = 'score'
  AND po.endpoint_path = 'core';

UPDATE course_phase_type_participation_provided_output_dto po
SET endpoint_path = '/interview-review/scoreLevel'
FROM course_phase_type cpt
WHERE po.course_phase_type_id = cpt.id
  AND cpt.name = 'Interview'
  AND po.dto_name = 'scoreLevel'
  AND po.endpoint_path = 'core';

COMMIT;
