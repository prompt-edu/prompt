import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'

import { useAuthStore } from '@tumaet/prompt-shared-state'
import { Form, FormMessage } from '@tumaet/prompt-ui-components'

import { useStudentAssessmentStore } from '../../../../zustand/useStudentAssessmentStore'
import { useTeamStore } from '../../../../zustand/useTeamStore'
import { useSelfEvaluationCategoryStore } from '../../../../zustand/useSelfEvaluationCategoryStore'
import { usePeerEvaluationCategoryStore } from '../../../../zustand/usePeerEvaluationCategoryStore'

import { Assessment, CreateOrUpdateAssessmentRequest } from '../../../../interfaces/assessment'
import { Competency } from '../../../../interfaces/competency'
import {
  ScoreLevel,
  mapNumberToScoreLevel,
  mapScoreLevelToNumber,
} from '@tumaet/prompt-shared-state'

import { CompetencyHeader } from '../../../components/CompetencyHeader'
import { DeleteAssessmentDialog } from '../../../components/DeleteAssessmentDialog'
import { ScoreLevelSelector } from '../../../components/ScoreLevelSelector'

import { EvaluationScoreDescriptionBadge } from './components/EvaluationScoreDescriptionBadge'
import { AssessmentTextField } from './components/AssessmentTextField'

import { useCreateOrUpdateAssessment } from './hooks/useCreateOrUpdateAssessment'
import { useDeleteAssessment } from './hooks/useDeleteAssessment'
import { JSX } from 'react/jsx-runtime'

interface AssessmentFormProps {
  courseParticipationID: string
  competency: Competency
  assessment?: Assessment
  completed?: boolean
  peerEvaluationAverageScore?: number
  selfEvaluationAverageScore?: number
  hidePeerEvaluationDetails?: boolean
}

