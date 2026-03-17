package interviewSlot

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interviewSlotDTO "github.com/prompt-edu/prompt/servers/interview/interviewSlot/interviewSlotDTO"
	log "github.com/sirupsen/logrus"
)

type InterviewSlotService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var InterviewSlotServiceSingleton *InterviewSlotService

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

func timeToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func pgTimestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}

func stringPtrToPgText(s *string) pgtype.Text {
	if s != nil && *s != "" {
		return pgtype.Text{String: *s, Valid: true}
	}
	return pgtype.Text{Valid: false}
}

func pgTextToStringPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func CreateInterviewSlot(ctx context.Context, coursePhaseID uuid.UUID, req interviewSlotDTO.CreateInterviewSlotRequest) (db.InterviewSlot, error) {
	if !req.EndTime.After(req.StartTime) {
		return db.InterviewSlot{}, newServiceError(http.StatusBadRequest, "End time must be after start time", nil)
	}

	slot, err := InterviewSlotServiceSingleton.queries.CreateInterviewSlot(ctx, db.CreateInterviewSlotParams{
		CoursePhaseID: coursePhaseID,
		StartTime:     timeToPgTimestamptz(req.StartTime),
		EndTime:       timeToPgTimestamptz(req.EndTime),
		Location:      stringPtrToPgText(req.Location),
		Capacity:      req.Capacity,
	})
	if err != nil {
		log.Errorf("Failed to create interview slot: %v", err)
		return db.InterviewSlot{}, newServiceError(http.StatusInternalServerError, "Failed to create interview slot", err)
	}

	return slot, nil
}

func GetAllInterviewSlots(ctx context.Context, coursePhaseID uuid.UUID, authHeader string) ([]interviewSlotDTO.InterviewSlotResponse, error) {
	rows, err := InterviewSlotServiceSingleton.queries.GetInterviewSlotWithAssignments(ctx, coursePhaseID)
	if err != nil {
		log.Errorf("Failed to get interview slots: %v", err)
		return nil, newServiceError(http.StatusInternalServerError, "Failed to get interview slots", err)
	}

	studentMap, err := fetchAllStudentsForCoursePhase(ctx, coursePhaseID, authHeader)
	if err != nil {
		log.Warnf("Failed to fetch students for course phase (may be due to permissions): %v", err)
		studentMap = make(map[uuid.UUID]*interviewSlotDTO.StudentInfo)
	}

	slotMap := make(map[uuid.UUID]*interviewSlotDTO.InterviewSlotResponse)
	slotOrder := []uuid.UUID{}

	for _, row := range rows {
		if _, exists := slotMap[row.SlotID]; !exists {
			slotMap[row.SlotID] = &interviewSlotDTO.InterviewSlotResponse{
				ID:            row.SlotID,
				CoursePhaseID: row.CoursePhaseID,
				StartTime:     pgTimestamptzToTime(row.StartTime),
				EndTime:       pgTimestamptzToTime(row.EndTime),
				Location:      pgTextToStringPtr(row.Location),
				Capacity:      row.Capacity,
				Assignments:   []interviewSlotDTO.AssignmentInfo{},
				CreatedAt:     pgTimestamptzToTime(row.CreatedAt),
				UpdatedAt:     pgTimestamptzToTime(row.UpdatedAt),
			}
			slotOrder = append(slotOrder, row.SlotID)
		}

		if row.AssignmentID.Valid {
			assignmentUUID, err := uuid.FromBytes(row.AssignmentID.Bytes[:])
			if err != nil {
				log.Warnf("Failed to parse assignment UUID: %v", err)
				continue
			}
			participationUUID, err := uuid.FromBytes(row.CourseParticipationID.Bytes[:])
			if err != nil {
				log.Warnf("Failed to parse course participation UUID: %v", err)
				continue
			}

			assignment := interviewSlotDTO.AssignmentInfo{
				ID:                    assignmentUUID,
				CourseParticipationID: participationUUID,
				AssignedAt:            pgTimestamptzToTime(row.AssignedAt),
				Student:               studentMap[participationUUID],
			}
			slotMap[row.SlotID].Assignments = append(slotMap[row.SlotID].Assignments, assignment)
		}
	}

	response := make([]interviewSlotDTO.InterviewSlotResponse, 0, len(slotOrder))
	for _, slotID := range slotOrder {
		slot := slotMap[slotID]
		slot.AssignedCount = int64(len(slot.Assignments))
		response = append(response, *slot)
	}

	return response, nil
}

