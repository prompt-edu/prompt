package tutorscope

// ponytail: single local copy of tutor scoping; promote to prompt-sdk when a second service needs it.

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/team/teamDTO"
)

const teamIDKey = "tutorTeamID"

// Middleware resolves the requesting tutor's assigned team and stores it in the gin
// context for handlers to scope their responses. Non-tutor roles pass through untouched.
func Middleware(q db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
		if !ok || !tokenUser.IsEditor || tokenUser.IsLecturer {
			c.Next()
			return
		}
		login := teamDTO.NormalizeUniversityLogin(tokenUser.UniversityLogin)
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
			UniversityLogin: teamDTO.UniversityLoginParam(login),
		})
		// No tutor row: genuine editors/organizers (and tutors whose university_login isn't recorded) keep full, unscoped access by design.
		if errors.Is(err, pgx.ErrNoRows) {
			c.Next()
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "access check failed"})
			return
		}
		c.Set(teamIDKey, teamID)
		c.Next()
	}
}

// TeamID returns the tutor's scoped team and whether scoping applies to this request.
func TeamID(c *gin.Context) (uuid.UUID, bool) {
	if v, exists := c.Get(teamIDKey); exists {
		if id, ok := v.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}
