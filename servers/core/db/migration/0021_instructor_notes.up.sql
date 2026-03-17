-- Migration: Create Instructor Notes

CREATE TABLE note (
  id             uuid PRIMARY KEY,
  for_student    uuid NOT NULL REFERENCES student(id),
  author         uuid NOT NULL,
  author_name    text NOT NULL,
  author_email   text NOT NULL,
  date_created   timestamptz NOT NULL default now(),
  date_deleted   timestamptz,
  deleted_by     uuid
);

CREATE TABLE note_version (
  id             uuid PRIMARY KEY,
  content        text NOT NULL,
  date_created   timestamptz NOT NULL default now(),
  version_number int NOT NULL,
  for_note       uuid NOT NULL REFERENCES note(id),
  UNIQUE (for_note, version_number)
);

CREATE TYPE note_tag_color AS ENUM ('blue', 'green', 'red', 'yellow', 'orange', 'pink');

CREATE TABLE note_tag (
  id    uuid           PRIMARY KEY,
  name  text           NOT NULL UNIQUE,
  color note_tag_color NOT NULL
);

CREATE TABLE note_tag_relation (
  note_id uuid NOT NULL REFERENCES note(id) ON DELETE CASCADE,
  tag_id  uuid NOT NULL REFERENCES note_tag(id) ON DELETE CASCADE,
  PRIMARY KEY(note_id, tag_id)
);

CREATE INDEX idx_note_tag_tag_id ON note_tag_relation(tag_id);
CREATE INDEX idx_note_tag_note_id ON note_tag_relation(note_id);
CREATE INDEX idx_note_for_student_deleted ON note(for_student, date_deleted);

-- view: reusable CTE

CREATE VIEW note_with_versions AS
SELECT
  n.id,
  n.author,
  n.author_name,
  n.author_email,
  n.for_student,
  n.date_created,
  n.date_deleted,
  n.deleted_by,
  CASE
    WHEN n.date_deleted IS NULL THEN
      jsonb_agg(
        jsonb_build_object(
          'id', nv.id,
          'content', nv.content,
          'dateCreated', nv.date_created,
          'versionNumber', nv.version_number
        )
        ORDER BY nv.version_number
      )
    ELSE '[]'::jsonb
  END AS versions,
  CASE
    WHEN n.date_deleted IS NULL THEN
      COALESCE(
        (
          SELECT jsonb_agg(jsonb_build_object('id', nt.id, 'name', nt.name, 'color', nt.color) ORDER BY nt.name)
          FROM note_tag_relation ntr
          JOIN note_tag nt ON nt.id = ntr.tag_id
          WHERE ntr.note_id = n.id
        ),
        '[]'::jsonb
      )
    ELSE '[]'::jsonb
  END AS tags
FROM note n
JOIN note_version nv ON nv.for_note = n.id
GROUP BY n.id;
