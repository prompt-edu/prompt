package presentation

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/presentation/storage"
	log "github.com/sirupsen/logrus"
)

const (
	targetModeIndividual = "individual"
	targetModeTeam       = "team"
	feedbackIndependent  = "independent"
	feedbackShared       = "shared"
)

type User struct {
	ID                    string
	Email                 string
	Name                  string
	CourseParticipationID uuid.UUID
	Staff                 bool
	CanRelease            bool
}

type Service struct {
	queries          *db.Queries
	conn             *pgxpool.Pool
	storage          storage.Adapter
	coreURL          string
	hub              *EventHub
	uploadTTLSeconds int
	downloadTTL      int
	maxUploadBytes   int64
	now              func() time.Time
}

func NewService(
	queries *db.Queries,
	conn *pgxpool.Pool,
	storageAdapter storage.Adapter,
	coreURL string,
	uploadTTLSeconds int,
	downloadTTLSeconds int,
	maxUploadBytes int64,
) *Service {
	return &Service{
		queries:          queries,
		conn:             conn,
		storage:          storageAdapter,
		coreURL:          strings.TrimRight(coreURL, "/"),
		hub:              NewEventHub(),
		uploadTTLSeconds: uploadTTLSeconds,
		downloadTTL:      downloadTTLSeconds,
		maxUploadBytes:   maxUploadBytes,
		now:              time.Now,
	}
}

func configResponse(config db.CoursePhaseConfig) SettingsResponse {
	return SettingsResponse{
		CoursePhaseID: config.CoursePhaseID,
		TargetMode:    config.TargetMode,
		FeedbackMode:  config.FeedbackMode,
	}
}

func categoryResponse(category db.FeedbackCategory) CategoryResponse {
	return CategoryResponse{
		ID:            category.ID,
		CoursePhaseID: category.CoursePhaseID,
		Name:          category.Name,
		Description:   category.Description,
		Position:      category.Position,
	}
}

func materialResponse(material db.PresentationMaterial) MaterialResponse {
	return MaterialResponse{
		ID:             material.ID,
		PresentationID: material.PresentationID,
		FileName:       material.OriginalFilename,
		ContentType:    material.ContentType,
		SizeBytes:      material.SizeBytes,
		UploadedByName: material.UploaderName,
		UploadedAt:     material.CreatedAt.Time,
	}
}

func (s *Service) GetConfig(ctx context.Context, coursePhaseID uuid.UUID) (SettingsResponse, error) {
	config, err := s.queries.EnsureCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		return SettingsResponse{}, fmt.Errorf("ensure presentation config: %w", err)
	}
	return configResponse(config), nil
}

func (s *Service) UpdateConfig(ctx context.Context, coursePhaseID uuid.UUID, request UpdateSettingsRequest) (SettingsResponse, error) {
	current, err := s.queries.EnsureCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		return SettingsResponse{}, fmt.Errorf("ensure presentation config: %w", err)
	}
	if current.TargetMode == request.TargetMode && current.FeedbackMode == request.FeedbackMode {
		return configResponse(current), nil
	}

	presentationCount, err := s.queries.CountPresentationsByPhase(ctx, coursePhaseID)
	if err != nil {
		return SettingsResponse{}, fmt.Errorf("count presentations: %w", err)
	}
	feedbackCount, err := s.queries.CountFeedbackFormsByPhase(ctx, coursePhaseID)
	if err != nil {
		return SettingsResponse{}, fmt.Errorf("count feedback forms: %w", err)
	}
	if (presentationCount > 0 || feedbackCount > 0) && !request.ResetExistingData {
		return SettingsResponse{}, apiError(
			409,
			"config_locked",
			"Changing target or feedback mode requires explicitly resetting existing presentation data",
			nil,
		)
	}
	var storageKeys []string
	if request.ResetExistingData {
		storageKeys, err = s.queries.ListMaterialStorageKeysByPhase(ctx, coursePhaseID)
		if err != nil {
			return SettingsResponse{}, fmt.Errorf("list materials during config reset: %w", err)
		}
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return SettingsResponse{}, fmt.Errorf("begin config update: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)
	if request.ResetExistingData {
		if err := qtx.DeletePresentationsByPhase(ctx, coursePhaseID); err != nil {
			return SettingsResponse{}, fmt.Errorf("delete presentations during config reset: %w", err)
		}
		if err := qtx.DeletePresentationSlotsByPhase(ctx, coursePhaseID); err != nil {
			return SettingsResponse{}, fmt.Errorf("delete slots during config reset: %w", err)
		}
	}
	config, err := qtx.UpdateCoursePhaseConfig(ctx, db.UpdateCoursePhaseConfigParams{
		CoursePhaseID: coursePhaseID,
		TargetMode:    request.TargetMode,
		FeedbackMode:  request.FeedbackMode,
	})
	if err != nil {
		return SettingsResponse{}, fmt.Errorf("update presentation config: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return SettingsResponse{}, fmt.Errorf("commit presentation config: %w", err)
	}
	for _, storageKey := range storageKeys {
		if err := s.storage.Delete(ctx, storageKey); err != nil {
			log.WithError(err).WithField("storageKey", storageKey).Warn("Failed to delete material object during presentation config reset")
		}
	}
	return configResponse(config), nil
}

func (s *Service) ListCategories(ctx context.Context, coursePhaseID uuid.UUID) ([]CategoryResponse, error) {
	if _, err := s.queries.EnsureCoursePhaseConfig(ctx, coursePhaseID); err != nil {
		return nil, fmt.Errorf("ensure presentation config: %w", err)
	}
	categories, err := s.queries.ListFeedbackCategories(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("list feedback categories: %w", err)
	}
	result := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		result = append(result, categoryResponse(category))
	}
	return result, nil
}

func (s *Service) CreateCategory(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	request CategoryRequest,
	resetExistingData bool,
) (CategoryResponse, error) {
	if strings.TrimSpace(request.Name) == "" {
		return CategoryResponse{}, apiError(400, "invalid_category", "Category name must not be empty", nil)
	}
	if _, err := s.queries.EnsureCoursePhaseConfig(ctx, coursePhaseID); err != nil {
		return CategoryResponse{}, fmt.Errorf("ensure presentation config: %w", err)
	}
	var category db.FeedbackCategory
	err := s.withCategoryMutation(ctx, coursePhaseID, resetExistingData, func(queries *db.Queries) error {
		var mutationErr error
		category, mutationErr = queries.CreateFeedbackCategory(ctx, db.CreateFeedbackCategoryParams{
			CoursePhaseID: coursePhaseID,
			Name:          strings.TrimSpace(request.Name),
			Description:   strings.TrimSpace(request.Description),
			Position:      request.Position,
		})
		return mutationErr
	})
	if err != nil {
		return CategoryResponse{}, fmt.Errorf("create feedback category: %w", err)
	}
	return categoryResponse(category), nil
}

func (s *Service) UpdateCategory(
	ctx context.Context,
	coursePhaseID, categoryID uuid.UUID,
	request CategoryRequest,
	resetExistingData bool,
) (CategoryResponse, error) {
	if strings.TrimSpace(request.Name) == "" {
		return CategoryResponse{}, apiError(400, "invalid_category", "Category name must not be empty", nil)
	}
	var category db.FeedbackCategory
	err := s.withCategoryMutation(ctx, coursePhaseID, resetExistingData, func(queries *db.Queries) error {
		var mutationErr error
		category, mutationErr = queries.UpdateFeedbackCategory(ctx, db.UpdateFeedbackCategoryParams{
			ID:            categoryID,
			CoursePhaseID: coursePhaseID,
			Name:          strings.TrimSpace(request.Name),
			Description:   strings.TrimSpace(request.Description),
			Position:      request.Position,
		})
		return mutationErr
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return CategoryResponse{}, apiError(404, "category_not_found", "Feedback category not found", err)
	}
	if err != nil {
		return CategoryResponse{}, fmt.Errorf("update feedback category: %w", err)
	}
	return categoryResponse(category), nil
}

func (s *Service) DeleteCategory(
	ctx context.Context,
	coursePhaseID, categoryID uuid.UUID,
	resetExistingData bool,
) error {
	var deleted int64
	err := s.withCategoryMutation(ctx, coursePhaseID, resetExistingData, func(queries *db.Queries) error {
		var mutationErr error
		deleted, mutationErr = queries.DeleteFeedbackCategory(ctx, db.DeleteFeedbackCategoryParams{
			ID:            categoryID,
			CoursePhaseID: coursePhaseID,
		})
		return mutationErr
	})
	if err != nil {
		return fmt.Errorf("delete feedback category: %w", err)
	}
	if deleted == 0 {
		return apiError(404, "category_not_found", "Feedback category not found", nil)
	}
	return nil
}

func (s *Service) withCategoryMutation(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	resetExistingData bool,
	mutate func(*db.Queries) error,
) error {
	count, err := s.queries.CountFeedbackFormsByPhase(ctx, coursePhaseID)
	if err != nil {
		return fmt.Errorf("count feedback forms: %w", err)
	}
	if count == 0 {
		return mutate(s.queries)
	}
	if !resetExistingData {
		return apiError(409, "categories_locked", "Feedback categories cannot change after feedback has been started", nil)
	}
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin category reset: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)
	if err := qtx.DeleteFeedbackFormsByPhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("delete phase feedback forms: %w", err)
	}
	if err := qtx.ClearFeedbackReleasesByPhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("clear phase feedback releases: %w", err)
	}
	if err := mutate(qtx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit category reset: %w", err)
	}
	return nil
}