export const AssessmentForm = ({
  courseParticipationID,
  competency,
  assessment,
  completed = false,
  peerEvaluationAverageScore,
  selfEvaluationAverageScore,
  hidePeerEvaluationDetails = false,
}: AssessmentFormProps) => {
  const [error, setError] = useState<string | undefined>(undefined)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  const { user } = useAuthStore()
  const userName = user ? `${user.firstName} ${user.lastName}` : 'Unknown User'

  const form = useForm<CreateOrUpdateAssessmentRequest>({
    mode: 'onChange',
    criteriaMode: 'all',
    reValidateMode: 'onChange',
    defaultValues: {
      courseParticipationID,
      competencyID: competency.id,
      scoreLevel: assessment?.scoreLevel,
      comment: assessment ? assessment.comment : '',
      examples: assessment ? assessment.examples : '',
      author: userName,
    },
  })

  const { mutate: createOrUpdateAssessment } = useCreateOrUpdateAssessment(setError)
  const deleteAssessment = useDeleteAssessment(setError)
  const selectedScore = form.watch('scoreLevel')
  const hasExample = (assessment?.examples ?? '').trim().length > 0
  const hasComment = (assessment?.comment ?? '').trim().length > 0
  const shouldHideCommentAndExample = completed && !hasExample && !hasComment

  useEffect(() => {
    form.reset({
      courseParticipationID,
      competencyID: competency.id,
      scoreLevel: assessment?.scoreLevel,
      comment: assessment ? assessment.comment : '',
      examples: assessment ? assessment.examples : '',
      author: userName,
    })
  }, [form, courseParticipationID, competency.id, assessment, userName])

  const saveAssessment = async () => {
    if (completed) return

    const isValid = await form.trigger()
    if (!isValid) return

    const data = form.getValues()
    if (!data.scoreLevel) return

    createOrUpdateAssessment(data)
  }

  useEffect(() => {
    if (completed) return

    const subscription = form.watch(async (_, { name }) => {
      if (name === 'scoreLevel') {
        await form.trigger(['comment', 'examples'])
      }
    })

    return () => {
      subscription.unsubscribe()
    }
  }, [form, completed])

  const handleScoreChange = async (value: ScoreLevel) => {
    if (completed) return
    form.setValue('scoreLevel', value, { shouldValidate: true })
    await saveAssessment()
  }

  const handleDelete = () => {
    if (assessment?.id) {
      deleteAssessment.mutate(assessment.id, {
        onSuccess: () => {
          setDeleteDialogOpen(false)

          form.reset({
            courseParticipationID,
            competencyID: competency.id,
            scoreLevel: undefined,
            comment: '',
            examples: '',
            author: userName,
          })
        },
      })
    }
  }

  const selfEvaluationCompetency =
    useSelfEvaluationCategoryStore().allSelfEvaluationCompetencies.find((c) =>
      competency.mappedFromCompetencies.includes(c.id),
    )
  const peerEvaluationCompetency =
    usePeerEvaluationCategoryStore().allPeerEvaluationCompetencies.find((c) =>
      competency.mappedFromCompetencies.includes(c.id),
    )

  const {
    selfEvaluations: allSelfEvaluationsForThisStudent,
    peerEvaluations: allPeerEvaluationsForThisStudent,
    assessmentParticipation,
  } = useStudentAssessmentStore()

  const { teams } = useTeamStore()
  const teamMembers = teams.find((t) =>
    t.members.map((m) => m.id).includes(courseParticipationID ?? ''),
  )?.members

  const selfEvaluationScoreLevel =
    selfEvaluationAverageScore !== undefined
      ? mapNumberToScoreLevel(selfEvaluationAverageScore)
      : allSelfEvaluationsForThisStudent.find(
          (se) => se.competencyID === selfEvaluationCompetency?.id,
        )?.scoreLevel

  const selfEvaluationStudentAnswers = [
    () => (
      <EvaluationScoreDescriptionBadge
        key={'self'}
        competency={selfEvaluationCompetency}
        scoreLevel={selfEvaluationScoreLevel}
        name={assessmentParticipation?.student.firstName ?? 'This Person'}
      />
    ),
  ]

  const peerEvaluations = allPeerEvaluationsForThisStudent.filter(
    (pe) => pe.competencyID === peerEvaluationCompetency?.id,
  )

  const peerEvaluationScore =
    peerEvaluationAverageScore !== undefined
      ? mapNumberToScoreLevel(peerEvaluationAverageScore)
      : peerEvaluations?.length
        ? mapNumberToScoreLevel(
            peerEvaluations.reduce((acc, pe) => acc + mapScoreLevelToNumber(pe.scoreLevel), 0) /
              peerEvaluations.length,
          )
        : undefined

  const peerEvaluationStudentAnswers: (() => JSX.Element)[] | undefined = hidePeerEvaluationDetails
    ? undefined
    : teamMembers
        ?.map((member) => {
          const memberScoreLevel = peerEvaluations.find(
            (pe) => pe.authorCourseParticipationID === member.id,
          )?.scoreLevel

          return memberScoreLevel !== undefined && peerEvaluationCompetency
            ? () => (
                <EvaluationScoreDescriptionBadge
                  key={member.id}
                  competency={peerEvaluationCompetency}
                  scoreLevel={memberScoreLevel}
                  name={`${member.firstName} ${member.lastName}`}
                />
              )
            : undefined
        })
        .filter((item): item is () => JSX.Element => item !== undefined)

  return (
    <Form {...form}>
      <div className={'grid grid-cols-1 lg:grid-cols-2 gap-4 items-start p-4 border rounded-md'}>
        <CompetencyHeader
          className='lg:col-span-2'
          competency={competency}
          competencyScore={assessment}
          completed={completed}
          onResetClick={() => setDeleteDialogOpen(true)}
        />

        <ScoreLevelSelector
          className='lg:col-span-2 grid grid-cols-1 lg:grid-cols-5 gap-1'
          competency={competency}
          selectedScore={selectedScore}
          onScoreChange={handleScoreChange}
          completed={completed}
          selfEvaluationCompetency={selfEvaluationCompetency}
          selfEvaluationScoreLevel={selfEvaluationScoreLevel}
          selfEvaluationStudentAnswers={selfEvaluationStudentAnswers}
          peerEvaluationCompetency={
            peerEvaluationCompetency && peerEvaluationCompetency.id
              ? {
                  ...peerEvaluationCompetency,
                  name:
                    peerEvaluationCompetency.name.replace(
                      /This person|this person/g,
                      assessmentParticipation?.student.firstName ?? 'This Person',
                    ) ?? '',
                }
              : undefined
          }
          peerEvaluationScoreLevel={peerEvaluationScore}
          peerEvaluationStudentAnswers={peerEvaluationStudentAnswers}
        />

        {!shouldHideCommentAndExample && (
          <>
            <AssessmentTextField
              control={form.control}
              name='examples'
              placeholder='Example'
              completed={completed}
              getScoreLevel={() => form.getValues('scoreLevel')}
              onBlur={saveAssessment}
            />

            <AssessmentTextField
              control={form.control}
              name='comment'
              placeholder='Additional comments'
              completed={completed}
              getScoreLevel={() => form.getValues('scoreLevel')}
              onBlur={saveAssessment}
            />
          </>
        )}

        {error && !completed && <FormMessage className='mt-2'>{error}</FormMessage>}

        {assessment && (
          <DeleteAssessmentDialog
            open={deleteDialogOpen}
            onOpenChange={setDeleteDialogOpen}
            onConfirm={handleDelete}
            isDeleting={deleteAssessment.isPending}
          />
        )}
      </div>
    </Form>
  )
}
