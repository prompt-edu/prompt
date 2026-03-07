package copy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

func InitCourseCopyModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {

	setupCourseCopyRouter(routerGroup, keycloakTokenVerifier.KeycloakMiddleware, permissionValidation.CheckAccessControlByRole, checkAccessControlByIDWrapper)
	CourseCopyServiceSingleton = &CourseCopyService{
		queries:                    queries,
		conn:                       conn,
		createCourseGroupsAndRoles: keycloakRealmManager.CreateCourseGroupsAndRoles,
	}

	// possibly more setup tasks
}

// initializes the handler func with CheckCoursePermissions
func checkAccessControlByIDWrapper(allowedRoles ...string) gin.HandlerFunc {
	return permissionValidation.CheckAccessControlByID(permissionValidation.CheckCoursePermission, "uuid", allowedRoles...)
}
