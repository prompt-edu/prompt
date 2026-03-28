package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/storage/privacyexport"
	log "github.com/sirupsen/logrus"
)

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

func DeleteExpiredExportFiles(c context.Context) {
	exports, err := PrivacyServiceSingleton.queries.GetInvalidExports(c)
	if err != nil {
		log.WithError(err).Error("failed to fetch invalid exports for deletion")
		return
	}

	for _, exp := range exports {
		exportPGID := pgtype.UUID{Bytes: exp.ID, Valid: true}

		objectKeys, err := PrivacyServiceSingleton.queries.GetExportDocObjectKeysByExportID(c, exportPGID)
		if err != nil {
			log.WithError(err).WithField("exportID", exp.ID).Error("failed to fetch object keys for export")
			continue
		}

		deleteErr := false
		for _, key := range objectKeys {
			if err := privacyexport.DeleteFile(c, key); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"exportID":  exp.ID,
					"objectKey": key,
				}).Error("failed to delete export file from S3")
				deleteErr = true
			}
		}

		if deleteErr {
			continue
		}

		if err := PrivacyServiceSingleton.queries.SetExportDocStatusByExportID(c, db.SetExportDocStatusByExportIDParams{
			ExportID: exportPGID,
			Status:   db.ExportStatusArchived,
		}); err != nil {
			log.WithError(err).WithField("exportID", exp.ID).Error("failed to set export docs status to archived")
		}

		if err := SetExportStatus(c, exp.ID, db.ExportStatusArchived); err != nil {
			log.WithError(err).WithField("exportID", exp.ID).Error("failed to set export status to archived")
		}
	}
}
