package announcement

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/prompt-edu/prompt/servers/core/announcement/announcementDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

const (
	maxTitleLength     = 100
	maxMessageLength   = 300
	maxLinkLabelLength = 50
)

func validateAnnouncement(request announcementDTO.UpsertAnnouncement) error {
	switch db.AnnouncementSeverity(request.Severity) {
	case db.AnnouncementSeverityInfo, db.AnnouncementSeverityWarning, db.AnnouncementSeverityCritical:
	default:
		return errors.New("severity must be one of info, warning, critical")
	}
	if strings.TrimSpace(request.Message) == "" {
		return errors.New("message is required")
	}
	if utf8.RuneCountInString(request.Message) > maxMessageLength {
		return errors.New("message must not exceed 300 characters")
	}
	if utf8.RuneCountInString(request.Title) > maxTitleLength {
		return errors.New("title must not exceed 100 characters")
	}
	if err := validateLink(request.LinkURL, request.LinkLabel); err != nil {
		return err
	}
	if request.StartsAt != nil && request.ExpiresAt != nil && !request.ExpiresAt.After(*request.StartsAt) {
		return errors.New("expiry must be after the start")
	}
	return nil
}

func validateLink(linkURL, linkLabel string) error {
	if linkURL == "" {
		if linkLabel != "" {
			return errors.New("a link label requires a link URL")
		}
		return nil
	}
	parsed, err := url.ParseRequestURI(linkURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return errors.New("link URL must be a valid http or https URL")
	}
	if utf8.RuneCountInString(linkLabel) > maxLinkLabelLength {
		return errors.New("link label must not exceed 50 characters")
	}
	return nil
}
