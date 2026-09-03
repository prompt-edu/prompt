-- PROMPT demo seed: certificate database.
--
-- The Certificate phase of iPraktikumDemo (f0000007-0000-0000-0000-000000000007):
-- a released Typst template (the module's own sample_template.typ) and one
-- recorded download.
--
-- certificate_download.id is a SERIAL, so it is left to the sequence: inserting
-- explicit ids would desync it. Nothing may depend on that id.

DELETE FROM certificate_download WHERE course_phase_id = 'f0000007-0000-0000-0000-000000000007';
DELETE FROM course_phase_config WHERE course_phase_id = 'f0000007-0000-0000-0000-000000000007';

INSERT INTO course_phase_config (course_phase_id, template_content, updated_by, release_date)
    VALUES ('f0000007-0000-0000-0000-000000000007', $typst$#let vars = json("vars.json")

#set page(margin: 1cm)
#set text(size: 12pt, font: "Times New Roman")

#align(center)[
  #text(size: 24pt, weight: "bold")[Certificate of Completion]

  #v(1cm)

  #text(size: 16pt)[This certifies that]

  #v(0.5cm)

  #text(size: 20pt, weight: "bold")[#vars.studentName]

  #v(0.5cm)

  #text(size: 16pt)[has successfully completed the course]

  #v(0.5cm)

  #text(size: 18pt, weight: "bold")[#vars.courseName]

  #v(0.5cm)

  #text(size: 14pt)[as a member of #vars.teamName]

  #v(1cm)

  #text(size: 12pt)[Date: #vars.date]
]
$typst$, 'seed-lecturer', '2026-09-30 12:00:00+02');

-- student_id, not course_participation_id: the certificate module is the only
-- one keyed on the core `student` row. This is Stan (the Keycloak `student` user).
INSERT INTO certificate_download (student_id, course_phase_id, download_count)
    VALUES ('e0000005-0000-0000-0000-000000000005', 'f0000007-0000-0000-0000-000000000007', 2);
