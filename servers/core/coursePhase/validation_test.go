package coursePhase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/stretchr/testify/assert"
)

func TestValidateCreateCoursePhase(t *testing.T) {
	tests := []struct {
		name          string
		input         coursePhaseDTO.CreateCoursePhase
		expectedError string
	}{
		{
			name: "valid course phase",
			input: coursePhaseDTO.CreateCoursePhase{
				CourseID:            uuid.New(),
				Name:                "Phase 1",
				IsInitialPhase:      true,
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"key": "value"},
				CoursePhaseTypeID:   uuid.New(),
			},
			expectedError: "",
		},
		{
			name: "missing name",
			input: coursePhaseDTO.CreateCoursePhase{
				CourseID:            uuid.New(),
				Name:                "",
				IsInitialPhase:      false,
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"key": "value"},
				CoursePhaseTypeID:   uuid.New(),
			},
			expectedError: "course phase name is required",
		},
		{
			name: "missing course ID",
			input: coursePhaseDTO.CreateCoursePhase{
				CourseID:            uuid.Nil,
				Name:                "Phase 1",
				IsInitialPhase:      true,
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"key": "value"},
				CoursePhaseTypeID:   uuid.New(),
			},
			expectedError: "course id is required",
		},
		{
			name: "missing name and course ID",
			input: coursePhaseDTO.CreateCoursePhase{
				CourseID:            uuid.Nil,
				Name:                "",
				IsInitialPhase:      false,
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"key": "value"},
				CoursePhaseTypeID:   uuid.New(),
			},
			expectedError: "course phase name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateCoursePhase(tt.input)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func TestValidateUpdateCoursePhase(t *testing.T) {
	tests := []struct {
		name          string
		input         coursePhaseDTO.UpdateCoursePhase
		expectedError string
	}{
		{
			name: "valid update",
			input: coursePhaseDTO.UpdateCoursePhase{
				ID:                  uuid.New(),
				Name:                pgtype.Text{Valid: true, String: "Updated Phase Name"},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"key": "value"},
			},
			expectedError: "",
		},
		{
			name: "missing name",
			input: coursePhaseDTO.UpdateCoursePhase{
				ID:                  uuid.New(),
				Name:                pgtype.Text{Valid: true, String: ""},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"key": "value"},
			},
			expectedError: "course phase name is required",
		},
		{
			name: "empty metadata",
			input: coursePhaseDTO.UpdateCoursePhase{
				ID:                  uuid.New(),
				Name:                pgtype.Text{Valid: true, String: "Phase with empty meta data"},
				StudentReadableData: meta.MetaData{},
				RestrictedData:      meta.MetaData{},
			},
			expectedError: "",
		},
		{
			name: "missing ID",
			input: coursePhaseDTO.UpdateCoursePhase{
				ID:                  uuid.Nil,
				Name:                pgtype.Text{Valid: true, String: "Valid Name"},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"key": "value"},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdateCoursePhase(tt.input)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
