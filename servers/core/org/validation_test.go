package org

import (
	"strings"
	"testing"

	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	"github.com/stretchr/testify/assert"
)

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name      string
		slug      string
		wantError bool
	}{
		{name: "single segment", slug: "ase"},
		{name: "hyphenated", slug: "ase-lab"},
		{name: "digits", slug: "i17"},
		{name: "surrounding spaces", slug: "  ase  "},
		{name: "minimum length", slug: strings.Repeat("a", minSlugLength)},
		{name: "maximum length", slug: strings.Repeat("a", maxSlugLength)},
		{name: "segment containing cg", slug: "cgi-lab"},
		{name: "segment ending in cg", slug: "lab-xcg"},
		{name: "empty", slug: "", wantError: true},
		{name: "blank", slug: "   ", wantError: true},
		{name: "too short", slug: "a", wantError: true},
		{name: "too long", slug: strings.Repeat("a", maxSlugLength+1), wantError: true},
		{name: "uppercase", slug: "ASE", wantError: true},
		{name: "leading hyphen", slug: "-ase", wantError: true},
		{name: "trailing hyphen", slug: "ase-", wantError: true},
		{name: "double hyphen", slug: "ase--lab", wantError: true},
		{name: "underscore", slug: "ase_lab", wantError: true},
		{name: "slash", slug: "ase/lab", wantError: true},
		{name: "non-ascii", slug: "läb", wantError: true},
		{name: "cg only", slug: "cg", wantError: true},
		{name: "leading cg segment", slug: "cg-lab", wantError: true},
		{name: "inner cg segment", slug: "ase-cg-lab", wantError: true},
		{name: "trailing cg segment", slug: "ase-cg", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSlug(tt.slug)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOrgDetails(t *testing.T) {
	valid := orgDTO.UpdateOrg{
		Name:         "Applied Software Engineering",
		School:       "School of Computation, Information and Technology",
		University:   "Technical University of Munich",
		Website:      "https://aet.cit.tum.de",
		ContactEmail: "office@example.com",
	}
	with := func(change func(*orgDTO.UpdateOrg)) orgDTO.UpdateOrg {
		input := valid
		change(&input)
		return input
	}

	tests := []struct {
		name      string
		input     orgDTO.UpdateOrg
		wantError bool
	}{
		{name: "all fields", input: valid},
		{name: "name only", input: orgDTO.UpdateOrg{Name: "Applied Software Engineering"}},
		{name: "surrounding spaces", input: with(func(o *orgDTO.UpdateOrg) {
			o.Website = "  https://example.com  "
			o.ContactEmail = " office@example.com "
		})},
		{name: "http website", input: with(func(o *orgDTO.UpdateOrg) { o.Website = "http://example.com/chair?tab=team" })},
		{name: "website with a fragment after the host", input: with(func(o *orgDTO.UpdateOrg) { o.Website = "https://aet.cit.tum.de#contact" })},
		{name: "name at the limit", input: with(func(o *orgDTO.UpdateOrg) { o.Name = strings.Repeat("ä", maxNameLength) })},
		{name: "school at the limit", input: with(func(o *orgDTO.UpdateOrg) { o.School = strings.Repeat("a", maxSchoolLength) })},
		{name: "university at the limit", input: with(func(o *orgDTO.UpdateOrg) { o.University = strings.Repeat("a", maxUniversityLength) })},
		{name: "empty name", input: with(func(o *orgDTO.UpdateOrg) { o.Name = "" }), wantError: true},
		{name: "blank name", input: with(func(o *orgDTO.UpdateOrg) { o.Name = "   " }), wantError: true},
		{name: "name too long", input: with(func(o *orgDTO.UpdateOrg) { o.Name = strings.Repeat("a", maxNameLength+1) }), wantError: true},
		{name: "school too long", input: with(func(o *orgDTO.UpdateOrg) { o.School = strings.Repeat("a", maxSchoolLength+1) }), wantError: true},
		{name: "university too long", input: with(func(o *orgDTO.UpdateOrg) { o.University = strings.Repeat("a", maxUniversityLength+1) }), wantError: true},
		{name: "NUL in the name", input: with(func(o *orgDTO.UpdateOrg) { o.Name = "Applied\x00Software" }), wantError: true},
		{name: "NUL in the school", input: with(func(o *orgDTO.UpdateOrg) { o.School = "CIT\x00" }), wantError: true},
		{name: "website too long", input: with(func(o *orgDTO.UpdateOrg) { o.Website = "https://example.com/" + strings.Repeat("a", maxWebsiteLength) }), wantError: true},
		{name: "website without scheme", input: with(func(o *orgDTO.UpdateOrg) { o.Website = "example.com" }), wantError: true},
		{name: "website with other scheme", input: with(func(o *orgDTO.UpdateOrg) { o.Website = "ftp://example.com" }), wantError: true},
		{name: "website without host", input: with(func(o *orgDTO.UpdateOrg) { o.Website = "https://" }), wantError: true},
		{name: "website with only a port", input: with(func(o *orgDTO.UpdateOrg) { o.Website = "https://:443" }), wantError: true},
		{name: "javascript website", input: with(func(o *orgDTO.UpdateOrg) { o.Website = "javascript:alert(1)" }), wantError: true},
		{name: "email without domain", input: with(func(o *orgDTO.UpdateOrg) { o.ContactEmail = "office" }), wantError: true},
		{name: "email with display name", input: with(func(o *orgDTO.UpdateOrg) { o.ContactEmail = "Office <office@example.com>" }), wantError: true},
		{name: "email in angle brackets", input: with(func(o *orgDTO.UpdateOrg) { o.ContactEmail = "<office@example.com>" }), wantError: true},
		{name: "email too long", input: with(func(o *orgDTO.UpdateOrg) {
			o.ContactEmail = strings.Repeat("a", maxContactEmailLength) + "@example.com"
		}), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdateOrg(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCreateOrgChecksSlugAndDetails(t *testing.T) {
	assert.NoError(t, validateCreateOrg(orgDTO.CreateOrg{Name: "Applied Software Engineering", Slug: "ase"}))
	assert.Error(t, validateCreateOrg(orgDTO.CreateOrg{Name: "Applied Software Engineering", Slug: "ase-cg"}))
	assert.Error(t, validateCreateOrg(orgDTO.CreateOrg{Name: "", Slug: "ase"}))
}
