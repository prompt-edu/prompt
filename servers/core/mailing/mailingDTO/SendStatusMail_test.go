package mailingDTO

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendStatusMailRecipientFilterPresence(t *testing.T) {
	t.Run("omitted filter", func(t *testing.T) {
		var request SendStatusMail

		err := json.Unmarshal([]byte(`{"statusMailToBeSend":"passed"}`), &request)

		require.NoError(t, err)
		assert.Nil(t, request.RecipientCourseParticipationIDs)
	})

	t.Run("empty filter", func(t *testing.T) {
		var request SendStatusMail

		err := json.Unmarshal(
			[]byte(`{"statusMailToBeSend":"passed","recipientCourseParticipationIDs":[]}`),
			&request,
		)

		require.NoError(t, err)
		require.NotNil(t, request.RecipientCourseParticipationIDs)
		assert.Empty(t, *request.RecipientCourseParticipationIDs)
	})
}
