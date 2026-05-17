package authDTO

type GetCourseRoles struct {
	CourseLecturerRole string `json:"courseLecturerRole"`
	CourseEditorRole   string `json:"courseEditorRole"`
	CustomRolePrefix   string `json:"customRolePrefix"`
}
