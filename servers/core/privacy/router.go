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

func setupPrivacyRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, premissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	privacyRouter := router.Group("/privacy", authMiddleware())

	privacyRouter.POST("/data-export", premissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), studentDataExport)
	privacyRouter.GET("/data-export", premissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), getLatestExport)

	privacyRouter.GET("/data-export/:uuid", premissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), getExport)
	privacyRouter.GET("/data-export/:uuid/docs/:docID/download-url", premissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), getExportDocDownloadURL)
}

// studentDataExport exports all student related data from core and all microservices.
//
// @Summary Export all student related data
// @Description Get an export of all student data from core and all microservices
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /privacy/data-export [post]
func studentDataExport(c *gin.Context) {
	subjectIdentifiers, errSI := service.GetSubjectIdentifiers(c)
	if errSI != nil {
    utils.HandleError(c, http.StatusBadRequest, errSI);
		return
	}

  if valErr := service.ValidateAllowedToExport(c, subjectIdentifiers.UserID); valErr != nil {
    utils.HandleError(c, http.StatusTooManyRequests, valErr)
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
// @Produce json
// @Success 200 {object} privacyDTO.PrivacyExport
// @Success 204 "No recent export"
// @Failure 500 {object} utils.ErrorResponse
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

// getExportDocDownloadURL returns a presigned download URL for a single export document.
//
// @Summary Get download URL for an export document
// @Description Returns a short-lived presigned URL to download a specific export ZIP
// @Produce json
// @Param uuid path string true "Export UUID"
// @Param docID path string true "Export Document UUID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 405 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
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
		utils.HandleError(c, http.StatusMethodNotAllowed, valErr)
		return
	}

  if valErr := service.ValidateExportValid(c, exportID); valErr != nil {
		utils.HandleError(c, http.StatusMethodNotAllowed, valErr)
		return
  }

	downloadURL, err := service.GetDownloadURLForDoc(c, exportID, docID)
	if err != nil {
		log.Error("get export doc download URL failed: ", err)
		utils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"downloadUrl": downloadURL})
}

// getExport returns export status and ids for documents related to an export
//
// @Summary Get all Ids and status of export and export docs
// @Description Get the current status and all ids of the export record and all export docs. needed so client knows what docs exist to be able to download them
// @Produce json
// @Success 200 {object} privacyDTO.PrivacyExport
// @Failure 500 {object} utils.ErrorResponse
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
