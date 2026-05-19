import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { AlertCircle } from 'lucide-react'

import { Form, cn } from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../../../interfaces/assessmentType'
import { Competency } from '../../../../interfaces/competency'
import { Evaluation, CreateOrUpdateEvaluationRequest } from '../../../../interfaces/evaluation'
import { ScoreLevel } from '@tumaet/prompt-shared-state'

import { CompetencyHeader } from '../../../components/CompetencyHeader'
import { DeleteAssessmentDialog } from '../../../components/DeleteAssessmentDialog'
import { ScoreLevelSelector } from '../../../components/ScoreLevelSelector'

import { useCreateOrUpdateEvaluation } from './hooks/useCreateOrUpdateEvaluation'
import { useDeleteEvaluation } from './hooks/useDeleteEvaluation'

interface EvaluationFormProps {
  type: AssessmentType
  courseParticipationID: string
  authorCourseParticipationID: string
  competency: Competency
  evaluation?: Evaluation
  completed?: boolean
}

export const EvaluationForm = ({
  type,
  courseParticipationID,
  authorCourseParticipationID,
  competency,
  evaluation,
  completed = false,
}: EvaluationFormProps) => {
  const [error, setError] = useState<string | undefined>(undefined)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  const form = useForm<CreateOrUpdateEvaluationRequest>({
    mode: 'onChange',
    defaultValues: {
      scoreLevel: evaluation?.scoreLevel,
    },
  })

  const { mutate: createOrUpdateEvaluation } = useCreateOrUpdateEvaluation(setError)
  const deleteEvaluation = useDeleteEvaluation(setError)
  const selectedScoreLevel = form.watch('scoreLevel')

  useEffect(() => {
    form.reset({
      scoreLevel: evaluation?.scoreLevel,
    })
  }, [form, evaluation])

  useEffect(() => {
    if (completed) return

    const subscription = form.watch(async (_, { name }) => {
      if (name) {
        const scoreLevel = form.getValues().scoreLevel
        createOrUpdateEvaluation({
          courseParticipationID: courseParticipationID,
          competencyID: competency.id,
          scoreLevel: scoreLevel as ScoreLevel,
          authorCourseParticipationID: authorCourseParticipationID,
          type: type,
        })
      }
    })
    return () => subscription.unsubscribe()
  }, [
    form,
    createOrUpdateEvaluation,
    completed,
    courseParticipationID,
    competency.id,
    authorCourseParticipationID,
    type,
  ])

  const handleScoreChange = (value: ScoreLevel) => {
    if (completed) return
    form.setValue('scoreLevel', value, { shouldValidate: true })
  }

  const handleDelete = () => {
    if (evaluation?.id) {
      deleteEvaluation.mutate(evaluation.id, {
        onSuccess: () => {
          setDeleteDialogOpen(false)

          form.reset({
            scoreLevel: undefined,
          })
        },
      })
    }
  }

  return (
    <Form {...form}>
      <div
        className={cn(
          'space-y-4 p-4 border rounded-md relative',
          completed && 'bg-gray-700 border-gray-700',
        )}
      >
        <CompetencyHeader
          competency={competency}
          competencyScore={evaluation}
          completed={completed}
          onResetClick={() => setDeleteDialogOpen(true)}
          assessmentType={type}
        />

        <ScoreLevelSelector
          className='grid grid-cols-1 gap-1 md:grid-cols-5'
          competency={competency}
          selectedScore={selectedScoreLevel}
          onScoreChange={handleScoreChange}
          completed={completed}
          assessmentType={type}
        />
      </div>

      {error && !completed && (
        <div className='flex items-center gap-2 text-destructive text-xs p-2 mt-2 bg-destructive/10 rounded-md'>
          <AlertCircle className='h-3 w-3' />
          <p>{error}</p>
        </div>
      )}

      {evaluation && (
        <div className='col-span-full'>
          <DeleteAssessmentDialog
            open={deleteDialogOpen}
            onOpenChange={setDeleteDialogOpen}
            onConfirm={handleDelete}
            isDeleting={deleteEvaluation.isPending}
          />
        </div>
      )}
    </Form>
  )
}
