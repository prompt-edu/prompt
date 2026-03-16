import { getPermissionString, Role, useAuthStore } from '@tumaet/prompt-shared-state'
import { PropsWithChildren } from 'react'

type CourseRole = Extract<Role, Role.COURSE_LECTURER | Role.COURSE_EDITOR | Role.COURSE_STUDENT>

interface RoleCheckOptions {
  roles?: Role[]
  anyCourseRole?: CourseRole[]
}

/**
 * Pure permission check (no React dependencies)
 */
export function hasRolePermission(
  permissions: string[],
  { roles = [], anyCourseRole = [] }: RoleCheckOptions,
) {
  const globalAllowed = roles.some((r) => permissions.includes(getPermissionString(r)))

  const courseAllowed = anyCourseRole.some((r) => permissions.some((p) => p.endsWith(r)))

  return globalAllowed || courseAllowed
}

/**
 * Hook version for React components
 */
export function useHasRolePermission(options: RoleCheckOptions) {
  const { permissions } = useAuthStore()
  return hasRolePermission(permissions, options)
}

interface ShowForRoleProps extends PropsWithChildren {
  roles?: Role[]
  anyCourseRole?: CourseRole[]
}

/**
 * UI wrapper
 */
export function ShowForRole({ roles = [], anyCourseRole = [], children }: ShowForRoleProps) {
  const allowed = useHasRolePermission({ roles, anyCourseRole })

  return allowed ? <>{children}</> : null
}
