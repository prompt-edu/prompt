-- PROMPT demo seed: example_server database.
--
-- The example phase module keeps a single row per course phase. Seeded for the
-- off-graph Example phase of iPraktikumDemo (f0000009-0000-0000-0000-000000000009).

DELETE FROM example_table WHERE course_phase_id = 'f0000009-0000-0000-0000-000000000009';

INSERT INTO example_table (course_phase_id, name)
    VALUES ('f0000009-0000-0000-0000-000000000009', 'iPraktikumDemo example phase');