func validateSlot(request SlotRequest) error {
	if !request.EndTime.After(request.StartTime) {
		return apiError(400, "invalid_slot", "Slot end time must be after its start time", nil)
	}
	return nil
}

func slotResponse(slot db.PresentationSlot, presentation *PresentationResponse) SlotResponse {
	location := ""
	if slot.Location.Valid {
		location = slot.Location.String
	}
	return SlotResponse{
		ID:            slot.ID,
		CoursePhaseID: slot.CoursePhaseID,
		StartTime:     slot.StartTime.Time,
		EndTime:       slot.EndTime.Time,
		Location:      location,
		Presentation:  presentation,
	}
}

func (s *Service) CreateSlot(ctx context.Context, coursePhaseID uuid.UUID, request SlotRequest) (SlotResponse, error) {
	if err := validateSlot(request); err != nil {
		return SlotResponse{}, err
	}
	slot, err := s.queries.CreatePresentationSlot(ctx, db.CreatePresentationSlotParams{
		CoursePhaseID: coursePhaseID,
		StartTime:     pgtype.Timestamptz{Time: request.StartTime, Valid: true},
		EndTime:       pgtype.Timestamptz{Time: request.EndTime, Valid: true},
		Location:      pgtype.Text{String: strings.TrimSpace(request.Location), Valid: strings.TrimSpace(request.Location) != ""},
	})
	if err != nil {
		return SlotResponse{}, fmt.Errorf("create presentation slot: %w", err)
	}
	return slotResponse(slot, nil), nil
}