func GetInterviewSlot(ctx context.Context, coursePhaseID uuid.UUID, slotID uuid.UUID, authHeader string) (interviewSlotDTO.InterviewSlotResponse, error) {
	slot, err := InterviewSlotServiceSingleton.queries.GetInterviewSlot(ctx, slotID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return interviewSlotDTO.InterviewSlotResponse{}, newServiceError(http.StatusNotFound, "Interview slot not found", err)
		}
		log.Errorf("Failed to get interview slot: %v", err)
		return interviewSlotDTO.InterviewSlotResponse{}, newServiceError(http.StatusInternalServerError, "Failed to get interview slot", err)
	}

	if slot.CoursePhaseID != coursePhaseID {
		return interviewSlotDTO.InterviewSlotResponse{}, newServiceError(http.StatusNotFound, "Interview slot not found", nil)
	}

	assignments, err := InterviewSlotServiceSingleton.queries.GetInterviewAssignmentsBySlot(ctx, slotID)
	if err != nil {
		log.Errorf("Failed to get assignments for slot %s: %v", slotID, err)
		return interviewSlotDTO.InterviewSlotResponse{}, newServiceError(http.StatusInternalServerError, "Failed to get assignments", err)
	}

	studentMap, err := fetchAllStudentsForCoursePhase(ctx, coursePhaseID, authHeader)
	if err != nil {
		log.Warnf("Failed to fetch students for course phase (may be due to permissions): %v", err)
		studentMap = make(map[uuid.UUID]*interviewSlotDTO.StudentInfo)
	}

	assignmentInfos := make([]interviewSlotDTO.AssignmentInfo, len(assignments))
	for i, assignment := range assignments {
		assignmentInfos[i] = interviewSlotDTO.AssignmentInfo{
			ID:                    assignment.ID,
			CourseParticipationID: assignment.CourseParticipationID,
			AssignedAt:            pgTimestamptzToTime(assignment.AssignedAt),
			Student:               studentMap[assignment.CourseParticipationID],
		}
	}

	response := interviewSlotDTO.InterviewSlotResponse{
		ID:            slot.ID,
		CoursePhaseID: slot.CoursePhaseID,
		StartTime:     pgTimestamptzToTime(slot.StartTime),
		EndTime:       pgTimestamptzToTime(slot.EndTime),
		Location:      pgTextToStringPtr(slot.Location),
		Capacity:      slot.Capacity,
		AssignedCount: int64(len(assignments)),
		Assignments:   assignmentInfos,
		CreatedAt:     pgTimestamptzToTime(slot.CreatedAt),
		UpdatedAt:     pgTimestamptzToTime(slot.UpdatedAt),
	}

	return response, nil
}

func UpdateInterviewSlot(ctx context.Context, coursePhaseID uuid.UUID, slotID uuid.UUID, req interviewSlotDTO.UpdateInterviewSlotRequest) (db.InterviewSlot, error) {
	if !req.EndTime.After(req.StartTime) {
		return db.InterviewSlot{}, newServiceError(http.StatusBadRequest, "End time must be after start time", nil)
	}

	existingSlot, err := InterviewSlotServiceSingleton.queries.GetInterviewSlot(ctx, slotID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.InterviewSlot{}, newServiceError(http.StatusNotFound, "Interview slot not found", err)
		}
		log.Errorf("Failed to get interview slot: %v", err)
		return db.InterviewSlot{}, newServiceError(http.StatusInternalServerError, "Failed to get interview slot", err)
	}

	if existingSlot.CoursePhaseID != coursePhaseID {
		return db.InterviewSlot{}, newServiceError(http.StatusNotFound, "Interview slot not found", nil)
	}

	currentAssignedCount, err := InterviewSlotServiceSingleton.queries.CountAssignmentsBySlot(ctx, slotID)
	if err != nil {
		log.Errorf("Failed to count assignments for slot: %v", err)
		return db.InterviewSlot{}, newServiceError(http.StatusInternalServerError, "Failed to validate capacity", err)
	}

	if req.Capacity < int32(currentAssignedCount) {
		return db.InterviewSlot{}, newServiceError(
			http.StatusBadRequest,
			fmt.Sprintf("Cannot reduce capacity to %d: slot has %d existing assignments", req.Capacity, currentAssignedCount),
			nil,
		)
	}

	slot, err := InterviewSlotServiceSingleton.queries.UpdateInterviewSlot(ctx, db.UpdateInterviewSlotParams{
		ID:        slotID,
		StartTime: timeToPgTimestamptz(req.StartTime),
		EndTime:   timeToPgTimestamptz(req.EndTime),
		Location:  stringPtrToPgText(req.Location),
		Capacity:  req.Capacity,
	})
	if err != nil {
		log.Errorf("Failed to update interview slot: %v", err)
		return db.InterviewSlot{}, newServiceError(http.StatusInternalServerError, "Failed to update interview slot", err)
	}

	return slot, nil
}

func DeleteInterviewSlot(ctx context.Context, coursePhaseID uuid.UUID, slotID uuid.UUID) error {
	existingSlot, err := InterviewSlotServiceSingleton.queries.GetInterviewSlot(ctx, slotID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newServiceError(http.StatusNotFound, "Interview slot not found", err)
		}
		log.Errorf("Failed to get interview slot: %v", err)
		return newServiceError(http.StatusInternalServerError, "Failed to get interview slot", err)
	}

	if existingSlot.CoursePhaseID != coursePhaseID {
		return newServiceError(http.StatusNotFound, "Interview slot not found", nil)
	}

	if err := InterviewSlotServiceSingleton.queries.DeleteInterviewSlot(ctx, slotID); err != nil {
		log.Errorf("Failed to delete interview slot: %v", err)
		return newServiceError(http.StatusInternalServerError, "Failed to delete interview slot", err)
	}

	return nil
}

func fetchAllStudentsForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID, authHeader string) (map[uuid.UUID]*interviewSlotDTO.StudentInfo, error) {
	coreURL := sdkUtils.GetCoreUrl()
	url := fmt.Sprintf("%s/api/course_phases/%s/participations", coreURL, coursePhaseID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for course phase participations: %w", err)
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch course phase participations: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core service returned status %d for course phase participations", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read course phase participations response: %w", err)
	}

	var participationsResponse struct {
		Participations []struct {
			CourseParticipationID uuid.UUID                    `json:"courseParticipationID"`
			Student               interviewSlotDTO.StudentInfo `json:"student"`
		} `json:"participations"`
	}

	if err := json.Unmarshal(body, &participationsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal course phase participations: %w", err)
	}

	studentMap := make(map[uuid.UUID]*interviewSlotDTO.StudentInfo)
	for _, participation := range participationsResponse.Participations {
		studentCopy := participation.Student
		studentMap[participation.CourseParticipationID] = &studentCopy
	}

	return studentMap, nil
}

