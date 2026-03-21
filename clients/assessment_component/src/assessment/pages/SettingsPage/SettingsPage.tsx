import { useState } from 'react'
import { Loader2 } from 'lucide-react'

import { ManagementPageHeader, ErrorPage, Button } from '@tumaet/prompt-ui-components'
import { useAuthStore, Role, getPermissionString } from '@tumaet/prompt-shared-state'

import { useParticipationStore } from '../../zustand/useParticipationStore'
import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'
import { useCategoryStore } from '../../zustand/useCategoryStore'
import { useScoreLevelStore } from '../../zustand/useScoreLevelStore'
import { useGetAllAssessments } from '../hooks/useGetAllAssessments'
import { useGetAllAssessmentCompletions } from '../hooks/useGetAllAssessmentCompletions'
import { useSchemaHasAssessmentData } from './hooks/usePhaseHasAssessmentData'
import { useReleaseResults } from './hooks/useReleaseResults'
import { ReleaseConfirmationDialog } from './components/ReleaseConfirmationDialog'

import { AssessmentType } from '../../interfaces/assessmentType'

import { CategoryDiagram } from '../components/diagrams/CategoryDiagram'
import { ScoreLevelDistributionDiagram } from '../components/diagrams/ScoreLevelDistributionDiagram'

import { CoursePhaseConfigSelection } from './components/CoursePhaseConfigSelection/CoursePhaseConfigSelection'
import { CategoryList } from './components/CategoryList/CategoryList'
import { AssessmentReminderCard } from './components/AssessmentReminderCard'

export const SettingsPage = () => {
  const [showReleaseDialog, setShowReleaseDialog] = useState(false)
  const { participations } = useParticipationStore()
  const { coursePhaseConfig: config } = useCoursePhaseConfigStore()
  const { categories } = useCategoryStore()
  const { scoreLevels } = useScoreLevelStore()
  const { permissions } = useAuthStore()

  const isPromptAdmin = permissions.includes(getPermissionString(Role.PROMPT_ADMIN))
  const isLecturer = permissions.includes(getPermissionString(Role.COURSE_LECTURER))

  const { mutate: releaseResults, isPending: isReleasing } = useReleaseResults()

  const {
    data: assessments,
    isPending: isAssessmentsPending,
    isError: isAssessmentsError,
    refetch: refetchAssessments,
  } = useGetAllAssessments()

  const { data: assessmentSchemaData } = useSchemaHasAssessmentData(config?.assessmentSchemaID)
  const { data: selfEvalSchemaData } = useSchemaHasAssessmentData(
    config?.selfEvaluationEnabled ? config?.selfEvaluationSchema : undefined,
  )
  const { data: peerEvalSchemaData } = useSchemaHasAssessmentData(
    config?.peerEvaluationEnabled ? config?.peerEvaluationSchema : undefined,
  )
  const { data: tutorEvalSchemaData } = useSchemaHasAssessmentData(
    config?.tutorEvaluationEnabled ? config?.tutorEvaluationSchema : undefined,
  )

  const { data: assessmentCompletions } = useGetAllAssessmentCompletions()

  const totalAssessments = participations.length
  const completedAssessments = assessmentCompletions?.filter((c) => c.completed).length ?? 0
  const allAssessmentsCompleted = totalAssessments > 0 && completedAssessments === totalAssessments

  const handleReleaseResults = () => {
    setShowReleaseDialog(true)
  }

  const confirmRelease = () => {
    releaseResults(undefined, {
      onSuccess: () => {
        setShowReleaseDialog(false)
      },
    })
  }

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>Assessment Settings</ManagementPageHeader>

      {isAssessmentsError ? (
        <div>
          <ErrorPage onRetry={refetchAssessments} description='Could not fetch assessments' />
        </div>
      ) : isAssessmentsPending ? (
        <div className='flex justify-center items-center h-64'>
          <Loader2 className='h-12 w-12 animate-spin text-primary' />
        </div>
      ) : (
        <div className='grid gap-6 grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 2xl:col-span-4 mb-4'>
          <ScoreLevelDistributionDiagram
            participations={participations}
            scoreLevels={scoreLevels}
          />
          <CategoryDiagram categories={categories} assessments={assessments} />
        </div>
      )}

      <CoursePhaseConfigSelection
        hasAssessmentData={assessmentSchemaData?.hasAssessmentData ?? false}
        hasSelfEvalData={selfEvalSchemaData?.hasAssessmentData ?? false}
        hasPeerEvalData={peerEvalSchemaData?.hasAssessmentData ?? false}
        hasTutorEvalData={tutorEvalSchemaData?.hasAssessmentData ?? false}
      />

      {(isPromptAdmin || isLecturer) && <AssessmentReminderCard />}

      {(isPromptAdmin || isLecturer) && !config?.resultsReleased && (
        <div className='w-full'>
          <Button
            onClick={handleReleaseResults}
            disabled={!allAssessmentsCompleted || isReleasing}
            className='w-full'
            size='lg'
          >
            {isReleasing
              ? 'Releasing...'
              : `Release Results to Students (${completedAssessments}/${totalAssessments} final)`}
          </Button>
          {!allAssessmentsCompleted && (
            <p className='text-sm text-muted-foreground mt-2 text-center'>
              All assessments must be marked as final before releasing results
            </p>
          )}
        </div>
      )}

      {config?.resultsReleased && (
        <div className='w-full p-4 bg-green-50 border border-green-200 rounded-lg'>
          <p className='text-center text-green-700 font-medium'>
            ✓ Results have been released to students
          </p>
        </div>
      )}

      {isPromptAdmin && (
        <>
          {config?.assessmentSchemaID && (
            <CategoryList
              assessmentSchemaID={config?.assessmentSchemaID}
              assessmentType={AssessmentType.ASSESSMENT}
              hasAssessmentData={assessmentSchemaData?.hasAssessmentData ?? false}
            />
          )}

          {config?.selfEvaluationEnabled &&
            config.selfEvaluationSchema &&
            config.selfEvaluationSchema !== config.assessmentSchemaID && (
              <CategoryList
                assessmentSchemaID={config?.selfEvaluationSchema}
                assessmentType={AssessmentType.SELF}
                hasAssessmentData={selfEvalSchemaData?.hasAssessmentData ?? false}
              />
            )}

          {config?.peerEvaluationEnabled &&
            config.peerEvaluationSchema &&
            config.peerEvaluationSchema !== config.assessmentSchemaID && (
              <CategoryList
                assessmentSchemaID={config?.peerEvaluationSchema}
                assessmentType={AssessmentType.PEER}
                hasAssessmentData={peerEvalSchemaData?.hasAssessmentData ?? false}
              />
            )}

          {config?.tutorEvaluationEnabled &&
            config.tutorEvaluationSchema &&
            config.tutorEvaluationSchema !== config.assessmentSchemaID && (
              <CategoryList
                assessmentSchemaID={config?.tutorEvaluationSchema}
                assessmentType={AssessmentType.TUTOR}
                hasAssessmentData={tutorEvalSchemaData?.hasAssessmentData ?? false}
              />
            )}
        </>
      )}

      <ReleaseConfirmationDialog
        open={showReleaseDialog}
        onOpenChange={setShowReleaseDialog}
        onConfirm={confirmRelease}
        isReleasing={isReleasing}
        completedAssessments={completedAssessments}
        totalAssessments={totalAssessments}
      />
    </div>
  )
}
