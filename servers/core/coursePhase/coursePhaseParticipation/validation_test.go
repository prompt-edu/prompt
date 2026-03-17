package coursePhaseParticipation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation/coursePhaseParticipationDTO"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		input       coursePhaseParticipationDTO.CreateCoursePhaseParticipation
		expectError bool
	}{
		{
			name: "Valid input",
			input: coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
				CourseParticipationID: uuid.New(),
				CoursePhaseID:         uuid.New(),
			},
			expectError: false,
		},
		{
			name: "Missing CourseParticipationID",
			input: coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
				CourseParticipationID: uuid.Nil,
				CoursePhaseID:         uuid.New(),
			},
			expectError: true,
		},
		{
			name: "Missing CoursePhaseID",
			input: coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
				CourseParticipationID: uuid.New(),
				CoursePhaseID:         uuid.Nil,
			},
			expectError: true,
		},
		{
			name: "Both IDs missing",
			input: coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
				CourseParticipationID: uuid.Nil,
				CoursePhaseID:         uuid.Nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError = %v", err, tt.expectError)
			}
		})
	}
}
