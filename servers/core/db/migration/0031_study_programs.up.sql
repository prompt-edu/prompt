BEGIN;

CREATE TABLE study_program (
  id         uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
  name       varchar(100) NOT NULL,
  short_name varchar(20)
);

CREATE UNIQUE INDEX study_program_name_unique ON study_program (lower(trim(name)));
CREATE UNIQUE INDEX study_program_label_unique ON study_program (lower(trim(coalesce(short_name, name))));

INSERT INTO study_program (name, short_name) VALUES
  ('Computer Science', 'CS'),
  ('Information Systems', 'IS'),
  ('Games Engineering', 'GE'),
  ('Management and Technology', 'M&T');

COMMIT;
