package presentation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
)

func RegisterRoutes(router *gin.RouterGroup, service *Service) {
	allRoles := []string{
		promptSDK.PromptAdmin,
		promptSDK.PromptLecturer,
		promptSDK.CourseLecturer,
		promptSDK.CourseEditor,
		promptSDK.CourseStudent,
	}
	staffRoles := []string{
		promptSDK.PromptAdmin,
		promptSDK.PromptLecturer,
		promptSDK.CourseLecturer,
		promptSDK.CourseEditor,
	}
	managerRoles := []string{
		promptSDK.PromptAdmin,
		promptSDK.CourseLecturer,
	}

	router.GET("/config", promptSDK.AuthenticationMiddleware(allRoles...), service.getConfig)
	router.PUT("/config", promptSDK.AuthenticationMiddleware(managerRoles...), service.updateConfig)

	router.GET("/categories", promptSDK.AuthenticationMiddleware(allRoles...), service.listCategories)
	router.POST("/categories", promptSDK.AuthenticationMiddleware(managerRoles...), service.createCategory)
	router.PUT("/categories/:categoryID", promptSDK.AuthenticationMiddleware(managerRoles...), service.updateCategory)
	router.DELETE("/categories/:categoryID", promptSDK.AuthenticationMiddleware(managerRoles...), service.deleteCategory)

	router.GET("/slots", promptSDK.AuthenticationMiddleware(allRoles...), service.listSlots)
	router.POST("/slots", promptSDK.AuthenticationMiddleware(managerRoles...), service.createSlot)
	router.PUT("/slots/:slotID", promptSDK.AuthenticationMiddleware(managerRoles...), service.updateSlot)
	router.DELETE("/slots/:slotID", promptSDK.AuthenticationMiddleware(managerRoles...), service.deleteSlot)
	router.GET("/targets", promptSDK.AuthenticationMiddleware(staffRoles...), service.listTargets)
	router.PUT("/slots/:slotID/assignment", promptSDK.AuthenticationMiddleware(managerRoles...), service.assignTarget)
	router.DELETE("/slots/:slotID/assignment", promptSDK.AuthenticationMiddleware(managerRoles...), service.unassignTarget)

	router.GET("/presentations", promptSDK.AuthenticationMiddleware(staffRoles...), service.listPresentations)
	router.GET("/presentations/me", promptSDK.AuthenticationMiddleware(promptSDK.CourseStudent), service.getOwnPresentation)

	router.GET("/presentations/:presentationID/materials", promptSDK.AuthenticationMiddleware(allRoles...), service.listMaterials)
	router.POST("/presentations/:presentationID/materials/presign", promptSDK.AuthenticationMiddleware(allRoles...), service.createUploadIntent)
	router.POST("/presentations/:presentationID/materials/:materialID/complete", promptSDK.AuthenticationMiddleware(allRoles...), service.completeUpload)
	router.GET("/presentations/:presentationID/materials/:materialID/download", promptSDK.AuthenticationMiddleware(allRoles...), service.getMaterialDownload)
	router.DELETE("/presentations/:presentationID/materials/:materialID", promptSDK.AuthenticationMiddleware(allRoles...), service.deleteMaterial)

	router.GET("/presentations/:presentationID/feedback", promptSDK.AuthenticationMiddleware(allRoles...), service.getFeedback)
	router.PUT("/presentations/:presentationID/feedback/answers/:categoryID", promptSDK.AuthenticationMiddleware(staffRoles...), service.putFeedbackAnswer)
	router.POST("/presentations/:presentationID/feedback/submit", promptSDK.AuthenticationMiddleware(staffRoles...), service.submitFeedback)
	router.POST("/presentations/:presentationID/feedback/reopen", promptSDK.AuthenticationMiddleware(staffRoles...), service.reopenFeedback)
	router.DELETE("/presentations/:presentationID/feedback/draft", promptSDK.AuthenticationMiddleware(staffRoles...), service.deleteDraft)
	router.POST("/presentations/:presentationID/feedback/release", promptSDK.AuthenticationMiddleware(managerRoles...), service.releaseFeedback)
	router.DELETE("/presentations/:presentationID/feedback/release", promptSDK.AuthenticationMiddleware(managerRoles...), service.unreleaseFeedback)
	router.DELETE("/presentations/:presentationID/feedback", promptSDK.AuthenticationMiddleware(managerRoles...), service.resetFeedback)
	router.GET("/presentations/:presentationID/feedback/events", promptSDK.AuthenticationMiddleware(staffRoles...), service.streamFeedbackEvents)
}

func coursePhaseID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		return uuid.Nil, apiError(http.StatusBadRequest, "invalid_course_phase_id", "Invalid course phase ID", err)
	}
	return id, nil
}

