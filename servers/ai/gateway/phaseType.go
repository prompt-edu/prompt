package gateway

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
)

// phaseTypes caches the type of each course phase, which never changes once the phase exists.
type phaseTypes struct {
	coreURL string
	known   sync.Map
}

func (p *phaseTypes) resolve(ctx context.Context, coursePhaseID uuid.UUID, authHeader string) (string, error) {
	if phaseType, ok := p.known.Load(coursePhaseID); ok {
		return phaseType.(string), nil
	}
	phase, err := promptSDK.FetchCoursePhase(ctx, p.coreURL, authHeader, coursePhaseID)
	if err != nil {
		return "", fmt.Errorf("resolve phase type from core: %w", err)
	}
	if phase.CoursePhaseTypeName == "" {
		return "", fmt.Errorf("core answered without a phase type for %s", coursePhaseID)
	}
	p.known.Store(coursePhaseID, phase.CoursePhaseTypeName)
	return phase.CoursePhaseTypeName, nil
}
