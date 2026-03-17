package student

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		input         studentDTO.CreateStudent
		expectedError string
	}{
		{
			name: "valid student data",
			input: studentDTO.CreateStudent{
				FirstName:            "John",
				LastName:             "Doe",
				Email:                "john.doe@example.com",
				HasUniversityAccount: true,
				MatriculationNumber:  "01234567",
				UniversityLogin:      "ab12xyz",
				Nationality:          "DE",
				CurrentSemester:      pgtype.Int4{Valid: true, Int32: 1},
				StudyProgram:         "Computer Science",
				StudyDegree:          "bachelor",
			},
			expectedError: "",
		},
		{
			name: "missing first name",
			input: studentDTO.CreateStudent{
				LastName:        "Doe",
				Email:           "john.doe@example.com",
				CurrentSemester: pgtype.Int4{Valid: true, Int32: 1},
				StudyProgram:    "Computer Science",
			},
			expectedError: "first name is required",
		},
		{
			name: "missing last name",
			input: studentDTO.CreateStudent{
				FirstName:       "John",
				Email:           "john.doe@example.com",
				CurrentSemester: pgtype.Int4{Valid: true, Int32: 1},
				StudyProgram:    "Computer Science",
			},
			expectedError: "last name is required",
		},
		{
			name: "invalid email",
			input: studentDTO.CreateStudent{
				FirstName:       "John",
				LastName:        "Doe",
				Email:           "invalid-email",
				CurrentSemester: pgtype.Int4{Valid: true, Int32: 1},
				StudyProgram:    "Computer Science",
			},
			expectedError: "invalid email address",
		},
		{
			name: "valid student without university account",
			input: studentDTO.CreateStudent{
				FirstName:            "John",
				LastName:             "Doe",
				Email:                "john.doe@example.com",
				HasUniversityAccount: false,
				Nationality:          "DE",
				CurrentSemester:      pgtype.Int4{Valid: true, Int32: 1},
				StudyProgram:         "Computer Science",
				StudyDegree:          "bachelor",
			},
			expectedError: "",
		},
		{
			name: "student with university account but invalid matriculation number",
			input: studentDTO.CreateStudent{
				FirstName:            "John",
				LastName:             "Doe",
				Email:                "john.doe@example.com",
				HasUniversityAccount: true,
				MatriculationNumber:  "1234567",
				UniversityLogin:      "ab12xyz",
				Nationality:          "DE",
				CurrentSemester:      pgtype.Int4{Valid: true, Int32: 1},
				StudyProgram:         "Computer Science",
			},
			expectedError: "invalid matriculation number",
		},
		{
			name: "student with university account but invalid university login",
			input: studentDTO.CreateStudent{
				FirstName:            "John",
				LastName:             "Doe",
				Email:                "john.doe@example.com",
				HasUniversityAccount: true,
				MatriculationNumber:  "01234567",
				UniversityLogin:      "xyz123",
				Nationality:          "DE",
				CurrentSemester:      pgtype.Int4{Valid: true, Int32: 1},
				StudyProgram:         "Computer Science",
			},
			expectedError: "invalid university login",
		},
		{
			name: "student without university account but with matriculation number",
			input: studentDTO.CreateStudent{
				FirstName:            "John",
				LastName:             "Doe",
				Email:                "john.doe@example.com",
				HasUniversityAccount: false,
				MatriculationNumber:  "01234567",
				Nationality:          "DE",
				CurrentSemester:      pgtype.Int4{Valid: true, Int32: 1},
				StudyProgram:         "Computer Science",
			},
			expectedError: "student has no university account but has university data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	err := validateName("", "Doe")
	assert.EqualError(t, err, "first name is required")

	err = validateName("John", "")
	assert.EqualError(t, err, "last name is required")

	err = validateName("John", "Doe")
	assert.NoError(t, err)
}

func TestValidateEmail(t *testing.T) {
	err := validateEmail("invalid-email")
	assert.EqualError(t, err, "invalid email address")

	err = validateEmail("john.doe@example.com")
	assert.NoError(t, err)
}

func TestValidateUniversityData(t *testing.T) {
	err := validateUniversityData(true, "1234567", "ab12xyz")
	assert.EqualError(t, err, "invalid matriculation number")

	err = validateUniversityData(true, "01234567", "xyz123")
	assert.EqualError(t, err, "invalid university login")

	err = validateUniversityData(true, "01234567", "ab12xyz")
	assert.NoError(t, err)

	err = validateUniversityData(false, "01234567", "")
	assert.EqualError(t, err, "student has no university account but has university data")

	err = validateUniversityData(false, "", "")
	assert.NoError(t, err)
}
