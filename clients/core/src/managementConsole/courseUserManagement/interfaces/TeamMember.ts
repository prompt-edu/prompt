export interface TeamMember {
  keycloakUserID: string
  username: string
  email: string
  firstName: string
  lastName: string
}

export interface CourseTeam {
  lecturers: TeamMember[]
  editors: TeamMember[]
}

export interface UserSearchResults {
  results: TeamMember[]
  truncated: boolean
}

export type CourseGroupName = 'Lecturer' | 'Editor'
