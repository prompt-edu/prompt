import { useState } from 'react'
import { User, Users, Check, Copy } from 'lucide-react'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
  Button,
} from '@tumaet/prompt-ui-components'

import type { FeedbackItem } from '../../../interfaces/feedbackItem'
import { useTeamStore } from '../../../zustand/useTeamStore'

interface FeedbackItemRowProps {
  feedbackItem: FeedbackItem
}

export const FeedbackItemRow = ({ feedbackItem }: FeedbackItemRowProps) => {
  const { teams } = useTeamStore()
  const [showCopied, setShowCopied] = useState(false)

  const isSelfFeedback =
    feedbackItem.courseParticipationID === feedbackItem.authorCourseParticipationID

  const getAuthorName = () => {
    if (isSelfFeedback) return null

    for (const team of teams) {
      const author = team.members.find(
        (member) => member.id === feedbackItem.authorCourseParticipationID,
      )
      if (author) {
        return `${author.firstName} ${author.lastName}`
      }
    }
    return 'Unknown Author'
  }

  const authorName = getAuthorName()

  const handleCopyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(feedbackItem.feedbackText)
      setShowCopied(true)
      setTimeout(() => setShowCopied(false), 750)
    } catch (err) {
      console.error('Failed to copy to clipboard:', err)
    }
  }

  return (
    <div className='p-3 border rounded-md bg-muted/50 relative'>
      {showCopied && (
        <div
          className={
            'absolute inset-0 bg-black/10 dark:bg-white/10 backdrop-blur-sm rounded-md ' +
            'flex items-center justify-center z-10 animate-in fade-in-0 duration-150 animate-out fade-out-0 duration-200'
          }
        >
          <div className='flex items-center gap-2 text-foreground font-medium text-sm'>
            <Check className='h-4 w-4' />
            <span>Copied to clipboard</span>
          </div>
        </div>
      )}

      <div className='flex items-start items-center justify-between gap-2'>
        <p className='text-sm text-foreground whitespace-pre-wrap flex-1'>
          {feedbackItem.feedbackText}
        </p>

        <div className='flex-shrink-0 mt-1 flex items-center gap-1.5'>
          <Button
            onClick={handleCopyToClipboard}
            variant='ghost'
            size='sm'
            className='h-6 w-6 p-0'
            title='Copy to clipboard'
            aria-label='Copy to clipboard'
          >
            <Copy className='h-4 w-4 text-muted-foreground' />
          </Button>
          {isSelfFeedback ? (
            <User
              className='h-4 w-4 text-blue-500 dark:text-blue-400'
              aria-label='Self feedback'
              role='img'
            />
          ) : (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Users
                    className='h-4 w-4 text-green-600 dark:text-green-400'
                    aria-label='Author information'
                    role='img'
                  />
                </TooltipTrigger>
                <TooltipContent>
                  <p>Author: {authorName}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}
        </div>
      </div>
    </div>
  )
}
