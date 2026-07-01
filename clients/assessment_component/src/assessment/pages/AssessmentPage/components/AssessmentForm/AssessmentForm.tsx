import {
  mapNumberToScoreLevel,
  mapScoreLevelToNumber,
  type ScoreLevel,
} from '@tumaet/prompt-shared-state'
import { Form, FormMessage } from '@tumaet/prompt-ui-components'
import { useEffect, useState } from 'react'
import type { JSX } from 'react/jsx-runtime'
import { useForm } from 'react-hook-form'
import type { Assessment, CreateOrUpdateAssessmentRequest } from '../../../../interfaces/assessment'
import type { Competency } from '../../../../interfaces/competency'
import { usePeerEvaluationCategoryStore } from '../../../../zustand/usePeerEvaluationCategoryStore'
import { useSelfEvaluationCategoryStore } from '../../../../zustand/useSelfEvaluationCategoryStore'
import { useStudentAssessmentStore } from '../../../../zustand/useStudentAssessmentStore'
import { useTeamStore } from '../../../../zustand/useTeamStore'
import { CompetencyHeader } from '../../../components/CompetencyHeader'
import { DeleteAssessmentDialog } from '../../../components/DeleteAssessmentDialog'
import { ScoreLevelSelector } from '../../../components/ScoreLevelSelector'
import { EvaluationScoreDescriptionBadge } from './components/EvaluationScoreDescriptionBadge'
import { useCreateOrUpdateAssessment } from './hooks/useCreateOrUpdateAssessment'
import { useDeleteAssessment } from './hooks/useDeleteAssessment'

interface AssessmentFormProps {
  courseParticipationID: string
  competency: Competency
  assessment?: Assessment
  completed?: boolean
  disabled?: boolean
  peerEvaluationAverageScore?: number
  selfEvaluationAverageScore?: number
  hidePeerEvaluationDetails?: boolean
}

export const AssessmentForm = ({
  courseParticipationID,
  competency,
  assessment,
  completed = false,
  disabled = false,
  peerEvaluationAverageScore,
  selfEvaluationAverageScore,
  hidePeerEvaluationDetails = false,
}: AssessmentFormProps) => {
  const [error, setError] = useState<string | undefined>(undefined)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  const form = useForm<CreateOrUpdateAssessmentRequest>({
    mode: 'onChange',
    criteriaMode: 'all',
    reValidateMode: 'onChange',
    defaultValues: {
      courseParticipationID,
      competencyID: competency.id,
      scoreLevel: assessment?.scoreLevel,
    },
  })

  const { mutate: createOrUpdateAssessment } = useCreateOrUpdateAssessment(setError)
  const deleteAssessment = useDeleteAssessment(setError)
  const selectedScore = form.watch('scoreLevel')
  const controlsDisabled = completed || disabled

  useEffect(() => {
    form.reset({
      courseParticipationID,
      competencyID: competency.id,
      scoreLevel: assessment?.scoreLevel,
    })
  }, [form, courseParticipationID, competency.id, assessment])

  const saveAssessment = () => {
    if (controlsDisabled) return
    const data = form.getValues()
    if (!data.scoreLevel) return
    createOrUpdateAssessment(data)
  }

  const handleScoreChange = (value: ScoreLevel) => {
    if (controlsDisabled) return
    form.setValue('scoreLevel', value)
    saveAssessment()
  }

  const handleDelete = () => {
    if (controlsDisabled) return
    if (assessment?.id) {
      deleteAssessment.mutate(assessment.id, {
        onSuccess: () => {
          setDeleteDialogOpen(false)

          form.reset({
            courseParticipationID,
            competencyID: competency.id,
            scoreLevel: undefined,
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
      <div className='space-y-4 p-4 border rounded-md'>
        <CompetencyHeader
          competency={competency}
          competencyScore={assessment}
          completed={controlsDisabled}
          onResetClick={() => setDeleteDialogOpen(true)}
        />

        <ScoreLevelSelector
          className='grid grid-cols-1 gap-1 md:grid-cols-5'
          competency={competency}
          selectedScore={selectedScore}
          onScoreChange={handleScoreChange}
          completed={controlsDisabled}
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

        {error && !controlsDisabled && <FormMessage className='mt-2'>{error}</FormMessage>}

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
