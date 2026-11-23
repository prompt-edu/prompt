package authDTO

type GetCourseRoles struct {
	CourseLecturerRole string `json:"courseLecturerRole"`
	CourseEditorRole   string `json:"courseEditorRole"`
	CustomRolePrefix   string `json:"customRolePrefix"`
	// OrgAdminRoles are the Admin roles of the course's org, whose holders are Lecturer-equivalent
	// on the course. Empty for an ungrouped course.
	OrgAdminRoles []string `json:"orgAdminRoles,omitempty"`
}
