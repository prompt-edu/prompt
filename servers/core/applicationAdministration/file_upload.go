package applicationAdministration

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/storage/files"
	log "github.com/sirupsen/logrus"
)

const externalUploaderID = "external"

type applicationPresignUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
}

type applicationCompleteUploadRequest struct {
	StorageKey       string `json:"storageKey"`
	OriginalFilename string `json:"originalFilename"`
	ContentType      string `json:"contentType"`
	Description      string `json:"description"`
	Tags             string `json:"tags"`
}

// presignApplicationUploadExternal godoc
// @Summary Create a presigned upload URL (external)
// @Description Returns a presigned URL for uploading a file directly to storage (external applicants)
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param body body applicationPresignUploadRequest true "Presign request"
// @Success 200 {object} files.PresignUploadResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply/{coursePhaseID}/files/presign [post]
func presignApplicationUploadExternal(c *gin.Context) {
	coursePhaseID, ok := parseCoursePhaseID(c)
	if !ok {
		return
	}
	if !ensureOpenApplicationPhase(c, coursePhaseID) {
		return
	}

	var body applicationPresignUploadRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := files.StorageServiceSingleton.PresignUpload(c.Request.Context(), files.PresignUploadRequest{
		Filename:      body.Filename,
		ContentType:   body.ContentType,
		CoursePhaseID: &coursePhaseID,
		Description:   body.Description,
		Tags:          parseTags(body.Tags),
	})
	if err != nil {
		log.WithError(err).Error("Failed to presign external upload")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate upload URL"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// completeApplicationUploadExternal godoc
// @Summary Complete a presigned upload (external)
// @Description Registers an uploaded file after a presigned upload completes (external applicants)
// @Tags applications
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param body body applicationCompleteUploadRequest true "Complete request"
// @Success 201 {object} files.FileResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply/{coursePhaseID}/files/complete [post]
func completeApplicationUploadExternal(c *gin.Context) {
	coursePhaseID, ok := parseCoursePhaseID(c)
	if !ok {
		return
	}
	if !ensureOpenApplicationPhase(c, coursePhaseID) {
		return
	}

	var body applicationCompleteUploadRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	fileResponse, err := files.StorageServiceSingleton.CreateFileFromStorageKey(c.Request.Context(), files.CreateFileFromStorageKeyRequest{
		StorageKey:       body.StorageKey,
		OriginalFilename: body.OriginalFilename,
		ContentType:      body.ContentType,
		CoursePhaseID:    &coursePhaseID,
		Description:      body.Description,
		Tags:             parseTags(body.Tags),
	}, externalUploaderID, "")
	if err != nil {
		log.WithError(err).Error("Failed to complete external upload")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete upload"})
		return
	}

	c.JSON(http.StatusCreated, fileResponse)
}

// presignApplicationUploadAuthenticated godoc
// @Summary Create a presigned upload URL (authenticated)
// @Description Returns a presigned URL for uploading a file directly to storage (authenticated applicants)
// @Tags applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param body body applicationPresignUploadRequest true "Presign request"
// @Success 200 {object} files.PresignUploadResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply/authenticated/{coursePhaseID}/files/presign [post]
func presignApplicationUploadAuthenticated(c *gin.Context) {
	if _, ok := getUserID(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	coursePhaseID, ok := parseCoursePhaseID(c)
	if !ok {
		return
	}

	if !ensureOpenApplicationPhase(c, coursePhaseID) {
		return
	}

	var body applicationPresignUploadRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := files.StorageServiceSingleton.PresignUpload(c.Request.Context(), files.PresignUploadRequest{
		Filename:      body.Filename,
		ContentType:   body.ContentType,
		CoursePhaseID: &coursePhaseID,
		Description:   body.Description,
		Tags:          parseTags(body.Tags),
	})
	if err != nil {
		log.WithError(err).Error("Failed to presign authenticated upload")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// completeApplicationUploadAuthenticated godoc
// @Summary Complete a presigned upload (authenticated)
// @Description Registers an uploaded file after a presigned upload completes (authenticated applicants)
// @Tags applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param body body applicationCompleteUploadRequest true "Complete request"
// @Success 201 {object} files.FileResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply/authenticated/{coursePhaseID}/files/complete [post]
func completeApplicationUploadAuthenticated(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	coursePhaseID, ok := parseCoursePhaseID(c)
	if !ok {
		return
	}

	if !ensureOpenApplicationPhase(c, coursePhaseID) {
		return
	}

	var body applicationCompleteUploadRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	email := ""
	if emailVal, exists := c.Get("userEmail"); exists {
		if emailStr, ok := emailVal.(string); ok {
			email = emailStr
		}
	}

	fileResponse, err := files.StorageServiceSingleton.CreateFileFromStorageKey(c.Request.Context(), files.CreateFileFromStorageKeyRequest{
		StorageKey:       body.StorageKey,
		OriginalFilename: body.OriginalFilename,
		ContentType:      body.ContentType,
		CoursePhaseID:    &coursePhaseID,
		Description:      body.Description,
		Tags:             parseTags(body.Tags),
	}, userID, email)
	if err != nil {
		log.WithError(err).Error("Failed to complete authenticated upload")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, fileResponse)
}

// deleteApplicationFileAuthenticated godoc
// @Summary Delete an uploaded application file (authenticated)
// @Description Deletes a file uploaded by the authenticated applicant for the given course phase
// @Tags applications
// @Security BearerAuth
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param fileId path string true "File UUID"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /apply/authenticated/{coursePhaseID}/files/{fileId} [delete]
func deleteApplicationFileAuthenticated(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	coursePhaseID, ok := parseCoursePhaseID(c)
	if !ok {
		return
	}

	fileID, err := uuid.Parse(c.Param("fileId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	fileResponse, err := files.StorageServiceSingleton.GetFileByID(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if fileResponse.CoursePhaseID == nil || *fileResponse.CoursePhaseID != coursePhaseID {
		c.JSON(http.StatusForbidden, gin.H{"error": "no permission to delete file"})
		return
	}

	if fileResponse.UploadedByUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "no permission to delete file"})
		return
	}

	if err := files.StorageServiceSingleton.DeleteFile(c.Request.Context(), fileID, true); err != nil {
		log.WithError(err).Error("Failed to delete application file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	c.Status(http.StatusNoContent)
}

// getApplicationFileDownloadURL godoc
// @Summary Get application file download URL
// @Description Creates a presigned download URL for an application file in the given course phase
// @Tags applications
// @Security BearerAuth
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param fileId path string true "File UUID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /applications/{coursePhaseID}/files/{fileId}/download-url [get]
func getApplicationFileDownloadURL(c *gin.Context) {
	coursePhaseID, ok := parseCoursePhaseID(c)
	if !ok {
		return
	}

	fileID, err := uuid.Parse(c.Param("fileId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	fileResponse, err := files.StorageServiceSingleton.GetFileByID(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if fileResponse.CoursePhaseID == nil || *fileResponse.CoursePhaseID != coursePhaseID {
		c.JSON(http.StatusForbidden, gin.H{"error": "no permission to access file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"downloadUrl": fileResponse.DownloadURL})
}

func parseCoursePhaseID(c *gin.Context) (uuid.UUID, bool) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course phase ID"})
		return uuid.Nil, false
	}
	return coursePhaseID, true
}

func getUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

func parseTags(tags string) []string {
	if tags == "" {
		return nil
	}
	parts := strings.Split(tags, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func ensureOpenApplicationPhase(c *gin.Context, coursePhaseID uuid.UUID) bool {
	ctxWithTimeout, cancel := db.GetTimeoutContext(c.Request.Context())
	defer cancel()

	applicationDetails, err := ApplicationServiceSingleton.queries.CheckIfCoursePhaseIsOpenApplicationPhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Could not validate application phase")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not validate application phase"})
		return false
	}
	if !applicationDetails.IsApplication {
		c.JSON(http.StatusForbidden, gin.H{"error": "course phase is not open for applications"})
		return false
	}
	return true
}
