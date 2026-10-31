package coursePhaseConfig

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func startCoreParticipations(t *testing.T, status int, participants ...uuid.UUID) (*string, uuid.UUID) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	t.Setenv("SERVER_CORE_HOST", server.URL)

	coursePhaseID := uuid.New()
	var authHeader string
	router.GET("/api/course_phases/"+coursePhaseID.String()+"/participations", func(c *gin.Context) {
		authHeader = c.GetHeader("Authorization")
		if status != http.StatusOK {
			c.Status(status)
			return
		}
		participations := make([]gin.H, 0, len(participants))
		for _, id := range participants {
			participations = append(participations, gin.H{"courseParticipationID": id})
		}
		c.JSON(http.StatusOK, gin.H{"participations": participations, "resolutions": []gin.H{}})
	})
	return &authHeader, coursePhaseID
}

func TestCheckParticipantInPhaseAcceptsListedParticipant(t *testing.T) {
	participant := uuid.New()
	authHeader, coursePhaseID := startCoreParticipations(t, http.StatusOK, uuid.New(), participant)

	err := CheckParticipantInPhase(context.Background(), "Bearer token", coursePhaseID, participant)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer token", *authHeader)
}

func TestCheckParticipantInPhaseRejectsUnlistedParticipant(t *testing.T) {
	_, coursePhaseID := startCoreParticipations(t, http.StatusOK, uuid.New())

	err := CheckParticipantInPhase(context.Background(), "Bearer token", coursePhaseID, uuid.New())
	assert.ErrorIs(t, err, ErrParticipantNotInPhase)
}

func TestCheckParticipantInPhaseReportsFailedLookup(t *testing.T) {
	_, coursePhaseID := startCoreParticipations(t, http.StatusInternalServerError)

	err := CheckParticipantInPhase(context.Background(), "Bearer token", coursePhaseID, uuid.New())
	assert.ErrorIs(t, err, ErrParticipantLookup)
}
