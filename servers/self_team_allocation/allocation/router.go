package allocation

import (
	"errors"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/tutorscope"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/allocation/allocationDTO"
	teams "github.com/prompt-edu/prompt/servers/self_team_allocation/team"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *AllocationService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	allocationRouter := routerGroup.Group("/allocation")
	scopingMW := teams.TutorScopeMiddleware(service.conn)

	allocationRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, service.getAllAllocations)
	allocationRouter.GET("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, service.getAllocationByCourseParticipationID)
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

	if tutorTeamIDs, scoped := tutorscope.TeamIDs(c); scoped {
		allocations = filterAllocationsByTeams(allocations, tutorTeamIDs)
	}

	c.JSON(http.StatusOK, allocations)
}

func filterAllocationsByTeams(allocations []allocationDTO.AllocationWithParticipation, teamIDs []uuid.UUID) []allocationDTO.AllocationWithParticipation {
	result := make([]allocationDTO.AllocationWithParticipation, 0)
	for _, a := range allocations {
		if slices.Contains(teamIDs, a.TeamAllocation) {
			result = append(result, a)
		}
	}
	return result
}

// getAllocationByCourseParticipationID godoc
// @Summary Get allocation by course participation ID
// @Description Get the team allocation for a specific course participation
// @Tags allocation
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Success 200 {object} allocationDTO.AllocationWithParticipation
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

	allocation, err := s.GetAllocationByCourseParticipationID(c, courseParticipationID, coursePhaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			sdkUtils.HandleError(c, http.StatusNotFound, err)
		} else {
			sdkUtils.HandleError(c, http.StatusInternalServerError, err)
		}
		return
	}

	if tutorTeamIDs, scoped := tutorscope.TeamIDs(c); scoped && !slices.Contains(tutorTeamIDs, allocation) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access restricted to assigned teams"})
		return
	}

	c.JSON(http.StatusOK, allocation)
}
