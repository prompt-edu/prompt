import type { Org } from '@core/interfaces/org'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { orgLabel } from '../utils/orgHierarchy'

interface ParentOrgSelectProps {
  id: string
  candidates: Org[]
  value: string | null
  disabled?: boolean
  onChange: (parentOrgID: string | null) => void
}

const NO_PARENT = 'none'

export const ParentOrgSelect = ({
  id,
  candidates,
  value,
  disabled = false,
  onChange,
}: ParentOrgSelectProps) => (
  <Select
    value={value ?? NO_PARENT}
    disabled={disabled}
    onValueChange={(selected) => onChange(selected === NO_PARENT ? null : selected)}
  >
    <SelectTrigger id={id}>
      <SelectValue />
    </SelectTrigger>
    <SelectContent>
      <SelectItem value={NO_PARENT}>None (top level)</SelectItem>
      {candidates.map((org) => (
        <SelectItem key={org.id} value={org.id}>
          {orgLabel(org)}
        </SelectItem>
      ))}
    </SelectContent>
  </Select>
)
