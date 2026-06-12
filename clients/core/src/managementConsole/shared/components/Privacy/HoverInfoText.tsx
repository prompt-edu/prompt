import { Tooltip, TooltipContent, TooltipTrigger } from '@tumaet/prompt-ui-components'
import { ReactNode } from 'react'

interface HoverInfoTextProps {
  children: ReactNode
  content: ReactNode
}

export function HoverInfoText({ children, content }: HoverInfoTextProps) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className='cursor-default underline underline-offset-4'>{children}</span>
      </TooltipTrigger>
      <TooltipContent>{content}</TooltipContent>
    </Tooltip>
  )
}
