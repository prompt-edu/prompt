-- Phase graph updates used to check only the target of an edge, so remove edges that leave their course.
DELETE FROM course_phase_graph graph
USING course_phase from_phase, course_phase to_phase
WHERE graph.from_course_phase_id = from_phase.id
  AND graph.to_course_phase_id = to_phase.id
  AND from_phase.course_id <> to_phase.course_id;
