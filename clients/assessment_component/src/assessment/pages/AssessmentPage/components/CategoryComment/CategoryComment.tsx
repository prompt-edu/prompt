import { useState, useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'

import { Alert, AlertDescription, Textarea } from '@tumaet/prompt-ui-components'
import { useAuthStore } from '@tumaet/prompt-shared-state'

import { CategoryAssessment } from '../../../../interfaces/categoryAssessment'

import { useCreateOrUpdateCategoryAssessment } from './hooks/useCreateOrUpdateCategoryAssessment'

interface CategoryCommentProps {
  categoryID: string
  courseParticipationID: string
  categoryAssessment?: CategoryAssessment
  completed?: boolean
}

export const CategoryComment = ({
  categoryID,
  courseParticipationID,
  categoryAssessment,
  completed = false,
}: CategoryCommentProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { user } = useAuthStore()
  const userName = user ? `${user.firstName} ${user.lastName}` : 'Unknown User'
  const userID = user?.username ?? ''

  const [comment, setComment] = useState(categoryAssessment?.comment ?? '')
  const [error, setError] = useState<string | undefined>(undefined)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    setComment(categoryAssessment?.comment ?? '')
  }, [categoryAssessment?.comment])

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
    }
  }, [comment])

  const { mutate: saveCategoryAssessment } = useCreateOrUpdateCategoryAssessment(setError)

  const handleBlur = () => {
    if (completed) return
    const trimmed = comment.trim()
    const previous = (categoryAssessment?.comment ?? '').trim()
    if (trimmed === previous) return

    saveCategoryAssessment({
      categoryID,
      coursePhaseID: phaseId ?? '',
      courseParticipationID,
      comment: trimmed,
      author: userName,
      authorID: userID,
    })
  }

  const updatedAtLabel = categoryAssessment?.updatedAt
    ? new Date(categoryAssessment.updatedAt).toLocaleString()
    : null

  return (
    <div className='space-y-2 mb-4'>
      <Textarea
        ref={textareaRef}
        placeholder='Add an overall comment for this category…'
        className='w-full resize-none min-h-[80px] overflow-hidden'
        rows={3}
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        onBlur={handleBlur}
        disabled={completed}
      />
      {categoryAssessment?.author && updatedAtLabel && (
        <p className='text-xs text-muted-foreground'>
          Last updated by {categoryAssessment.author} on {updatedAtLabel}
        </p>
      )}
      {error && (
        <Alert variant='destructive'>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
    </div>
  )
}
