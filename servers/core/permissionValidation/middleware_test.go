package permissionValidation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func statusForRoles(userRoles map[string]bool, allowedRoles ...string) int {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		c.Set("userRoles", userRoles)
	}, CheckAccessControlByRole(allowedRoles...), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	return w.Code
}

func TestCheckAccessControlByRoleMatchesOrgRolesExactly(t *testing.T) {
	tests := []struct {
		name         string
		userRoles    map[string]bool
		allowedRoles []string
		want         int
	}{
		{"org admin for the admin marker", map[string]bool{"org-ase-Admin": true}, []string{OrgAdmin}, http.StatusOK},
		{"org member for the member marker", map[string]bool{"org-ase-lab-Member": true}, []string{OrgMember}, http.StatusOK},
		{"org admin for either marker", map[string]bool{"org-ase-Admin": true}, []string{OrgAdmin, OrgMember}, http.StatusOK},
		{"org member for the admin marker", map[string]bool{"org-ase-Member": true}, []string{OrgAdmin}, http.StatusForbidden},
		{"org admin for the member marker", map[string]bool{"org-ase-Admin": true}, []string{OrgMember}, http.StatusForbidden},
		{"custom-group role ending in Admin", map[string]bool{"org-ios-cg-Admin": true}, []string{OrgAdmin}, http.StatusForbidden},
		{"custom-group role ending in Member", map[string]bool{"ws25-ios-cg-Member": true}, []string{OrgMember}, http.StatusForbidden},
		{"global admin for the admin marker", map[string]bool{PromptAdmin: true}, []string{OrgAdmin}, http.StatusForbidden},
		{"a role that is not held", map[string]bool{"org-ase-Admin": false}, []string{OrgAdmin}, http.StatusForbidden},
		{"org roles never count as course roles", map[string]bool{"org-ase-Admin": true, "org-ase-Member": true}, []string{CourseLecturer, CourseEditor, CourseStudent}, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, statusForRoles(tt.userRoles, tt.allowedRoles...))
		})
	}
}
