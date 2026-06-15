package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/core/storage/privacyexport"
	log "github.com/sirupsen/logrus"
)

const deletionBatchLimit = 500

func StartExportDeletionRoutine(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				runCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
				DeleteExpiredExportFiles(runCtx)
				cancel()

			case <-ctx.Done():
				return
			}
		}
	}()
}

func DeleteExpiredExportFiles(ctx context.Context) {
	log.Info("gdpr export deletion started.")
	exports, err := PrivacyServiceSingleton.queries.GetInvalidExports(ctx, deletionBatchLimit)
	if err != nil {
		log.WithError(err).Error("failed to fetch invalid exports for deletion")
		return
	}

	for _, exp := range exports {
		if err := ArchiveExport(ctx, exp.ID); err != nil {
			log.WithError(err).WithField("exportID", exp.ID).Error("failed to archive expired export")
		}
	}
}

func ArchiveExport(ctx context.Context, exportID uuid.UUID) error {
	exportPGID := pgtype.UUID{Bytes: exportID, Valid: true}

	objectKeys, err := PrivacyServiceSingleton.queries.GetExportDocObjectKeysByExportID(ctx, exportPGID)
	if err != nil {
		return fmt.Errorf("fetch object keys: %w", err)
	}

	for _, key := range objectKeys {
		if err := privacyexport.DeleteFile(ctx, key); err != nil {
			return fmt.Errorf("delete s3 object %s: %w", key, err)
		}
	}

	if err := PrivacyServiceSingleton.queries.ArchiveCompletedExportDocs(ctx, exportPGID); err != nil {
		return fmt.Errorf("archive completed docs: %w", err)
	}

	if err := PrivacyServiceSingleton.queries.ArchiveExportRecord(ctx, exportID); err != nil {
		return fmt.Errorf("archive export record: %w", err)
	}

	return nil
}
