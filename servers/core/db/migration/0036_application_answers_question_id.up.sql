-- Application answers carry the question id and title next to the access key, so a
-- consuming phase can reference a question by id and show its title (#2348). The key
-- stays until every consumer has moved to the id.
BEGIN;

UPDATE course_phase_type_participation_provided_output_dto po
SET specification = '
    {
      "type": "array",
      "items": {
        "oneOf": [
          {
            "type": "object",
            "properties": {
              "questionID": {
                "type": "string",
                "format": "uuid"
              },
              "title": {
                "type": "string"
              },
              "answer": {
                "type": "string"
              },
              "key": {
                "type": "string"
              },
              "order_num": {
                "type": "integer"
              },
              "type": {
                "type": "string",
                "enum": [
                  "text"
                ]
              }
            },
            "required": [
              "questionID",
              "title",
              "answer",
              "key",
              "order_num",
              "type"
            ]
          },
          {
            "type": "object",
            "properties": {
              "questionID": {
                "type": "string",
                "format": "uuid"
              },
              "title": {
                "type": "string"
              },
              "answer": {
                "type": "array",
                "items": {
                  "type": "string"
                }
              },
              "key": {
                "type": "string"
              },
              "order_num": {
                "type": "integer"
              },
              "type": {
                "type": "string",
                "enum": [
                  "multiselect"
                ]
              }
            },
            "required": [
              "questionID",
              "title",
              "answer",
              "key",
              "order_num",
              "type"
            ]
          }
        ]
      }
    }
'::jsonb
FROM course_phase_type cpt
WHERE cpt.id = po.course_phase_type_id
  AND cpt.name = 'Application'
  AND po.dto_name = 'applicationAnswers'
  AND po.endpoint_path = 'core';

COMMIT;
