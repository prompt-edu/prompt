package phaseconfig

import (
	"errors"
	"regexp"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/phaseconfig/phaseconfigDTO"
	log "github.com/sirupsen/logrus"
)

const maxSemesterTagLength = 32

// semesterTagPattern matches lowercase letters and digits — the tag is interpolated
// into provider resource names that often disallow spaces or special characters.
var semesterTagPattern = regexp.MustCompile(`^[a-z0-9]+$`)

// validateUpsertRequest checks that the semester tag, when provided, is in a format
// the execution worker can safely interpolate into provider resource names.
// An empty tag is allowed — phases that do not yet use templated names skip this check.
// Source phase IDs are optional; FK violations surface as DB errors from the upsert
// query when the IDs do not refer to existing phases.
func validateUpsertRequest(req phaseconfigDTO.UpsertRequest) error {
	tag := strings.TrimSpace(req.SemesterTag)
	if tag == "" {
		return nil
	}
	if len(tag) > maxSemesterTagLength {
		return logAndReturnError("semesterTag is too long")
	}
	if !semesterTagPattern.MatchString(tag) {
		return logAndReturnError("semesterTag must contain only lowercase letters and digits")
	}
	return nil
}

func logAndReturnError(msg string) error {
	log.Error(msg)
	return errors.New(msg)
}
