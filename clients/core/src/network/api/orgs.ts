import type { CreateOrg, Org, UpdateOrg, UpdateOrgParent } from '@core/interfaces/org'
import { API_PREFIX, coreRequest } from '../client'

const path = `${API_PREFIX}/orgs`
const ofOrg = (orgID: string) => `${path}/${orgID}`

export const orgs = {
  list: (): Promise<Org[]> => coreRequest.get(`${path}/`),

  get: (orgID: string): Promise<Org> => coreRequest.get(ofOrg(orgID)),

  create: (org: CreateOrg): Promise<Org> => coreRequest.post(`${path}/`, org),

  update: (orgID: string, org: UpdateOrg): Promise<Org> => coreRequest.put(ofOrg(orgID), org),

  updateParent: (orgID: string, parent: UpdateOrgParent): Promise<Org> =>
    coreRequest.put(`${ofOrg(orgID)}/parent`, parent),

  remove: (orgID: string): Promise<void> => coreRequest.del(ofOrg(orgID)),
}
