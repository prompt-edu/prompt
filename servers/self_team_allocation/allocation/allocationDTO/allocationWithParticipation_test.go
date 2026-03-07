package allocationDTO

import (
	"testing"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/assert"
)

func TestGetAllocationsFromDBModels(t *testing.T) {
	assignments := []db.Assignment{
		{
			CourseParticipationID: uuid.MustParse("aaaa1111-1111-1111-1111-111111111111"),
			TeamID:                uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		},
		{
			CourseParticipationID: uuid.MustParse("bbbb1111-1111-1111-1111-111111111111"),
			TeamID:                uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		},
	}

	result := GetAllocationsFromDBModels(assignments)

	assert.Len(t, result, 2)
	assert.Equal(t, uuid.MustParse("aaaa1111-1111-1111-1111-111111111111"), result[0].CourseParticipationID)
	assert.Equal(t, uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), result[0].TeamAllocation)
	assert.Equal(t, uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), result[1].TeamAllocation)
}
