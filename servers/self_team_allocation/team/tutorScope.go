package teams

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/tutorscope"
)

// tutorScopeTable is the view from migration 0006 that holds only the tutor rows with a login. A
// row without one matches no editor, so counting it as a tutor of the phase would lock every
// editor out of its reads.
const tutorScopeTable = "tutor_with_login"

// TutorScopeMiddleware scopes a course editor's reads to the teams they lead in the course phase.
func TutorScopeMiddleware(conn *pgxpool.Pool) gin.HandlerFunc {
	return tutorscope.TeamsMiddleware(tutorscope.NewPgxTeamsResolver(conn, tutorscope.WithTable(tutorScopeTable)))
}
