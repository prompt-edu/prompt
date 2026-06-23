export interface StaffMember {
  keycloakUserID: string
  username: string
  email: string
  firstName: string
  lastName: string
}

export interface CourseStaff {
  lecturers: StaffMember[]
  editors: StaffMember[]
}

export interface UserSearchResults {
  results: StaffMember[]
  truncated: boolean
}

export type CourseGroupName = 'Lecturer' | 'Editor'

export const staffMemberFullName = (m: StaffMember): string =>
  [m.firstName, m.lastName].filter(Boolean).join(' ') || m.username
