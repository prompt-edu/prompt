package org

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
)

const (
	minSlugLength         = 2
	maxSlugLength         = 50
	maxNameLength         = 255
	maxSchoolLength       = 255
	maxUniversityLength   = 255
	maxWebsiteLength      = 2048
	maxContactEmailLength = 254
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func validateCreateOrg(input orgDTO.CreateOrg) error {
	if err := validateSlug(input.Slug); err != nil {
		return err
	}
	return validateOrgDetails(input.Name, input.School, input.University, input.Website, input.ContactEmail)
}

func validateUpdateOrg(input orgDTO.UpdateOrg) error {
	return validateOrgDetails(input.Name, input.School, input.University, input.Website, input.ContactEmail)
}

// validateSlug mirrors the check_org_slug_format constraint. The slug is embedded in
// Keycloak role names, so a "cg" segment is rejected: a course in semester "org" mints
// custom group roles "org-<course>-cg-<name>", which could otherwise equal an org role.
func validateSlug(slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return errors.New("slug is required")
	}
	if !slugPattern.MatchString(slug) {
		return errors.New("slug may only contain lowercase letters and digits, separated by single hyphens")
	}
	if len(slug) < minSlugLength || len(slug) > maxSlugLength {
		return fmt.Errorf("slug must be between %d and %d characters", minSlugLength, maxSlugLength)
	}
	if slices.Contains(strings.Split(slug, "-"), "cg") {
		return errors.New(`slug must not contain a "cg" segment`)
	}
	return nil
}

func validateOrgDetails(name, school, university, website, contactEmail string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("org name is required")
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		return fmt.Errorf("org name must be at most %d characters", maxNameLength)
	}
	if utf8.RuneCountInString(strings.TrimSpace(school)) > maxSchoolLength {
		return fmt.Errorf("school must be at most %d characters", maxSchoolLength)
	}
	if utf8.RuneCountInString(strings.TrimSpace(university)) > maxUniversityLength {
		return fmt.Errorf("university must be at most %d characters", maxUniversityLength)
	}
	if err := validateWebsite(strings.TrimSpace(website)); err != nil {
		return err
	}
	return validateContactEmail(strings.TrimSpace(contactEmail))
}

func validateWebsite(website string) error {
	if website == "" {
		return nil
	}
	if utf8.RuneCountInString(website) > maxWebsiteLength {
		return fmt.Errorf("website must be at most %d characters", maxWebsiteLength)
	}
	parsed, err := url.ParseRequestURI(website)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return errors.New("website must be an absolute http or https URL")
	}
	return nil
}

// validateContactEmail accepts a bare address only: the display-name form
// "Chair <chair@example.com>" parses as well, but would not be stored as given.
func validateContactEmail(contactEmail string) error {
	if contactEmail == "" {
		return nil
	}
	if utf8.RuneCountInString(contactEmail) > maxContactEmailLength {
		return fmt.Errorf("contact email must be at most %d characters", maxContactEmailLength)
	}
	address, err := mail.ParseAddress(contactEmail)
	if err != nil || address.Name != "" || address.Address != contactEmail {
		return errors.New("contact email must be a valid email address")
	}
	return nil
}
