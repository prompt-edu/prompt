import type { ReactNode } from 'react'

interface AnswerBlockProps {
  isEmpty?: boolean
  children?: ReactNode
}

// Wraps an answer in a soft panel with a subtle left accent so it reads as the
// primary content, clearly separated from the question. Empty answers get a
// lighter dashed placeholder to distinguish "not answered" from a real answer.
export const AnswerBlock = ({ isEmpty = false, children }: AnswerBlockProps) => {
  if (isEmpty) {
    return (
      <div className='rounded-md border border-dashed border-border px-3 py-2.5 text-sm italic text-muted-foreground'>
        Not answered
      </div>
    )
  }

  return (
    <div className='rounded-md border-l-2 border-primary/40 bg-muted/50 px-3 py-2.5 text-sm'>
      {children}
    </div>
  )
}
