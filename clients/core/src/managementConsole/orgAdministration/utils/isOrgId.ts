// Org ids are UUIDs, as the server parses them
const ORG_ID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

/**
 * Whether an id from the URL can name an org. Anything else is no org at all: requesting it could
 * even reach another endpoint, as `/orgs/self` lists the caller's own orgs.
 */
export const isOrgId = (orgId: string | undefined): orgId is string =>
  orgId !== undefined && ORG_ID_PATTERN.test(orgId)
