package coursePhase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	sdkTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

const (
	// coursePhaseRoute is the path the SDK's RegisterCoursePhaseDeletionEndpoint serves the
	// deletion under: DELETE <baseURL>/course_phase/<coursePhaseID>.
	coursePhaseRoute = "course_phase"

	// coreBaseURL marks a phase type core implements itself, so there is no module to call.
	coreBaseURL = "core"

	moduleRequestTimeout = 10 * time.Second
)

// moduleClient never follows redirects: a 3xx must surface as a failed deletion instead of being
// replayed as a GET whose 200 would look like success.
var moduleClient = &http.Client{
	Timeout:       moduleRequestTimeout,
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

// deleteModuleData asks every course phase module holding data for one of the given phases to drop
// it. It returns an error unless every module either deleted its data or reported that it does not
// implement phase deletion, so callers can keep core's rows and let the whole operation be retried.
func (s *CoursePhaseService) deleteModuleData(ctx context.Context, authHeader string, coursePhaseIDs []uuid.UUID) error {
	targets, err := s.queries.GetCoursePhaseDeletionTargets(ctx, coursePhaseIDs)
	if err != nil {
		return fmt.Errorf("failed to load course phase deletion targets: %w", err)
	}

	byBaseURL := make(map[string][]db.GetCoursePhaseDeletionTargetsRow)
	for _, target := range targets {
		if target.BaseUrl == coreBaseURL {
			continue
		}

		baseURL := s.resolutions.ResolveBaseURL(target.BaseUrl)
		if _, err := url.ParseRequestURI(baseURL); err != nil {
			return fmt.Errorf("course phase type %q has an unusable base url %q: %w", target.CoursePhaseTypeName, baseURL, err)
		}
		byBaseURL[baseURL] = append(byBaseURL[baseURL], target)
	}

	var mu sync.Mutex
	var failures []error
	var wg sync.WaitGroup
	for baseURL, moduleTargets := range byBaseURL {
		wg.Go(func() {
			if err := deleteModuleDataAt(ctx, authHeader, baseURL, moduleTargets); err != nil {
				mu.Lock()
				failures = append(failures, err)
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	return errors.Join(failures...)
}

// DeleteModuleDataForCourse drops the module-held data of every phase of the course and returns the
// phases it covered. It is the course deletion counterpart of DeleteCoursePhase: the course row
// cascades into its phases, which leaves the modules no chance to be asked afterwards. The caller
// must check under a course row lock that no phase was added since, before deleting the course.
func (s *CoursePhaseService) DeleteModuleDataForCourse(ctx context.Context, authHeader string, courseID uuid.UUID) ([]uuid.UUID, error) {
	phases, err := s.queries.GetAllCoursePhaseForCourse(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to load course phases: %w", err)
	}

	coursePhaseIDs := make([]uuid.UUID, 0, len(phases))
	for _, phase := range phases {
		coursePhaseIDs = append(coursePhaseIDs, phase.ID)
	}

	if err := s.deleteModuleData(ctx, authHeader, coursePhaseIDs); err != nil {
		return nil, err
	}
	return coursePhaseIDs, nil
}

// deleteModuleDataAt asks one module to delete the data of each of its phases, once it has reported
// that it implements phase deletion at all.
func deleteModuleDataAt(ctx context.Context, authHeader, baseURL string, targets []db.GetCoursePhaseDeletionTargetsRow) error {
	supported, err := moduleSupportsPhaseDeletion(ctx, baseURL)
	if err != nil {
		return err
	}
	if !supported {
		log.Warnf("%s does not support phase deletion, its data for %d phase(s) is kept", baseURL, len(targets))
		return nil
	}

	for _, target := range targets {
		if err := deletePhaseDataAt(ctx, authHeader, baseURL, target.ID); err != nil {
			return fmt.Errorf("course phase type %q failed to delete its data for phase %s: %w", target.CoursePhaseTypeName, target.ID, err)
		}
	}
	return nil
}

// moduleSupportsPhaseDeletion reads the module's advertised capabilities. Everything other than a
// well-formed answer is an error: an unreachable or misrouted service must not be mistaken for one
// that has no data to delete, and a reverse proxy without a route for the module answers the probe
// with the client's SPA document under status 200.
func moduleSupportsPhaseDeletion(ctx context.Context, baseURL string) (bool, error) {
	infoURL, err := url.JoinPath(baseURL, "info")
	if err != nil {
		return false, fmt.Errorf("failed to build info url for %q: %w", baseURL, err)
	}

	resp, err := sendModuleRequest(ctx, http.MethodGet, infoURL, "")
	if err != nil {
		return false, err
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%s answered %d, so its phase deletion support is unknown", infoURL, resp.StatusCode)
	}

	var info sdkTypes.ServiceInfo
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&info); err != nil {
		return false, fmt.Errorf("%s did not answer with service info: %w", infoURL, err)
	}

	return info.Capabilities[sdkTypes.CapabilityPhaseDeletion], nil
}

func deletePhaseDataAt(ctx context.Context, authHeader, baseURL string, coursePhaseID uuid.UUID) error {
	deletionURL, err := url.JoinPath(baseURL, coursePhaseRoute, coursePhaseID.String())
	if err != nil {
		return fmt.Errorf("failed to build deletion url: %w", err)
	}

	resp, err := sendModuleRequest(ctx, http.MethodDelete, deletionURL, authHeader)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("%s answered %d: %s", deletionURL, resp.StatusCode, string(body))
}

func sendModuleRequest(ctx context.Context, method, requestURL, authHeader string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", requestURL, err)
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := moduleClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach %s: %w", requestURL, err)
	}
	return resp, nil
}

func closeBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Error(err)
	}
}
