package presentation

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
)

func TestRequestUserRecognizesGlobalStaffRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		roles       map[string]bool
		wantStaff   bool
		wantRelease bool
	}{
		{
			name:        "prompt admin can manage and release",
			roles:       map[string]bool{promptSDK.PromptAdmin: true},
			wantStaff:   true,
			wantRelease: true,
		},
		{
			name:        "prompt lecturer can edit but not release",
			roles:       map[string]bool{promptSDK.PromptLecturer: true},
			wantStaff:   true,
			wantRelease: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			keycloakTokenVerifier.SetTokenUser(context, keycloakTokenVerifier.TokenUser{
				ID:        "staff-user",
				Email:     "staff@example.com",
				FirstName: "Staff",
				LastName:  "Member",
				Roles:     test.roles,
			})

			user, err := requestUser(context)
			if err != nil {
				t.Fatalf("requestUser returned an error: %v", err)
			}
			if user.Staff != test.wantStaff {
				t.Fatalf("Staff = %v, want %v", user.Staff, test.wantStaff)
			}
			if user.CanRelease != test.wantRelease {
				t.Fatalf("CanRelease = %v, want %v", user.CanRelease, test.wantRelease)
			}
		})
	}
}
