import type {
  CreateOrg,
  Org,
  OrgMember,
  OwnOrg,
  UpdateOrg,
  UpdateOrgMemberRole,
  UpdateOrgParent,
} from '@core/interfaces/org'
import { API_PREFIX, coreRequest } from '../client'

const path = `${API_PREFIX}/orgs`
const ofOrg = (orgID: string) => `${path}/${orgID}`

export const orgs = {
  list: (): Promise<Org[]> => coreRequest.get(`${path}/`),

  own: (): Promise<OwnOrg[]> => coreRequest.get(`${path}/self`),

  get: (orgID: string): Promise<Org> => coreRequest.get(ofOrg(orgID)),

  create: (org: CreateOrg): Promise<Org> => coreRequest.post(`${path}/`, org),

  update: (orgID: string, org: UpdateOrg): Promise<Org> => coreRequest.put(ofOrg(orgID), org),

  updateParent: (orgID: string, parent: UpdateOrgParent): Promise<Org> =>
    coreRequest.put(`${ofOrg(orgID)}/parent`, parent),

  remove: (orgID: string): Promise<void> => coreRequest.del(ofOrg(orgID)),

  members: (orgID: string): Promise<OrgMember[]> => coreRequest.get(`${ofOrg(orgID)}/members`),

  setMemberRole: (
    orgID: string,
    keycloakUserID: string,
    role: UpdateOrgMemberRole,
  ): Promise<void> => coreRequest.put(`${ofOrg(orgID)}/members/${keycloakUserID}`, role),

  removeMember: (orgID: string, keycloakUserID: string): Promise<void> =>
    coreRequest.del(`${ofOrg(orgID)}/members/${keycloakUserID}`),
}
