package config

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/ls1intum/prompt-sdk"
	"github.com/ls1intum/prompt-sdk/keycloakTokenVerifier"
	"github.com/ls1intum/prompt2/servers/certificate/config/configDTO"
	log "github.com/sirupsen/logrus"
)

func setupConfigRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	configRouter := routerGroup.Group("/config")

	configRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), getConfig)
	configRouter.PUT("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), updateConfig)
	configRouter.PUT("/release-date", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), updateReleaseDate)
	configRouter.GET("/template", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTemplate)
}

func getConfig(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	config, err := GetCoursePhaseConfig(c, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get course phase config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve course phase config"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func updateConfig(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	var request configDTO.UpdateConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get the name of the user making the update
	var updatedBy string
	if user, exists := keycloakTokenVerifier.GetTokenUser(c); exists {
		updatedBy = user.FirstName + " " + user.LastName
	}

	config, err := UpdateCoursePhaseConfig(c, coursePhaseID, request.TemplateContent, updatedBy)
	if err != nil {
		log.WithError(err).Error("Failed to update course phase config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update course phase config"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func updateReleaseDate(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	var request configDTO.UpdateReleaseDateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get the name of the user making the update
	var updatedBy string
	if user, exists := keycloakTokenVerifier.GetTokenUser(c); exists {
		updatedBy = user.FirstName + " " + user.LastName
	}

	config, err := UpdateReleaseDate(c, coursePhaseID, request.ReleaseDate, updatedBy)
	if err != nil {
		log.WithError(err).Error("Failed to update release date")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update release date"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func getTemplate(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	templateContent, err := GetTemplateContent(c, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get template content")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=template.typ")
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, templateContent)
}
