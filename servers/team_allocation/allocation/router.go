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
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation/allocationDTO"
	log "github.com/sirupsen/logrus"
)

const maxAllocationBodyBytes = 4 << 10

func RegisterRoutes(routerGroup *gin.RouterGroup, service *AllocationService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	allocationRouter := routerGroup.Group("/allocation")
	scopingMW := tutorscope.Middleware(tutorscope.NewPgxResolver(service.conn))

	allocationRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, service.getAllAllocations)
	allocationRouter.GET("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, service.getAllocationByCourseParticipationID)
	allocationRouter.PUT("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), scopingMW, service.updateAllocation)
	allocationRouter.DELETE("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), scopingMW, service.deleteAllocation)
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
		handleError(c, http.StatusBadRequest, err)
		return
	}

	allocations, err := s.GetAllAllocations(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
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
		handleError(c, http.StatusBadRequest, err)
		return
	}

	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	teamID, err := s.GetAllocationByCourseParticipationID(c, courseParticipationID, coursePhaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			handleError(c, http.StatusNotFound, err)
		} else {
			handleError(c, http.StatusInternalServerError, err)
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
// @Description Assign a course participation to a team. Lecturers and admins may write any team; a tutor may only write the team they are assigned to, and only for a participant who is unallocated or already in that team.
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
		handleError(c, http.StatusBadRequest, errors.New("teamID is required"))
		return
	}

	access, allowed := authorizeWrite(c)
	if !allowed {
		return
	}
	if !access.AllowsTeam(request.TeamID) {
		denyAllocationWrite(c)
		return
	}

	err := s.UpsertAllocation(c, c.GetHeader("Authorization"), coursePhaseID, courseParticipationID, request.TeamID, access.Guard())
	switch {
	case err == nil:
		c.JSON(http.StatusOK, allocationDTO.Allocation{TeamAllocation: request.TeamID})
	case errors.Is(err, tutorscope.ErrWriteDenied):
		denyAllocationWrite(c)
	case errors.Is(err, ErrParticipantNotInPhase), errors.Is(err, ErrInvalidTeamForPhase):
		handleError(c, http.StatusBadRequest, err)
	case errors.Is(err, ErrParticipantLookup):
		handleError(c, http.StatusBadGateway, err)
	default:
		handleError(c, http.StatusInternalServerError, err)
	}
}

// deleteAllocation godoc
// @Summary Remove a participant's team allocation
// @Description Remove the team allocation of a course participation. Lecturers and admins may remove any; a tutor may only remove one from the team they are assigned to.
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

	access, allowed := authorizeWrite(c)
	if !allowed {
		return
	}

	err := s.DeleteAllocation(c, coursePhaseID, courseParticipationID, access.Guard())
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, ErrAllocationNotFound):
		if s.isForeignTeamAllocation(c, coursePhaseID, courseParticipationID, access) {
			denyAllocationWrite(c)
			return
		}
		handleError(c, http.StatusNotFound, err)
	default:
		handleError(c, http.StatusInternalServerError, err)
	}
}

func parseAllocationParams(c *gin.Context) (coursePhaseID, courseParticipationID uuid.UUID, ok bool) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil || coursePhaseID == uuid.Nil {
		handleError(c, http.StatusBadRequest, errors.New("invalid course phase id"))
		return uuid.Nil, uuid.Nil, false
	}

	courseParticipationID, err = uuid.Parse(c.Param("courseParticipationID"))
	if err != nil || courseParticipationID == uuid.Nil {
		handleError(c, http.StatusBadRequest, errors.New("invalid course participation id"))
		return uuid.Nil, uuid.Nil, false
	}

	return coursePhaseID, courseParticipationID, true
}

// authorizeWrite answers the request itself when the write is refused.
func authorizeWrite(c *gin.Context) (tutorscope.Access, bool) {
	access, err := tutorscope.AuthorizeWrite(c)
	switch {
	case err == nil:
		return access, true
	case errors.Is(err, tutorscope.ErrNotAuthenticated):
		handleError(c, http.StatusUnauthorized, err)
	case errors.Is(err, tutorscope.ErrWriteDenied):
		denyAllocationWrite(c)
	default:
		handleError(c, http.StatusInternalServerError, err)
	}
	return tutorscope.Access{}, false
}

// isForeignTeamAllocation classifies a delete that affected no rows. It only picks
// the status code, it never grants access: the scoped delete has already not happened.
func (s *AllocationService) isForeignTeamAllocation(c *gin.Context, coursePhaseID, courseParticipationID uuid.UUID, access tutorscope.Access) bool {
	if !access.Confined {
		return false
	}
	_, err := s.GetAllocationByCourseParticipationID(c, courseParticipationID, coursePhaseID)
	return err == nil
}

func denyAllocationWrite(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": tutorscope.ErrWriteDenied.Error()})
}

func bindAllocationJSON(c *gin.Context, target any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAllocationBodyBytes)
	if err := c.ShouldBindJSON(target); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			handleError(c, http.StatusRequestEntityTooLarge, fmt.Errorf("request body exceeds %d bytes", maxAllocationBodyBytes))
			return false
		}
		if errors.Is(err, io.EOF) {
			handleError(c, http.StatusBadRequest, errors.New("request body is required"))
			return false
		}
		handleError(c, http.StatusBadRequest, err)
		return false
	}
	return true
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
