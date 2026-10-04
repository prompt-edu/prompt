package service

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	coreutils "github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

const adminInitiatedAuditorNote = "Admin-initiated bulk deletion"

func (s *PrivacyService) CreateAdminInitiatedDeletionRequests(c *gin.Context, studentIDs []uuid.UUID) ([]privacyDTO.PrivacyDeletionRequest, error) {
	auditorID, err := coreutils.GetUserUUIDFromContext(c)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve admin identity: %w", err)
	}
	auditorName := coreutils.GetUserNameFromContext(c)
	auditorEmail := coreutils.GetUserEmailFromContext(c)

	records := make([]privacyDTO.PrivacyDeletionRequest, 0, len(studentIDs))
	for _, sid := range studentIDs {
		recipientEmail := ""
		stud, err := s.students.GetStudentByID(c, sid)
		if err != nil {
			log.WithError(err).WithField("studentID", sid).Warn("failed to load student for confirmation email; proceeding without")
		} else {
			recipientEmail = stud.Email
		}

		row, err := s.queries.CreateAdminInitiatedDeletionRequest(c, db.CreateAdminInitiatedDeletionRequestParams{
			ID:             uuid.New(),
			StudentID:      pgtype.UUID{Bytes: sid, Valid: true},
			AuditorID:      pgtype.UUID{Bytes: auditorID, Valid: true},
			AuditorName:    auditorName,
			AuditorEmail:   auditorEmail,
			AuditorNote:    adminInitiatedAuditorNote,
			RecipientEmail: recipientEmail,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create admin-initiated deletion request for student %s: %w", sid, err)
		}
		records = append(records, privacyDTO.GetPrivacyDeletionRequestDTOFromDBModel(row))
	}
	return records, nil
}

func (s *PrivacyService) RunAdminInitiatedDeletions(ctx context.Context, authHeader string, records []privacyDTO.PrivacyDeletionRequest) {
	for _, rec := range records {
		state, err := s.PrepareDataDeletion(ctx, rec)
		if err != nil {
			log.WithError(err).WithField("requestID", rec.ID).Error("admin-initiated deletion: prepare failed")
			s.MarkDeletionRequestFailed(ctx, rec.ID)
			continue
		}
		s.RunDataDeletion(ctx, authHeader, state)
	}
}
