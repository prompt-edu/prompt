package teams

// ponytail: local copy of tutor scoping logic pending prompt-sdk issue for the reusable version.
// Interface: ResolveTutorTeam(ctx, coursePhaseID, universityLogin) → (uuid.UUID, error)

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

const tutorTeamIDKey = "tutorTeamID"

func tutorScopingMiddleware(q db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
		if !ok || !tokenUser.IsEditor || tokenUser.IsLecturer {
			c.Next()
			return
		}
		login := strings.TrimSpace(strings.ToLower(tokenUser.UniversityLogin))
		if login == "" {
			c.Next()
			return
		}
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid course phase id"})
			return
		}
		teamID, err := q.GetTutorTeamByUniversityLogin(c.Request.Context(), db.GetTutorTeamByUniversityLoginParams{
			CoursePhaseID:   coursePhaseID,
			UniversityLogin: pgtype.Text{String: login, Valid: true},
		})
		if errors.Is(err, pgx.ErrNoRows) {
			c.Next()
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "access check failed"})
			return
		}
		c.Set(tutorTeamIDKey, teamID)
		c.Next()
	}
}

func getTutorTeamID(c *gin.Context) (uuid.UUID, bool) {
	if v, exists := c.Get(tutorTeamIDKey); exists {
		if id, ok := v.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}
