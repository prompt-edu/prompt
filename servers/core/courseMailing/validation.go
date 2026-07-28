package courseMailing

import (
	"fmt"
	"strings"

	"github.com/prompt-edu/prompt/servers/core/courseMailing/courseMailingDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// validateCampaignRequest validates a create/update payload. Drafts may leave the
// subject, body, target phase, and statuses empty; those are enforced on send.
func validateCampaignRequest(req courseMailingDTO.MailCampaignRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("%w: campaign name is required", ErrValidation)
	}
	if _, err := validateStatusValues(req.TargetPassStatuses); err != nil {
		return err
	}
	if req.ReplyToOverride != nil && req.ReplyToOverride.Email == "" {
		return fmt.Errorf("%w: reply-to override requires an email address", ErrValidation)
	}
	return nil
}

func validateCampaignForSend(base db.MailCampaign) error {
	if err := validateCampaignForTest(base); err != nil {
		return err
	}
	if _, err := expandStatuses(base.TargetPassStatuses); err != nil {
		return err
	}
	return nil
}

func validateCampaignForTest(base db.MailCampaign) error {
	if strings.TrimSpace(base.Subject) == "" {
		return fmt.Errorf("%w: a subject is required before sending", ErrValidation)
	}
	if !isASCII(base.Subject) {
		return fmt.Errorf("%w: the subject must contain only ASCII characters", ErrValidation)
	}
	if strings.TrimSpace(base.Body) == "" {
		return fmt.Errorf("%w: a message body is required before sending", ErrValidation)
	}
	if !base.TargetCoursePhaseID.Valid {
		return fmt.Errorf("%w: a target course phase must be selected before sending", ErrValidation)
	}
	return nil
}

// validateStatusValues checks that each status is one of the accepted values and
// returns the deduplicated list.
func validateStatusValues(statuses []string) ([]string, error) {
	seen := make(map[string]bool, len(statuses))
	out := make([]string, 0, len(statuses))
	for _, s := range statuses {
		switch s {
		case "all", string(db.PassStatusPassed), string(db.PassStatusFailed), string(db.PassStatusNotAssessed):
			if !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		default:
			return nil, fmt.Errorf("%w: invalid target status %q", ErrValidation, s)
		}
	}
	return out, nil
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}
