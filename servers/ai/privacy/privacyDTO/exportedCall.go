package privacyDTO

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/ai/calls/callDTO"
)

// ContentWithheld marks content that also concerns other people, so it is not handed out
// (GDPR Art. 15(4)).
const ContentWithheld = "withheld"

// ExportedCall is one call in the privacy export of the person it was made by or about.
type ExportedCall struct {
	ID               uuid.UUID        `json:"id"`
	CoursePhaseID    uuid.UUID        `json:"coursePhaseId"`
	MadeBySubject    bool             `json:"madeBySubject"`
	ActorRole        string           `json:"actorRole"`
	CallerService    string           `json:"callerService"`
	Feature          string           `json:"feature"`
	Template         *string          `json:"template"`
	TemplateVersion  *string          `json:"templateVersion"`
	RequestedModel   *string          `json:"requestedModel"`
	ServedModel      *string          `json:"servedModel"`
	Outcome          string           `json:"outcome"`
	FinishReason     *string          `json:"finishReason"`
	PromptTokens     *int32           `json:"promptTokens"`
	CompletionTokens *int32           `json:"completionTokens"`
	RequestedAt      time.Time        `json:"requestedAt"`
	CompletedAt      *time.Time       `json:"completedAt"`
	ContentState     string           `json:"contentState"`
	Request          json.RawMessage  `json:"request,omitempty"`
	Choices          []callDTO.Choice `json:"choices,omitempty"`
	Events           []ExportedEvent  `json:"events"`
}
