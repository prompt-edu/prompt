import { Alert, AlertDescription, Textarea } from '@tumaet/prompt-ui-components'
import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'

import type { CategoryAssessment } from '../../../../interfaces/categoryAssessment'

import { useCreateOrUpdateCategoryAssessment } from './hooks/useCreateOrUpdateCategoryAssessment'

interface CategoryCommentProps {
  categoryID: string
  courseParticipationID: string
  categoryAssessment?: CategoryAssessment
  completed?: boolean
  disabled?: boolean
}

export const CategoryComment = ({
  categoryID,
  courseParticipationID,
  categoryAssessment,
  completed = false,
  disabled = false,
}: CategoryCommentProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()

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
    if (completed || disabled) return
    const trimmed = comment.trim()
    const previous = (categoryAssessment?.comment ?? '').trim()
    if (trimmed === previous) return

    saveCategoryAssessment({
      categoryID,
      coursePhaseID: phaseId ?? '',
      courseParticipationID,
      comment: trimmed,
    })
  }

  const hasComment = (categoryAssessment?.comment ?? '').trim().length > 0
  // In read-only contexts (released results), don't render an empty textarea with an
  // assessor-targeted placeholder. Just hide the whole block if nothing was written.
  if (completed && !hasComment) {
    return null
  }

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
        disabled={completed || disabled}
      />
      {error && (
        <Alert variant='destructive'>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
    </div>
  )
}
