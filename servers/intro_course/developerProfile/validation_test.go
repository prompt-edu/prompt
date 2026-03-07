package developerProfile

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/intro_course/developerProfile/developerProfileDTO"
	"github.com/stretchr/testify/assert"
)

func TestIsValidUDID(t *testing.T) {
	validUDID := pgtype.Text{String: "ABCDEF12-34567890ABCDEF12", Valid: true}
	invalidUDID := pgtype.Text{String: "INVALID-UDID", Valid: true}
	emptyUDID := pgtype.Text{Valid: false}

	assert.True(t, isValidUDID(validUDID))
	assert.True(t, isValidUDID(emptyUDID))
	assert.False(t, isValidUDID(invalidUDID))
}

func TestValidateDeveloperProfileUDIDs(t *testing.T) {
	err := validateDeveloperProfileUDIDs(profileWithUDIDs("AAAABBBB-CCCCDDDDEEEEFFFF", "11112222-3333444455556666", "12345678-90ABCDEF12345678"))
	assert.NoError(t, err)

	err = validateDeveloperProfileUDIDs(profileWithUDIDs("invalid", "", ""))
	assert.Error(t, err)
}

func profileWithUDIDs(iPhone, iPad, watch string) developerProfileDTO.PostDeveloperProfile {
	profile := developerProfileDTO.PostDeveloperProfile{
		IPhoneUDID: pgtype.Text{String: iPhone, Valid: iPhone != ""},
		IPadUDID:   pgtype.Text{String: iPad, Valid: iPad != ""},
	}
	if watch != "" {
		profile.AppleWatchUDID = pgtype.Text{String: watch, Valid: true}
	}
	return profile
}
