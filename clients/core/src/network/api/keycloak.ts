import type {
  CourseGroupName,
  CourseStaff,
  UserSearchResults,
} from '@core/managementConsole/courseUserManagement/interfaces/StaffMember'
import type { KeycloakStatus } from '@core/managementConsole/pages/SystemStatusPage/interfaces/keycloakStatus'
import { API_PREFIX, coreRequest } from '../client'

const path = `${API_PREFIX}/keycloak`
const member = (courseID: string, groupName: CourseGroupName, keycloakUserID: string) =>
  `${path}/${courseID}/group/${groupName}/members/${keycloakUserID}`

const DEFAULT_SEARCH_LIMIT = 20

export const keycloak = {
  status: (): Promise<KeycloakStatus> => coreRequest.get(`${path}/status`),

  courseStaff: (courseID: string): Promise<CourseStaff> =>
    coreRequest.get(`${path}/${courseID}/group/staff`),

  searchUsers: (query: string, limit: number = DEFAULT_SEARCH_LIMIT): Promise<UserSearchResults> =>
    coreRequest.get(`${path}/users/search`, { params: { q: query, limit } }),

  addStaffMember: (
    courseID: string,
    groupName: CourseGroupName,
    keycloakUserID: string,
  ): Promise<void> => coreRequest.put(member(courseID, groupName, keycloakUserID)),

  removeStaffMember: (
    courseID: string,
    groupName: CourseGroupName,
    keycloakUserID: string,
  ): Promise<void> => coreRequest.del(member(courseID, groupName, keycloakUserID)),
}
