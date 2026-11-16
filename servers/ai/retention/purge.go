package retention

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
	"github.com/prompt-edu/prompt/servers/ai/feature"
	log "github.com/sirupsen/logrus"
)

const (
	purgeInterval = 24 * time.Hour
	// Longer than any request timeout: such a call lost its server mid-request.
	abandonedAfter = time.Hour
)

func Validate(metadataRetentionDays int) error {
	if longest := feature.MaxContentRetentionDays(); metadataRetentionDays <= longest {
		return fmt.Errorf("AI_AUDIT_METADATA_RETENTION_DAYS (%d) must be longer than the longest content retention (%d days)",
			metadataRetentionDays, longest)
	}
	return nil
}

// The first run catches up after the service was switched off.
func Start(ctx context.Context, queries *db.Queries, metadataRetentionDays int) {
	run := func() {
		if err := Purge(ctx, queries, time.Now(), metadataRetentionDays); err != nil {
			log.WithError(err).Error("Could not purge expired AI audit data")
		}
	}
	go func() {
		run()
		ticker := time.NewTicker(purgeInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}

func Purge(ctx context.Context, queries *db.Queries, now time.Time, metadataRetentionDays int) error {
	if _, err := queries.AbandonPendingCalls(ctx, timestamp(now.Add(-abandonedAfter))); err != nil {
		return fmt.Errorf("complete abandoned calls: %w", err)
	}
	features, err := queries.ListCallFeatures(ctx)
	if err != nil {
		return fmt.Errorf("list call features: %w", err)
	}
	for _, name := range features {
		cutoff := daysBefore(now, feature.PolicyOf(name).ContentRetentionDays)
		if _, err := queries.PurgeCallContents(ctx, db.PurgeCallContentsParams{Feature: name, Cutoff: cutoff}); err != nil {
			return fmt.Errorf("purge content of %q: %w", name, err)
		}
	}
	if _, err := queries.PurgeCalls(ctx, daysBefore(now, metadataRetentionDays)); err != nil {
		return fmt.Errorf("purge call metadata: %w", err)
	}
	return nil
}

func daysBefore(now time.Time, days int) pgtype.Timestamptz {
	return timestamp(now.AddDate(0, 0, -days))
}

func timestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}
