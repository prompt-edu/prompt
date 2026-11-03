export interface CourseStaffMember {
  keycloakUserID: string
  username: string
  email: string
  firstName: string
  lastName: string
  universityLogin: string
}

export interface CourseStaff {
  lecturers: CourseStaffMember[]
  editors: CourseStaffMember[]
}
