import type { UpdateOrg } from '@core/interfaces/org'
import { Input, Label } from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import type { OrgFormErrors } from '../utils/validateOrgForm'

interface OrgDetailsFieldsProps {
  idPrefix: string
  values: UpdateOrg
  errors: OrgFormErrors
  fields?: (keyof UpdateOrg)[]
  disabled?: boolean
  onChange: (values: UpdateOrg) => void
}

interface FieldConfig {
  field: keyof UpdateOrg
  label: string
  placeholder: string
  type?: string
}

const FIELDS: FieldConfig[] = [
  {
    field: 'name',
    label: 'Name',
    placeholder: 'e.g. Applied Education Technologies',
  },
  {
    field: 'school',
    label: 'School (optional)',
    placeholder: 'e.g. School of Computation, Information and Technology',
  },
  {
    field: 'university',
    label: 'University (optional)',
    placeholder: 'e.g. Technical University of Munich',
  },
  {
    field: 'website',
    label: 'Website (optional)',
    placeholder: 'https://',
    type: 'url',
  },
  {
    field: 'contactEmail',
    label: 'Contact Email (optional)',
    placeholder: 'e.g. chair@example.com',
    type: 'email',
  },
]

export const OrgDetailsFields = ({
  idPrefix,
  values,
  errors,
  fields,
  disabled = false,
  onChange,
}: OrgDetailsFieldsProps) => {
  const [editedFields, setEditedFields] = useState<(keyof UpdateOrg)[]>([])

  return (
    <>
      {FIELDS.filter(({ field }) => fields === undefined || fields.includes(field)).map(
        ({ field, label, placeholder, type }) => {
          const id = `${idPrefix}-${field}`
          const error =
            editedFields.includes(field) || values[field].trim() !== '' ? errors[field] : undefined
          return (
            <div key={field} className='flex flex-col gap-1.5'>
              <Label htmlFor={id}>{label}</Label>
              <Input
                id={id}
                type={type}
                value={values[field]}
                placeholder={placeholder}
                disabled={disabled}
                aria-invalid={error !== undefined}
                onChange={(e) => {
                  if (!editedFields.includes(field)) {
                    setEditedFields([...editedFields, field])
                  }
                  onChange({ ...values, [field]: e.target.value })
                }}
              />
              {error && <p className='text-sm text-destructive'>{error}</p>}
            </div>
          )
        },
      )}
    </>
  )
}
