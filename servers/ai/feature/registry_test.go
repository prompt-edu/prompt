package feature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	suggestions, ok := Lookup("assessment.action_item_suggestions")
	assert.True(t, ok, "the assessment rule covers every assessment feature")
	assert.Equal(t, 730, suggestions.ContentRetentionDays)

	for _, name := range []string{"assessment", "assessment.", ".x", "interview.summary", "adhoc", ""} {
		_, ok := Lookup(name)
		assert.False(t, ok, "%q must not be registered", name)
	}
}

func TestPolicyOfUnknownFeatureIsStrictest(t *testing.T) {
	assert.Equal(t, Policy{ContentRetentionDays: 730, HighRisk: true}, PolicyOf("removed.feature"))
}

func TestSlug(t *testing.T) {
	assert.Equal(t, "assessment", Slug("Assessment"))
	assert.Equal(t, "self-team-allocation", Slug("Self Team Allocation"))
	assert.Equal(t, "intro-course-developer", Slug(" Intro  Course Developer "))
}

func TestCheckFor(t *testing.T) {
	assert.NoError(t, CheckFor("assessment.action_item_suggestions", "Assessment"))
	assert.ErrorIs(t, CheckFor("assessment.action_item_suggestions", "Interview"), ErrNotAllowed,
		"a feature must not be recorded under another phase type")
	assert.ErrorIs(t, CheckFor("interview.summary", "Interview"), ErrNotAllowed,
		"an unregistered feature has no retention, so it is refused")
	assert.ErrorIs(t, CheckFor("", "Assessment"), ErrNotAllowed, "a call must name its feature")
}
