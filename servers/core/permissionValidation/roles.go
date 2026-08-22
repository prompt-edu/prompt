package permissionValidation

const PromptAdmin = "PROMPT_Admin"
const PromptLecturer = "PROMPT_Lecturer"
const CourseLecturer = "Lecturer"
const CourseEditor = "Editor"
const CourseStudent = "Student"

// CourseIdentifier builds the "<semesterTag>-<name>" prefix shared by a course's
// Keycloak group and role names.
func CourseIdentifier(semesterTag, name string) string {
	return semesterTag + "-" + name
}

// CourseRoleName builds the full "<semesterTag>-<name>-<role>" Keycloak role string
// from a course identifier (see CourseIdentifier) and a course role such as CourseLecturer.
func CourseRoleName(courseIdentifier, role string) string {
	return courseIdentifier + "-" + role
}
