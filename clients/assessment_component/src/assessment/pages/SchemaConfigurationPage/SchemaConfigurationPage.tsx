import { Link, useParams } from 'react-router-dom'
import {
  Button,
  Card,
  CardContent,
  ErrorPage,
  LoadingPage,
  ManagementPageHeader,
} from '@tumaet/prompt-ui-components'
import { ArrowLeft, Lock } from 'lucide-react'

import { AssessmentType } from '../../interfaces/assessmentType'
import { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'
import { useGetAllAssessmentSchemas } from '../hooks/useGetAllAssessmentSchemas'
import { useSchemaHasAssessmentData } from '../hooks/useSchemaHasAssessmentData'
import { schemaSectionContent } from '../schemaSectionContent'
import { CategoryList } from './components/CategoryList/CategoryList'

const BackToSettingsButton = () => (
  <Button asChild variant='outline'>
    <Link to='../..' relative='path'>
      <ArrowLeft className='mr-2 h-4 w-4' />
      Settings
    </Link>
  </Button>
)

const LOCK_BADGE_CLASS = [
  'inline-flex items-center gap-1 rounded-full bg-amber-100 px-3 py-1 text-xs font-medium',
  'text-amber-900 dark:bg-amber-900/40 dark:text-amber-200',
].join(' ')

const getAssessmentTypeForSchema = (
  config: CoursePhaseConfig | undefined,
  schemaID: string | undefined,
): AssessmentType | undefined => {
  if (!config || !schemaID) {
    return undefined
  }

  if (config.assessmentSchemaID === schemaID) {
    return AssessmentType.ASSESSMENT
  }

  if (config.selfEvaluationSchema === schemaID) {
    return AssessmentType.SELF
  }

  if (config.peerEvaluationSchema === schemaID) {
    return AssessmentType.PEER
  }

  if (config.tutorEvaluationSchema === schemaID) {
    return AssessmentType.TUTOR
  }

  return undefined
}

const isAssessmentTypeEnabled = (
  config: CoursePhaseConfig | undefined,
  assessmentType: AssessmentType,
): boolean => {
  if (!config) {
    return false
  }

  switch (assessmentType) {
    case AssessmentType.SELF:
      return config.selfEvaluationEnabled
    case AssessmentType.PEER:
      return config.peerEvaluationEnabled
    case AssessmentType.TUTOR:
      return config.tutorEvaluationEnabled
    case AssessmentType.ASSESSMENT:
    default:
      return true
  }
}

export const SchemaConfigurationPage = () => {
  const { schemaId } = useParams<{ schemaId: string }>()
  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const {
    data: schemas,
    isPending: isSchemasPending,
    isError: isSchemasError,
  } = useGetAllAssessmentSchemas()
  const {
    data: schemaData,
    isPending: isSchemaDataPending,
    isError: isSchemaDataError,
  } = useSchemaHasAssessmentData(schemaId)

  const assessmentType = getAssessmentTypeForSchema(coursePhaseConfig, schemaId)
  const content = assessmentType ? schemaSectionContent[assessmentType] : undefined
  const isEnabled = assessmentType
    ? isAssessmentTypeEnabled(coursePhaseConfig, assessmentType)
    : false

  if (!schemaId) {
    return (
      <div className='space-y-4'>
        <BackToSettingsButton />
        <ErrorPage message='The requested schema does not exist.' />
      </div>
    )
  }

  if (isSchemasError) {
    return (
      <div className='space-y-4'>
        <BackToSettingsButton />
        <ErrorPage message='Could not load assessment schemas.' />
      </div>
    )
  }

  if (isSchemasPending) {
    return <LoadingPage />
  }

  if (isSchemaDataPending) {
    return <LoadingPage />
  }

  if (isSchemaDataError || !schemaData) {
    return (
      <div className='space-y-4'>
        <BackToSettingsButton />
        <ErrorPage message='Could not determine whether this schema is locked by submitted data.' />
      </div>
    )
  }

  if (!assessmentType || !content) {
    return (
      <div className='space-y-4'>
        <BackToSettingsButton />
        <ErrorPage message='This schema is currently not configured for this course phase.' />
      </div>
    )
  }

  if (assessmentType !== AssessmentType.ASSESSMENT && !isEnabled) {
    return (
      <div className='space-y-4'>
        <BackToSettingsButton />
        <ErrorPage message={`${content.title} is currently disabled for this phase.`} />
      </div>
    )
  }

  const schema = schemas?.find((item) => item.id === schemaId)

  return (
    <div className='space-y-6'>
      <div className='flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between'>
        <div className='space-y-2'>
          <BackToSettingsButton />
          <ManagementPageHeader>{content.detailTitle}</ManagementPageHeader>
          <p className='max-w-3xl text-sm leading-6 text-muted-foreground'>
            {content.detailDescription}
          </p>
        </div>
      </div>

      <Card className='border-border shadow-xs'>
        <CardContent className='space-y-3 p-6'>
          <div className='flex flex-wrap items-center gap-2'>
            <h2 className='text-xl font-semibold text-foreground'>
              {schema?.name ?? 'Configured schema'}
            </h2>
            {schemaData.hasAssessmentData && (
              <span className={LOCK_BADGE_CLASS}>
                <Lock className='h-3.5 w-3.5' />
                Locked by submitted data
              </span>
            )}
          </div>
          <p className='text-sm leading-6 text-muted-foreground'>
            {schema?.description ||
              'This schema is assigned to the current phase. Review the categories and competencies below.'}
          </p>
          <p className='text-xs leading-5 text-muted-foreground'>
            Changes here affect the schema currently connected to this assessment type for the
            current course phase.
          </p>
        </CardContent>
      </Card>

      <CategoryList
        assessmentSchemaID={schemaId}
        assessmentType={assessmentType}
        hasAssessmentData={schemaData.hasAssessmentData}
      />
    </div>
  )
}
