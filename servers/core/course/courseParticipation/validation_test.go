package courseParticipation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation/courseParticipationDTO"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		input         courseParticipationDTO.CreateCourseParticipation
		expectedError string
	}{
		{
			name: "valid course participation",
			input: courseParticipationDTO.CreateCourseParticipation{
				CourseID:  uuid.New(),
				StudentID: uuid.New(),
			},
			expectedError: "",
		},
		{
			name: "missing course ID",
			input: courseParticipationDTO.CreateCourseParticipation{
				CourseID:  uuid.Nil,
				StudentID: uuid.New(),
			},
			expectedError: "validation error: course id is required",
		},
		{
			name: "missing student ID",
			input: courseParticipationDTO.CreateCourseParticipation{
				CourseID:  uuid.New(),
				StudentID: uuid.Nil,
			},
			expectedError: "validation error: student id is required",
		},
		{
			name: "missing both course ID and student ID",
			input: courseParticipationDTO.CreateCourseParticipation{
				CourseID:  uuid.Nil,
				StudentID: uuid.Nil,
			},
			expectedError: "validation error: course id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
