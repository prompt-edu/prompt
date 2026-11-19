import type { Org, UpdateOrg } from '@core/interfaces/org'
import { useUpdateOrg } from '@core/network/hooks/useOrgs'
import {
  Alert,
  AlertDescription,
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
import { type OrgDetailsField, rebaseOrgDetails } from '../utils/rebaseOrgDetails'
import { trimOrgDetails, validateOrgDetails } from '../utils/validateOrgForm'
import { OrgDetailsFields } from './OrgDetailsFields'

interface OrgDetailsSettingsProps {
  org: Org
  readOnly?: boolean
}

const toDetails = (org: Org): UpdateOrg => ({
  name: org.name,
  school: org.school ?? '',
  university: org.university ?? '',
  website: org.website ?? '',
  contactEmail: org.contactEmail ?? '',
})

const differ = (a: UpdateOrg, b: UpdateOrg): boolean =>
  (Object.keys(a) as (keyof UpdateOrg)[]).some((field) => a[field] !== b[field])

const FIELD_NAMES: Record<OrgDetailsField, string> = {
  name: 'name',
  school: 'school',
  university: 'university',
  website: 'website',
  contactEmail: 'contact email',
}

export const OrgDetailsSettings = ({ org, readOnly = false }: OrgDetailsSettingsProps) => {
  const updateOrg = useUpdateOrg()
  const saved = toDetails(org)
  const [details, setDetails] = useState(saved)
  // The org can change while the card is shown: a cached org renders first, and its refetch may bring
  // another Admin's edits. A field the draft left alone then shows the new value, so a save never
  // writes an old one back, and a field it edited keeps the edit.
  const [shown, setShown] = useState(saved)
  const [conflicts, setConflicts] = useState<OrgDetailsField[]>([])
  if (differ(saved, shown)) {
    const rebased = rebaseOrgDetails(shown, saved, details)
    setShown(saved)
    setDetails(rebased.draft)
    setConflicts((previous) => [...new Set([...previous, ...rebased.conflicts])])
  }

  const values = trimOrgDetails(details)
  const errors = validateOrgDetails(values)
  const isDirty = differ(saved, values)
  const canSave = isDirty && Object.keys(errors).length === 0 && !updateOrg.isPending
  // Only the conflicting fields the draft still overrides
  const openConflicts = conflicts.filter((field) => values[field] !== saved[field])

  return (
    <Card>
      <CardHeader>
        <CardTitle>Details</CardTitle>
        <CardDescription>
          {readOnly
            ? "The org's name and contact information. Admins of the org can change them."
            : "The org's name and contact information. Leaving an optional field empty clears it."}
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
          disabled={readOnly || updateOrg.isPending}
          onChange={setDetails}
        />
        {!readOnly && openConflicts.length > 0 && (
          <Alert>
            <AlertDescription className='flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
              <span>
                Someone else changed the{' '}
                {openConflicts.map((field) => FIELD_NAMES[field]).join(', ')} while you were
                editing. Saving replaces their changes with yours.
              </span>
              <Button
                variant='outline'
                size='sm'
                onClick={() => {
                  setDetails(saved)
                  setConflicts([])
                }}
              >
                Discard my changes
              </Button>
            </AlertDescription>
          </Alert>
        )}
        {!readOnly && (
          <div className='flex justify-end'>
            <Button
              disabled={!canSave}
              onClick={() =>
                updateOrg.mutate(
                  { orgId: org.id, org: values },
                  { onSuccess: () => setConflicts([]) },
                )
              }
            >
              Save
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
