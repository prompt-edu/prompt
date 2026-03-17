package interviewAssignment

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interviewAssignmentDTO "github.com/prompt-edu/prompt/servers/interview/interviewAssignment/interviewAssignmentDTO"
	interviewSlotDTO "github.com/prompt-edu/prompt/servers/interview/interviewSlot/interviewSlotDTO"
	log "github.com/sirupsen/logrus"
)

type InterviewAssignmentService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var InterviewAssignmentServiceSingleton *InterviewAssignmentService

type ServiceError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *ServiceError) Error() string {
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

func newServiceError(statusCode int, message string, err error) *ServiceError {
	return &ServiceError{StatusCode: statusCode, Message: message, Err: err}
}

func pgTimestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}

func pgTextToStringPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func CreateInterviewAssignment(ctx context.Context, coursePhaseID uuid.UUID, participationUUID uuid.UUID, req interviewAssignmentDTO.CreateInterviewAssignmentRequest) (interviewAssignmentDTO.InterviewAssignmentResponse, error) {
	if participationUUID == uuid.Nil {
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusBadRequest, "Invalid course participation ID", nil)
	}

	tx, err := InterviewAssignmentServiceSingleton.conn.Begin(ctx)
	if err != nil {
		log.Errorf("Failed to begin transaction: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to process request", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := InterviewAssignmentServiceSingleton.queries.WithTx(tx)

	slot, err := qtx.GetInterviewSlotForUpdate(ctx, req.InterviewSlotID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusNotFound, "Interview slot not found", err)
		}
		log.Errorf("Failed to get interview slot: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to get interview slot", err)
	}

	if slot.CoursePhaseID != coursePhaseID {
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusNotFound, "Interview slot not found", nil)
	}

	count, err := qtx.CountAssignmentsBySlot(ctx, req.InterviewSlotID)
	if err != nil {
		log.Errorf("Failed to count assignments: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to check slot availability", err)
	}
	if count >= int64(slot.Capacity) {
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusConflict, "Interview slot is full", nil)
	}

	existingAssignment, err := qtx.GetInterviewAssignmentByParticipation(ctx, db.GetInterviewAssignmentByParticipationParams{
		CourseParticipationID: participationUUID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("Failed to check existing assignment: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to check existing assignment", err)
	}
	if err == nil && existingAssignment.ID != uuid.Nil {
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusConflict, "You already have an interview slot assigned", nil)
	}

	assignment, err := qtx.CreateInterviewAssignment(ctx, db.CreateInterviewAssignmentParams{
		InterviewSlotID:       req.InterviewSlotID,
		CourseParticipationID: participationUUID,
	})
	if err != nil {
		log.Errorf("Failed to create interview assignment: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to create interview assignment", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Errorf("Failed to commit transaction: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to create assignment", err)
	}

	return interviewAssignmentDTO.InterviewAssignmentResponse{
		ID:                    assignment.ID,
		InterviewSlotID:       assignment.InterviewSlotID,
		CourseParticipationID: assignment.CourseParticipationID,
		AssignedAt:            pgTimestamptzToTime(assignment.AssignedAt),
	}, nil
}

func CreateInterviewAssignmentAdmin(ctx context.Context, coursePhaseID uuid.UUID, req interviewAssignmentDTO.CreateInterviewAssignmentAdminRequest) (interviewAssignmentDTO.InterviewAssignmentResponse, error) {
	if req.CourseParticipationID == uuid.Nil {
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusBadRequest, "Invalid course participation ID", nil)
	}

	tx, err := InterviewAssignmentServiceSingleton.conn.Begin(ctx)
	if err != nil {
		log.Errorf("Failed to begin transaction: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to process request", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := InterviewAssignmentServiceSingleton.queries.WithTx(tx)

	slot, err := qtx.GetInterviewSlotForUpdate(ctx, req.InterviewSlotID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusNotFound, "Interview slot not found", err)
		}
		log.Errorf("Failed to get interview slot: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to get interview slot", err)
	}

	if slot.CoursePhaseID != coursePhaseID {
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusNotFound, "Interview slot not found", nil)
	}

	count, err := qtx.CountAssignmentsBySlot(ctx, req.InterviewSlotID)
	if err != nil {
		log.Errorf("Failed to count assignments: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to check slot availability", err)
	}
	if count >= int64(slot.Capacity) {
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusConflict, "Interview slot is full", nil)
	}

	existingAssignment, err := qtx.GetInterviewAssignmentByParticipation(ctx, db.GetInterviewAssignmentByParticipationParams{
		CourseParticipationID: req.CourseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("Failed to check existing assignment: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to check existing assignment", err)
	}
	if err == nil && existingAssignment.ID != uuid.Nil {
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusConflict, "Student already has an interview slot assigned", nil)
	}

	assignment, err := qtx.CreateInterviewAssignment(ctx, db.CreateInterviewAssignmentParams{
		InterviewSlotID:       req.InterviewSlotID,
		CourseParticipationID: req.CourseParticipationID,
	})
	if err != nil {
		log.Errorf("Failed to create interview assignment: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to create interview assignment", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Errorf("Failed to commit transaction: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to create assignment", err)
	}

	return interviewAssignmentDTO.InterviewAssignmentResponse{
		ID:                    assignment.ID,
		InterviewSlotID:       assignment.InterviewSlotID,
		CourseParticipationID: assignment.CourseParticipationID,
		AssignedAt:            pgTimestamptzToTime(assignment.AssignedAt),
	}, nil
}

