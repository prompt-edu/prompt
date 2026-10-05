package key

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
	"github.com/prompt-edu/prompt/servers/ai/encryption"
	"github.com/prompt-edu/prompt/servers/ai/key/keyDTO"
)

var ErrNotConfigured = errors.New("AI not configured")

type Service struct {
	queries *db.Queries
	coreURL string
}

func NewService(queries *db.Queries, coreURL string) *Service {
	return &Service{queries: queries, coreURL: coreURL}
}

type PhaseKey struct {
	Key       string
	PhaseType string
}

func (s *Service) Get(ctx context.Context, coursePhaseID uuid.UUID) (keyDTO.Status, error) {
	row, err := s.queries.GetPhaseKey(ctx, coursePhaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return keyDTO.Status{}, nil
	}
	if err != nil {
		return keyDTO.Status{}, fmt.Errorf("load phase key: %w", err)
	}
	return keyDTO.GetStatusDTOFromDBModel(row), nil
}

func (s *Service) Set(ctx context.Context, coursePhaseID uuid.UUID, logosKey, setBy, authHeader string) (keyDTO.Status, error) {
	phaseType, err := s.phaseType(ctx, coursePhaseID, authHeader)
	if err != nil {
		return keyDTO.Status{}, err
	}
	encrypted, err := encryption.Encrypt([]byte(logosKey))
	if err != nil {
		return keyDTO.Status{}, fmt.Errorf("encrypt phase key: %w", err)
	}
	row, err := s.queries.UpsertPhaseKey(ctx, db.UpsertPhaseKeyParams{
		CoursePhaseID: coursePhaseID,
		PhaseType:     phaseType,
		EncryptedKey:  encrypted,
		Last4:         logosKey[len(logosKey)-4:],
		SetBy:         setBy,
	})
	if err != nil {
		return keyDTO.Status{}, fmt.Errorf("store phase key: %w", err)
	}
	return keyDTO.GetStatusDTOFromDBModel(row), nil
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

func (s *Service) phaseType(ctx context.Context, coursePhaseID uuid.UUID, authHeader string) (string, error) {
	phase, err := promptSDK.FetchCoursePhase(ctx, s.coreURL, authHeader, coursePhaseID)
	if err != nil {
		return "", fmt.Errorf("resolve phase type from core: %w", err)
	}
	if phase.CoursePhaseTypeName == "" {
		return "", fmt.Errorf("core answered without a phase type for %s", coursePhaseID)
	}
	return phase.CoursePhaseTypeName, nil
}
