import { Label, Switch } from '@tumaet/prompt-ui-components'

interface SettingsSwitchFieldProps {
  checked: boolean
  onCheckedChange?: (checked: boolean) => void
  title: string
  description: string
  disabled?: boolean
}

export const SettingsSwitchField = ({
  checked,
  onCheckedChange,
  title,
  description,
  disabled = false,
}: SettingsSwitchFieldProps) => {
  const switchId = title.toLowerCase().replace(/[^a-z0-9]+/g, '-')

  return (
    <div className='flex items-start justify-between gap-4 rounded-xl border border-border bg-muted/40 p-4'>
      <div className='space-y-1'>
        <Label htmlFor={switchId} className='text-sm font-semibold text-foreground'>
          {title}
        </Label>
        <p className='text-sm text-muted-foreground'>{description}</p>
      </div>
      <Switch
        id={switchId}
        checked={checked}
        onCheckedChange={onCheckedChange}
        disabled={disabled}
      />
    </div>
  )
}
