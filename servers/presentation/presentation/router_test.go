package presentation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"

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

// The auth middleware admits PROMPT lecturers before it ever looks up their course roles,
// so a PROMPT lecturer who also runs the course used to lose the release controls.
func TestResolveCourseLecturerFillsInTheCourseRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const courseLecturerRole = "ios2526-Lecturer"
	core := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(writer).Encode(map[string]string{
			"courseLecturerRole": courseLecturerRole,
			"courseEditorRole":   "ios2526-Editor",
			"customRolePrefix":   "ios2526-",
		})
	}))
	defer core.Close()
	coreURL, err := url.Parse(core.URL)
	if err != nil {
		t.Fatalf("could not parse the test core URL: %v", err)
	}

	previous := keycloakTokenVerifier.KeycloakTokenVerifierSingleton
	keycloakTokenVerifier.KeycloakTokenVerifierSingleton = &keycloakTokenVerifier.KeycloakTokenVerifier{CoreURL: *coreURL}
	t.Cleanup(func() { keycloakTokenVerifier.KeycloakTokenVerifierSingleton = previous })

	tests := []struct {
		name        string
		roles       map[string]bool
		wantRelease bool
	}{
		{
			name:        "prompt lecturer who also runs the course can release",
			roles:       map[string]bool{promptSDK.PromptLecturer: true, courseLecturerRole: true},
			wantRelease: true,
		},
		{
			name:        "prompt lecturer without the course role cannot release",
			roles:       map[string]bool{promptSDK.PromptLecturer: true},
			wantRelease: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			context.Params = gin.Params{{Key: "coursePhaseID", Value: uuid.NewString()}}
			keycloakTokenVerifier.SetTokenUser(context, keycloakTokenVerifier.TokenUser{
				ID:    "course-lecturer",
				Email: "lecturer@example.com",
				Roles: test.roles,
			})

			resolveCourseLecturer()(context)

			user, err := requestUser(context)
			if err != nil {
				t.Fatalf("requestUser returned an error: %v", err)
			}
			if user.CanRelease != test.wantRelease {
				t.Fatalf("CanRelease = %v, want %v", user.CanRelease, test.wantRelease)
			}
		})
	}
}
