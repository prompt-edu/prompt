import type { CreateOrg, Org, UpdateOrg } from '@core/interfaces/org'
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
} from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import { deriveOrgSlug } from '../utils/deriveOrgSlug'
import { MAX_SLUG_LENGTH, trimOrgDetails, validateCreateOrg } from '../utils/validateOrgForm'
import { OrgDetailsFields } from './OrgDetailsFields'
import { ParentOrgSelect } from './ParentOrgSelect'

interface CreateOrgDialogProps {
  orgs: Org[]
  isPending: boolean
  onClose: () => void
  onSubmit: (org: CreateOrg) => void
}

const OPTIONAL_FIELDS: (keyof UpdateOrg)[] = ['school', 'university', 'website', 'contactEmail']

const EMPTY_DETAILS: UpdateOrg = {
  name: '',
  school: '',
  university: '',
  website: '',
  contactEmail: '',
}

export const CreateOrgDialog = ({ orgs, isPending, onClose, onSubmit }: CreateOrgDialogProps) => {
  const [details, setDetails] = useState(EMPTY_DETAILS)
  const [editedSlug, setEditedSlug] = useState<string | null>(null)
  const [parentOrgID, setParentOrgID] = useState<string | null>(null)

  const slugInput = editedSlug ?? deriveOrgSlug(details.name)
  const values: CreateOrg = { ...trimOrgDetails(details), slug: slugInput.trim(), parentOrgID }
  const errors = validateCreateOrg(values)
  const isValid = Object.keys(errors).length === 0

  return (
    <Dialog open onOpenChange={(isOpen) => !isOpen && onClose()}>
      <DialogContent className='sm:max-w-xl max-h-[90vh] overflow-y-auto'>
        <DialogHeader>
          <DialogTitle>Add Org</DialogTitle>
          <DialogDescription>
            Creating an org also creates its Admin and Member roles in Keycloak. Nobody holds them
            yet, including you.
          </DialogDescription>
        </DialogHeader>
        <div className='flex flex-col gap-4 py-2'>
          <OrgDetailsFields
            idPrefix='create-org'
            values={details}
            errors={errors}
            fields={['name']}
            disabled={isPending}
            onChange={setDetails}
          />
          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='create-org-slug'>Slug</Label>
            <Input
              id='create-org-slug'
              value={slugInput}
              maxLength={MAX_SLUG_LENGTH}
              placeholder='e.g. aet'
              disabled={isPending}
              aria-invalid={values.slug !== '' && errors.slug !== undefined}
              onChange={(e) => setEditedSlug(e.target.value)}
            />
            {values.slug !== '' && errors.slug ? (
              <p className='text-sm text-destructive'>{errors.slug}</p>
            ) : (
              <p className='text-sm text-muted-foreground'>
                Names the org&apos;s Keycloak roles, such as{' '}
                <code>org-{values.slug || '<slug>'}-Admin</code>. It cannot be changed later.
              </p>
            )}
          </div>
          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='create-org-parent'>Parent Org (optional)</Label>
            <ParentOrgSelect
              id='create-org-parent'
              candidates={orgs}
              value={parentOrgID}
              disabled={isPending}
              onChange={setParentOrgID}
            />
          </div>
          <OrgDetailsFields
            idPrefix='create-org'
            values={details}
            errors={errors}
            fields={OPTIONAL_FIELDS}
            disabled={isPending}
            onChange={setDetails}
          />
        </div>
        <DialogFooter>
          <Button variant='outline' onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={() => onSubmit(values)} disabled={!isValid || isPending}>
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
