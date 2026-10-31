import type { UpdateOrg } from '@core/interfaces/org'
import { Input, Label } from '@tumaet/prompt-ui-components'
import {
  MAX_CONTACT_EMAIL_LENGTH,
  MAX_NAME_LENGTH,
  MAX_SCHOOL_LENGTH,
  MAX_UNIVERSITY_LENGTH,
  MAX_WEBSITE_LENGTH,
  type OrgFormErrors,
} from '../utils/validateOrgForm'

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
  maxLength: number
  placeholder: string
  type?: string
}

const FIELDS: FieldConfig[] = [
  {
    field: 'name',
    label: 'Name',
    maxLength: MAX_NAME_LENGTH,
    placeholder: 'e.g. Applied Education Technologies',
  },
  {
    field: 'school',
    label: 'School (optional)',
    maxLength: MAX_SCHOOL_LENGTH,
    placeholder: 'e.g. School of Computation, Information and Technology',
  },
  {
    field: 'university',
    label: 'University (optional)',
    maxLength: MAX_UNIVERSITY_LENGTH,
    placeholder: 'e.g. Technical University of Munich',
  },
  {
    field: 'website',
    label: 'Website (optional)',
    maxLength: MAX_WEBSITE_LENGTH,
    placeholder: 'https://',
    type: 'url',
  },
  {
    field: 'contactEmail',
    label: 'Contact Email (optional)',
    maxLength: MAX_CONTACT_EMAIL_LENGTH,
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
}: OrgDetailsFieldsProps) => (
  <>
    {FIELDS.filter(({ field }) => fields === undefined || fields.includes(field)).map(
      ({ field, label, maxLength, placeholder, type }) => {
        const id = `${idPrefix}-${field}`
        const error = values[field].trim() === '' ? undefined : errors[field]
        return (
          <div key={field} className='flex flex-col gap-1.5'>
            <Label htmlFor={id}>{label}</Label>
            <Input
              id={id}
              type={type}
              value={values[field]}
              maxLength={maxLength}
              placeholder={placeholder}
              disabled={disabled}
              aria-invalid={error !== undefined}
              onChange={(e) => onChange({ ...values, [field]: e.target.value })}
            />
            {error && <p className='text-sm text-destructive'>{error}</p>}
          </div>
        )
      },
    )}
  </>
)
