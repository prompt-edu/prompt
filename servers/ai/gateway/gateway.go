package gateway

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/ai/calls"
	"github.com/prompt-edu/prompt/servers/ai/feature"
	"github.com/prompt-edu/prompt/servers/ai/key"
	log "github.com/sirupsen/logrus"
)

const (
	CallIDHeader = "X-Prompt-AI-Call-ID"

	maxRequestBytes   = 4 << 20
	maxSubjects       = 100
	totalTimeout      = 10 * time.Minute
	idleTimeout       = 2 * time.Minute
	completionTimeout = 10 * time.Second
)

type Gateway struct {
	providerURL   *url.URL
	allowedModels map[string]bool
	keys          *key.Service
	calls         *calls.Service
	modelsClient  *http.Client
}

func New(providerURL *url.URL, allowedModels []string, keys *key.Service, callService *calls.Service) *Gateway {
	allowed := make(map[string]bool, len(allowedModels))
	for _, model := range allowedModels {
		allowed[model] = true
	}
	return &Gateway{
		providerURL:   providerURL,
		allowedModels: allowed,
		keys:          keys,
		calls:         callService,
		modelsClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// PromptLecturer is never allowed: the SDK admits it for every course phase.
func RegisterRoutes(coursePhaseAPI *gin.RouterGroup, gateway *Gateway) {
	v1 := coursePhaseAPI.Group("/v1",
		promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor))
	v1.POST("/chat/completions", gateway.chatCompletions)
	v1.GET("/models", gateway.listModels)
}

func (g *Gateway) chatCompletions(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "invalid course phase id"})
		return
	}
	user, _ := keycloakTokenVerifier.GetTokenUser(c)
	request := calls.Request{
		CoursePhaseID:   coursePhaseID,
		ActorID:         user.ID,
		ActorRole:       actorRole(user),
		Feature:         feature.Adhoc,
		Template:        c.GetHeader("X-Prompt-Template"),
		TemplateVersion: c.GetHeader("X-Prompt-Template-Version"),
	}
	if name := c.GetHeader("X-Prompt-Feature"); name != "" {
		request.Feature = name
	}

	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		g.deny(c, request, http.StatusRequestEntityTooLarge, "request_too_large", "the request body is larger than 4 MiB")
		return
	}
	if err != nil {
		g.deny(c, request, http.StatusBadRequest, "invalid_request", "the request body could not be read")
		return
	}
	prepared, err := prepare(body)
	if err != nil {
		g.deny(c, request, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	request.Model, request.Params, request.Streamed, request.Body = prepared.model, prepared.params, prepared.streamed, prepared.body
	if !g.allowedModels[request.Model] {
		g.deny(c, request, http.StatusBadRequest, "model_not_allowed", "the model is not allowed")
		return
	}
	if request.Subjects, err = parseSubjects(c.GetHeader("X-Prompt-Subjects")); err != nil {
		g.deny(c, request, http.StatusBadRequest, "invalid_subjects", "X-Prompt-Subjects must list at most 100 course participation ids")
		return
	}
	phaseKey, err := g.keys.Resolve(c.Request.Context(), coursePhaseID)
	if errors.Is(err, key.ErrNotConfigured) {
		g.deny(c, request, http.StatusConflict, "ai_not_configured", key.ErrNotConfigured.Error())
		return
	}
	if err != nil {
		log.WithError(err).Error("Could not resolve the phase key")
		c.JSON(http.StatusInternalServerError, sdkUtils.ErrorResponse{Error: "could not resolve the phase key"})
		return
	}
	if err := feature.CheckFor(request.Feature, phaseKey.PhaseType); err != nil {
		g.deny(c, request, http.StatusBadRequest, "feature_not_allowed", err.Error())
		return
	}

	callID, err := g.calls.Begin(c.Request.Context(), request)
	if err != nil {
		log.WithError(err).Error("Could not record the AI call, so it is not forwarded")
		c.JSON(http.StatusServiceUnavailable, sdkUtils.ErrorResponse{Error: "the AI call could not be recorded"})
		return
	}
	c.Header(CallIDHeader, callID.String())
	g.forward(c, callID, phaseKey.Key, request)
}

func (g *Gateway) deny(c *gin.Context, request calls.Request, status int, errorCode, message string) {
	callID, err := g.calls.Deny(c.Request.Context(), request, status, errorCode)
	if err != nil {
		log.WithError(err).Error("Could not record the denied AI call")
		c.JSON(http.StatusServiceUnavailable, sdkUtils.ErrorResponse{Error: "the AI call could not be recorded"})
		return
	}
	c.Header(CallIDHeader, callID.String())
	c.JSON(status, sdkUtils.ErrorResponse{Error: message})
}

func (g *Gateway) listModels(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "invalid course phase id"})
		return
	}
	phaseKey, err := g.keys.Resolve(c.Request.Context(), coursePhaseID)
	if errors.Is(err, key.ErrNotConfigured) {
		c.JSON(http.StatusConflict, sdkUtils.ErrorResponse{Error: key.ErrNotConfigured.Error()})
		return
	}
	if err != nil {
		log.WithError(err).Error("Could not resolve the phase key")
		c.JSON(http.StatusInternalServerError, sdkUtils.ErrorResponse{Error: "could not resolve the phase key"})
		return
	}

	available, err := g.providerModels(c, phaseKey.Key)
	if err != nil {
		log.WithError(err).Warn("Could not list the provider's models")
		c.JSON(http.StatusBadGateway, sdkUtils.ErrorResponse{Error: "the AI provider did not list its models"})
		return
	}
	models := make([]json.RawMessage, 0, len(available))
	for _, model := range available {
		var entry struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(model, &entry) == nil && g.allowedModels[entry.ID] {
			models = append(models, model)
		}
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": models})
}

func (g *Gateway) providerModels(c *gin.Context, logosKey string) ([]json.RawMessage, error) {
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, g.providerURL.JoinPath("models").String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+logosKey)
	response, err := g.modelsClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("the provider answered " + response.Status)
	}
	var listing struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&listing); err != nil {
		return nil, err
	}
	return listing.Data, nil
}

func actorRole(user keycloakTokenVerifier.TokenUser) string {
	switch {
	case user.Roles[promptSDK.PromptAdmin]:
		return promptSDK.PromptAdmin
	case user.IsLecturer:
		return promptSDK.CourseLecturer
	default:
		return promptSDK.CourseEditor
	}
}
