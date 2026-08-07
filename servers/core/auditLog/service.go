package auditLog

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/audit"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/auditLog/auditLogDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPageLimit = 50
	maxPageLimit     = 200
)

// AuditLogService owns audit-log reads and maintenance for core.
type AuditLogService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var AuditLogServiceSingleton *AuditLogService

// ListAuditLog returns one keyset-paginated page of audit entries matching the
// filters, plus the cursor for the next (older) page.
func (s *AuditLogService) ListAuditLog(ctx context.Context, f auditLogDTO.ListFilters) (auditLogDTO.AuditLogPage, error) {
	limit := f.Limit
	if limit <= 0 || limit > maxPageLimit {
		limit = defaultPageLimit
	}

	rows, err := s.queries.ListAuditLog(ctx, db.ListAuditLogParams{
		CourseID:      pgUUID(f.CourseID),
		ActorRole:     pgText(f.ActorRole),
		Outcome:       pgText(f.Outcome),
		ActionKey:     pgText(f.ActionKey),
		EntityType:    pgText(f.EntityType),
		SourceService: pgText(f.SourceService),
		CoursePhaseID: pgUUID(f.CoursePhaseID),
		FromTime:      pgTimestamptz(f.From),
		ToTime:        pgTimestamptz(f.To),
		Search:        pgText(f.Search),
		CursorTs:      pgTimestamptz(f.CursorCreatedAt),
		CursorID:      pgUUID(f.CursorID),
		PageLimit:     int32(limit + 1), // fetch one extra to detect a next page
	})
	if err != nil {
		return auditLogDTO.AuditLogPage{}, err
	}

	page := auditLogDTO.AuditLogPage{Entries: make([]auditLogDTO.AuditEntry, 0, limit)}
	for i, row := range rows {
		if i >= limit {
			last := page.Entries[len(page.Entries)-1]
			page.NextCursor = &auditLogDTO.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
			break
		}
		page.Entries = append(page.Entries, toAuditEntry(row))
	}
	return page, nil
}

func toAuditEntry(row db.AuditLog) auditLogDTO.AuditEntry {
	return auditLogDTO.AuditEntry{
		ID:            row.ID.String(),
		CreatedAt:     row.CreatedAt.Time,
		ActorID:       uuidToString(row.ActorID),
		ActorName:     row.ActorName,
		ActorEmail:    row.ActorEmail,
		ActorRoles:    row.ActorRoles,
		ActorRole:     row.ActorRole,
		Action:        row.Action,
		ActionKey:     row.ActionKey,
		Outcome:       row.Outcome,
		EntityType:    textToString(row.EntityType),
		EntityID:      textToString(row.EntityID),
		EntityName:    textToString(row.EntityName),
		CourseID:      uuidToString(row.CourseID),
		CoursePhaseID: uuidToString(row.CoursePhaseID),
		SourceService: row.SourceService,
		HTTPMethod:    textToString(row.HttpMethod),
		HTTPPath:      textToString(row.HttpPath),
		HTTPStatus:    int4ToInt(row.HttpStatus),
		Metadata:      row.Metadata,
	}
}

// RecordTx writes an audit entry inside an existing transaction, so a
// high-stakes change and its audit row commit atomically. Callers pass the
// transaction-bound queries (queries.WithTx(tx)).
func RecordTx(ctx context.Context, qtx *db.Queries, e audit.Event) error {
	if !audit.Enabled() {
		return nil
	}
	if e.SourceService == "" {
		e.SourceService = "core"
	}
	_, err := qtx.CreateAuditEntry(ctx, eventToParams(ctx, qtx, e))
	return err
}

// StartRetentionPruner launches a daily goroutine that deletes audit entries
// older than AUDIT_RETENTION_DAYS. When the variable is unset, no pruning ever
// runs (fail-safe toward retention). Mirrors the privacy export-deletion job.
func StartRetentionPruner(ctx context.Context) {
	raw := sdkUtils.GetEnv("AUDIT_RETENTION_DAYS", "")
	if raw == "" {
		log.Info("audit retention disabled (AUDIT_RETENTION_DAYS not set); entries are never pruned")
		return
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days <= 0 {
		log.WithField("AUDIT_RETENTION_DAYS", raw).Warn("invalid AUDIT_RETENTION_DAYS; audit retention disabled")
		return
	}

	pruneOnce(ctx, days)
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				pruneOnce(ctx, days)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func pruneOnce(ctx context.Context, days int) {
	runCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	cutoff := time.Now().AddDate(0, 0, -days)
	if err := AuditLogServiceSingleton.queries.DeleteExpiredAuditEntries(runCtx, pgtype.Timestamptz{Time: cutoff, Valid: true}); err != nil {
		log.WithError(err).Error("failed to prune expired audit entries")
		return
	}
	log.WithField("olderThan", cutoff).Info("pruned expired audit entries")
}
