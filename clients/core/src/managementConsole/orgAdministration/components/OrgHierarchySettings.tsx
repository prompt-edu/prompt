import type { Org } from '@core/interfaces/org'
import { useUpdateOrgParent } from '@core/network/hooks/useOrgs'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Label,
} from '@tumaet/prompt-ui-components'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { childOrgs, orgLabel, parentCandidates } from '../utils/orgHierarchy'
import { ParentOrgSelect } from './ParentOrgSelect'

interface OrgHierarchySettingsProps {
  org: Org
  orgs: Org[]
}

export const OrgHierarchySettings = ({ org, orgs }: OrgHierarchySettingsProps) => {
  const updateParent = useUpdateOrgParent()
  const [parentOrgID, setParentOrgID] = useState(org.parentOrgID)

  const candidates = useMemo(() => parentCandidates(orgs, org.id), [orgs, org.id])
  const children = useMemo(() => childOrgs(orgs, org.id), [orgs, org.id])
  const canSave = parentOrgID !== org.parentOrgID && !updateParent.isPending

  return (
    <Card>
      <CardHeader>
        <CardTitle>Hierarchy</CardTitle>
        <CardDescription>
          Where the org sits among the other orgs. The Admins and Members of a parent org get no
          rights on its child orgs.
        </CardDescription>
      </CardHeader>
      <CardContent className='flex flex-col gap-6'>
        <div className='flex flex-col gap-1.5'>
          <Label htmlFor='org-parent'>Parent Org</Label>
          <div className='flex gap-2'>
            <ParentOrgSelect
              id='org-parent'
              candidates={candidates}
              value={parentOrgID}
              disabled={updateParent.isPending}
              onChange={setParentOrgID}
            />
            <Button
              disabled={!canSave}
              onClick={() => updateParent.mutate({ orgId: org.id, parentOrgID })}
            >
              Move
            </Button>
          </div>
        </div>
        <div className='flex flex-col gap-1.5'>
          <h3 className='text-sm font-medium'>Child Orgs</h3>
          {children.length === 0 ? (
            <p className='text-sm text-muted-foreground'>This org has no child orgs.</p>
          ) : (
            <ul className='flex flex-col gap-1 text-sm'>
              {children.map((child) => (
                <li key={child.id}>
                  <Link
                    to={`/management/admin/orgs/${child.id}`}
                    className='underline underline-offset-4 hover:text-muted-foreground'
                  >
                    {orgLabel(child)}
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
