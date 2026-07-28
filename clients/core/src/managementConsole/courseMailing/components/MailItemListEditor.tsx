import { Button, Input, Label } from '@tumaet/prompt-ui-components'
import { Plus, Trash2 } from 'lucide-react'
import type { MailItem } from '../interfaces/mailCampaign'

interface MailItemListEditorProps {
  label: string
  items: MailItem[]
  onChange: (items: MailItem[]) => void
}

export const MailItemListEditor = ({ label, items, onChange }: MailItemListEditorProps) => {
  const update = (index: number, patch: Partial<MailItem>) =>
    onChange(items.map((item, i) => (i === index ? { ...item, ...patch } : item)))
  const remove = (index: number) => onChange(items.filter((_, i) => i !== index))
  const add = () => onChange([...items, { name: '', email: '' }])

  return (
    <div className='space-y-2'>
      <Label>{label} override (optional)</Label>
      {items.map((item, index) => (
        <div key={`${label}-${index}`} className='flex gap-2'>
          <Input
            placeholder='email'
            value={item.email}
            onChange={(e) => update(index, { email: e.target.value })}
          />
          <Input
            placeholder='name (optional)'
            value={item.name}
            onChange={(e) => update(index, { name: e.target.value })}
          />
          <Button variant='ghost' size='icon' type='button' onClick={() => remove(index)}>
            <Trash2 className='h-4 w-4' />
          </Button>
        </div>
      ))}
      <Button variant='outline' size='sm' type='button' onClick={add}>
        <Plus className='mr-2 h-4 w-4' />
        Add {label}
      </Button>
    </div>
  )
}
