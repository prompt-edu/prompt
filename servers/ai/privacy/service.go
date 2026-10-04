package privacy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/ai/calls"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
	"github.com/prompt-edu/prompt/servers/ai/encryption"
	"github.com/prompt-edu/prompt/servers/ai/feature"
)

// GDPR Art. 15(4): content that also concerns other people is not handed out.
const contentWithheld = "withheld"

type Service struct {
	queries  *db.Queries
	conn     *pgxpool.Pool
	policyOf func(string) feature.Policy
}

func NewService(queries *db.Queries, conn *pgxpool.Pool) *Service {
	return &Service{queries: queries, conn: conn, policyOf: feature.PolicyOf}
}

type exportedCall struct {
	ID               uuid.UUID       `json:"id"`
	CoursePhaseID    uuid.UUID       `json:"coursePhaseId"`
	MadeBySubject    bool            `json:"madeBySubject"`
	ActorRole        string          `json:"actorRole"`
	Feature          string          `json:"feature"`
	Template         *string         `json:"template"`
	TemplateVersion  *string         `json:"templateVersion"`
	RequestedModel   *string         `json:"requestedModel"`
	ServedModel      *string         `json:"servedModel"`
	Outcome          string          `json:"outcome"`
	FinishReason     *string         `json:"finishReason"`
	PromptTokens     *int32          `json:"promptTokens"`
	CompletionTokens *int32          `json:"completionTokens"`
	RequestedAt      time.Time       `json:"requestedAt"`
	CompletedAt      *time.Time      `json:"completedAt"`
	ContentState     string          `json:"contentState"`
	Request          json.RawMessage `json:"request,omitempty"`
	Response         string          `json:"response,omitempty"`
	Events           []exportedEvent `json:"events"`
}

type exportedEvent struct {
	Type          string          `json:"type"`
	Data          json.RawMessage `json:"data"`
	MadeBySubject bool            `json:"madeBySubject"`
	CreatedAt     time.Time       `json:"createdAt"`
}

func (s *Service) Export(c *gin.Context, export *sdkUtils.Export, subject keycloakTokenVerifier.SubjectIdentifiers) error {
	export.AddJSON("AI calls", "ai-calls.json", func() (any, error) {
		exported, err := s.exportCalls(c.Request.Context(), actorID(subject), subject.CourseParticipationIDs)
		if err != nil || len(exported) == 0 {
			return nil, err
		}
		return exported, nil
	})
	return export.Err()
}

func (s *Service) exportCalls(ctx context.Context, actorID string, participationIDs []uuid.UUID) ([]exportedCall, error) {
	rows, err := s.queries.ListPrivacyCalls(ctx, db.ListPrivacyCallsParams{ActorID: actorID, CourseParticipationIds: participationIDs})
	if err != nil {
		return nil, fmt.Errorf("list calls: %w", err)
	}
	exported := make([]exportedCall, 0, len(rows))
	callIDs := make([]uuid.UUID, 0, len(rows))
	byID := map[uuid.UUID]int{}
	for _, row := range rows {
		call, err := exportedCallOf(row)
		if err != nil {
			return nil, err
		}
		byID[row.ID] = len(exported)
		exported = append(exported, call)
		callIDs = append(callIDs, row.ID)
	}

	events, err := s.queries.ListPrivacyEvents(ctx, db.ListPrivacyEventsParams{ActorID: actorID, CallIds: callIDs})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	for _, event := range events {
		call := &exported[byID[event.CallID]]
		call.Events = append(call.Events, exportedEvent{
			Type: event.Type, Data: event.Data, MadeBySubject: event.IsActor, CreatedAt: event.CreatedAt.Time,
		})
	}
	return exported, nil
}

func exportedCallOf(row db.ListPrivacyCallsRow) (exportedCall, error) {
	call := exportedCall{
		ID:              row.ID,
		CoursePhaseID:   row.CoursePhaseID,
		MadeBySubject:   row.IsActor,
		ActorRole:       row.ActorRole,
		Feature:         row.Feature,
		Template:        textOf(row.Template.String, row.Template.Valid),
		TemplateVersion: textOf(row.TemplateVersion.String, row.TemplateVersion.Valid),
		RequestedModel:  textOf(row.RequestedModel.String, row.RequestedModel.Valid),
		ServedModel:     textOf(row.ServedModel.String, row.ServedModel.Valid),
		Outcome:         row.Outcome,
		FinishReason:    textOf(row.FinishReason.String, row.FinishReason.Valid),
		RequestedAt:     row.RequestedAt.Time,
		ContentState:    calls.ContentUnavailable,
		Events:          []exportedEvent{},
	}
	if row.PromptTokens.Valid {
		call.PromptTokens = &row.PromptTokens.Int32
	}
	if row.CompletionTokens.Valid {
		call.CompletionTokens = &row.CompletionTokens.Int32
	}
	if row.CompletedAt.Valid {
		call.CompletedAt = &row.CompletedAt.Time
	}

	switch {
	case row.Restricted:
		call.ContentState = calls.ContentRestricted
	case row.Request != nil && (!row.IsSubject || row.HasOtherSubjects):
		call.ContentState = contentWithheld
	case row.Request != nil:
		request, err := encryption.Decrypt(row.Request)
		if err != nil {
			return exportedCall{}, fmt.Errorf("decrypt request of call %s: %w", row.ID, err)
		}
		call.ContentState, call.Request = calls.ContentAvailable, request
		if row.Response != nil {
			response, err := encryption.Decrypt(row.Response)
			if err != nil {
				return exportedCall{}, fmt.Errorf("decrypt response of call %s: %w", row.ID, err)
			}
			call.Response = calls.Summarize(response, row.Streamed).Text
		}
	}
	return call, nil
}

// High-risk content is restricted until its retention ends (GDPR Art. 17(3)(b)), the rest deleted.
func (s *Service) Delete(c *gin.Context, subject keycloakTokenVerifier.SubjectIdentifiers) error {
	if len(subject.CourseParticipationIDs) == 0 {
		return nil
	}
	ctx := c.Request.Context()
	subjectCalls, err := s.queries.ListSubjectCalls(ctx, subject.CourseParticipationIDs)
	if err != nil {
		return fmt.Errorf("list the subject's calls: %w", err)
	}
	var restricted, deleted []uuid.UUID
	for _, call := range subjectCalls {
		if s.policyOf(call.Feature).HighRisk {
			restricted = append(restricted, call.ID)
		} else {
			deleted = append(deleted, call.ID)
		}
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin erasure: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)
	if err := qtx.RestrictCallContents(ctx, restricted); err != nil {
		return fmt.Errorf("restrict call contents: %w", err)
	}
	if err := qtx.DeleteSubjectCallContents(ctx, db.DeleteSubjectCallContentsParams{
		CallIds:                deleted,
		CourseParticipationIds: subject.CourseParticipationIDs,
	}); err != nil {
		return fmt.Errorf("delete call contents: %w", err)
	}
	return tx.Commit(ctx)
}

func actorID(subject keycloakTokenVerifier.SubjectIdentifiers) string {
	if subject.UserID == uuid.Nil {
		return ""
	}
	return subject.UserID.String()
}

func textOf(value string, valid bool) *string {
	if !valid {
		return nil
	}
	return &value
}
