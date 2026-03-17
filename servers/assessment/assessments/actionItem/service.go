package actionItem

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type ActionItemService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var ActionItemServiceSingleton *ActionItemService

func GetActionItem(ctx context.Context, actionItemID uuid.UUID) (*actionItemDTO.ActionItem, error) {
	actionItem, err := ActionItemServiceSingleton.queries.GetActionItem(ctx, actionItemID)
	if err != nil {
		log.Error("could not get action item: ", err)
		return nil, errors.New("could not get action item")
	}
	dto := actionItemDTO.MapDBActionItemToActionItemDTO(actionItem)
	return &dto, nil
}

func ListActionItemsForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]actionItemDTO.ActionItem, error) {
	actionItems, err := ActionItemServiceSingleton.queries.ListActionItemsForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not list action items for course phase: ", err)
		return nil, errors.New("could not list action items for course phase")
	}
	return actionItemDTO.GetActionItemDTOsFromDBModels(actionItems), nil
}

func GetAllActionItemsForCoursePhaseCommunication(ctx context.Context, coursePhaseID uuid.UUID) ([]actionItemDTO.ActionItemWithParticipation, error) {
	actionItems, err := ActionItemServiceSingleton.queries.GetAllActionItemsForCoursePhaseCommunication(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not list action items for course phase: ", err)
		return nil, errors.New("could not list action items for course phase")
	}
	return actionItemDTO.GetActionItemsFromDBActionItemsWithParticipation(actionItems), nil
}

func GetStudentActionItemsForCoursePhaseCommunication(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]string, error) {
	actionItems, err := ActionItemServiceSingleton.queries.GetStudentActionItemsForCoursePhaseCommunication(ctx, db.GetStudentActionItemsForCoursePhaseCommunicationParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not list action items for student in phase: ", err)
		return nil, errors.New("could not list action items for student in phase")
	}
	return actionItems, nil
}

func CreateActionItem(ctx context.Context, req actionItemDTO.CreateActionItemRequest) error {
	err := assessmentCompletion.CheckAssessmentIsEditable(ctx, &ActionItemServiceSingleton.queries, req.CourseParticipationID, req.CoursePhaseID)
	if err != nil {
		return err
	}
	err = ActionItemServiceSingleton.queries.CreateActionItem(ctx, req.GetDBModel())
	if err != nil {
		log.Error("could not create action item: ", err)
		return errors.New("could not create action item")
	}
	return nil
}

func UpdateActionItem(ctx context.Context, req actionItemDTO.UpdateActionItemRequest) error {
	err := assessmentCompletion.CheckAssessmentIsEditable(ctx, &ActionItemServiceSingleton.queries, req.CourseParticipationID, req.CoursePhaseID)
	if err != nil {
		return err
	}
	err = ActionItemServiceSingleton.queries.UpdateActionItem(ctx, req.GetDBModel())
	if err != nil {
		log.Error("could not update action item: ", err)
		return errors.New("could not update action item")
	}
	return nil
}

func DeleteActionItem(ctx context.Context, actionItemID uuid.UUID) error {
	actionItem, err := ActionItemServiceSingleton.queries.GetActionItem(ctx, actionItemID)
	if err != nil {
		log.Error("could not get action item: ", err)
		return errors.New("could not get action item")
	}

	err = assessmentCompletion.CheckAssessmentIsEditable(ctx, &ActionItemServiceSingleton.queries, actionItem.CourseParticipationID, actionItem.CoursePhaseID)
	if err != nil {
		return err
	}

	err = ActionItemServiceSingleton.queries.DeleteActionItem(ctx, actionItemID)
	if err != nil {
		log.Error("could not delete action item: ", err)
		return errors.New("could not delete action item")
	}
	return nil
}

func ListActionItemsForStudentInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]actionItemDTO.ActionItem, error) {
	actionItems, err := ActionItemServiceSingleton.queries.ListActionItemsForStudentInPhase(ctx, db.ListActionItemsForStudentInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not list action items for student in phase: ", err)
		return nil, errors.New("could not list action items for student in phase")
	}
	return actionItemDTO.GetActionItemDTOsFromDBModels(actionItems), nil
}

func CountActionItemsForStudentInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (int64, error) {
	count, err := ActionItemServiceSingleton.queries.CountActionItemsForStudentInPhase(ctx, db.CountActionItemsForStudentInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not count action items for student in phase: ", err)
		return 0, errors.New("could not count action items for student in phase")
	}
	return count, nil
}
