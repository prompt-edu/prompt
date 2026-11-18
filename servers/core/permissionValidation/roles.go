package permissionValidation

const PromptAdmin = "PROMPT_Admin"
const PromptLecturer = "PROMPT_Lecturer"
const CourseLecturer = "Lecturer"
const CourseEditor = "Editor"
const CourseStudent = "Student"
const OrgAdmin = "Admin"
const OrgMember = "Member"

// UnaddressableCharacters are the characters that keep Keycloak from addressing the groups and
// roles named after a course identifier that contains one. Keycloak splits a group path at every
// "/". The identifier goes unescaped into the admin URLs, which a "%" makes invalid and a "?" or
// "#" cuts short, so the lookup of "ws24-lab#1" reaches "ws24-lab".
const UnaddressableCharacters = "/%?#"

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

func OrgRoleName(slug, role string) string {
	return "org-" + slug + "-" + role
}
