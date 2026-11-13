import { ORG_ROLE_LABELS, ORG_ROLES, type OrgRole } from '@core/interfaces/org'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'

interface OrgRoleSelectProps {
  value: OrgRole
  ariaLabel: string
  disabled?: boolean
  className?: string
  onChange: (role: OrgRole) => void
}

export const OrgRoleSelect = ({
  value,
  ariaLabel,
  disabled,
  className,
  onChange,
}: OrgRoleSelectProps) => (
  <Select
    value={value}
    disabled={disabled}
    onValueChange={(selected) => {
      const role = ORG_ROLES.find((orgRole) => orgRole === selected)
      if (role) {
        onChange(role)
      }
    }}
  >
    <SelectTrigger aria-label={ariaLabel} className={className}>
      <SelectValue />
    </SelectTrigger>
    <SelectContent>
      {ORG_ROLES.map((role) => (
        <SelectItem key={role} value={role}>
          {ORG_ROLE_LABELS[role]}
        </SelectItem>
      ))}
    </SelectContent>
  </Select>
)
