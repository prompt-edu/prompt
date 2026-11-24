package permissionValidation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOrgRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		wantSlug string
		wantRole string
	}{
		{name: "admin", role: "org-ase-Admin", wantSlug: "ase", wantRole: OrgAdmin},
		{name: "member", role: "org-ase-Member", wantSlug: "ase", wantRole: OrgMember},
		{name: "hyphenated slug", role: "org-ase-lab-Admin", wantSlug: "ase-lab", wantRole: OrgAdmin},
		{name: "slug ending in a role name", role: "org-team-admin-Member", wantSlug: "team-admin", wantRole: OrgMember},
		{name: "slug segment containing cg", role: "org-cgi-lab-Member", wantSlug: "cgi-lab", wantRole: OrgMember},
		{name: "longest slug", role: OrgRoleName(strings.Repeat("a", MaxOrgSlugLength), OrgAdmin), wantSlug: strings.Repeat("a", MaxOrgSlugLength), wantRole: OrgAdmin},

		{name: "custom-group role of a course in semester org", role: "org-ios-cg-Admin"},
		{name: "custom-group role named Member", role: "org-ios-cg-Member"},
		{name: "custom-group role with a hyphenated group name", role: "org-ios-cg-team-Admin"},
		{name: "course lecturer role in semester org", role: "org-ios-Lecturer"},
		{name: "course editor role in semester org", role: "org-ios-Editor"},
		{name: "course student role in semester org", role: "org-ios-Student"},
		{name: "course role of another semester", role: "ws25-ase-Lecturer"},
		{name: "custom-group role of another semester", role: "ws25-ase-cg-Admin"},
		{name: "global admin role", role: PromptAdmin},
		{name: "global lecturer role", role: PromptLecturer},
		{name: "no slug", role: "org--Admin"},
		{name: "prefix and suffix only", role: "org-Admin"},
		{name: "one-character slug", role: "org-a-Admin"},
		{name: "slug too long", role: OrgRoleName(strings.Repeat("a", MaxOrgSlugLength+1), OrgAdmin)},
		{name: "uppercase slug", role: "org-ASE-Admin"},
		{name: "course name with spaces", role: "org-iOS Praktikum-Admin"},
		{name: "lowercase role suffix", role: "org-ase-admin"},
		{name: "unknown role suffix", role: "org-ase-Owner"},
		{name: "missing prefix", role: "ase-Admin"},
		{name: "uppercase prefix", role: "Org-ase-Admin"},
		{name: "empty", role: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug, orgRole, ok := ParseOrgRole(tt.role)
			assert.Equal(t, tt.wantRole != "", ok)
			assert.Equal(t, tt.wantSlug, slug)
			assert.Equal(t, tt.wantRole, orgRole)
		})
	}
}

func TestParseOrgRoleInvertsOrgRoleName(t *testing.T) {
	for _, slug := range []string{"ase", "ase-lab", "i17", "a1-b2-c3"} {
		for _, orgRole := range []string{OrgAdmin, OrgMember} {
			parsedSlug, parsedRole, ok := ParseOrgRole(OrgRoleName(slug, orgRole))
			assert.True(t, ok)
			assert.Equal(t, slug, parsedSlug)
			assert.Equal(t, orgRole, parsedRole)
		}
	}
}
