package auditLog

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/audit"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// InitAuditLogModule wires up audit logging for core: it registers the
// auto-capture middleware on the API group (so every mutating core route is
// recorded), mounts the read and ingest endpoints, and starts the retention
// pruner. It must be called before other modules so the middleware wraps their
// routes. It is a no-op unless the AUDIT_ENABLED feature toggle is set.
func InitAuditLogModule(api *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	if !audit.Enabled() {
		log.Info("audit logging disabled (AUDIT_ENABLED not set)")
		return
	}

	AuditLogServiceSingleton = &AuditLogService{queries: queries, conn: conn}
	sink := NewDBSink(queries)

	api.Use(audit.Middleware(sink,
		audit.WithActorExtractor(CoreActorExtractor),
		audit.WithSourceService("core")))

	setupAuditLogRouter(api, sink)
	StartRetentionPruner(context.Background())
	log.Info("audit logging enabled")
}
