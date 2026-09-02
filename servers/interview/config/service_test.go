package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlePhaseConfigReturnsEmptyMap(t *testing.T) {
	handler := &configHandler{}

	config, err := handler.HandlePhaseConfig(nil)

	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, map[string]bool{}, config)
}
