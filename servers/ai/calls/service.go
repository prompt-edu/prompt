package calls

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/ai/calls/callDTO"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
	"github.com/prompt-edu/prompt/servers/ai/encryption"
	log "github.com/sirupsen/logrus"
)

const (
	OutcomePending   = "pending"
	OutcomeSuccess   = "success"
	OutcomeError     = "error"
	OutcomeTimeout   = "timeout"
	OutcomeCancelled = "cancelled"
	OutcomeDenied    = "denied"
)

var (
	ErrNotFound     = errors.New("call not found")
	ErrInvalidEvent = errors.New("an edited event needs an edit distance")
)

type Service struct {
	queries       *db.Queries
	conn          *pgxpool.Pool
	provider      string
	serverVersion string
	box           *encryption.Box
}

func NewService(queries *db.Queries, conn *pgxpool.Pool, provider, serverVersion string, box *encryption.Box) *Service {
	return &Service{queries: queries, conn: conn, provider: provider, serverVersion: serverVersion, box: box}
}

type Request struct {
	CoursePhaseID   uuid.UUID
	ActorID         string
	ActorRole       string
	Issuer          string
	CallerService   string
	Feature         string
	Template        string
	TemplateVersion string
	Model           string
	Params          []byte
	Streamed        bool
	Subjects        []uuid.UUID
	Body            []byte
}

type Completion struct {
	Outcome     string
	HTTPStatus  int
	ErrorCode   string
	FirstByteAt time.Time
	Response    []byte
	Streamed    bool
}

// Begin must succeed before the provider is called, so no model call goes unrecorded.
func (s *Service) Begin(ctx context.Context, request Request) (uuid.UUID, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin call transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	callID, err := qtx.CreateCall(ctx, s.callParams(request, OutcomePending))
	if err != nil {
		return uuid.Nil, fmt.Errorf("create call: %w", err)
	}
	encryptedRequest, err := s.box.Encrypt(callID, encryption.Request, request.Body)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encrypt request: %w", err)
	}
	if err := qtx.CreateCallContent(ctx, db.CreateCallContentParams{CallID: callID, Request: encryptedRequest}); err != nil {
		return uuid.Nil, fmt.Errorf("create call content: %w", err)
	}
	if len(request.Subjects) > 0 {
		if err := qtx.CreateCallSubjects(ctx, db.CreateCallSubjectsParams{CallID: callID, CourseParticipationIds: request.Subjects}); err != nil {
			return uuid.Nil, fmt.Errorf("create call subjects: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit call: %w", err)
	}
	return callID, nil
}

func (s *Service) Deny(ctx context.Context, request Request, httpStatus int, errorCode string) (uuid.UUID, error) {
	params := s.callParams(request, OutcomeDenied)
	params.HttpStatus = optionalInt(int32(httpStatus))
	params.ErrorCode = optionalText(errorCode)
	params.CompletedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	return s.queries.CreateCall(ctx, params)
}

