import { Checkbox } from '@tumaet/prompt-ui-components'

interface TableCheckboxProps {
  checked: boolean
  onToggle: (arg0: boolean) => void
  location: 'row' | 'header'
}

export const TableCheckbox = ({ checked, onToggle, location }: TableCheckboxProps) => {
  return (
    <div className='h-full flex flex-col items-center'>
      <Checkbox
        checked={checked}
        onCheckedChange={onToggle}
        aria-label={location == 'row' ? 'Select row' : 'Select all'}
        onClick={(e) => e.stopPropagation()}
        onFocus={(e) => e.stopPropagation()}
      />
    </div>
  )
}
