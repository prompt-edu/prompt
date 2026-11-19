package privacy

import (
	"testing"

	"github.com/google/uuid"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/stretchr/testify/require"
)

func TestTutorIDsCoverEveryIDATutorRowCanCarry(t *testing.T) {
	participationID, studentID, userID := uuid.New(), uuid.New(), uuid.New()

	ids := tutorIDs(sdkAuth.SubjectIdentifiers{
		CourseParticipationIDs: []uuid.UUID{participationID},
		StudentID:              studentID,
		UserID:                 userID,
	})

	require.Equal(t, []uuid.UUID{participationID, studentID, userID}, ids)
	require.Empty(t, tutorIDs(sdkAuth.SubjectIdentifiers{}))
}