func (s *Service) UpdateSlot(ctx context.Context, coursePhaseID, slotID uuid.UUID, request SlotRequest) (SlotResponse, error) {
	if err := validateSlot(request); err != nil {
		return SlotResponse{}, err
	}
	slot, err := s.queries.UpdatePresentationSlot(ctx, db.UpdatePresentationSlotParams{
		ID:            slotID,
		CoursePhaseID: coursePhaseID,
		StartTime:     pgtype.Timestamptz{Time: request.StartTime, Valid: true},
		EndTime:       pgtype.Timestamptz{Time: request.EndTime, Valid: true},
		Location:      pgtype.Text{String: strings.TrimSpace(request.Location), Valid: strings.TrimSpace(request.Location) != ""},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return SlotResponse{}, apiError(404, "slot_not_found", "Presentation slot not found", err)
	}
	if err != nil {
		return SlotResponse{}, fmt.Errorf("update presentation slot: %w", err)
	}
	var presentationResponse *PresentationResponse
	presentation, err := s.queries.GetPresentationBySlot(ctx, db.GetPresentationBySlotParams{
		CoursePhaseID: coursePhaseID,
		SlotID:        slotID,
	})
	if err == nil {
		response, responseErr := s.presentationResponse(ctx, presentation, slot)
		if responseErr != nil {
			return SlotResponse{}, responseErr
		}
		presentationResponse = &response
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return SlotResponse{}, fmt.Errorf("get slot assignment: %w", err)
	}
	return slotResponse(slot, presentationResponse), nil
}

func (s *Service) DeleteSlot(ctx context.Context, coursePhaseID, slotID uuid.UUID) error {
	deleted, err := s.queries.DeletePresentationSlot(ctx, db.DeletePresentationSlotParams{
		ID:            slotID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		return fmt.Errorf("delete presentation slot: %w", err)
	}
	if deleted == 0 {
		return apiError(409, "slot_assigned", "Unassign the presentation before deleting this slot", nil)
	}
	return nil
}

func (s *Service) ListSlots(ctx context.Context, coursePhaseID uuid.UUID) ([]SlotResponse, error) {
	slots, err := s.queries.ListPresentationSlots(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("list presentation slots: %w", err)
	}
	presentations, err := s.queries.ListPresentations(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("list slot assignments: %w", err)
	}
	bySlot := make(map[uuid.UUID]PresentationResponse, len(presentations))
	for _, presentation := range presentations {
		response, responseErr := s.listPresentationResponse(ctx, presentation)
		if responseErr != nil {
			return nil, responseErr
		}
		bySlot[presentation.SlotID] = response
	}
	result := make([]SlotResponse, 0, len(slots))
	for _, slot := range slots {
		var assignment *PresentationResponse
		if presentation, exists := bySlot[slot.ID]; exists {
			presentationCopy := presentation
			assignment = &presentationCopy
		}
		result = append(result, slotResponse(slot, assignment))
	}
	return result, nil
}

func (s *Service) ListTargets(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) ([]TargetResponse, error) {
	config, err := s.GetConfig(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	var targets []TargetResponse
	if config.TargetMode == targetModeTeam {
		targets, err = s.fetchTeamTargets(authHeader, coursePhaseID)
	} else {
		targets, err = s.fetchIndividualTargets(authHeader, coursePhaseID)
	}
	if err != nil {
		return nil, err
	}
	presentations, err := s.queries.ListPresentations(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("list assigned presentation targets: %w", err)
	}
	assigned := make(map[uuid.UUID]uuid.UUID, len(presentations))
	for _, presentation := range presentations {
		assigned[presentation.TargetID] = presentation.ID
	}
	for index := range targets {
		if presentationID, exists := assigned[targets[index].ID]; exists {
			presentationIDCopy := presentationID
			targets[index].AssignedPresentationID = &presentationIDCopy
		}
	}
	return targets, nil
}

func (s *Service) AssignTarget(
	ctx context.Context,
	authHeader string,
	coursePhaseID, slotID uuid.UUID,
	request AssignmentRequest,
) (PresentationResponse, error) {
	config, err := s.GetConfig(ctx, coursePhaseID)
	if err != nil {
		return PresentationResponse{}, err
	}
	if request.TargetType != config.TargetMode {
		return PresentationResponse{}, apiError(400, "target_mode_mismatch", "Assignment target does not match the configured target mode", nil)
	}
	targets, err := s.ListTargets(ctx, authHeader, coursePhaseID)
	if err != nil {
		return PresentationResponse{}, err
	}
	targetName := ""
	for _, target := range targets {
		if target.ID == request.TargetID {
			targetName = target.Name
			break
		}
	}
	if targetName == "" {
		return PresentationResponse{}, apiError(400, "invalid_target", "The selected presentation target is not available", nil)
	}

	slot, err := s.queries.GetPresentationSlot(ctx, db.GetPresentationSlotParams{
		ID:            slotID,
		CoursePhaseID: coursePhaseID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return PresentationResponse{}, apiError(404, "slot_not_found", "Presentation slot not found", err)
	}
	if err != nil {
		return PresentationResponse{}, fmt.Errorf("get presentation slot: %w", err)
	}
	if _, err := s.queries.GetPresentationBySlot(ctx, db.GetPresentationBySlotParams{
		CoursePhaseID: coursePhaseID,
		SlotID:        slotID,
	}); err == nil {
		return PresentationResponse{}, apiError(409, "slot_assigned", "This slot already has a presentation assigned", nil)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return PresentationResponse{}, fmt.Errorf("check slot assignment: %w", err)
	}
	if _, err := s.queries.GetPresentationByTarget(ctx, db.GetPresentationByTargetParams{
		CoursePhaseID: coursePhaseID,
		TargetType:    request.TargetType,
		TargetID:      request.TargetID,
	}); err == nil {
		return PresentationResponse{}, apiError(409, "target_assigned", "This target already has a presentation slot", nil)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return PresentationResponse{}, fmt.Errorf("check target assignment: %w", err)
	}

	presentation, err := s.queries.CreatePresentation(ctx, db.CreatePresentationParams{
		CoursePhaseID: coursePhaseID,
		SlotID:        slotID,
		TargetType:    request.TargetType,
		TargetID:      request.TargetID,
		TargetName:    targetName,
	})
	if err != nil {
		return PresentationResponse{}, fmt.Errorf("assign presentation target: %w", err)
	}
	return s.presentationResponse(ctx, presentation, slot)
}

func (s *Service) UnassignTarget(ctx context.Context, coursePhaseID, slotID uuid.UUID) error {
	presentation, err := s.queries.GetPresentationBySlot(ctx, db.GetPresentationBySlotParams{
		CoursePhaseID: coursePhaseID,
		SlotID:        slotID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apiError(404, "assignment_not_found", "Presentation assignment not found", err)
	}
	if err != nil {
		return fmt.Errorf("get presentation assignment: %w", err)
	}
	dependencies, err := s.queries.CountPresentationDependencies(ctx, presentation.ID)
	if err != nil {
		return fmt.Errorf("count presentation dependencies: %w", err)
	}
	if dependencies.MaterialCount > 0 || dependencies.FeedbackCount > 0 {
		return apiError(409, "presentation_has_data", "Delete presentation materials and feedback before unassigning it", nil)
	}
	deleted, err := s.queries.DeletePresentation(ctx, db.DeletePresentationParams{
		ID:            presentation.ID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		return fmt.Errorf("unassign presentation target: %w", err)
	}
	if deleted == 0 {
		return apiError(404, "assignment_not_found", "Presentation assignment not found", nil)
	}
	return nil
}

func (s *Service) ListPresentations(ctx context.Context, coursePhaseID uuid.UUID) ([]PresentationResponse, error) {
	presentations, err := s.queries.ListPresentations(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("list presentations: %w", err)
	}
	result := make([]PresentationResponse, 0, len(presentations))
	for _, presentation := range presentations {
		response, responseErr := s.listPresentationResponse(ctx, presentation)
		if responseErr != nil {
			return nil, responseErr
		}
		result = append(result, response)
	}
	return result, nil
}

func (s *Service) GetOwnPresentation(
	ctx context.Context,
	authHeader string,
	coursePhaseID uuid.UUID,
	user User,
) (*PresentationResponse, error) {
	config, err := s.GetConfig(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	targetID := user.CourseParticipationID
	if config.TargetMode == targetModeTeam {
		targetID, err = s.fetchOwnTeamID(authHeader, coursePhaseID, user.CourseParticipationID)
		if err != nil {
			return nil, err
		}
	}
	presentation, err := s.queries.GetPresentationByTarget(ctx, db.GetPresentationByTargetParams{
		CoursePhaseID: coursePhaseID,
		TargetType:    config.TargetMode,
		TargetID:      targetID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get own presentation: %w", err)
	}
	slot, err := s.queries.GetPresentationSlot(ctx, db.GetPresentationSlotParams{
		ID:            presentation.SlotID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		return nil, fmt.Errorf("get own presentation slot: %w", err)
	}
	response, err := s.presentationResponse(ctx, presentation, slot)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *Service) getPresentationAndAuthorize(
	ctx context.Context,
	authHeader string,
	coursePhaseID, presentationID uuid.UUID,
	user User,
) (db.Presentation, db.PresentationSlot, error) {
	presentation, err := s.queries.GetPresentation(ctx, db.GetPresentationParams{
		ID:            presentationID,
		CoursePhaseID: coursePhaseID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Presentation{}, db.PresentationSlot{}, apiError(404, "presentation_not_found", "Presentation not found", err)
	}
	if err != nil {
		return db.Presentation{}, db.PresentationSlot{}, fmt.Errorf("get presentation: %w", err)
	}
	if !user.Staff {
		targetID := user.CourseParticipationID
		if presentation.TargetType == targetModeTeam {
			targetID, err = s.fetchOwnTeamID(authHeader, coursePhaseID, user.CourseParticipationID)
			if err != nil {
				return db.Presentation{}, db.PresentationSlot{}, err
			}
		}
		if presentation.TargetID != targetID {
			return db.Presentation{}, db.PresentationSlot{}, apiError(403, "presentation_forbidden", "You cannot access this presentation", nil)
		}
	}
	slot, err := s.queries.GetPresentationSlot(ctx, db.GetPresentationSlotParams{
		ID:            presentation.SlotID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		return db.Presentation{}, db.PresentationSlot{}, fmt.Errorf("get presentation slot: %w", err)
	}
	return presentation, slot, nil
}

func (s *Service) presentationResponse(
	ctx context.Context,
	presentation db.Presentation,
	slot db.PresentationSlot,
) (PresentationResponse, error) {
	dependencies, err := s.queries.CountPresentationDependencies(ctx, presentation.ID)
	if err != nil {
		return PresentationResponse{}, fmt.Errorf("count presentation data: %w", err)
	}
	submittedCount, err := s.queries.CountSubmittedFeedbackForms(ctx, presentation.ID)
	if err != nil {
		return PresentationResponse{}, fmt.Errorf("count submitted feedback: %w", err)
	}
	location := ""
	if slot.Location.Valid {
		location = slot.Location.String
	}
	response := PresentationResponse{
		ID:                     presentation.ID,
		CoursePhaseID:          presentation.CoursePhaseID,
		SlotID:                 presentation.SlotID,
		TargetType:             presentation.TargetType,
		TargetID:               presentation.TargetID,
		TargetName:             presentation.TargetName,
		StartTime:              slot.StartTime.Time,
		EndTime:                slot.EndTime.Time,
		Location:               location,
		MaterialCount:          dependencies.MaterialCount,
		FeedbackCount:          dependencies.FeedbackCount,
		SubmittedFeedbackCount: submittedCount,
	}
	if presentation.FeedbackReleasedAt.Valid {
		value := presentation.FeedbackReleasedAt.Time
		response.FeedbackReleasedAt = &value
	}
	if presentation.FeedbackReleaseName.Valid {
		value := presentation.FeedbackReleaseName.String
		response.FeedbackReleaseName = &value
	}
	if presentation.FeedbackReleasedByName.Valid {
		value := presentation.FeedbackReleasedByName.String
		response.FeedbackReleasedByName = &value
	}
	return response, nil
}

func (s *Service) listPresentationResponse(
	ctx context.Context,
	presentation db.ListPresentationsRow,
) (PresentationResponse, error) {
	model := db.Presentation{
		ID:                       presentation.ID,
		CoursePhaseID:            presentation.CoursePhaseID,
		SlotID:                   presentation.SlotID,
		TargetType:               presentation.TargetType,
		TargetID:                 presentation.TargetID,
		TargetName:               presentation.TargetName,
		FeedbackReleaseName:      presentation.FeedbackReleaseName,
		FeedbackReleasedAt:       presentation.FeedbackReleasedAt,
		FeedbackReleasedByUserID: presentation.FeedbackReleasedByUserID,
		FeedbackReleasedByName:   presentation.FeedbackReleasedByName,
		CreatedAt:                presentation.CreatedAt,
		UpdatedAt:                presentation.UpdatedAt,
	}
	slot := db.PresentationSlot{
		ID:            presentation.SlotID,
		CoursePhaseID: presentation.CoursePhaseID,
		StartTime:     presentation.StartTime,
		EndTime:       presentation.EndTime,
		Location:      presentation.Location,
	}
	return s.presentationResponse(ctx, model, slot)
}

func (s *Service) ensureMaterialMutationAllowed(user User, slot db.PresentationSlot) error {
	if user.Staff {
		return nil
	}
	if !s.now().Before(slot.StartTime.Time) {
		return apiError(409, "material_deadline_passed", "Students can only change materials before the presentation starts", nil)
	}
	return nil
}

func (s *Service) ListMaterials(
	ctx context.Context,
	authHeader string,
	coursePhaseID, presentationID uuid.UUID,
	user User,
) ([]MaterialResponse, error) {
	if _, _, err := s.getPresentationAndAuthorize(ctx, authHeader, coursePhaseID, presentationID, user); err != nil {
		return nil, err
	}
	materials, err := s.queries.ListPresentationMaterials(ctx, presentationID)
	if err != nil {
		return nil, fmt.Errorf("list presentation materials: %w", err)
	}
	result := make([]MaterialResponse, 0, len(materials))
	for _, material := range materials {
		result = append(result, materialResponse(material))
	}
	return result, nil
}

func (s *Service) CreateUploadIntent(
	ctx context.Context,
	authHeader string,
	coursePhaseID, presentationID uuid.UUID,
	user User,
	request PresignMaterialRequest,
) (PresignMaterialResponse, error) {
	_, slot, err := s.getPresentationAndAuthorize(ctx, authHeader, coursePhaseID, presentationID, user)
	if err != nil {
		return PresignMaterialResponse{}, err
	}
	if err := s.ensureMaterialMutationAllowed(user, slot); err != nil {
		return PresignMaterialResponse{}, err
	}
	if request.SizeBytes <= 0 || request.SizeBytes > s.maxUploadBytes {
		return PresignMaterialResponse{}, apiError(
			413,
			"material_too_large",
			fmt.Sprintf("Material must be at most %d bytes", s.maxUploadBytes),
			nil,
		)
	}
	fileName := filepath.Base(strings.TrimSpace(request.FileName))
	if fileName == "" || fileName == "." {
		return PresignMaterialResponse{}, apiError(400, "invalid_file_name", "Material file name is invalid", nil)
	}
	contentType := strings.TrimSpace(request.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	uploadID := uuid.New()
	expiresAt := s.now().Add(time.Duration(s.uploadTTLSeconds) * time.Second)
	storageKey := fmt.Sprintf(
		"presentations/%s/%s/%s/%s",
		coursePhaseID,
		presentationID,
		uploadID,
		fileName,
	)
	material, err := s.queries.CreatePendingMaterial(ctx, db.CreatePendingMaterialParams{
		PresentationID:   presentationID,
		OriginalFilename: fileName,
		ContentType:      contentType,
		StorageKey:       storageKey,
		UploaderUserID:   user.ID,
		UploaderName:     user.Name,
		UploaderEmail:    user.Email,
		ExpiresAt:        pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return PresignMaterialResponse{}, fmt.Errorf("create pending presentation material: %w", err)
	}
	uploadURL, err := s.storage.GetUploadURL(ctx, storageKey, contentType, s.uploadTTLSeconds)
	if err != nil {
		_, _ = s.queries.DeletePresentationMaterial(ctx, db.DeletePresentationMaterialParams{
			ID:             material.ID,
			PresentationID: presentationID,
		})
		return PresignMaterialResponse{}, fmt.Errorf("create material upload URL: %w", err)
	}
	return PresignMaterialResponse{
		UploadID:  material.ID,
		UploadURL: uploadURL,
		Headers:   map[string]string{"Content-Type": contentType},
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) CompleteUpload(
	ctx context.Context,
	authHeader string,
	coursePhaseID, presentationID, uploadID uuid.UUID,
	user User,
) (MaterialResponse, error) {
	_, slot, err := s.getPresentationAndAuthorize(ctx, authHeader, coursePhaseID, presentationID, user)
	if err != nil {
		return MaterialResponse{}, err
	}
	if err := s.ensureMaterialMutationAllowed(user, slot); err != nil {
		return MaterialResponse{}, err
	}
	pending, err := s.queries.GetPresentationMaterial(ctx, db.GetPresentationMaterialParams{
		ID:            uploadID,
		CoursePhaseID: coursePhaseID,
	})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && pending.PresentationID != presentationID) {
		return MaterialResponse{}, apiError(404, "upload_not_found", "Pending material upload not found", err)
	}
	if err != nil {
		return MaterialResponse{}, fmt.Errorf("get pending material upload: %w", err)
	}
	if pending.State != "pending" {
		return MaterialResponse{}, apiError(409, "upload_already_completed", "Material upload has already been completed", nil)
	}
	metadata, err := s.storage.GetMetadata(ctx, pending.StorageKey)
	if errors.Is(err, storage.ErrObjectNotFound) {
		return MaterialResponse{}, apiError(409, "upload_incomplete", "The material object has not been uploaded yet", err)
	}
	if err != nil {
		return MaterialResponse{}, fmt.Errorf("read uploaded material metadata: %w", err)
	}
	if metadata.Size <= 0 || metadata.Size > s.maxUploadBytes {
		_ = s.storage.Delete(ctx, pending.StorageKey)
		_, _ = s.queries.DeletePresentationMaterial(ctx, db.DeletePresentationMaterialParams{
			ID:             pending.ID,
			PresentationID: presentationID,
		})
		return MaterialResponse{}, apiError(413, "material_too_large", "Uploaded material exceeds the configured size limit", nil)
	}
	contentType := metadata.ContentType
	if contentType == "" {
		contentType = pending.ContentType
	}
	material, err := s.queries.CompletePresentationMaterial(ctx, db.CompletePresentationMaterialParams{
		ID:             pending.ID,
		PresentationID: presentationID,
		SizeBytes:      metadata.Size,
		ContentType:    contentType,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MaterialResponse{}, apiError(409, "upload_expired", "Pending material upload has expired", err)
	}
	if err != nil {
		return MaterialResponse{}, fmt.Errorf("complete presentation material: %w", err)
	}
	return materialResponse(material), nil
}

func (s *Service) GetMaterialDownload(
	ctx context.Context,
	authHeader string,
	coursePhaseID, presentationID, materialID uuid.UUID,
	user User,
) (MaterialDownloadResponse, error) {
	if _, _, err := s.getPresentationAndAuthorize(ctx, authHeader, coursePhaseID, presentationID, user); err != nil {
		return MaterialDownloadResponse{}, err
	}
	material, err := s.queries.GetPresentationMaterial(ctx, db.GetPresentationMaterialParams{
		ID:            materialID,
		CoursePhaseID: coursePhaseID,
	})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && material.PresentationID != presentationID) {
		return MaterialDownloadResponse{}, apiError(404, "material_not_found", "Presentation material not found", err)
	}
	if err != nil {
		return MaterialDownloadResponse{}, fmt.Errorf("get presentation material: %w", err)
	}
	if material.State != "ready" {
		return MaterialDownloadResponse{}, apiError(404, "material_not_found", "Presentation material not found", nil)
	}
	downloadURL, err := s.storage.GetDownloadURL(ctx, material.StorageKey, s.downloadTTL)
	if err != nil {
		return MaterialDownloadResponse{}, fmt.Errorf("create material download URL: %w", err)
	}
	return MaterialDownloadResponse{
		DownloadURL: downloadURL,
		FileName:    material.OriginalFilename,
		ExpiresAt:   s.now().Add(time.Duration(s.downloadTTL) * time.Second),
	}, nil
}

func (s *Service) DeleteMaterial(
	ctx context.Context,
	authHeader string,
	coursePhaseID, presentationID, materialID uuid.UUID,
	user User,
) error {
	_, slot, err := s.getPresentationAndAuthorize(ctx, authHeader, coursePhaseID, presentationID, user)
	if err != nil {
		return err
	}
	if err := s.ensureMaterialMutationAllowed(user, slot); err != nil {
		return err
	}
	material, err := s.queries.GetPresentationMaterial(ctx, db.GetPresentationMaterialParams{
		ID:            materialID,
		CoursePhaseID: coursePhaseID,
	})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && material.PresentationID != presentationID) {
		return apiError(404, "material_not_found", "Presentation material not found", err)
	}
	if err != nil {
		return fmt.Errorf("get presentation material: %w", err)
	}
	deleted, err := s.queries.DeletePresentationMaterial(ctx, db.DeletePresentationMaterialParams{
		ID:             materialID,
		PresentationID: presentationID,
	})
	if err != nil {
		return fmt.Errorf("delete presentation material: %w", err)
	}
	if deleted == 0 {
		return apiError(404, "material_not_found", "Presentation material not found", nil)
	}
	if err := s.storage.Delete(ctx, material.StorageKey); err != nil {
		return fmt.Errorf("delete stored presentation material: %w", err)
	}
	return nil
}

func answerResponse(answer db.FeedbackAnswer) FeedbackAnswerResponse {
	return FeedbackAnswerResponse{
		CategoryID:    answer.CategoryID,
		Value:         answer.Value,
		Revision:      answer.Revision,
		UpdatedByName: answer.UpdatedByName,
		UpdatedAt:     answer.UpdatedAt.Time,
	}
}

func (s *Service) feedbackFormResponse(ctx context.Context, form db.FeedbackForm, ownUserID string) (FeedbackFormResponse, error) {
	answers, err := s.queries.ListFeedbackAnswers(ctx, form.ID)
	if err != nil {
		return FeedbackFormResponse{}, fmt.Errorf("list feedback answers: %w", err)
	}
	contributors, err := s.queries.ListFeedbackContributors(ctx, form.ID)
	if err != nil {
		return FeedbackFormResponse{}, fmt.Errorf("list feedback contributors: %w", err)
	}
	answerResponses := make([]FeedbackAnswerResponse, 0, len(answers))
	for _, answer := range answers {
		answerResponses = append(answerResponses, answerResponse(answer))
	}
	contributorResponses := make([]ContributorResponse, 0, len(contributors))
	for _, contributor := range contributors {
		contributorResponses = append(contributorResponses, ContributorResponse{
			UserID:             contributor.UserID,
			Name:               contributor.Name,
			FirstContributedAt: contributor.FirstContributedAt.Time,
			LastContributedAt:  contributor.LastContributedAt.Time,
		})
	}
	response := FeedbackFormResponse{
		ID:            form.ID,
		EvaluatorName: form.EvaluatorName,
		Status:        form.Status,
		Answers:       answerResponses,
		Contributors:  contributorResponses,
		IsOwn:         form.EvaluatorUserID.Valid && form.EvaluatorUserID.String == ownUserID,
	}
	if form.SubmittedAt.Valid {
		value := form.SubmittedAt.Time
		response.SubmittedAt = &value
	}
	return response, nil
}

func (s *Service) feedbackScope(config SettingsResponse, user User) (string, pgtype.Text, string) {
	if config.FeedbackMode == feedbackShared {
		return "shared", pgtype.Text{}, "Shared feedback"
	}
	return user.ID, pgtype.Text{String: user.ID, Valid: true}, user.Name
}

func (s *Service) getOrCreateFeedbackForm(
	ctx context.Context,
	config SettingsResponse,
	presentationID uuid.UUID,
	user User,
) (db.FeedbackForm, error) {
	scope, evaluatorID, evaluatorName := s.feedbackScope(config, user)
	form, err := s.queries.GetFeedbackFormByScope(ctx, db.GetFeedbackFormByScopeParams{
		PresentationID: presentationID,
		ScopeKey:       scope,
	})
	if err == nil {
		return form, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.FeedbackForm{}, fmt.Errorf("get feedback form: %w", err)
	}
	form, err = s.queries.CreateFeedbackForm(ctx, db.CreateFeedbackFormParams{
		PresentationID:  presentationID,
		ScopeKey:        scope,
		EvaluatorUserID: evaluatorID,
		EvaluatorName:   evaluatorName,
		EvaluatorEmail:  user.Email,
	})
	if err != nil {
		return db.FeedbackForm{}, fmt.Errorf("create feedback form: %w", err)
	}
	return form, nil
}

func (s *Service) GetFeedback(
	ctx context.Context,
	authHeader string,
	coursePhaseID, presentationID uuid.UUID,
	user User,
) (FeedbackDocumentResponse, error) {
	presentation, slot, err := s.getPresentationAndAuthorize(ctx, authHeader, coursePhaseID, presentationID, user)
	if err != nil {
		return FeedbackDocumentResponse{}, err
	}
	if !user.Staff && !presentation.FeedbackReleasedAt.Valid {
		return FeedbackDocumentResponse{}, apiError(403, "feedback_not_released", "Feedback has not been released yet", nil)
	}
	config, err := s.GetConfig(ctx, coursePhaseID)
	if err != nil {
		return FeedbackDocumentResponse{}, err
	}
	categories, err := s.ListCategories(ctx, coursePhaseID)
	if err != nil {
		return FeedbackDocumentResponse{}, err
	}
	presentationResponse, err := s.presentationResponse(ctx, presentation, slot)
	if err != nil {
		return FeedbackDocumentResponse{}, err
	}

	document := FeedbackDocumentResponse{
		Presentation:  presentationResponse,
		Mode:          config.FeedbackMode,
		Categories:    categories,
		Forms:         []FeedbackFormResponse{},
		ActiveEditors: []ActiveEditorResponse{},
		CanEdit:       user.Staff && !presentation.FeedbackReleasedAt.Valid,
		CanRelease:    user.CanRelease && !presentation.FeedbackReleasedAt.Valid,
	}
	if config.FeedbackMode == feedbackShared {
		document.ActiveEditors = s.hub.ActiveEditors(presentationID)
	}

	var forms []db.FeedbackForm
	if user.Staff {
		forms, err = s.queries.ListSubmittedFeedbackForms(ctx, presentationID)
	} else {
		forms, err = s.queries.ListSubmittedFeedbackForms(ctx, presentationID)
	}
	if err != nil {
		return FeedbackDocumentResponse{}, fmt.Errorf("list visible feedback forms: %w", err)
	}
	for _, form := range forms {
		response, responseErr := s.feedbackFormResponse(ctx, form, user.ID)
		if responseErr != nil {
			return FeedbackDocumentResponse{}, responseErr
		}
		document.Forms = append(document.Forms, response)
	}

	if user.Staff {
		scope, _, _ := s.feedbackScope(config, user)
		ownForm, formErr := s.queries.GetFeedbackFormByScope(ctx, db.GetFeedbackFormByScopeParams{
			PresentationID: presentationID,
			ScopeKey:       scope,
		})
		if formErr == nil {
			response, responseErr := s.feedbackFormResponse(ctx, ownForm, user.ID)
			if responseErr != nil {
				return FeedbackDocumentResponse{}, responseErr
			}
			if config.FeedbackMode == feedbackShared {
				response.IsOwn = true
			}
			document.OwnForm = &response
			if config.FeedbackMode == feedbackShared && ownForm.Status == "draft" {
				document.Forms = []FeedbackFormResponse{response}
			}
		} else if !errors.Is(formErr, pgx.ErrNoRows) {
			return FeedbackDocumentResponse{}, fmt.Errorf("get own feedback form: %w", formErr)
		}
	}
	return document, nil
}

func (s *Service) PutFeedbackAnswer(
	ctx context.Context,
	coursePhaseID, presentationID, categoryID uuid.UUID,
	user User,
	request PutAnswerRequest,
) (FeedbackAnswerResponse, error) {
	presentation, err := s.queries.GetPresentation(ctx, db.GetPresentationParams{
		ID:            presentationID,
		CoursePhaseID: coursePhaseID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return FeedbackAnswerResponse{}, apiError(404, "presentation_not_found", "Presentation not found", err)
	}
	if err != nil {
		return FeedbackAnswerResponse{}, fmt.Errorf("get presentation: %w", err)
	}
	if presentation.FeedbackReleasedAt.Valid {
		return FeedbackAnswerResponse{}, apiError(409, "feedback_released", "Unrelease feedback before editing it", nil)
	}
	if _, err := s.queries.GetFeedbackCategory(ctx, db.GetFeedbackCategoryParams{
		ID:            categoryID,
		CoursePhaseID: coursePhaseID,
	}); errors.Is(err, pgx.ErrNoRows) {
		return FeedbackAnswerResponse{}, apiError(404, "category_not_found", "Feedback category not found", err)
	} else if err != nil {
		return FeedbackAnswerResponse{}, fmt.Errorf("get feedback category: %w", err)
	}
	config, err := s.GetConfig(ctx, coursePhaseID)
	if err != nil {
		return FeedbackAnswerResponse{}, err
	}
	form, err := s.getOrCreateFeedbackForm(ctx, config, presentationID, user)
	if err != nil {
		return FeedbackAnswerResponse{}, err
	}
	if form.Status != "draft" {
		return FeedbackAnswerResponse{}, apiError(409, "feedback_submitted", "Reopen submitted feedback before editing it", nil)
	}
	answer, err := s.queries.PutFeedbackAnswer(ctx, db.PutFeedbackAnswerParams{
		FeedbackFormID:   form.ID,
		CategoryID:       categoryID,
		Value:            request.Value,
		ExpectedRevision: request.ExpectedRevision,
		UpdatedByUserID:  user.ID,
		UpdatedByName:    user.Name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		current, currentErr := s.queries.GetFeedbackAnswer(ctx, db.GetFeedbackAnswerParams{
			FeedbackFormID: form.ID,
			CategoryID:     categoryID,
		})
		if currentErr != nil {
			return FeedbackAnswerResponse{}, apiError(409, "feedback_conflict", "Feedback was updated by another editor", nil)
		}
		currentResponse := answerResponse(current)
		return FeedbackAnswerResponse{}, &APIError{
			Status:  409,
			Code:    "feedback_conflict",
			Message: "Feedback was updated by another editor",
			Details: currentResponse,
			Err:     err,
		}
	}
	if err != nil {
		return FeedbackAnswerResponse{}, fmt.Errorf("update feedback answer: %w", err)
	}
	if _, err := s.queries.UpsertFeedbackContributor(ctx, db.UpsertFeedbackContributorParams{
		FeedbackFormID: form.ID,
		UserID:         user.ID,
		Name:           user.Name,
		Email:          user.Email,
	}); err != nil {
		return FeedbackAnswerResponse{}, fmt.Errorf("record feedback contributor: %w", err)
	}
	response := answerResponse(answer)
	if config.FeedbackMode == feedbackShared {
		s.hub.Publish(FeedbackEvent{
			Type:           "answer.updated",
			PresentationID: presentationID,
			Answer:         &response,
		})
	}
	return response, nil
}

func (s *Service) setFeedbackStatus(
	ctx context.Context,
	coursePhaseID, presentationID uuid.UUID,
	user User,
	status string,
) error {
	presentation, err := s.queries.GetPresentation(ctx, db.GetPresentationParams{
		ID:            presentationID,
		CoursePhaseID: coursePhaseID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apiError(404, "presentation_not_found", "Presentation not found", err)
	}
	if err != nil {
		return fmt.Errorf("get presentation: %w", err)
	}
	if presentation.FeedbackReleasedAt.Valid {
		return apiError(409, "feedback_released", "Unrelease feedback before changing its status", nil)
	}
	config, err := s.GetConfig(ctx, coursePhaseID)
	if err != nil {
		return err
	}
	form, err := s.getOrCreateFeedbackForm(ctx, config, presentationID, user)
	if err != nil {
		return err
	}
	if _, err := s.queries.SetFeedbackFormStatus(ctx, db.SetFeedbackFormStatusParams{
		ID:             form.ID,
		PresentationID: presentationID,
		Status:         status,
	}); err != nil {
		return fmt.Errorf("set feedback status: %w", err)
	}
	if config.FeedbackMode == feedbackShared {
		s.hub.Publish(FeedbackEvent{Type: "form.status.changed", PresentationID: presentationID})
	}
	return nil
}

func (s *Service) SubmitFeedback(ctx context.Context, coursePhaseID, presentationID uuid.UUID, user User) error {
	return s.setFeedbackStatus(ctx, coursePhaseID, presentationID, user, "submitted")
}

func (s *Service) ReopenFeedback(ctx context.Context, coursePhaseID, presentationID uuid.UUID, user User) error {
	return s.setFeedbackStatus(ctx, coursePhaseID, presentationID, user, "draft")
}

func (s *Service) DeleteDraft(ctx context.Context, coursePhaseID, presentationID uuid.UUID, user User) error {
	presentation, err := s.queries.GetPresentation(ctx, db.GetPresentationParams{
		ID:            presentationID,
		CoursePhaseID: coursePhaseID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apiError(404, "presentation_not_found", "Presentation not found", err)
	}
	if err != nil {
		return fmt.Errorf("get presentation: %w", err)
	}
	if presentation.FeedbackReleasedAt.Valid {
		return apiError(409, "feedback_released", "Unrelease feedback before deleting drafts", nil)
	}
	config, err := s.GetConfig(ctx, coursePhaseID)
	if err != nil {
		return err
	}
	scope, _, _ := s.feedbackScope(config, user)
	form, err := s.queries.GetFeedbackFormByScope(ctx, db.GetFeedbackFormByScopeParams{
		PresentationID: presentationID,
		ScopeKey:       scope,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get feedback draft: %w", err)
	}
	deleted, err := s.queries.DeleteDraftFeedbackForm(ctx, db.DeleteDraftFeedbackFormParams{
		ID:             form.ID,
		PresentationID: presentationID,
	})
	if err != nil {
		return fmt.Errorf("delete feedback draft: %w", err)
	}
	if deleted == 0 {
		return apiError(409, "feedback_submitted", "Submitted feedback must be reopened before deletion", nil)
	}
	if config.FeedbackMode == feedbackShared {
		s.hub.Publish(FeedbackEvent{Type: "form.status.changed", PresentationID: presentationID})
	}
	return nil
}

func (s *Service) ReleaseFeedback(
	ctx context.Context,
	coursePhaseID, presentationID uuid.UUID,
	user User,
	releaseName string,
) error {
	releaseName = strings.TrimSpace(releaseName)
	if releaseName == "" {
		return apiError(400, "invalid_release_name", "Feedback release name must not be empty", nil)
	}
	presentation, err := s.queries.GetPresentation(ctx, db.GetPresentationParams{
		ID:            presentationID,
		CoursePhaseID: coursePhaseID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apiError(404, "presentation_not_found", "Presentation not found", err)
	}
	if err != nil {
		return fmt.Errorf("get presentation: %w", err)
	}
	if presentation.FeedbackReleasedAt.Valid {
		return nil
	}
	draftCount, err := s.queries.CountDraftFeedbackForms(ctx, presentationID)
	if err != nil {
		return fmt.Errorf("count feedback drafts: %w", err)
	}
	if draftCount > 0 {
		return apiError(409, "feedback_drafts_exist", "All instructor drafts must be submitted or deleted before release", nil)
	}
	submittedCount, err := s.queries.CountSubmittedFeedbackForms(ctx, presentationID)
	if err != nil {
		return fmt.Errorf("count submitted feedback: %w", err)
	}
	if submittedCount == 0 {
		return apiError(409, "feedback_empty", "At least one submitted feedback form is required before release", nil)
	}
	if _, err := s.queries.SetFeedbackRelease(ctx, db.SetFeedbackReleaseParams{
		ID:                       presentationID,
		CoursePhaseID:            coursePhaseID,
		FeedbackReleaseName:      pgtype.Text{String: releaseName, Valid: true},
		FeedbackReleasedByUserID: pgtype.Text{String: user.ID, Valid: true},
		FeedbackReleasedByName:   pgtype.Text{String: user.Name, Valid: true},
	}); err != nil {
		return fmt.Errorf("release presentation feedback: %w", err)
	}
	s.hub.Publish(FeedbackEvent{Type: "released", PresentationID: presentationID})
	return nil
}

func (s *Service) UnreleaseFeedback(ctx context.Context, coursePhaseID, presentationID uuid.UUID) error {
	if _, err := s.queries.ClearFeedbackRelease(ctx, db.ClearFeedbackReleaseParams{
		ID:            presentationID,
		CoursePhaseID: coursePhaseID,
	}); errors.Is(err, pgx.ErrNoRows) {
		return apiError(404, "presentation_not_found", "Presentation not found", err)
	} else if err != nil {
		return fmt.Errorf("unrelease presentation feedback: %w", err)
	}
	s.hub.Publish(FeedbackEvent{Type: "released", PresentationID: presentationID})
	return nil
}

func (s *Service) ResetFeedback(ctx context.Context, coursePhaseID, presentationID uuid.UUID) error {
	if _, err := s.queries.GetPresentation(ctx, db.GetPresentationParams{
		ID:            presentationID,
		CoursePhaseID: coursePhaseID,
	}); errors.Is(err, pgx.ErrNoRows) {
		return apiError(404, "presentation_not_found", "Presentation not found", err)
	} else if err != nil {
		return fmt.Errorf("get presentation: %w", err)
	}
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin feedback reset: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)
	if err := qtx.ResetPresentationFeedback(ctx, presentationID); err != nil {
		return fmt.Errorf("reset presentation feedback: %w", err)
	}
	if _, err := qtx.ClearFeedbackRelease(ctx, db.ClearFeedbackReleaseParams{
		ID:            presentationID,
		CoursePhaseID: coursePhaseID,
	}); err != nil {
		return fmt.Errorf("clear feedback release: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit feedback reset: %w", err)
	}
	s.hub.Publish(FeedbackEvent{Type: "form.status.changed", PresentationID: presentationID})
	return nil
}
