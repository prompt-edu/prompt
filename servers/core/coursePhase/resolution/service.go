package resolution

import (
	"context"
	"strings"

	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution/resolutionDTO"
)

type ResolutionService struct {
	coreHost string
}

func NewResolutionService(coreHost string) *ResolutionService {
	return &ResolutionService{
		coreHost: coreHost,
	}
}

// ReplaceResolutionURLs rewrites BaseURL for each resolution.
//
// • If resolveLocally == true and a local resolution exists, its URL is preferred; otherwise we fall back to the core host. (For example, when requesting server is in same docker network as providing server.)
// • The placeholder {CORE_HOST} is always replaced by the normalised core host.
func (s *ResolutionService) ReplaceResolutionURLs(ctx context.Context, resolutions []resolutionDTO.Resolution) ([]resolutionDTO.Resolution, error) {
	if len(resolutions) == 0 {
		return resolutions, nil
	}

	for i, r := range resolutions {
		resolutions[i].BaseURL = s.ResolveBaseURL(r.BaseURL)
	}

	return resolutions, nil
}

// ResolveBaseURL replaces the {CORE_HOST} placeholder in a course phase type base URL.
func (s *ResolutionService) ResolveBaseURL(baseURL string) string {
	return strings.ReplaceAll(baseURL, "{CORE_HOST}", NormaliseHost(s.coreHost))
}

// NormaliseHost ensures the host string starts with a scheme.
func NormaliseHost(host string) string {
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}
	return "https://" + host
}
