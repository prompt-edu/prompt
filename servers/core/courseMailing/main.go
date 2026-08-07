package courseMailing

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

func InitCourseMailingModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	CourseMailingServiceSingleton = &CourseMailingService{
		queries:   queries,
		conn:      conn,
		clientURL: sdkUtils.GetEnv("CORE_HOST", "localhost:3000"),
	}
	setupCourseMailingRouter(routerGroup, keycloakTokenVerifier.KeycloakMiddleware, checkAccessControlByIDWrapper)

	// Recover any campaigns left mid-send by a previous crash/restart.
	CourseMailingServiceSingleton.ReconcileStuckCampaigns(context.Background())
}

// checkAccessControlByIDWrapper enforces course-level permissions on the :uuid param.
func checkAccessControlByIDWrapper(allowedRoles ...string) gin.HandlerFunc {
	return permissionValidation.CheckAccessControlByID(permissionValidation.CheckCoursePermission, "uuid", allowedRoles...)
}
