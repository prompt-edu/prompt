package developerProfile

import (
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/intro_course/developerProfile/developerProfileDTO"
)

// udidRegex validates a UDID of the form "XXXXXXXX-XXXXXXXXXXXXXXXX"
// where X is a hexadecimal digit.
var udidRegex = regexp.MustCompile(`^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{16}$`)

// isValidUDID returns true if the given pgtype.Text UDID is valid.
// If udid.Valid is false, meaning no UDID was provided, it returns true.
func isValidUDID(udid pgtype.Text) bool {
	if !udid.Valid {
		// No UDID provided is considered valid.
		return true
	}
	return udidRegex.MatchString(udid.String)
}

// validateDeveloperProfileUDIDs checks all UDID fields in the profile DTO.
func validateDeveloperProfileUDIDs(profile developerProfileDTO.PostDeveloperProfile) error {
	if !isValidUDID(profile.IPhoneUDID) {
		return fmt.Errorf("invalid iPhone UDID")
	}
	if !isValidUDID(profile.IPadUDID) {
		return fmt.Errorf("invalid iPad UDID")
	}
	if !isValidUDID(profile.AppleWatchUDID) {
		return fmt.Errorf("invalid Apple Watch UDID")
	}
	return nil
}
