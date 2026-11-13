package keycloakRealmManager

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGroupMembersReturnsEveryMemberAcrossPages(t *testing.T) {
	for _, count := range []int{0, groupMembersPageSize - 1, groupMembersPageSize, 2*groupMembersPageSize + 1} {
		t.Run(fmt.Sprintf("%d members", count), func(t *testing.T) {
			userIDs := make([]string, count)
			for i := range userIDs {
				userIDs[i] = fmt.Sprintf("student-%03d", i)
			}
			keycloak := &fakeKeycloak{members: map[string][]string{"custom-group": userIDs}}

			members, err := keycloak.start(t).GetGroupMembers(context.Background(), "token", "custom-group")

			require.NoError(t, err)
			memberIDs := make([]string, len(members))
			for i, member := range members {
				memberIDs[i] = *member.ID
			}
			assert.Equal(t, userIDs, memberIDs)
		})
	}
}

func TestGetGroupMembersReturnsTheKeycloakError(t *testing.T) {
	keycloak := &fakeKeycloak{membersFail: true}

	members, err := keycloak.start(t).GetGroupMembers(context.Background(), "token", "custom-group")

	assert.Error(t, err)
	assert.Nil(t, members)
}
