package callDTO

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
)

// Call is the audit metadata of one model call.
type Call struct {
	ID                uuid.UUID       `json:"id"`
	ActorID           string          `json:"actorId"`
	ActorRole         string          `json:"actorRole"`
	Issuer            string          `json:"issuer"`
	Feature           string          `json:"feature"`
	Template          *string         `json:"template"`
	TemplateVersion   *string         `json:"templateVersion"`
	RequestedModel    *string         `json:"requestedModel"`
	ServedModel       *string         `json:"servedModel"`
	SystemFingerprint *string         `json:"systemFingerprint"`
	Provider          string          `json:"provider"`
	Params            json.RawMessage `json:"params"`
	ContextHash       *string         `json:"contextHash"`
	ResponseHash      *string         `json:"responseHash"`
	Outcome           string          `json:"outcome"`
	HTTPStatus        *int32          `json:"httpStatus"`
	FinishReason      *string         `json:"finishReason"`
	ErrorCode         *string         `json:"errorCode"`
	PromptTokens      *int32          `json:"promptTokens"`
	CompletionTokens  *int32          `json:"completionTokens"`
	Streamed          bool            `json:"streamed"`
	ServerVersion     string          `json:"serverVersion"`
	RequestedAt       time.Time       `json:"requestedAt"`
	FirstTokenAt      *time.Time      `json:"firstTokenAt"`
	CompletedAt       *time.Time      `json:"completedAt"`
}

func GetCallDTOFromDBModel(row db.AiCall) Call {
	return Call{
		ID:                row.ID,
		ActorID:           row.ActorID,
		ActorRole:         row.ActorRole,
		Issuer:            row.Issuer,
		Feature:           row.Feature,
		Template:          textOf(row.Template),
		TemplateVersion:   textOf(row.TemplateVersion),
		RequestedModel:    textOf(row.RequestedModel),
		ServedModel:       textOf(row.ServedModel),
		SystemFingerprint: textOf(row.SystemFingerprint),
		Provider:          row.Provider,
		Params:            row.Params,
		ContextHash:       textOf(row.ContextHash),
		ResponseHash:      textOf(row.ResponseHash),
		Outcome:           row.Outcome,
		HTTPStatus:        intOf(row.HttpStatus),
		FinishReason:      textOf(row.FinishReason),
		ErrorCode:         textOf(row.ErrorCode),
		PromptTokens:      intOf(row.PromptTokens),
		CompletionTokens:  intOf(row.CompletionTokens),
		Streamed:          row.Streamed,
		ServerVersion:     row.ServerVersion,
		RequestedAt:       row.RequestedAt.Time,
		FirstTokenAt:      timeOf(row.FirstTokenAt),
		CompletedAt:       timeOf(row.CompletedAt),
	}
}
