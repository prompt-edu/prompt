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
func StatusPlaceholderValues(courseName string, courseStartDate, courseEndDate pgtype.Date, participant db.GetParticipantMailingInformationByIDsRow) map[string]string {
	return getStatusEmailPlaceholderValues(courseName, courseStartDate, courseEndDate, participant)
}

// CoursePlaceholderValues builds only the course-level placeholders, for mails with
// no single recipient (e.g. an archive copy sent to a campaign's CC/BCC addresses).
// Student placeholders are intentionally omitted rather than filled with fake values,
// so they are left untouched in the rendered mail.
func CoursePlaceholderValues(courseName string, courseStartDate, courseEndDate pgtype.Date) map[string]string {
	return map[string]string{
		"courseName":      courseName,
		"courseStartDate": getPgtypeDateValue(courseStartDate),
		"courseEndDate":   getPgtypeDateValue(courseEndDate),
	}
}
