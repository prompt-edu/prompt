package permissionValidation

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const PromptAdmin = "PROMPT_Admin"
const PromptLecturer = "PROMPT_Lecturer"
const CourseLecturer = "Lecturer"
const CourseEditor = "Editor"
const CourseStudent = "Student"
const OrgAdmin = "Admin"
const OrgMember = "Member"

const (
	MinOrgSlugLength = 2
	MaxOrgSlugLength = 50
	orgRolePrefix    = "org-"
)

var orgSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// CourseIdentifier builds the "<semesterTag>-<name>" prefix shared by a course's
// Keycloak group and role names.
func CourseIdentifier(semesterTag, name string) string {
	return semesterTag + "-" + name
}

// CourseRoleName builds the full "<semesterTag>-<name>-<role>" Keycloak role string
// from a course identifier (see CourseIdentifier) and a course role such as CourseLecturer.
func CourseRoleName(courseIdentifier, role string) string {
	return courseIdentifier + "-" + role
}

func OrgRoleName(slug, role string) string {
	return orgRolePrefix + slug + "-" + role
}

// ParseOrgRole is the inverse of OrgRoleName: it accepts "org-<slug>-Admin" and
// "org-<slug>-Member" with a valid org slug, and nothing else. A prefix and suffix match is not
// enough, because the custom-group role "org-<course>-cg-Admin" of a course in semester "org" has
// both; the slug rules forbid the "cg" segment that tells the two apart. Slugs may contain hyphens,
// so the role is never split on "-".
func ParseOrgRole(role string) (slug, orgRole string, ok bool) {
	rest, found := strings.CutPrefix(role, orgRolePrefix)
	if !found {
		return "", "", false
	}
	for _, candidate := range []string{OrgAdmin, OrgMember} {
		if slug, found := strings.CutSuffix(rest, "-"+candidate); found && ValidateOrgSlug(slug) == nil {
			return slug, candidate, true
		}
	}
	return "", "", false
}

// ValidateOrgSlug checks the rules every org slug follows. They keep org roles apart from course
// and custom-group roles, see ParseOrgRole.
func ValidateOrgSlug(slug string) error {
	if slug == "" {
		return errors.New("slug is required")
	}
	if !orgSlugPattern.MatchString(slug) {
		return errors.New("slug may only contain lowercase letters and digits, separated by single hyphens")
	}
	if len(slug) < MinOrgSlugLength || len(slug) > MaxOrgSlugLength {
		return fmt.Errorf("slug must be between %d and %d characters", MinOrgSlugLength, MaxOrgSlugLength)
	}
	if slices.Contains(strings.Split(slug, "-"), "cg") {
		return errors.New(`slug must not contain a "cg" segment`)
	}
	return nil
}
