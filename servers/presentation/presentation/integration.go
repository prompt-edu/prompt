package presentation

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
)

// AuditCopyAction names the copy route and the event its handler records, so both
// describe the same action in the audit log.
const AuditCopyAction = "Copied course phase"

type CopyHandler struct {
	Service *Service
}

func (h *CopyHandler) HandlePhaseCopy(c *gin.Context, request promptTypes.PhaseCopyRequest) error {
	// Core probes this endpoint by posting a copy of a phase onto itself to find out
	// whether the service supports copying. Copying in place clears the target first, so
	// it would wipe the phase's own slots, presentations and uploaded material.
	if request.SourceCoursePhaseID == request.TargetCoursePhaseID {
		audit.Suppress(c)
		return nil
	}
	recordCopyAudit(c, request)

	sourceConfig, err := h.Service.queries.GetCoursePhaseConfig(c, request.SourceCoursePhaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Nothing configured on the source phase, so there is nothing to copy.
		return nil
	}
	if err != nil {
		return fmt.Errorf("get source presentation config: %w", err)
	}
	sourceCategories, err := h.Service.queries.ListFeedbackCategories(c, request.SourceCoursePhaseID)
	if err != nil {
		return fmt.Errorf("get source feedback categories: %w", err)
	}
	targetStorageKeys, err := h.Service.queries.ListMaterialStorageKeysByPhase(c, request.TargetCoursePhaseID)
	if err != nil {
		return fmt.Errorf("get target presentation materials: %w", err)
	}

	tx, err := h.Service.conn.Begin(c)
	if err != nil {
		return fmt.Errorf("begin presentation copy: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, c)
	qtx := h.Service.queries.WithTx(tx)
	if err := qtx.DeletePresentationsByPhase(c, request.TargetCoursePhaseID); err != nil {
		return fmt.Errorf("clear target presentations: %w", err)
	}
	if err := qtx.DeletePresentationSlotsByPhase(c, request.TargetCoursePhaseID); err != nil {
		return fmt.Errorf("clear target presentation slots: %w", err)
	}
	if err := qtx.DeleteFeedbackCategoriesByPhase(c, request.TargetCoursePhaseID); err != nil {
		return fmt.Errorf("clear target feedback categories: %w", err)
	}
	if _, err := qtx.EnsureCoursePhaseConfig(c, request.TargetCoursePhaseID); err != nil {
		return fmt.Errorf("ensure target presentation config: %w", err)
	}
	if _, err := qtx.UpdateCoursePhaseConfig(c, db.UpdateCoursePhaseConfigParams{
		CoursePhaseID:         request.TargetCoursePhaseID,
		TargetMode:            sourceConfig.TargetMode,
		FeedbackMode:          sourceConfig.FeedbackMode,
		RequiredMaterialTypes: sourceConfig.RequiredMaterialTypes,
	}); err != nil {
		return fmt.Errorf("copy presentation config: %w", err)
	}
	for _, category := range sourceCategories {
		if _, err := qtx.CreateFeedbackCategory(c, db.CreateFeedbackCategoryParams{
			CoursePhaseID: request.TargetCoursePhaseID,
			Name:          category.Name,
			Description:   category.Description,
			Position:      category.Position,
		}); err != nil {
			return fmt.Errorf("copy feedback category: %w", err)
		}
	}
	if err := tx.Commit(c); err != nil {
		return fmt.Errorf("commit presentation copy: %w", err)
	}
	h.Service.deleteStoredMaterials(c, targetStorageKeys)
	return nil
}

// recordCopyAudit scopes the event to the target phase. The route sits outside
// :coursePhaseID, so an automatically captured event would carry no phase and never reach
// the course audit log. A blank target is left to that automatic entry rather than pinning
// the log to the nil phase.
func recordCopyAudit(c *gin.Context, request promptTypes.PhaseCopyRequest) {
	if request.TargetCoursePhaseID == uuid.Nil {
		return
	}
	audit.Record(c, audit.Event{
		Action:        AuditCopyAction,
		EntityType:    "coursePhase",
		EntityID:      request.TargetCoursePhaseID.String(),
		CoursePhaseID: request.TargetCoursePhaseID.String(),
		Metadata:      map[string]any{"sourceCoursePhaseID": request.SourceCoursePhaseID.String()},
	})
}

func (s *Service) PrivacyExportHandler(
	c *gin.Context,
	export *sdkUtils.Export,
	subject keycloakTokenVerifier.SubjectIdentifiers,
) error {
	userID := ""
	if subject.UserID != [16]byte{} {
		userID = subject.UserID.String()
	}
	export.AddJSON("Presentation materials", "presentation-materials.json", func() (any, error) {
		return s.queries.GetPrivacyMaterials(c, db.GetPrivacyMaterialsParams{
			UploaderUserID:         userID,
			CourseParticipationIds: subject.CourseParticipationIDs,
		})
	})
	if userID != "" {
		// Independent feedback is attributable through evaluator_user_id. Shared feedback
		// is not: those forms carry a NULL evaluator, so the subject's own work is only
		// reachable through their contributor rows and the answers they last wrote. Both
		// are exported per record rather than as whole forms, which would also disclose
		// co-instructors' text.
		export.AddJSON("Presentation feedback", "presentation-feedback.json", func() (any, error) {
			return s.queries.GetPrivacyFeedbackForms(c, pgtype.Text{String: userID, Valid: true})
		})
		export.AddJSON(
			"Presentation feedback contributions",
			"presentation-feedback-contributions.json",
			func() (any, error) { return s.queries.GetPrivacyFeedbackContributions(c, userID) },
		)
		export.AddJSON(
			"Presentation feedback answers",
			"presentation-feedback-answers.json",
			func() (any, error) { return s.queries.GetPrivacyFeedbackAnswers(c, userID) },
		)
	}
	return export.Err()
}

func (s *Service) PrivacyDeletionHandler(
	c *gin.Context,
	subject keycloakTokenVerifier.SubjectIdentifiers,
) error {
	if len(subject.CourseParticipationIDs) > 0 {
		if err := s.queries.AnonymizePrivacyPresentationTargets(c, subject.CourseParticipationIDs); err != nil {
			return fmt.Errorf("anonymize presentation targets: %w", err)
		}
	}
	if subject.UserID == [16]byte{} {
		return nil
	}
	userID := subject.UserID.String()
	if err := s.queries.DeletePrivacyDrafts(c, pgtype.Text{String: userID, Valid: true}); err != nil {
		return fmt.Errorf("delete private feedback drafts: %w", err)
	}
	if err := s.queries.AnonymizePrivacyFeedback(c, pgtype.Text{String: userID, Valid: true}); err != nil {
		return fmt.Errorf("anonymize feedback forms: %w", err)
	}
	if err := s.queries.AnonymizePrivacyContributions(c, userID); err != nil {
		return fmt.Errorf("anonymize feedback contributions: %w", err)
	}
	if err := s.queries.AnonymizePrivacyAnswers(c, userID); err != nil {
		return fmt.Errorf("anonymize feedback answers: %w", err)
	}
	if err := s.queries.AnonymizePrivacyMaterials(c, userID); err != nil {
		return fmt.Errorf("anonymize presentation materials: %w", err)
	}
	return nil
}
