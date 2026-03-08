-- name: GetAllCoursePhaseTypes :many
SELECT * FROM course_phase_type;

-- name: GetCoursePhaseTypeByID :one
SELECT * FROM course_phase_type WHERE id = $1;

-- name: GetCoursePhaseRequiredParticipationInputs :many
SELECT *
FROM
    course_phase_type_participation_required_input_dto
WHERE
    course_phase_type_id = $1;

-- name: GetCoursePhaseProvidedParticipationOutputs :many
SELECT *
FROM
    course_phase_type_participation_provided_output_dto
WHERE
    course_phase_type_id = $1;

-- name: GetCoursePhaseRequiredPhaseInputs :many
SELECT *
FROM
    course_phase_type_phase_required_input_dto
WHERE
    course_phase_type_id = $1;

-- name: GetCoursePhaseProvidedPhaseOutputs :many
SELECT *
FROM
    course_phase_type_phase_provided_output_dto
WHERE
    course_phase_type_id = $1;

-- name: TestApplicationPhaseTypeExists :one
SELECT EXISTS (
        SELECT 1
        FROM course_phase_type
        WHERE
            name = 'Application'
    ) AS does_exist;

-- name: TestInterviewPhaseTypeExists :one
SELECT EXISTS (
        SELECT 1
        FROM course_phase_type
        WHERE
            name = 'Interview'
    ) AS does_exist;

-- name: TestMatchingPhaseTypeExists :one
SELECT EXISTS (
        SELECT 1
        FROM course_phase_type
        WHERE
            name = 'Matching'
    ) AS does_exist;

-- name: TestIntroCourseDeveloperPhaseTypeExists :one
SELECT EXISTS (SELECT 1
               FROM course_phase_type
               WHERE name = 'Intro Course Developer') AS does_exist;

-- name: TestIntroCourseTutorPhaseTypeExists :one
SELECT EXISTS (
        SELECT 1
        FROM course_phase_type
        WHERE
            name = 'IntroCourseTutor'
    ) AS does_exist;

-- name: TestDevOpsChallengeTypeExists :one
SELECT EXISTS (SELECT 1
               FROM course_phase_type
               WHERE name = 'DevOps Challenge') AS does_exist;

-- name: TestAssessmentTypeExists :one
SELECT EXISTS (
        SELECT 1
        FROM course_phase_type
        WHERE
            name = 'Assessment'
    ) AS does_exist;

-- name: TestTeamAllocationTypeExists :one
SELECT EXISTS (
        SELECT 1
        FROM course_phase_type
        WHERE
            name = 'Team Allocation'
    ) AS does_exist;

-- name: TestSelfTeamAllocationTypeExists :one
SELECT EXISTS (
        SELECT 1
        FROM course_phase_type
        WHERE
            name = 'Self Team Allocation'
    ) AS does_exist;

-- name: TestCertificateTypeExists :one
SELECT EXISTS (
        SELECT 1
        FROM course_phase_type
        WHERE
            name = 'Certificate'
    ) AS does_exist;

-- name: CreateCoursePhaseType :exec
INSERT INTO course_phase_type (id, name, initial_phase, base_url, description)
VALUES ($1, $2, $3, $4, $5);

-- name: CreateCoursePhaseTypeRequiredInput :exec
INSERT INTO
    course_phase_type_participation_required_input_dto (
        id,
        course_phase_type_id,
        dto_name,
        specification
    )
VALUES ($1, $2, $3, $4);

-- name: CreateCoursePhaseTypeProvidedOutput :exec
INSERT INTO
    course_phase_type_participation_provided_output_dto (
        id,
        course_phase_type_id,
        dto_name,
        version_number,
        endpoint_path,
        specification
    )
VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreateRequiredApplicationAnswers :exec
INSERT INTO course_phase_type_participation_required_input_dto (id, course_phase_type_id, dto_name, specification)
VALUES (gen_random_uuid(),
        $1,
        'applicationAnswers',
        '{
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
        }'::jsonb);

-- name: InsertCourseProvidedApplicationAnswers :exec
INSERT INTO course_phase_type_participation_provided_output_dto (id, course_phase_type_id, dto_name, version_number,
                                                                 endpoint_path, specification)
VALUES (gen_random_uuid(),
        $1,
        'applicationAnswers',
        1,
        'core',
        '{
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
        }'::jsonb);

-- name: InsertCourseProvidedAdditionalScores :exec
INSERT INTO course_phase_type_participation_provided_output_dto (id, course_phase_type_id, dto_name, version_number,
                                                                 endpoint_path, specification)
