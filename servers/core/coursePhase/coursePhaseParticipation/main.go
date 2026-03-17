package coursePhaseParticipation

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

func InitCoursePhaseParticipationModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupCoursePhaseParticipationRouter(routerGroup, keycloakTokenVerifier.KeycloakMiddleware, checkAccessControlByIDWrapper)
	CoursePhaseParticipationServiceSingleton = &CoursePhaseParticipationService{
		queries: queries,
		conn:    conn,
	}
}

// initializes the handler func with CheckCoursePermissions
func checkAccessControlByIDWrapper(allowedRoles ...string) gin.HandlerFunc {
	return permissionValidation.CheckAccessControlByID(permissionValidation.CheckCoursePhasePermission, "uuid", allowedRoles...)
}
