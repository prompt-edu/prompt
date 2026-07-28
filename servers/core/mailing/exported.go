package mailing

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// ReplacePlaceholders substitutes {{placeholder}} tokens in a template with the
// provided values. Exported wrapper around replacePlaceholders for reuse by the
// courseMailing module.
func ReplacePlaceholders(template string, values map[string]string) string {
	return replacePlaceholders(template, values)
}

// StatusPlaceholderValues builds the standard student/course placeholder map used
// for status and manual mails. Exported wrapper around getStatusEmailPlaceholderValues.
func StatusPlaceholderValues(courseName string, courseStartDate, courseEndDate pgtype.Date, participant db.GetParticipantMailingInformationRow) map[string]string {
	return getStatusEmailPlaceholderValues(courseName, courseStartDate, courseEndDate, participant)
}
