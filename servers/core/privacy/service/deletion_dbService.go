package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	authService "github.com/prompt-edu/prompt/servers/core/auth/service"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	coreutils "github.com/prompt-edu/prompt/servers/core/utils"
)

func CreateDeletionRequest(c *gin.Context) (privacyDTO.PrivacyDeletionRequest, error) {
	subjectIdentifiers, err := authService.GetSubjectIdentifiers(c)
	if err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, err
	}

	record, err := PrivacyServiceSingleton.queries.CreateNewDeletionRequest(c, db.CreateNewDeletionRequestParams{
		ID:             uuid.New(),
		UserID:         pgtype.UUID{Bytes: subjectIdentifiers.UserID, Valid: subjectIdentifiers.UserID != uuid.Nil},
		StudentID:      pgtype.UUID{Bytes: subjectIdentifiers.StudentID, Valid: subjectIdentifiers.StudentID != uuid.Nil},
		Status:         db.PrivacyDeletionRequestStatusPendingApproval,
		RecipientEmail: coreutils.GetUserEmailFromContext(c),
	})
	if err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to create deletion request: %w", err)
	}
	return privacyDTO.GetPrivacyDeletionRequestDTOFromDBModel(record), nil
}

func GetDeletionRequestWithSubrequests(c *gin.Context, requestID uuid.UUID) (privacyDTO.PrivacyDeletionRequest, error) {
	record, err := PrivacyServiceSingleton.queries.GetDeletionRequestByIDWithSubrequests(c, requestID)
	if err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, err
	}
	return privacyDTO.GetPrivacyDeletionRequestWithSubrequestsDTOFromDBModel(record)
}

func GetLatestDeletionRequestForUser(c *gin.Context) (*privacyDTO.PrivacyDeletionRequest, error) {
	userID, err := coreutils.GetUserUUIDFromContext(c)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user identity: %w", err)
	}

	record, err := PrivacyServiceSingleton.queries.GetLatestDeletionRequestForUserWithSubrequests(c, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch latest deletion request: %w", err)
	}

	dto, err := privacyDTO.GetPrivacyDeletionRequestWithSubrequestsDTOFromDBModel(record)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func GetDeletionRequestsByIDs(c context.Context, ids []uuid.UUID) ([]privacyDTO.PrivacyDeletionRequest, error) {
	rows, err := PrivacyServiceSingleton.queries.GetDeletionRequestsByIDsWithSubrequests(c, ids)
	if err != nil {
		return nil, err
	}
	results := make([]privacyDTO.PrivacyDeletionRequest, 0, len(rows))
	for _, row := range rows {
		dto, err := privacyDTO.GetPrivacyDeletionRequestWithSubrequestsDTOFromDBModel(row)
		if err != nil {
			return nil, err
		}
		results = append(results, dto)
	}
	return results, nil
}

func GetAllDeletionRequests(c context.Context) ([]privacyDTO.AdminPrivacyDeletionRequest, error) {
	dbRecords, err := PrivacyServiceSingleton.queries.GetAllDeletionRequests(c)
	if err != nil {
		return nil, err
	}

	results := make([]privacyDTO.AdminPrivacyDeletionRequest, 0, len(dbRecords))
	for _, dbRec := range dbRecords {
		dto, err := privacyDTO.GetAdminPrivacyDeletionRequestDTOFromDBModel(dbRec)
		if err != nil {
			return nil, err
		}
		results = append(results, dto)
	}
	return results, nil
}

func CreateDeletionSubrequest(ctx context.Context, q *db.Queries, deletionRequestID uuid.UUID, sourceName string, apiURL string) (ServiceDeletionRequest, error) {
	sub, err := q.CreateNewDeletionSubrequest(ctx, db.CreateNewDeletionSubrequestParams{
		ID:                uuid.New(),
		DeletionRequestID: deletionRequestID,
		SourceName:        sourceName,
		Status:            db.PrivacyDeletionSubrequestStatusPending,
	})
	if err != nil {
		return ServiceDeletionRequest{}, fmt.Errorf("failed to create deletion subrequest: %w", err)
	}
	return ServiceDeletionRequest{
		APIURL:     apiURL,
		Subrequest: privacyDTO.GetPrivacyDeletionSubrequestDTOFromDBModel(sub),
		Result:     db.PrivacyDeletionSubrequestStatusPending,
	}, nil
}

func AcceptDeletionRequest(c *gin.Context, requestID uuid.UUID, note string) (privacyDTO.PrivacyDeletionRequest, error) {
	auditorID, err := coreutils.GetUserUUIDFromContext(c)
	if err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to resolve auditor identity: %w", err)
	}

	tx, err := PrivacyServiceSingleton.conn.Begin(c)
	if err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(c) }()
	txQueries := PrivacyServiceSingleton.queries.WithTx(tx)

	if err := txQueries.SetDeletionRequestAuditor(c, db.SetDeletionRequestAuditorParams{
		ID:           requestID,
		AuditorID:    pgtype.UUID{Bytes: auditorID, Valid: true},
		AuditorName:  coreutils.GetUserNameFromContext(c),
		AuditorEmail: coreutils.GetUserEmailFromContext(c),
		AuditorNote:  note,
	}); err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to set auditor: %w", err)
	}

	record, err := txQueries.SetDeletionRequestStatus(c, db.SetDeletionRequestStatusParams{
		ID:     requestID,
		Status: db.PrivacyDeletionRequestStatusInProgress,
	})
	if err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to set status: %w", err)
	}

	if err := tx.Commit(c); err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return privacyDTO.GetPrivacyDeletionRequestDTOFromDBModel(record), nil
}

func RejectDeletionRequest(c *gin.Context, requestID uuid.UUID, note string) (privacyDTO.PrivacyDeletionRequest, error) {
	auditorID, err := coreutils.GetUserUUIDFromContext(c)
	if err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to resolve auditor identity: %w", err)
	}

	tx, err := PrivacyServiceSingleton.conn.Begin(c)
	if err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(c) }()
	txQueries := PrivacyServiceSingleton.queries.WithTx(tx)

	if err := txQueries.SetDeletionRequestAuditor(c, db.SetDeletionRequestAuditorParams{
		ID:           requestID,
		AuditorID:    pgtype.UUID{Bytes: auditorID, Valid: true},
		AuditorName:  coreutils.GetUserNameFromContext(c),
		AuditorEmail: coreutils.GetUserEmailFromContext(c),
		AuditorNote:  note,
	}); err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to set auditor: %w", err)
	}

	record, err := txQueries.SetDeletionRequestStatus(c, db.SetDeletionRequestStatusParams{
		ID:     requestID,
		Status: db.PrivacyDeletionRequestStatusRejected,
	})
	if err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to set status: %w", err)
	}

	if err := tx.Commit(c); err != nil {
		return privacyDTO.PrivacyDeletionRequest{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return privacyDTO.GetPrivacyDeletionRequestDTOFromDBModel(record), nil
}
