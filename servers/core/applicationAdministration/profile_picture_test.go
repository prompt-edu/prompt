package applicationAdministration

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Seeded in database_dumps/application_administration.sql.
var (
	phaseWithoutPictureSettings = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	phaseWithRequiredPicture    = uuid.MustParse("d0000093-0000-0000-0000-000000000093")
	userWithPictureWithoutLogin = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000093")
)

func (suite *ApplicationAdminServiceTestSuite) TestApplicationForm_ProfilePictureDefaults() {
	form, err := suite.applicationAdminService.GetApplicationFormWithDetails(suite.ctx, phaseWithoutPictureSettings)

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), form.ApplicationPhase.ProfilePicture)
	assert.Equal(suite.T(), "off", form.ApplicationPhase.ProfilePicture.Requirement)
	assert.True(suite.T(), form.ApplicationPhase.ProfilePicture.HiddenUntilAccepted,
		"pictures are hidden unless a lecturer turned it off")
}

func (suite *ApplicationAdminServiceTestSuite) TestApplicationForm_ProfilePictureSettings() {
	form, err := suite.applicationAdminService.GetApplicationFormWithDetails(suite.ctx, phaseWithRequiredPicture)

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), form.ApplicationPhase.ProfilePicture)
	assert.Equal(suite.T(), "required", form.ApplicationPhase.ProfilePicture.Requirement)
	assert.Equal(suite.T(), "For the team board", form.ApplicationPhase.ProfilePicture.Explanation)
	assert.False(suite.T(), form.ApplicationPhase.ProfilePicture.HiddenUntilAccepted)
}

func (suite *ApplicationAdminServiceTestSuite) TestValidateProfilePictureRequirement() {
	applicantWithoutPicture := uuid.New()

	err := suite.applicationAdminService.ValidateProfilePictureRequirement(suite.ctx, phaseWithRequiredPicture, applicantWithoutPicture)
	assert.ErrorIs(suite.T(), err, ErrProfilePictureRequired)

	assert.NoError(suite.T(), suite.applicationAdminService.ValidateProfilePictureRequirement(suite.ctx, phaseWithRequiredPicture, userWithPictureWithoutLogin))

	assert.NoError(suite.T(), suite.applicationAdminService.ValidateProfilePictureRequirement(suite.ctx, phaseWithoutPictureSettings, applicantWithoutPicture),
		"without the requirement, nobody is held back")
}

func (suite *ApplicationAdminServiceTestSuite) TestSyncProfilePictureUniversityLogin() {
	storedLogin := func() pgtype.Text {
		picture, err := suite.applicationAdminService.queries.GetProfilePictureByUserID(suite.ctx, userWithPictureWithoutLogin)
		require.NoError(suite.T(), err)
		return picture.UniversityLogin
	}

	require.NoError(suite.T(), suite.applicationAdminService.SyncProfilePictureUniversityLogin(suite.ctx, userWithPictureWithoutLogin, ""))
	assert.False(suite.T(), storedLogin().Valid, "an account without a login has nothing to link")

	require.NoError(suite.T(), suite.applicationAdminService.SyncProfilePictureUniversityLogin(suite.ctx, userWithPictureWithoutLogin, "ap10pli"))
	assert.Equal(suite.T(), pgtype.Text{String: "ap10pli", Valid: true}, storedLogin())

	assert.NoError(suite.T(), suite.applicationAdminService.SyncProfilePictureUniversityLogin(suite.ctx, uuid.New(), "ap11pli"),
		"applicants without a picture are not an error")
}
