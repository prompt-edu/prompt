package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/storage/files"
	log "github.com/sirupsen/logrus"
)

func ExecuteCoreDeletion(ctx context.Context, subject sdk.SubjectIdentifiers) error {

	// collect application file IDs
	// must happen before the student delete
	fileIDs, err := collectApplicationFileIDs(ctx, subject.CourseParticipationIDs)
	if err != nil {
		return fmt.Errorf("collect application file IDs: %w", err)
	}

	// begin transaction
	tx, err := PrivacyServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin deletion transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := PrivacyServiceSingleton.queries.WithTx(tx)

	// Student record related
	if err := deleteStudentScopedData(ctx, q, subject.StudentID); err != nil {
		return fmt.Errorf("delete student-scoped data: %w", err)
	}

	// user / keycloak record related
	if err := deleteUserScopedData(ctx, q, subject.UserID); err != nil {
		return fmt.Errorf("delete user-scoped data: %w", err)
	}

	// commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit deletion transaction: %w", err)
	}

	// applicationFiles — non-transactional (DB row + blob)
	if err := deleteApplicationFiles(ctx, fileIDs); err != nil {
		return fmt.Errorf("delete application files: %w", err)
	}

	// Privacy Exports — non-transactional (S3 + own DB writes)
	if err := archiveSubjectPrivacyExports(ctx, subject.UserID); err != nil {
		return fmt.Errorf("archive privacy exports: %w", err)
	}

	return nil
}

func collectApplicationFileIDs(ctx context.Context, courseParticipationIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(courseParticipationIDs) == 0 {
		return nil, nil
	}
	return PrivacyServiceSingleton.queries.GetApplicationFileIDsByCourseParticipationIDs(ctx, courseParticipationIDs)
}

func deleteStudentScopedData(ctx context.Context, q *db.Queries, studentID uuid.UUID) error {

	if studentID == uuid.Nil {
		return nil
	}

	// Cascade chain (migration 0025 added cascade to note FKs):
	//   student -> course_participation -> course_phase_participation
	//     -> application_answer_{text,multi_select,file_upload}
	//     -> application_assessment
	//   student -> note -> note_version, note_tag_relation
	if err := q.DeleteStudentByID(ctx, studentID); err != nil {
		return fmt.Errorf("delete student row: %w", err)
	}
	return nil
}

func deleteUserScopedData(ctx context.Context, q *db.Queries, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return nil
	}

	if err := q.AnonymizeNotesByAuthor(ctx, userID); err != nil {
		return fmt.Errorf("anonymize authored notes: %w", err)
	}

	if err := q.AnonymizeFilesByUploader(ctx, userID.String()); err != nil {
		return fmt.Errorf("anonymize uploaded files: %w", err)
	}

	if err := q.ScrubDeletionRequestAuditorByID(ctx, pgtype.UUID{Bytes: userID, Valid: true}); err != nil {
		return fmt.Errorf("scrub auditor identity: %w", err)
	}
	return nil
}

func deleteApplicationFiles(ctx context.Context, fileIDs []uuid.UUID) error {
	for _, fileID := range fileIDs {
		if err := files.StorageServiceSingleton.DeleteFile(ctx, fileID, true); err != nil {
			log.WithError(err).WithField("fileID", fileID).Warn("failed to delete application file during privacy deletion")
		}
	}
	return nil
}

func archiveSubjectPrivacyExports(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return nil
	}

	exportIDs, err := PrivacyServiceSingleton.queries.GetExportIDsForUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return fmt.Errorf("list exports: %w", err)
	}
	for _, exportID := range exportIDs {
		if err := ArchiveExport(ctx, exportID); err != nil {
			return fmt.Errorf("archive export %s: %w", exportID, err)
		}
	}
	return nil
}
