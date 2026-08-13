package auditLog

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/audit"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// InitAuditLogCapture registers the auto-capture middleware on the API group so
// every mutating core route is recorded. Gin snapshots the middleware chain when
// a route or subgroup is registered, so this MUST run before any module (in
// particular initKeycloak) registers its routes; otherwise those routes keep the
// pre-audit chain and their mutations are never captured. It is a no-op unless
// the AUDIT_ENABLED feature toggle is set.
func InitAuditLogCapture(api *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	if !audit.Enabled() {
		log.Info("audit logging disabled (AUDIT_ENABLED not set)")
		return
	}

	AuditLogServiceSingleton = &AuditLogService{queries: queries, conn: conn}

	api.Use(audit.Middleware(NewDBSink(queries),
		audit.WithActorExtractor(CoreActorExtractor),
		audit.WithSourceService("core")))
}

// InitAuditLogRoutes mounts the audit read and ingest endpoints and starts the
// retention pruner. Call it after permissionValidation is initialized (the read
// routes use its access-control middleware). It is a no-op unless AUDIT_ENABLED
// is set, and pairs with InitAuditLogCapture.
func InitAuditLogRoutes(api *gin.RouterGroup) {
	if !audit.Enabled() {
		return
	}

	setupAuditLogRouter(api, NewDBSink(AuditLogServiceSingleton.queries))
	StartRetentionPruner(context.Background())
	log.Info("audit logging enabled")
}
