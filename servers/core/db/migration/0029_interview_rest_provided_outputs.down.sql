BEGIN;

UPDATE course_phase_type_participation_provided_output_dto po
SET endpoint_path = 'core'
FROM course_phase_type cpt
WHERE po.course_phase_type_id = cpt.id
  AND cpt.name = 'Interview'
  AND po.dto_name = 'score'
  AND po.endpoint_path = '/interview-review/score';

UPDATE course_phase_type_participation_provided_output_dto po
SET endpoint_path = 'core'
FROM course_phase_type cpt
WHERE po.course_phase_type_id = cpt.id
  AND cpt.name = 'Interview'
  AND po.dto_name = 'scoreLevel'
  AND po.endpoint_path = '/interview-review/scoreLevel';

UPDATE course_phase_type
SET base_url = 'core'
WHERE name = 'Interview'
  AND base_url = '{CORE_HOST}/interview/api';

COMMIT;
