// Seeded Keycloak users from keycloakConfig.json. Passwords are the literal
// dev credentials baked into the realm import (username === password for all).
// `permission` is the role that appears in resourceAccess['prompt-server'].roles.

export type Role =
  | 'admin'
  | 'lecturer'
  | 'course-lecturer'
  | 'course-editor'
  | 'student'
  | 'student2'
  | 'privacy-student'
  | 'org-admin'
  | 'org-member'

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
  // Second student enrolled in the seeded iPraktikum course, for two-user
  // journeys (e.g. one student creates a team, the other joins it).
  student2: {
    username: 'student2',
    password: 'student2',
    permission: 'PROMPT_Student',
    email: 'second_student@example.com',
  },
  // Exists only to be deleted by the privacy deletion-approval spec, which is the
  // one journey that genuinely removes its subject. Never use it elsewhere.
  'privacy-student': {
    username: 'privacy-student',
    password: 'privacy-student',
    permission: 'PROMPT_Student',
    email: 'privacy_subject@example.com',
  },
  // Admin and Member of the seeded demo org (SEEDED_ORG). They hold no other role,
  // as the chair's office typically does not.
  'org-admin': {
    username: 'org-admin',
    password: 'org-admin',
    permission: 'org-demo-chair-Admin',
    email: 'org_admin@example.com',
  },
  'org-member': {
    username: 'org-member',
    password: 'org-member',
    permission: 'org-demo-chair-Member',
    email: 'org_member@example.com',
  },
}

// Roles we pre-authenticate in global-setup (storageState reused by tests).
export const SEEDED_ROLES: Role[] = [
  'admin',
  'lecturer',
  'course-lecturer',
  'course-editor',
  'student',
  'student2',
  'privacy-student',
  'org-admin',
  'org-member',
]
