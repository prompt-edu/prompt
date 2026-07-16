package coursePhaseTypeDTO

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequiredInputDTOsSerializeOptionalFlag(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		optional bool
	}{
		{
			name:     "participation input",
			input:    ParticipationInputDTO{Optional: true},
			optional: true,
		},
		{
			name:     "phase input",
			input:    PhaseInputDTO{Optional: false},
			optional: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			data, err := json.Marshal(testCase.input)
			require.NoError(t, err)

			var serialized map[string]any
			require.NoError(t, json.Unmarshal(data, &serialized))
			assert.Equal(t, testCase.optional, serialized["optional"])
		})
	}
}
