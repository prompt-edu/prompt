// Seeded Keycloak users from keycloakConfig.json. Passwords are the literal
// dev credentials baked into the realm import (username === password for all).
// `permission` is the role that appears in resourceAccess['prompt-server'].roles.

export type Role =
  | 'admin'
  | 'lecturer'
  | 'course-lecturer'
  | 'course-editor'
  | 'student'

export interface RoleAccount {
  username: string
  password: string
  permission: string
  email: string
}

export const ROLES: Record<Role, RoleAccount> = {
  admin: {
    username: 'admin',
    password: 'admin',
    permission: 'PROMPT_Admin',
    email: 'admin@example.com',
  },
  lecturer: {
    username: 'lecturer',
    password: 'lecturer',
    permission: 'PROMPT_Lecturer',
    email: 'lecturer@example.com',
  },
  'course-lecturer': {
    username: 'course-lecturer',
    password: 'course-lecturer',
    permission: 'PROMPT_Course_Lecturer',
    email: 'of_course-lecturer@example.com',
  },
  'course-editor': {
    username: 'course-editor',
    password: 'course-editor',
    permission: 'PROMPT_Course_Editor',
    email: 'best_edits@example.com',
  },
  student: {
    username: 'student',
    password: 'student',
    permission: 'PROMPT_Student',
    email: 'pgdp_enjoyer@example.com',
  },
}

// Roles we pre-authenticate in global-setup (storageState reused by tests).
export const SEEDED_ROLES: Role[] = [
  'admin',
  'lecturer',
  'course-lecturer',
  'course-editor',
  'student',
]