VALUES (gen_random_uuid(),
        $1,
        'additionalScores',
        1,
        'core',
        '{
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "score": {
                "type": "number"
              },
              "key": {
                "type": "string"
              }
            },
            "required": [
              "score",
              "key"
            ]
          }
        }'::jsonb);

-- name: CreateRequiredDevices :exec
INSERT INTO course_phase_type_participation_required_input_dto (id, course_phase_type_id, dto_name, specification)
VALUES (gen_random_uuid(),
        $1,
        'devices',
        '{
          "type": "array",
          "items": {
            "type": "string",
            "enum": [
              "IPhone",
              "IPad",
              "MacBook",
              "AppleWatch"
            ]
          }
        }'::jsonb);

-- name: InsertProvidedOutputDevices :exec
INSERT INTO course_phase_type_participation_provided_output_dto (id, course_phase_type_id, dto_name, version_number,
                                                                 endpoint_path, specification)
VALUES (gen_random_uuid(),
        $1,
        'devices',
        1,
        '/devices',
        '{
          "type": "array",
          "items": {
            "type": "string",
            "enum": [
              "IPhone",
              "IPad",
              "MacBook",
              "AppleWatch"
            ]
          }
        }'::jsonb);

-- This returns the teamID for a given courseParticipationID, to which the user is assigned
-- name: InsertTeamAllocationOutput :exec
INSERT INTO course_phase_type_participation_provided_output_dto (id, course_phase_type_id, dto_name, version_number,
                                                                 endpoint_path, specification)
VALUES (gen_random_uuid(),
        $1,
        'teamAllocation',
        1,
        '/allocation',
        '{
          "type": "string"
        }'::jsonb);

-- name: InsertTeamAllocationRequiredInput :exec
INSERT INTO course_phase_type_participation_required_input_dto(id, course_phase_type_id, dto_name, specification)
VALUES (gen_random_uuid(),
        $1,
        'teamAllocation',
        '{
          "type": "string"
        }'::jsonb);

-- name: InsertTeamOutput :exec
INSERT INTO course_phase_type_phase_provided_output_dto (id, course_phase_type_id, dto_name, version_number,
                                                         endpoint_path, specification)
VALUES (gen_random_uuid(),
        $1,
        'teams',
        1,
        '/team',
        '{
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "id": {
                "type": "string"
              },
              "name": {
                "type": "string"
              }
            },
            "required": [
              "id",
              "name"
            ]
          }
        }'::jsonb);

-- name: InsertTeamRequiredInput :exec
INSERT INTO course_phase_type_phase_required_input_dto (id, course_phase_type_id, dto_name, specification)
VALUES (gen_random_uuid(),
        $1,
        'teams',
        '{
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "id": {
                "type": "string"
              },
              "name": {
                "type": "string"
              }
            },
            "required": [
              "id",
              "name"
            ]
          }
        }'::jsonb);

-- name: InsertAssessmentScoreOutput :exec
INSERT INTO course_phase_type_participation_provided_output_dto (id, course_phase_type_id, dto_name, version_number,
                                                                 endpoint_path, specification)
VALUES (gen_random_uuid(),
        $1,
        'scoreLevel',
        1,
        '/student-assessment/scoreLevel',
        '{
          "type": "string",
          "enum": [
            "veryBad",
            "Bad",
            "Ok",
            "Good",
            "VeryGood"
          ]
        }'::jsonb);

-- name: InsertActionItemsOutput :exec
INSERT INTO course_phase_type_participation_provided_output_dto (id, course_phase_type_id, dto_name, version_number,
                                                                 endpoint_path, specification)
VALUES (gen_random_uuid(),
        $1,
        'actionItems',
        1,
        '/student-assessment/action-item/action',
        '{
          "type": "array",
          "items": {
            "type": "string"
          }
        }'::jsonb);

-- name: InsertGradeOutput :exec
INSERT INTO course_phase_type_participation_provided_output_dto (id, course_phase_type_id, dto_name, version_number,
                                                                 endpoint_path, specification)
VALUES (gen_random_uuid(),
        $1,
        'grade',
        1,
        '/student-assessment/completed/grade',
        '{
          "type": "number"
        }'::jsonb);

-- name: InsertAssessmentScoreRequiredInput :exec
INSERT INTO course_phase_type_participation_required_input_dto (id, course_phase_type_id, dto_name, specification)
VALUES (gen_random_uuid(),
        $1,
        'scoreLevel',
        '{
          "type": "string",
          "enum": [
            "veryBad",
            "Bad",
            "Ok",
            "Good",
            "VeryGood"
          ]
        }'::jsonb);