package privacy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	"github.com/prompt-edu/prompt/servers/core/privacy/service"
	coreutils "github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// @Summary Privacy Endpoints
// @Description Endpoints for managing GDPR data deletion requests
// @Tags privacy
// @Security BearerAuth
func setupPrivacyDeletionRouter(privacyRouter *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	privacyRouter.POST("/data-deletion", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), createNewSubjectDataDeletionRequest)
	privacyRouter.GET("/data-deletion", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), getLatestDeletionRequest)
	privacyRouter.GET("/data-deletion/:uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), getDeletionRequest)

	privacyRouter.GET("/admin/data-deletions", permissionRoleMiddleware(permissionValidation.PromptAdmin), getAllDeletionRequests)
	privacyRouter.POST("/admin/data-deletions/:uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin), decideDeletionRequest)

	// Admin-initiated deletion
	privacyRouter.POST("/admin/data-deletions", permissionRoleMiddleware(permissionValidation.PromptAdmin), adminInitiateDeletionRequests)
	privacyRouter.POST("/admin/data-deletions/status", permissionRoleMiddleware(permissionValidation.PromptAdmin), adminInitiatedDeletionsStatus)
}

// @Summary Submit a new data deletion request
// @Description Creates a deletion request in pending_approval state. Execution only starts after an admin approves.
// @Tags privacy
// @Produce json
// @Success 200 {object} privacyDTO.PrivacyDeletionRequest
// @Failure 400 {object} coreutils.ErrorResponse
// @Failure 409 {object} coreutils.ErrorResponse "An open request already exists for this user"
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/data-deletion [post]
func createNewSubjectDataDeletionRequest(c *gin.Context) {
	if valErr := service.ValidateUserMayCreateDeletionRequest(c); valErr != nil {
		handleError(c, http.StatusConflict, valErr)
		return
	}

	record, err := service.CreateDeletionRequest(c)
	if err != nil {
		log.Error("data deletion request creation failed: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, record)
}

// @Summary Get the current user's latest deletion request
// @Description Returns {status: "exists", request: ...} when a request exists, or 204 when none exists.
// @Tags privacy
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Success 204 "No deletion request on file"
// @Failure 400 {object} coreutils.ErrorResponse
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/data-deletion [get]
func getLatestDeletionRequest(c *gin.Context) {
	request, err := service.GetLatestDeletionRequestForUser(c)
	if err != nil {
		log.Error("get latest deletion request failed: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	if request == nil {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "exists", "request": request})
}

// @Summary Get a deletion request by id
// @Description Returns the deletion request together with its subrequests. Used by the client to poll for status while execution is in progress.
// @Tags privacy
// @Produce json
// @Param uuid path string true "Deletion Request UUID"
// @Success 200 {object} privacyDTO.PrivacyDeletionRequest
// @Failure 400 {object} coreutils.ErrorResponse
// @Failure 403 {object} coreutils.ErrorResponse
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/data-deletion/{uuid} [get]
func getDeletionRequest(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if valErr := service.ValidateDeletionRequestBelongsToCaller(c, requestID); valErr != nil {
		handleError(c, http.StatusForbidden, valErr)
		return
	}

	record, err := service.GetDeletionRequestWithSubrequests(c, requestID)
	if err != nil {
		log.Error("get deletion request failed: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, record)
}

// @Summary List all deletion requests (admin only)
// @Tags privacy
// @Produce json
// @Success 200 {array} privacyDTO.AdminPrivacyDeletionRequest
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/admin/data-deletions [get]
func getAllDeletionRequests(c *gin.Context) {
	records, err := service.GetAllDeletionRequests(c)
	if err != nil {
		log.Error("get all deletion requests failed: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, records)
}

// @Summary Decide on a deletion request (admin only)
// @Description Records the auditor decision on a pending_approval request. Approval starts the deletion fan-out; rejection is a database-only update.
// @Tags privacy
// @Accept json
// @Produce json
// @Param uuid path string true "Deletion Request UUID"
// @Param body body privacyDTO.AuditorDecisionRequest true "Auditor decision and optional note"
// @Success 200 {object} privacyDTO.PrivacyDeletionRequest
// @Failure 400 {object} coreutils.ErrorResponse "Invalid payload"
// @Failure 409 {object} coreutils.ErrorResponse "Request is no longer in pending_approval state"
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/admin/data-deletions/{uuid} [post]
func decideDeletionRequest(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var decision privacyDTO.AuditorDecisionRequest
	if err := c.ShouldBindJSON(&decision); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if valErr := service.ValidateDeletionRequestPending(c, requestID); valErr != nil {
		handleError(c, http.StatusConflict, valErr)
		return
	}

	switch decision.Decision {
	case privacyDTO.AuditorDecisionReject:
		record, err := service.RejectDeletionRequest(c, requestID, decision.Note)
		if err != nil {
			log.Error("deletion request rejection failed: ", err)
			handleError(c, http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, record)

	case privacyDTO.AuditorDecisionApprove:
		record, err := service.AcceptDeletionRequest(c, requestID, decision.Note)
		if err != nil {
			log.Error("deletion request approval failed: ", err)
			handleError(c, http.StatusInternalServerError, err)
			return
		}
		state, err := service.PrepareDataDeletion(c, record)
		if err != nil {
			log.Error("deletion preparation failed: ", err)
			handleError(c, http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, record)

		authHeader := c.GetHeader("Authorization")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), service.DeletionRunTimeout)
			defer cancel()
			service.RunDataDeletion(ctx, authHeader, state)
		}()

	default:
		handleError(c, http.StatusBadRequest, fmt.Errorf("unknown decision: %q", decision.Decision))
	}
}

// Admin-initiated deletion

// @Summary Admin-initiate one or more deletion requests
// @Description Creates one privacy_deletion_request row per student_id, all already in
// @Description 'in_progress' state with the calling admin recorded as the auditor.
// @Description Returns immediately; the actual fan-out runs in the background.
// @Description Use POST /admin/data-deletions/status to poll the batch to completion.
// @Tags privacy
// @Accept json
// @Produce json
// @Param body body privacyDTO.AdminInitiateDeletionBody true "List of student IDs to delete"
// @Success 200 {array} privacyDTO.PrivacyDeletionRequest
// @Failure 400 {object} coreutils.ErrorResponse "Invalid payload or unknown student_id"
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/admin/data-deletions [post]
func adminInitiateDeletionRequests(c *gin.Context) {
	var body privacyDTO.AdminInitiateDeletionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := service.ValidateStudentsExist(c, body.StudentIDs); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	records, err := service.CreateAdminInitiatedDeletionRequests(c, body.StudentIDs)
	if err != nil {
		log.Error("admin-initiated deletion creation failed: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, records)

	authHeader := c.GetHeader("Authorization")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), service.DeletionRunTimeout)
		defer cancel()
		service.RunAdminInitiatedDeletions(ctx, authHeader, records)
	}()
}

// @Summary Fetch the current status of a set of deletion requests
// @Description Targeted polling endpoint used by the bulk-deletion dialog to track when
// @Description each record reaches a terminal state. Returns only the records whose IDs
// @Description are in the request body.
// @Tags privacy
// @Accept json
// @Produce json
// @Param body body privacyDTO.DeletionStatusBody true "List of deletion request IDs"
// @Success 200 {array} privacyDTO.PrivacyDeletionRequest
// @Failure 400 {object} coreutils.ErrorResponse
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/admin/data-deletions/status [post]
func adminInitiatedDeletionsStatus(c *gin.Context) {
	var body privacyDTO.DeletionStatusBody
	if err := c.ShouldBindJSON(&body); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	records, err := service.GetDeletionRequestsByIDs(c, body.IDs)
	if err != nil {
		log.Error("deletion status fetch failed: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, records)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, coreutils.ErrorResponse{
		Error: err.Error(),
	})
}
