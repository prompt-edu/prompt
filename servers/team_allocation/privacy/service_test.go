package privacy

import (
	"testing"

	"github.com/google/uuid"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/stretchr/testify/assert"
)

func TestTutorIDs(t *testing.T) {
	participationA, participationB := uuid.New(), uuid.New()
	studentID, userID := uuid.New(), uuid.New()

	tests := []struct {
		name    string
		subject sdkAuth.SubjectIdentifiers
		want    []uuid.UUID
	}{
		{
			name:    "student with participations",
			subject: sdkAuth.SubjectIdentifiers{CourseParticipationIDs: []uuid.UUID{participationA, participationB}, StudentID: studentID, UserID: userID},
			want:    []uuid.UUID{participationA, participationB, studentID, userID},
		},
		{
			name:    "student without a Keycloak account",
			subject: sdkAuth.SubjectIdentifiers{CourseParticipationIDs: []uuid.UUID{participationA}, StudentID: studentID},
			want:    []uuid.UUID{participationA, studentID},
		},
		{
			name:    "course editor without a student record",
			subject: sdkAuth.SubjectIdentifiers{UserID: userID},
			want:    []uuid.UUID{userID},
		},
		{
			name:    "no identifiers",
			subject: sdkAuth.SubjectIdentifiers{},
			want:    []uuid.UUID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tutorIDs(tt.subject))
		})
	}
}
