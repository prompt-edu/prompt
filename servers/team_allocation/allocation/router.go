package allocation

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/tutorscope"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation/allocationDTO"
)

const maxAllocationBodyBytes = 4 << 10

func RegisterRoutes(routerGroup *gin.RouterGroup, service *AllocationService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	allocationRouter := routerGroup.Group("/allocation")
	scopingMW := tutorscope.Middleware(tutorscope.NewPgxResolver(service.conn))

	allocationRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, service.getAllAllocations)
	allocationRouter.GET("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, service.getAllocationByCourseParticipationID)
	allocationRouter.PUT("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.updateAllocation)
	allocationRouter.DELETE("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.deleteAllocation)
}

// getAllAllocations godoc
// @Summary Get all allocations
// @Description Get all team allocations for a course phase
// @Tags allocation
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} allocationDTO.AllocationWithParticipation
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/allocation [get]
func (s *AllocationService) getAllAllocations(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	allocations, err := s.GetAllAllocations(c, coursePhaseID)
	if err != nil {
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	if tutorTeamID, scoped := tutorscope.TeamID(c); scoped {
		allocations = filterAllocationsByTeam(allocations, tutorTeamID)
	}

	c.JSON(http.StatusOK, allocations)
}

// getAllocationByCourseParticipationID godoc
// @Summary Get allocation by course participation ID
// @Description Get the team allocation for a specific course participation
// @Tags allocation
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Success 200 {object} allocationDTO.Allocation
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/allocation/{courseParticipationID} [get]
func (s *AllocationService) getAllocationByCourseParticipationID(c *gin.Context) {
	courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	teamID, err := s.GetAllocationByCourseParticipationID(c, courseParticipationID, coursePhaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			sdkUtils.HandleError(c, http.StatusNotFound, err)
		} else {
			sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		}
		return
	}

	if tutorTeamID, scoped := tutorscope.TeamID(c); scoped && teamID != tutorTeamID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access restricted to assigned team"})
		return
	}

	c.JSON(http.StatusOK, allocationDTO.Allocation{TeamAllocation: teamID})
}

func filterAllocationsByTeam(allocations []allocationDTO.AllocationWithParticipation, teamID uuid.UUID) []allocationDTO.AllocationWithParticipation {
	result := make([]allocationDTO.AllocationWithParticipation, 0)
	for _, a := range allocations {
		if a.TeamAllocation == teamID {
			result = append(result, a)
		}
	}
	return result
}

// updateAllocation godoc
// @Summary Assign a participant to a team
// @Description Assign a course participation to a team. Only lecturers and admins may change allocations; tutors have read access only.
// @Tags allocation
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Param request body allocationDTO.UpdateAllocationRequest true "Target team"
// @Success 200 {object} allocationDTO.Allocation
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 413 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 502 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/allocation/{courseParticipationID} [put]
func (s *AllocationService) updateAllocation(c *gin.Context) {
	coursePhaseID, courseParticipationID, ok := parseAllocationParams(c)
	if !ok {
		return
	}

	var request allocationDTO.UpdateAllocationRequest
	if !bindAllocationJSON(c, &request) {
		return
	}
	if request.TeamID == uuid.Nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, errors.New("teamID is required"))
		return
	}

	err := s.UpsertAllocation(c, c.GetHeader("Authorization"), coursePhaseID, courseParticipationID, request.TeamID)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, allocationDTO.Allocation{TeamAllocation: request.TeamID})
	case errors.Is(err, ErrParticipantNotInPhase), errors.Is(err, ErrInvalidTeamForPhase):
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
	case errors.Is(err, ErrParticipantLookup):
		sdkUtils.HandleError(c, http.StatusBadGateway, err)
	default:
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
	}
}

// deleteAllocation godoc
// @Summary Remove a participant's team allocation
// @Description Remove the team allocation of a course participation. Only lecturers and admins may change allocations; tutors have read access only.
// @Tags allocation
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/allocation/{courseParticipationID} [delete]
func (s *AllocationService) deleteAllocation(c *gin.Context) {
	coursePhaseID, courseParticipationID, ok := parseAllocationParams(c)
	if !ok {
		return
	}

	err := s.DeleteAllocation(c, coursePhaseID, courseParticipationID)
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, ErrAllocationNotFound):
		sdkUtils.HandleError(c, http.StatusNotFound, err)
	default:
		sdkUtils.HandleError(c, http.StatusInternalServerError, err)
	}
}

func parseAllocationParams(c *gin.Context) (coursePhaseID, courseParticipationID uuid.UUID, ok bool) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil || coursePhaseID == uuid.Nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, errors.New("invalid course phase id"))
		return uuid.Nil, uuid.Nil, false
	}

	courseParticipationID, err = uuid.Parse(c.Param("courseParticipationID"))
	if err != nil || courseParticipationID == uuid.Nil {
		sdkUtils.HandleError(c, http.StatusBadRequest, errors.New("invalid course participation id"))
		return uuid.Nil, uuid.Nil, false
	}

	return coursePhaseID, courseParticipationID, true
}

func bindAllocationJSON(c *gin.Context, target any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAllocationBodyBytes)
	if err := c.ShouldBindJSON(target); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			sdkUtils.HandleError(c, http.StatusRequestEntityTooLarge, fmt.Errorf("request body exceeds %d bytes", maxAllocationBodyBytes))
			return false
		}
		if errors.Is(err, io.EOF) {
			sdkUtils.HandleError(c, http.StatusBadRequest, errors.New("request body is required"))
			return false
		}
		sdkUtils.HandleError(c, http.StatusBadRequest, err)
		return false
	}
	return true
}
