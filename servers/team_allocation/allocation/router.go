package allocation

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation/allocationDTO"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/tutorscope"
	log "github.com/sirupsen/logrus"
)

const maxAllocationBodyBytes = 4 << 10

func setupAllocationRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc, queries db.Queries) {
	allocationRouter := routerGroup.Group("/allocation")
	scopingMW := promptSDK.TutorScopingMiddleware(tutorscope.NewResolver(queries))

	allocationRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, getAllAllocations)
	allocationRouter.GET("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), scopingMW, getAllocationByCourseParticipationID)
	allocationRouter.PUT("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), scopingMW, updateAllocation)
	allocationRouter.DELETE("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), scopingMW, deleteAllocation)
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
func updateAllocation(c *gin.Context) {
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

	expectedTeamID, allowed := authorizeAllocationWrite(c)
	if !allowed {
		return
	}
	if expectedTeamID.Valid && request.TeamID != uuid.UUID(expectedTeamID.Bytes) {
		denyAllocationWrite(c)
		return
	}

	err := UpsertAllocation(c, c.GetHeader("Authorization"), coursePhaseID, courseParticipationID, request.TeamID, expectedTeamID)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, allocationDTO.Allocation{TeamAllocation: request.TeamID})
	case errors.Is(err, ErrTeamWriteDenied):
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
func deleteAllocation(c *gin.Context) {
	coursePhaseID, courseParticipationID, ok := parseAllocationParams(c)
	if !ok {
		return
	}

	expectedTeamID, allowed := authorizeAllocationWrite(c)
	if !allowed {
		return
	}

	err := DeleteAllocation(c, coursePhaseID, courseParticipationID, expectedTeamID)
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, ErrAllocationNotFound):
		if isForeignTeamAllocation(c, coursePhaseID, courseParticipationID, expectedTeamID) {
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

// authorizeAllocationWrite reports whether the requester may write allocations and,
// for a tutor, which source team the write is confined to. Reads deliberately fail
// open for editors the scoping middleware cannot resolve; writes fail closed.
// Authorization is decided from the team resolved at the start of the request.
// It answers the request itself when the write is refused.
func authorizeAllocationWrite(c *gin.Context) (pgtype.UUID, bool) {
	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if !ok {
		handleError(c, http.StatusUnauthorized, keycloakTokenVerifier.ErrUserNotInContext)
		return pgtype.UUID{}, false
	}

	// PromptLecturer is deliberately absent: the routes do not admit it directly, so
	// such a user arrives as a course editor and their reads are tutor-scoped. Writes
	// must be scoped with them.
	if tokenUser.Roles[promptSDK.PromptAdmin] || tokenUser.IsLecturer {
		return pgtype.UUID{}, true
	}

	if !tokenUser.IsEditor {
		denyAllocationWrite(c)
		return pgtype.UUID{}, false
	}

	tutorTeamID, scoped := promptSDK.GetTutorTeamID(c)
	if !scoped {
		denyAllocationWrite(c)
		return pgtype.UUID{}, false
	}

	return pgtype.UUID{Bytes: tutorTeamID, Valid: true}, true
}

// isForeignTeamAllocation classifies a delete that affected no rows. It only picks
// the status code, it never grants access: the scoped delete has already not happened.
func isForeignTeamAllocation(c *gin.Context, coursePhaseID, courseParticipationID uuid.UUID, expectedTeamID pgtype.UUID) bool {
	if !expectedTeamID.Valid {
		return false
	}
	_, err := GetAllocationByCourseParticipationID(c, courseParticipationID, coursePhaseID)
	return err == nil
}

func denyAllocationWrite(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": ErrTeamWriteDenied.Error()})
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