func pathID(c *gin.Context, parameter, code, message string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(parameter))
	if err != nil {
		return uuid.Nil, apiError(http.StatusBadRequest, code, message, err)
	}
	return id, nil
}

func requestUser(c *gin.Context) (User, error) {
	tokenUser, exists := keycloakTokenVerifier.GetTokenUser(c)
	if !exists {
		return User{}, apiError(http.StatusUnauthorized, "unauthorized", "Authenticated user is unavailable", nil)
	}
	name := strings.TrimSpace(tokenUser.FirstName + " " + tokenUser.LastName)
	if name == "" {
		name = tokenUser.Email
	}
	isPromptAdmin := tokenUser.Roles[promptSDK.PromptAdmin]
	isPromptLecturer := tokenUser.Roles[promptSDK.PromptLecturer]
	return User{
		ID:                    tokenUser.ID,
		Email:                 tokenUser.Email,
		Name:                  name,
		CourseParticipationID: tokenUser.CourseParticipationID,
		Staff:                 isPromptAdmin || isPromptLecturer || tokenUser.IsLecturer || tokenUser.IsEditor,
		CanRelease:            isPromptAdmin || tokenUser.IsLecturer,
	}, nil
}

func (s *Service) getConfig(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.GetConfig(c, phaseID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) updateConfig(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	var request UpdateSettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, apiError(http.StatusBadRequest, "invalid_config", "Invalid presentation config", err))
		return
	}
	response, err := s.UpdateConfig(c, phaseID, request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) listCategories(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.ListCategories(c, phaseID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) createCategory(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	var request CategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, apiError(http.StatusBadRequest, "invalid_category", "Invalid feedback category", err))
		return
	}
	response, err := s.CreateCategory(c, phaseID, request, c.Query("resetExistingData") == "true")
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (s *Service) updateCategory(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	categoryID, err := pathID(c, "categoryID", "invalid_category_id", "Invalid feedback category ID")
	if err != nil {
		writeError(c, err)
		return
	}
	var request CategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, apiError(http.StatusBadRequest, "invalid_category", "Invalid feedback category", err))
		return
	}
	response, err := s.UpdateCategory(c, phaseID, categoryID, request, c.Query("resetExistingData") == "true")
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) deleteCategory(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	categoryID, err := pathID(c, "categoryID", "invalid_category_id", "Invalid feedback category ID")
	if err != nil {
		writeError(c, err)
		return
	}
	if err := s.DeleteCategory(c, phaseID, categoryID, c.Query("resetExistingData") == "true"); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) listSlots(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.ListSlots(c, phaseID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) createSlot(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	var request SlotRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, apiError(http.StatusBadRequest, "invalid_slot", "Invalid presentation slot", err))
		return
	}
	response, err := s.CreateSlot(c, phaseID, request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (s *Service) updateSlot(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	slotID, err := pathID(c, "slotID", "invalid_slot_id", "Invalid presentation slot ID")
	if err != nil {
		writeError(c, err)
		return
	}
	var request SlotRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, apiError(http.StatusBadRequest, "invalid_slot", "Invalid presentation slot", err))
		return
	}
	response, err := s.UpdateSlot(c, phaseID, slotID, request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) deleteSlot(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	slotID, err := pathID(c, "slotID", "invalid_slot_id", "Invalid presentation slot ID")
	if err != nil {
		writeError(c, err)
		return
	}
	if err := s.DeleteSlot(c, phaseID, slotID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) listTargets(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.ListTargets(c, c.GetHeader("Authorization"), phaseID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) assignTarget(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	slotID, err := pathID(c, "slotID", "invalid_slot_id", "Invalid presentation slot ID")
	if err != nil {
		writeError(c, err)
		return
	}
	var request AssignmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, apiError(http.StatusBadRequest, "invalid_assignment", "Invalid presentation assignment", err))
		return
	}
	response, err := s.AssignTarget(c, c.GetHeader("Authorization"), phaseID, slotID, request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) unassignTarget(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	slotID, err := pathID(c, "slotID", "invalid_slot_id", "Invalid presentation slot ID")
	if err != nil {
		writeError(c, err)
		return
	}
	if err := s.UnassignTarget(c, phaseID, slotID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) listPresentations(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.ListPresentations(c, phaseID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) getOwnPresentation(c *gin.Context) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	user, err := requestUser(c)
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.GetOwnPresentation(c, c.GetHeader("Authorization"), phaseID, user)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func presentationRequest(c *gin.Context) (uuid.UUID, uuid.UUID, User, error) {
	phaseID, err := coursePhaseID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, User{}, err
	}
	presentationID, err := pathID(c, "presentationID", "invalid_presentation_id", "Invalid presentation ID")
	if err != nil {
		return uuid.Nil, uuid.Nil, User{}, err
	}
	user, err := requestUser(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, User{}, err
	}
	return phaseID, presentationID, user, nil
}

func (s *Service) listMaterials(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.ListMaterials(c, c.GetHeader("Authorization"), phaseID, presentationID, user)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) createUploadIntent(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	var request PresignMaterialRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, apiError(http.StatusBadRequest, "invalid_material", "Invalid material upload request", err))
		return
	}
	response, err := s.CreateUploadIntent(c, c.GetHeader("Authorization"), phaseID, presentationID, user, request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (s *Service) completeUpload(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	materialID, err := pathID(c, "materialID", "invalid_upload_id", "Invalid material upload ID")
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.CompleteUpload(c, c.GetHeader("Authorization"), phaseID, presentationID, materialID, user)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (s *Service) getMaterialDownload(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	materialID, err := pathID(c, "materialID", "invalid_material_id", "Invalid presentation material ID")
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.GetMaterialDownload(c, c.GetHeader("Authorization"), phaseID, presentationID, materialID, user)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) deleteMaterial(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	materialID, err := pathID(c, "materialID", "invalid_material_id", "Invalid presentation material ID")
	if err != nil {
		writeError(c, err)
		return
	}
	if err := s.DeleteMaterial(c, c.GetHeader("Authorization"), phaseID, presentationID, materialID, user); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) getFeedback(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	response, err := s.GetFeedback(c, c.GetHeader("Authorization"), phaseID, presentationID, user)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) putFeedbackAnswer(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	categoryID, err := pathID(c, "categoryID", "invalid_category_id", "Invalid feedback category ID")
	if err != nil {
		writeError(c, err)
		return
	}
	var request PutAnswerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, apiError(http.StatusBadRequest, "invalid_feedback_answer", "Invalid feedback answer", err))
		return
	}
	response, err := s.PutFeedbackAnswer(c, phaseID, presentationID, categoryID, user, request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) submitFeedback(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	if err := s.SubmitFeedback(c, phaseID, presentationID, user); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) reopenFeedback(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	if err := s.ReopenFeedback(c, phaseID, presentationID, user); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) deleteDraft(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	if err := s.DeleteDraft(c, phaseID, presentationID, user); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) releaseFeedback(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	var request ReleaseFeedbackRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, apiError(http.StatusBadRequest, "invalid_release", "Invalid feedback release", err))
		return
	}
	if err := s.ReleaseFeedback(c, phaseID, presentationID, user, request.ReleaseName); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) unreleaseFeedback(c *gin.Context) {
	phaseID, presentationID, _, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	if err := s.UnreleaseFeedback(c, phaseID, presentationID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) resetFeedback(c *gin.Context) {
	phaseID, presentationID, _, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	if err := s.ResetFeedback(c, phaseID, presentationID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) streamFeedbackEvents(c *gin.Context) {
	phaseID, presentationID, user, err := presentationRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	presentation, err := s.queries.GetPresentation(c, db.GetPresentationParams{
		ID:            presentationID,
		CoursePhaseID: phaseID,
	})
	if err != nil {
		writeError(c, apiError(http.StatusNotFound, "presentation_not_found", "Presentation not found", err))
		return
	}
	config, err := s.GetConfig(c, phaseID)
	if err != nil {
		writeError(c, err)
		return
	}
	if config.FeedbackMode != feedbackShared {
		writeError(c, apiError(http.StatusConflict, "feedback_not_shared", "Live editing is only available in shared feedback mode", nil))
		return
	}
	if presentation.FeedbackReleasedAt.Valid {
		writeError(c, apiError(http.StatusConflict, "feedback_released", "Released feedback cannot be edited live", nil))
		return
	}

	_, events, unsubscribe := s.hub.Subscribe(presentationID, user.ID, user.Name)
	defer unsubscribe()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.Flush()

	writeEvent := func(event FeedbackEvent) error {
		data, marshalErr := json.Marshal(event)
		if marshalErr != nil {
			return marshalErr
		}
		if _, writeErr := fmt.Fprintf(c.Writer, "data: %s\n\n", data); writeErr != nil {
			return writeErr
		}
		c.Writer.Flush()
		return nil
	}
	if err := writeEvent(FeedbackEvent{
		Type:           "snapshot",
		PresentationID: presentationID,
		ActiveEditors:  s.hub.ActiveEditors(presentationID),
	}); err != nil {
		return
	}

	keepAlive := time.NewTicker(20 * time.Second)
	defer keepAlive.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event := <-events:
			if err := writeEvent(event); err != nil {
				return
			}
		case <-keepAlive.C:
			if _, err := fmt.Fprint(c.Writer, ": keepalive\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}