func GetMyInterviewAssignment(ctx context.Context, coursePhaseID uuid.UUID, participationUUID uuid.UUID) (interviewAssignmentDTO.InterviewAssignmentResponse, error) {
	if participationUUID == uuid.Nil {
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusBadRequest, "Invalid course participation ID", nil)
	}

	assignment, err := InterviewAssignmentServiceSingleton.queries.GetInterviewAssignmentByParticipation(ctx, db.GetInterviewAssignmentByParticipationParams{
		CourseParticipationID: participationUUID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusNotFound, "No interview assignment found", err)
		}
		log.Errorf("Failed to get interview assignment: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to get interview assignment", err)
	}

	slot, err := InterviewAssignmentServiceSingleton.queries.GetInterviewSlot(ctx, assignment.InterviewSlotID)
	if err != nil {
		log.Errorf("Failed to get slot details: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to get slot details", err)
	}

	count, err := InterviewAssignmentServiceSingleton.queries.CountAssignmentsBySlot(ctx, slot.ID)
	if err != nil {
		log.Errorf("Failed to count assignments for slot: %v", err)
		return interviewAssignmentDTO.InterviewAssignmentResponse{}, newServiceError(http.StatusInternalServerError, "Failed to count assignments", err)
	}

	return interviewAssignmentDTO.InterviewAssignmentResponse{
		ID:                    assignment.ID,
		InterviewSlotID:       assignment.InterviewSlotID,
		CourseParticipationID: assignment.CourseParticipationID,
		AssignedAt:            pgTimestamptzToTime(assignment.AssignedAt),
		SlotDetails: &interviewSlotDTO.InterviewSlotResponse{
			ID:            slot.ID,
			CoursePhaseID: slot.CoursePhaseID,
			StartTime:     pgTimestamptzToTime(slot.StartTime),
			EndTime:       pgTimestamptzToTime(slot.EndTime),
			Location:      pgTextToStringPtr(slot.Location),
			Capacity:      slot.Capacity,
			AssignedCount: count,
			CreatedAt:     pgTimestamptzToTime(slot.CreatedAt),
			UpdatedAt:     pgTimestamptzToTime(slot.UpdatedAt),
		},
	}, nil
}

func DeleteInterviewAssignment(ctx context.Context, coursePhaseID uuid.UUID, assignmentID uuid.UUID, selfParticipationID uuid.UUID, isPrivileged bool) error {
	assignment, err := InterviewAssignmentServiceSingleton.queries.GetInterviewAssignment(ctx, assignmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newServiceError(http.StatusNotFound, "Assignment not found", err)
		}
		log.Errorf("Failed to get interview assignment: %v", err)
		return newServiceError(http.StatusInternalServerError, "Failed to get interview assignment", err)
	}

	slot, err := InterviewAssignmentServiceSingleton.queries.GetInterviewSlot(ctx, assignment.InterviewSlotID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newServiceError(http.StatusNotFound, "Slot not found", err)
		}
		log.Errorf("Failed to get interview slot: %v", err)
		return newServiceError(http.StatusInternalServerError, "Failed to get interview slot", err)
	}

	if slot.CoursePhaseID != coursePhaseID {
		return newServiceError(http.StatusForbidden, "Assignment does not belong to the specified course phase", nil)
	}

	if !isPrivileged {
		if selfParticipationID == uuid.Nil {
			return newServiceError(http.StatusUnauthorized, "Course participation ID not found", nil)
		}

		if assignment.CourseParticipationID != selfParticipationID {
			return newServiceError(http.StatusForbidden, "Cannot delete another user's assignment", nil)
		}
	}

	if err := InterviewAssignmentServiceSingleton.queries.DeleteInterviewAssignment(ctx, assignmentID); err != nil {
		log.Errorf("Failed to delete interview assignment: %v", err)
		return newServiceError(http.StatusInternalServerError, "Failed to delete interview assignment", err)
	}

	return nil
}
