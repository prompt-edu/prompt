package privacy

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/privacy/service"
	coreutils "github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// setupPrivacyRouter sets up the privacy endpoints
// @Summary Privacy Endpoints
// @Description Endpoints for managing GDPR data exports and deletion requests
// @Tags privacy
// @Security BearerAuth
func setupPrivacyRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	privacyRouter := router.Group("/privacy", authMiddleware())

	privacyRouter.POST("/data-export", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), studentDataExport)
	privacyRouter.GET("/data-export", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), getLatestExport)

	privacyRouter.GET("/data-export/:uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), getExport)
	privacyRouter.GET("/data-export/:uuid/docs/:docID/download-url", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), getExportDocDownloadURL)

	// Admin-only routes
	privacyRouter.GET("/admin/data-exports", permissionRoleMiddleware(permissionValidation.PromptAdmin), getAllExports)
}

// studentDataExport exports all student related data from core and all microservices.
//
// @Summary Export all student related data
// @Description Get an export of all student data from core and all microservices
// @Tags privacy
// @Produce json
// @Success 200 {object} privacyDTO.PrivacyExport
// @Failure 400 {object} coreutils.ErrorResponse
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/data-export [post]
func studentDataExport(c *gin.Context) {
	subjectIdentifiers, errSI := service.GetSubjectIdentifiers(c)
	if errSI != nil {
    utils.HandleError(c, http.StatusBadRequest, errSI);
		return
	}

  prep, err := service.PrepareStudentDataExport(c, subjectIdentifiers)
	if err != nil {
    log.Error("student data export failed: ", err)
    utils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, prep.Record)

  service.RunStudentDataExport(c, prep, subjectIdentifiers)
}

// getLatestExport returns the most recent export for the requesting user if one exists
// or 204 No Content if none.
//
// @Summary Get the current user's latest export
// @Description Returns {status: "exists", export: ...} when a valid export exists, {status: "rate_limited", retry_after: ...} when rate limited, or 204 when no export exists
// @Tags privacy
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Success 204 "No recent export"
// @Failure 400 {object} coreutils.ErrorResponse
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/data-export [get]
func getLatestExport(c *gin.Context) {
	userID, err := coreutils.GetUserUUIDFromContext(c)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	availability, exp, err := service.GetExportAvailability(c, userID)
	if err != nil {
		log.Error("get latest export failed: ", err)
		utils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	switch availability {
	case service.ExportExistsAndValid:
		c.JSON(http.StatusOK, gin.H{"status": "exists", "export": exp})
	case service.ExportRateLimited:
		c.JSON(http.StatusOK, gin.H{"status": "rate_limited", "retry_after": service.RateLimitEndForExport(*exp)})
	default:
		c.Status(http.StatusNoContent)
	}
}

// getExportDocDownloadURL returns a presigned download URL for a single document belonging to an export.
// It validates that the export belongs to the requester, that the document belongs to the export,
// and that the export is still valid before generating the URL.
//
// @Summary Get download URL for an export document
// @Description Returns a short-lived presigned URL to download a specific export document after validating ownership and validity
// @Tags privacy
// @Produce json
// @Param uuid path string true "Export UUID"
// @Param docID path string true "Export Document UUID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} coreutils.ErrorResponse
// @Failure 403 {object} coreutils.ErrorResponse
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/data-export/{uuid}/docs/{docID}/download-url [get]
func getExportDocDownloadURL(c *gin.Context) {
	exportID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	docID, err := uuid.Parse(c.Param("docID"))
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if valErr := service.ValidateExportBelongsToRequester(c, exportID); valErr != nil {
		utils.HandleError(c, http.StatusForbidden, valErr)
		return
	}

  if valErr := service.ValidateExportDocBelongsToExport(c, docID, exportID); valErr != nil {
		utils.HandleError(c, http.StatusForbidden, valErr)
		return
  }

  if valErr := service.ValidateExportValid(c, exportID); valErr != nil {
		utils.HandleError(c, http.StatusForbidden, valErr)
		return
  }

	downloadURL, err := service.GetDownloadURLForDoc(c, docID)
	if err != nil {
		log.Error("get export doc download URL failed: ", err)
		utils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"downloadUrl": downloadURL})
}

// getAllExports returns all export records for admin monitoring (no documents included).
//
// @Summary List all data exports (admin only)
// @Tags privacy
// @Produce json
// @Success 200 {array} privacyDTO.AdminPrivacyExport
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/admin/data-exports [get]
func getAllExports(c *gin.Context) {
	exports, err := service.GetAllExports(c)
	if err != nil {
		log.Error("get all exports failed: ", err)
		utils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, exports)
}

// getExport returns export status and ids for documents related to an export
//
// @Summary Get all Ids and status of export and export docs
// @Description Get the current status and all ids of the export record and all export docs. needed so client knows what docs exist to be able to download them
// @Tags privacy
// @Produce json
// @Param uuid path string true "Export UUID"
// @Success 200 {object} privacyDTO.PrivacyExport
// @Failure 400 {object} coreutils.ErrorResponse
// @Failure 405 {object} coreutils.ErrorResponse
// @Failure 500 {object} coreutils.ErrorResponse
// @Security BearerAuth
// @Router /privacy/data-export/{uuid} [get]
func getExport(c *gin.Context) {
	exportID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err)
		return
	}

  valErr := service.ValidateExportBelongsToRequester(c, exportID)
	if valErr != nil {
		utils.HandleError(c, http.StatusMethodNotAllowed, valErr)
		return
	}

  if valErr := service.ValidateExportValid(c, exportID); valErr != nil {
		utils.HandleError(c, http.StatusMethodNotAllowed, valErr)
		return
  }

  expWithDocs, expErr := service.GetExportWithDocs(c, exportID)
	if expErr != nil {
		utils.HandleError(c, http.StatusInternalServerError, expErr)
		return
	}

  c.JSON(200, expWithDocs)
}
