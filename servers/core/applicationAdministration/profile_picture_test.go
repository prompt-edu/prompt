package applicationAdministration

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Open application phases seeded in database_dumps/application_administration.sql.
var (
	phaseWithoutPictureSettings = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	phaseForPictureRequirement  = uuid.MustParse("d0000099-0000-0000-0000-000000000099")
)

func (suite *ApplicationAdminServiceTestSuite) TestApplicationForm_ProfilePictureDefaults() {
	form, err := suite.applicationAdminService.GetApplicationFormWithDetails(suite.ctx, phaseWithoutPictureSettings)

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), form.ApplicationPhase.ProfilePicture)
	assert.Equal(suite.T(), "off", form.ApplicationPhase.ProfilePicture.Requirement)
	assert.True(suite.T(), form.ApplicationPhase.ProfilePicture.HiddenUntilAccepted,
		"pictures are hidden unless a lecturer turned it off")
}

func (suite *ApplicationAdminServiceTestSuite) TestValidateProfilePictureRequirement() {
	conn := suite.applicationAdminService.conn
	_, err := conn.Exec(suite.ctx,
		`UPDATE course_phase SET restricted_data = restricted_data || '{"profilePictureRequirement": "required", "profilePictureExplanation": "For the team board"}'::jsonb WHERE id = $1`,
		phaseForPictureRequirement)
	require.NoError(suite.T(), err)
	defer func() {
		_, err := conn.Exec(suite.ctx,
			`UPDATE course_phase SET restricted_data = restricted_data - 'profilePictureRequirement' - 'profilePictureExplanation' WHERE id = $1`,
			phaseForPictureRequirement)
		require.NoError(suite.T(), err)
	}()

	form, err := suite.applicationAdminService.GetApplicationFormWithDetails(suite.ctx, phaseForPictureRequirement)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "required", form.ApplicationPhase.ProfilePicture.Requirement)
	assert.Equal(suite.T(), "For the team board", form.ApplicationPhase.ProfilePicture.Explanation)

	applicantWithoutPicture := uuid.New()
	err = suite.applicationAdminService.ValidateProfilePictureRequirement(suite.ctx, phaseForPictureRequirement, applicantWithoutPicture)
	assert.ErrorIs(suite.T(), err, ErrProfilePictureRequired)

	applicantWithPicture := uuid.New()
	fileID := uuid.New()
	_, err = conn.Exec(suite.ctx,
		`INSERT INTO files (id, filename, original_filename, content_type, size_bytes, storage_key, uploaded_by_user_id)
		 VALUES ($1, 'p.jpg', 'p.jpg', 'image/jpeg', 100, $2, $3)`,
		fileID, "profile-picture/"+applicantWithPicture.String()+"/p.jpg", applicantWithPicture.String())
	require.NoError(suite.T(), err)
	_, err = conn.Exec(suite.ctx, `INSERT INTO profile_picture (user_id, file_id) VALUES ($1, $2)`, applicantWithPicture, fileID)
	require.NoError(suite.T(), err)

	assert.NoError(suite.T(), suite.applicationAdminService.ValidateProfilePictureRequirement(suite.ctx, phaseForPictureRequirement, applicantWithPicture))

	// Without the requirement, nobody is held back
	assert.NoError(suite.T(), suite.applicationAdminService.ValidateProfilePictureRequirement(suite.ctx, phaseWithoutPictureSettings, applicantWithoutPicture))
}
