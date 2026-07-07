import { Button, Textarea } from '@tumaet/prompt-ui-components'
import { Trash2 } from 'lucide-react'
import { useEffect, useRef } from 'react'

import type { ActionItem } from '../../interfaces/actionItem'
import type { FeedbackItem } from '../../interfaces/feedbackItem'

interface BaseItemRowProps {
  value: string
  onTextChange: (itemId: string, value: string) => void
  onTextBlur?: (itemId: string) => void
  onDelete: (itemId: string) => void
  isSaving: boolean
  isPending: boolean
  isDisabled?: boolean
  placeholder?: string
}

interface ActionItemRowProps extends BaseItemRowProps {
  type: 'action'
  item: ActionItem
}

interface FeedbackItemRowProps extends BaseItemRowProps {
  type: 'feedback'
  item: FeedbackItem
}

type ItemRowProps = ActionItemRowProps | FeedbackItemRowProps

export function ItemRow({
  type,
  item,
  value,
  onTextChange,
  onTextBlur,
  onDelete,
  isPending,
  isDisabled = false,
  placeholder,
}: ItemRowProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
    }
  }, [value])

  const getDefaultPlaceholder = () => {
    if (placeholder) return placeholder
    return type === 'action' ? 'Enter action item...' : 'Enter feedback...'
  }

  const getDeleteTitle = () => {
    if (isDisabled) return 'Assessment completed - editing disabled'
    return type === 'action' ? 'Delete action item' : 'Delete feedback item'
  }

  return (
    <div
      className={`flex items-center gap-2 rounded-md group relative ${
        isDisabled ? 'opacity-60' : ''
      }`}
    >
      <div className='flex-1 relative'>
        <Textarea
          ref={textareaRef}
          className='w-full resize-none min-h-[24px] overflow-hidden'
          rows={1}
          value={value}
          onChange={(e) => {
            if (!isDisabled) {
              const cleanup = onTextChange(item.id, e.target.value)
              return cleanup
            }
          }}
          onBlur={() => {
            if (!isDisabled && onTextBlur) {
              onTextBlur(item.id)
            }
          }}
          placeholder={getDefaultPlaceholder()}
          readOnly={isDisabled}
        />
      </div>

      {!isDisabled && (
        <Button
          variant='ghost'
          size='icon'
          className=''
          onClick={() => onDelete(item.id)}
          disabled={isPending}
          title={getDeleteTitle()}
          aria-label={getDeleteTitle()}
        >
          <Trash2 className='h-4 w-4 text-destructive' aria-hidden />
        </Button>
      )}
    </div>
  )
}
