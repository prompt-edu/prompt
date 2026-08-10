package keycloakRealmManager

import "testing"

func TestIsCourseClientRole(t *testing.T) {
	const group = "ws24-ios"
	cases := []struct {
		name     string
		roleName string
		want     bool
	}{
		{"lecturer role", group + "-Lecturer", true},
		{"editor role", group + "-Editor", true},
		{"custom group role", group + "-cg-teamA", true},
		{"unrelated role", "PromptAdmin", false},
		{"other course lecturer", "ws24-android-Lecturer", false},
		{"prefix course is not matched", "ws24-ios-advanced-Lecturer", false},
		{"prefix course custom role is not matched", "ws24-ios-advanced-cg-teamA", false},
		{"bare group name", group, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isCourseClientRole(tc.roleName, group); got != tc.want {
				t.Errorf("isCourseClientRole(%q, %q) = %v, want %v", tc.roleName, group, got, tc.want)
			}
		})
	}
}
