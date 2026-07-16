package allocation

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation/allocationDTO"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/tutorscope"
	log "github.com/sirupsen/logrus"
)

func setupAllocationRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc, queries db.Queries) {
	allocationRouter := routerGroup.Group("/allocation")
	scopingMW := promptSDK.TutorScopingMiddleware(tutorscope.NewResolver(queries))

	allocationRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, getAllAllocations)
	allocationRouter.GET("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, getAllocationByCourseParticipationID)
}

// getAllAllocations godoc
// @Summary Get all allocations
// @Description Get all team allocations for a course phase
// @Tags allocation
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} allocationDTO.AllocationWithParticipation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/allocation [get]
func getAllAllocations(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	allocations, err := GetAllAllocations(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	if tutorTeamID, scoped := promptSDK.GetTutorTeamID(c); scoped {
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
func getAllocationByCourseParticipationID(c *gin.Context) {
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

	teamID, err := GetAllocationByCourseParticipationID(c, courseParticipationID, coursePhaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			handleError(c, http.StatusNotFound, err)
		} else {
			handleError(c, http.StatusInternalServerError, err)
		}
		return
	}

	if tutorTeamID, scoped := promptSDK.GetTutorTeamID(c); scoped && teamID != tutorTeamID {
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

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
