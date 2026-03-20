import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'
import { Send } from 'lucide-react'

const isMac = /Mac/i.test(navigator.platform)
const submitShortcut = isMac ? '⌘+Enter' : 'Ctrl+Enter'

interface InstructorNoteComposerSubmitButtonProps {
  onClick: () => void
  disabled: boolean
  isEditMode: boolean
}

export function InstructorNoteComposerSubmitButton({
  onClick,
  disabled,
  isEditMode,
}: InstructorNoteComposerSubmitButtonProps) {
  const buttonClass = `shrink-0 p-1 rounded transition-colors disabled:cursor-not-allowed ${
    disabled ? 'text-muted-foreground' : 'text-black dark:text-white'
  } hover:bg-muted`

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <button className={buttonClass} onClick={onClick} disabled={disabled}>
            <Send className='w-4 h-4' />
          </button>
        </TooltipTrigger>
        <TooltipContent>
          {isEditMode ? 'Save' : 'Send'}{' '}
          <span className='text-muted-foreground ml-1'>{submitShortcut}</span>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