func (s *Service) Finish(ctx context.Context, callID uuid.UUID, completion Completion) error {
	summary := Summarize(completion.Response, completion.Streamed)
	encryptedResponse, err := s.box.Encrypt(callID, encryption.Response, completion.Response)
	if err != nil {
		return fmt.Errorf("encrypt response: %w", err)
	}
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin completion transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	completed, err := qtx.CompleteCall(ctx, db.CompleteCallParams{
		ID:                callID,
		ServedModel:       optionalText(summary.Model),
		SystemFingerprint: optionalText(summary.SystemFingerprint),
		Outcome:           completion.Outcome,
		HttpStatus:        optionalInt(int32(completion.HTTPStatus)),
		FinishReason:      optionalText(summary.FinishReason),
		ErrorCode:         optionalText(completion.ErrorCode),
		PromptTokens:      pointerInt(summary.PromptTokens),
		CompletionTokens:  pointerInt(summary.CompletionTokens),
		FirstTokenAt:      pgtype.Timestamptz{Time: completion.FirstByteAt, Valid: !completion.FirstByteAt.IsZero()},
	})
	if err != nil {
		return fmt.Errorf("complete call: %w", err)
	}
	if completed == 0 {
		return fmt.Errorf("call %s was already completed, most likely as abandoned", callID)
	}
	if err := qtx.SetCallResponse(ctx, db.SetCallResponseParams{CallID: callID, Response: encryptedResponse}); err != nil {
		return fmt.Errorf("store call response: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Service) callParams(request Request, outcome string) db.CreateCallParams {
	params := request.Params
	if params == nil {
		params = []byte("{}")
	}
	return db.CreateCallParams{
		CoursePhaseID:   request.CoursePhaseID,
		ActorID:         request.ActorID,
		ActorRole:       request.ActorRole,
		Issuer:          request.Issuer,
		CallerService:   request.CallerService,
		Feature:         request.Feature,
		Template:        optionalText(request.Template),
		TemplateVersion: optionalText(request.TemplateVersion),
		RequestedModel:  optionalText(request.Model),
		Provider:        s.provider,
		Params:          params,
		Outcome:         outcome,
		Streamed:        request.Streamed,
		ServerVersion:   s.serverVersion,
	}
}

func (s *Service) List(ctx context.Context, coursePhaseID uuid.UUID, cursor *callDTO.Cursor, limit int32) (callDTO.Page, error) {
	params := db.ListCallsParams{CoursePhaseID: coursePhaseID, PageSize: limit + 1}
	if cursor != nil {
		params.CursorRequestedAt = pgtype.Timestamptz{Time: cursor.RequestedAt, Valid: true}
		params.CursorID = pgtype.UUID{Bytes: cursor.ID, Valid: true}
	}
	rows, err := s.queries.ListCalls(ctx, params)
	if err != nil {
		return callDTO.Page{}, fmt.Errorf("list calls: %w", err)
	}
	page := callDTO.Page{Calls: make([]callDTO.Call, 0, len(rows))}
	if len(rows) > int(limit) {
		rows = rows[:limit]
		last := rows[len(rows)-1]
		page.NextCursor = &callDTO.Cursor{RequestedAt: last.RequestedAt.Time, ID: last.ID}
	}
	for _, row := range rows {
		page.Calls = append(page.Calls, callDTO.GetCallDTOFromDBModel(row))
	}
	return page, nil
}

// Get returns the content only once its content_viewed event is recorded.
func (s *Service) Get(ctx context.Context, coursePhaseID, callID uuid.UUID, viewerID string) (callDTO.Detail, error) {
	row, err := s.queries.GetCall(ctx, db.GetCallParams{ID: callID, CoursePhaseID: coursePhaseID})
	if errors.Is(err, pgx.ErrNoRows) {
		return callDTO.Detail{}, ErrNotFound
	}
	if err != nil {
		return callDTO.Detail{}, fmt.Errorf("load call: %w", err)
	}
	detail := callDTO.Detail{Call: callDTO.GetCallDTOFromDBModel(row.AiCall), ContentState: callDTO.ContentUnavailable, Subjects: []uuid.UUID{}}

	switch {
	case row.Restricted:
		detail.ContentState = callDTO.ContentRestricted
	case row.Request != nil:
		content, err := s.decryptContent(callID, row.Request, row.Response, row.AiCall.Streamed, row.AiCall.Outcome)
		if err != nil {
			log.WithError(err).WithField("callID", callID).Error("Could not decrypt AI call content")
			break
		}
		if _, err := s.recordEvent(ctx, coursePhaseID, callID, viewerID, true, callDTO.EventContentViewed, nil); err != nil {
			return callDTO.Detail{}, fmt.Errorf("record content view: %w", err)
		}
		detail.ContentState = callDTO.ContentAvailable
		detail.Content = &content
	}

	if !row.Restricted {
		subjects, err := s.queries.ListCallSubjects(ctx, callID)
		if err != nil {
			return callDTO.Detail{}, fmt.Errorf("list call subjects: %w", err)
		}
		detail.Subjects = append(detail.Subjects, subjects...)
	}
	events, err := s.queries.ListCallEvents(ctx, callID)
	if err != nil {
		return callDTO.Detail{}, fmt.Errorf("list call events: %w", err)
	}
	detail.Events = make([]callDTO.Event, 0, len(events))
	for _, event := range events {
		detail.Events = append(detail.Events, callDTO.GetEventDTOFromDBModel(event))
	}
	return detail, nil
}

// Only an admin may report events on a call someone else made.
func (s *Service) AddEvent(ctx context.Context, coursePhaseID, callID uuid.UUID, actorID string, byAdmin bool, request callDTO.EventRequest) (callDTO.Event, error) {
	if request.Type == callDTO.EventEdited && request.EditDistance == nil {
		return callDTO.Event{}, ErrInvalidEvent
	}
	data, err := json.Marshal(request.Data())
	if err != nil {
		return callDTO.Event{}, err
	}
	return s.recordEvent(ctx, coursePhaseID, callID, actorID, byAdmin, request.Type, data)
}

func (s *Service) recordEvent(ctx context.Context, coursePhaseID, callID uuid.UUID, actorID string, byAdmin bool, eventType string, data []byte) (callDTO.Event, error) {
	if data == nil {
		data = []byte("{}")
	}
	row, err := s.queries.CreateCallEvent(ctx, db.CreateCallEventParams{
		ActorID:       actorID,
		Type:          eventType,
		Data:          data,
		CallID:        callID,
		CoursePhaseID: coursePhaseID,
		ByAdmin:       byAdmin,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return callDTO.Event{}, ErrNotFound
	}
	if err != nil {
		return callDTO.Event{}, err
	}
	return callDTO.GetEventDTOFromDBModel(row), nil
}

func (s *Service) decryptContent(callID uuid.UUID, encryptedRequest, encryptedResponse []byte, streamed bool, outcome string) (callDTO.Content, error) {
	request, err := s.box.Decrypt(callID, encryption.Request, encryptedRequest)
	if err != nil {
		return callDTO.Content{}, fmt.Errorf("decrypt request: %w", err)
	}
	content := callDTO.Content{Request: request, Choices: []callDTO.Choice{}}
	if encryptedResponse != nil {
		response, err := s.box.Decrypt(callID, encryption.Response, encryptedResponse)
		if err != nil {
			return callDTO.Content{}, fmt.Errorf("decrypt response: %w", err)
		}
		if choices := Summarize(response, streamed).Choices; choices != nil {
			content.Choices = choices
		} else if outcome != OutcomeSuccess {
			content.ProviderResponse = string(response)
		}
	}
	return content, nil
}

func optionalText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func optionalInt(value int32) pgtype.Int4 {
	return pgtype.Int4{Int32: value, Valid: value != 0}
}

func pointerInt(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}
