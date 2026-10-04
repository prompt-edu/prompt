package key

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
	"github.com/prompt-edu/prompt/servers/ai/encryption"
)

var ErrNotConfigured = errors.New("AI not configured")

type Service struct {
	queries *db.Queries
	coreURL string
}

func NewService(queries *db.Queries, coreURL string) *Service {
	return &Service{queries: queries, coreURL: coreURL}
}

type Status struct {
	Configured bool       `json:"configured"`
	Last4      string     `json:"last4,omitempty"`
	SetBy      string     `json:"setBy,omitempty"`
	SetAt      *time.Time `json:"setAt,omitempty"`
}

type PhaseKey struct {
	Key       string
	PhaseType string
}

func statusOf(row db.AiPhaseKey) Status {
	setAt := row.SetAt.Time
	return Status{Configured: true, Last4: row.Last4, SetBy: row.SetBy, SetAt: &setAt}
}

func (s *Service) Get(ctx context.Context, coursePhaseID uuid.UUID) (Status, error) {
	row, err := s.queries.GetPhaseKey(ctx, coursePhaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Status{}, nil
	}
	if err != nil {
		return Status{}, fmt.Errorf("load phase key: %w", err)
	}
	return statusOf(row), nil
}

func (s *Service) Set(ctx context.Context, coursePhaseID uuid.UUID, logosKey, setBy, authHeader string) (Status, error) {
	phaseType, err := s.phaseType(coursePhaseID, authHeader)
	if err != nil {
		return Status{}, err
	}
	encrypted, err := encryption.Encrypt([]byte(logosKey))
	if err != nil {
		return Status{}, fmt.Errorf("encrypt phase key: %w", err)
	}
	row, err := s.queries.UpsertPhaseKey(ctx, db.UpsertPhaseKeyParams{
		CoursePhaseID: coursePhaseID,
		PhaseType:     phaseType,
		EncryptedKey:  encrypted,
		Last4:         logosKey[len(logosKey)-4:],
		SetBy:         setBy,
	})
	if err != nil {
		return Status{}, fmt.Errorf("store phase key: %w", err)
	}
	return statusOf(row), nil
}

func (s *Service) Delete(ctx context.Context, coursePhaseID uuid.UUID) error {
	return s.queries.DeletePhaseKey(ctx, coursePhaseID)
}

// Audit records of the phase stay until their retention ends.
func (s *Service) HandleCoursePhaseDeletion(c *gin.Context, coursePhaseID uuid.UUID) error {
	return s.Delete(c.Request.Context(), coursePhaseID)
}

func (s *Service) Resolve(ctx context.Context, coursePhaseID uuid.UUID) (PhaseKey, error) {
	row, err := s.queries.GetPhaseKey(ctx, coursePhaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return PhaseKey{}, ErrNotConfigured
	}
	if err != nil {
		return PhaseKey{}, fmt.Errorf("load phase key: %w", err)
	}
	logosKey, err := encryption.Decrypt(row.EncryptedKey)
	if err != nil {
		return PhaseKey{}, fmt.Errorf("decrypt phase key: %w", err)
	}
	return PhaseKey{Key: string(logosKey), PhaseType: row.PhaseType}, nil
}

func (s *Service) phaseType(coursePhaseID uuid.UUID, authHeader string) (string, error) {
	phaseURL, err := url.JoinPath(s.coreURL, "api/course_phases", coursePhaseID.String())
	if err != nil {
		return "", err
	}
	body, err := sdkUtils.FetchJSON(phaseURL, authHeader)
	if err != nil {
		return "", fmt.Errorf("resolve phase type from core: %w", err)
	}
	var phase struct {
		CoursePhaseTypeName string `json:"coursePhaseTypeName"`
	}
	if err := json.Unmarshal(body, &phase); err != nil || phase.CoursePhaseTypeName == "" {
		return "", fmt.Errorf("core answered without a phase type for %s", coursePhaseID)
	}
	return phase.CoursePhaseTypeName, nil
}
