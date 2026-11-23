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
              "answer",
              "key",
              "order_num",
              "type"
            ]
          },
          {
            "type": "object",
            "properties": {
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
