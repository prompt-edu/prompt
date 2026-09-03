package copy

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/stretchr/testify/require"
)

func TestHandlePhaseCopyReturnsNoError(t *testing.T) {
	handler := &interviewCopyHandler{}
	c, _ := gin.CreateTestContext(nil)

	err := handler.HandlePhaseCopy(c, promptTypes.PhaseCopyRequest{})

	require.NoError(t, err)
}
