package actionItem

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type assessmentCompletionProvider interface {
	CheckAssessmentIsEditable(ctx context.Context, qtx *db.Queries, courseParticipationID, coursePhaseID uuid.UUID) error
	CheckAssessmentCompletionExists(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (bool, error)
	GetAssessmentCompletion(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (db.AssessmentCompletion, error)
}

type ActionItemService struct {
	queries              db.Queries
	assessmentCompletion assessmentCompletionProvider
}

func NewActionItemService(queries db.Queries, assessmentCompletion assessmentCompletionProvider) *ActionItemService {
	return &ActionItemService{
		queries:              queries,
		assessmentCompletion: assessmentCompletion,
	}
}

func (s *ActionItemService) GetActionItem(ctx context.Context, actionItemID uuid.UUID) (*actionItemDTO.ActionItem, error) {
	actionItem, err := s.queries.GetActionItem(ctx, actionItemID)
	if err != nil {
		log.Error("could not get action item: ", err)
		return nil, errors.New("could not get action item")
	}
	dto := actionItemDTO.MapDBActionItemToActionItemDTO(actionItem)
	return &dto, nil
}

func (s *ActionItemService) ListActionItemsForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]actionItemDTO.ActionItem, error) {
	actionItems, err := s.queries.ListActionItemsForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not list action items for course phase: ", err)
		return nil, errors.New("could not list action items for course phase")
	}
	return actionItemDTO.GetActionItemDTOsFromDBModels(actionItems), nil
}

func (s *ActionItemService) GetAllActionItemsForCoursePhaseCommunication(ctx context.Context, coursePhaseID uuid.UUID) ([]actionItemDTO.ActionItemWithParticipation, error) {
	actionItems, err := s.queries.GetAllActionItemsForCoursePhaseCommunication(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not list action items for course phase: ", err)
		return nil, errors.New("could not list action items for course phase")
	}
	return actionItemDTO.GetActionItemsFromDBActionItemsWithParticipation(actionItems), nil
}

func (s *ActionItemService) GetStudentActionItemsForCoursePhaseCommunication(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]string, error) {
	actionItems, err := s.queries.GetStudentActionItemsForCoursePhaseCommunication(ctx, db.GetStudentActionItemsForCoursePhaseCommunicationParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not list action items for student in phase: ", err)
		return nil, errors.New("could not list action items for student in phase")
	}
	return actionItems, nil
}

func (s *ActionItemService) CreateActionItem(ctx context.Context, req actionItemDTO.CreateActionItemRequest) error {
	err := s.assessmentCompletion.CheckAssessmentIsEditable(ctx, &s.queries, req.CourseParticipationID, req.CoursePhaseID)
	if err != nil {
		return err
	}
	err = s.queries.CreateActionItem(ctx, req.GetDBModel())
	if err != nil {
		log.Error("could not create action item: ", err)
		return errors.New("could not create action item")
	}
	return nil
}

func (s *ActionItemService) UpdateActionItem(ctx context.Context, req actionItemDTO.UpdateActionItemRequest) error {
	err := s.assessmentCompletion.CheckAssessmentIsEditable(ctx, &s.queries, req.CourseParticipationID, req.CoursePhaseID)
	if err != nil {
		return err
	}
	err = s.queries.UpdateActionItem(ctx, req.GetDBModel())
	if err != nil {
		log.Error("could not update action item: ", err)
		return errors.New("could not update action item")
	}
	return nil
}

func (s *ActionItemService) DeleteActionItem(ctx context.Context, actionItemID uuid.UUID) error {
	actionItem, err := s.queries.GetActionItem(ctx, actionItemID)
	if err != nil {
		log.Error("could not get action item: ", err)
		return errors.New("could not get action item")
	}

	err = s.assessmentCompletion.CheckAssessmentIsEditable(ctx, &s.queries, actionItem.CourseParticipationID, actionItem.CoursePhaseID)
	if err != nil {
		return err
	}

	err = s.queries.DeleteActionItem(ctx, actionItemID)
	if err != nil {
		log.Error("could not delete action item: ", err)
		return errors.New("could not delete action item")
	}
	return nil
}

func (s *ActionItemService) ListActionItemsForStudentInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]actionItemDTO.ActionItem, error) {
	actionItems, err := s.queries.ListActionItemsForStudentInPhase(ctx, db.ListActionItemsForStudentInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not list action items for student in phase: ", err)
		return nil, errors.New("could not list action items for student in phase")
	}
	return actionItemDTO.GetActionItemDTOsFromDBModels(actionItems), nil
}

func (s *ActionItemService) CountActionItemsForStudentInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (int64, error) {
	count, err := s.queries.CountActionItemsForStudentInPhase(ctx, db.CountActionItemsForStudentInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not count action items for student in phase: ", err)
		return 0, errors.New("could not count action items for student in phase")
	}
	return count, nil
}
