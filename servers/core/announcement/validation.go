package announcement

import (
	"errors"
	"fmt"
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

func normalizeAnnouncement(request announcementDTO.UpsertAnnouncement) announcementDTO.UpsertAnnouncement {
	request.Title = strings.TrimSpace(request.Title)
	request.Message = strings.TrimSpace(request.Message)
	request.LinkURL = strings.TrimSpace(request.LinkURL)
	request.LinkLabel = strings.TrimSpace(request.LinkLabel)
	return request
}

func validateAnnouncement(request announcementDTO.UpsertAnnouncement) error {
	switch db.AnnouncementSeverity(request.Severity) {
	case db.AnnouncementSeverityInfo, db.AnnouncementSeverityWarning, db.AnnouncementSeverityCritical:
	default:
		return errors.New("severity must be one of info, warning, critical")
	}
	if request.Message == "" {
		return errors.New("message is required")
	}
	if utf8.RuneCountInString(request.Message) > maxMessageLength {
		return fmt.Errorf("message must not exceed %d characters", maxMessageLength)
	}
	if utf8.RuneCountInString(request.Title) > maxTitleLength {
		return fmt.Errorf("title must not exceed %d characters", maxTitleLength)
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
		return fmt.Errorf("link label must not exceed %d characters", maxLinkLabelLength)
	}
	return nil
}
