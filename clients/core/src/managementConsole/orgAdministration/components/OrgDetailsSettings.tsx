import type { Org, UpdateOrg } from '@core/interfaces/org'
import { useUpdateOrg } from '@core/network/hooks/useOrgs'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Input,
  Label,
} from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import { trimOrgDetails, validateOrgDetails } from '../utils/validateOrgForm'
import { OrgDetailsFields } from './OrgDetailsFields'

interface OrgDetailsSettingsProps {
  org: Org
}

const toDetails = (org: Org): UpdateOrg => ({
  name: org.name,
  school: org.school ?? '',
  university: org.university ?? '',
  website: org.website ?? '',
  contactEmail: org.contactEmail ?? '',
})

export const OrgDetailsSettings = ({ org }: OrgDetailsSettingsProps) => {
  const updateOrg = useUpdateOrg()
  const [details, setDetails] = useState(() => toDetails(org))

  const values = trimOrgDetails(details)
  const errors = validateOrgDetails(values)
  const saved = toDetails(org)
  const isDirty = (Object.keys(saved) as (keyof UpdateOrg)[]).some(
    (field) => values[field] !== saved[field],
  )
  const canSave = isDirty && Object.keys(errors).length === 0 && !updateOrg.isPending

  return (
    <Card>
      <CardHeader>
        <CardTitle>Details</CardTitle>
        <CardDescription>
          The org&apos;s name and contact information. Leaving an optional field empty clears it.
        </CardDescription>
      </CardHeader>
      <CardContent className='flex flex-col gap-4'>
        <div className='flex flex-col gap-1.5'>
          <Label htmlFor='org-slug'>Slug</Label>
          <Input id='org-slug' value={org.slug} disabled />
          <p className='text-sm text-muted-foreground'>
            Names the org&apos;s Keycloak roles <code>org-{org.slug}-Admin</code> and{' '}
            <code>org-{org.slug}-Member</code>, so it cannot be changed.
          </p>
        </div>
        <OrgDetailsFields
          idPrefix='org'
          values={details}
          errors={errors}
          disabled={updateOrg.isPending}
          onChange={setDetails}
        />
        <div className='flex justify-end'>
          <Button
            disabled={!canSave}
            onClick={() => updateOrg.mutate({ orgId: org.id, org: values })}
          >
            Save
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
