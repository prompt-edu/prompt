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
import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'
import { useSchemaHasAssessmentData } from './hooks/usePhaseHasAssessmentData'
import { schemaSectionContent, getSchemaIdForType, isSchemaTypeEnabled } from './schemaConfig'
import { CategoryList } from './components/CategoryList/CategoryList'
import { useGetAllAssessmentSchemas } from './components/CoursePhaseConfigSelection/hooks/useGetAllAssessmentSchemas'

const isAssessmentType = (value: string | undefined): value is AssessmentType => {
  return value !== undefined && Object.values(AssessmentType).includes(value as AssessmentType)
}

export const SchemaDetailPage = () => {
  const { assessmentType: assessmentTypeParam } = useParams<{ assessmentType: string }>()
  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const {
    data: schemas,
    isPending: isSchemasPending,
    isError: isSchemasError,
  } = useGetAllAssessmentSchemas()
  const assessmentType = isAssessmentType(assessmentTypeParam) ? assessmentTypeParam : undefined
  const content = assessmentType ? schemaSectionContent[assessmentType] : undefined
  const schemaId = assessmentType
    ? getSchemaIdForType(coursePhaseConfig, assessmentType)
    : undefined
  const isEnabled = assessmentType ? isSchemaTypeEnabled(coursePhaseConfig, assessmentType) : false
  const { data: schemaData } = useSchemaHasAssessmentData(schemaId)

  if (!assessmentType || !content) {
    return (
      <div className='space-y-4'>
        <Button asChild variant='outline'>
          <Link to='../..' relative='path'>
            <ArrowLeft className='mr-2 h-4 w-4' />
            Back to settings
          </Link>
        </Button>
        <ErrorPage message='The requested schema type does not exist.' />
      </div>
    )
  }

  if (isSchemasError) {
    return (
      <div className='space-y-4'>
        <Button asChild variant='outline'>
          <Link to='../..' relative='path'>
            <ArrowLeft className='mr-2 h-4 w-4' />
            Back to settings
          </Link>
        </Button>
        <ErrorPage message='Could not load assessment schemas.' />
      </div>
    )
  }

  if (isSchemasPending) {
    return <LoadingPage />
  }

  if (assessmentType !== AssessmentType.ASSESSMENT && !isEnabled) {
    return (
      <div className='space-y-4'>
        <Button asChild variant='outline'>
          <Link to='../..' relative='path'>
            <ArrowLeft className='mr-2 h-4 w-4' />
            Back to settings
          </Link>
        </Button>
        <ErrorPage message={`${content.title} is currently disabled for this phase.`} />
      </div>
    )
  }

  if (!schemaId) {
    return (
      <div className='space-y-4'>
        <Button asChild variant='outline'>
          <Link to='../..' relative='path'>
            <ArrowLeft className='mr-2 h-4 w-4' />
            Back to settings
          </Link>
        </Button>
        <ErrorPage message='No schema is configured for this assessment type yet.' />
      </div>
    )
  }

  const schema = schemas?.find((item) => item.id === schemaId)

  return (
    <div className='space-y-6'>
      <div className='flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between'>
        <div className='space-y-2'>
          <Button asChild variant='outline'>
            <Link to='../..' relative='path'>
              <ArrowLeft className='mr-2 h-4 w-4' />
              Back to settings
            </Link>
          </Button>
          <ManagementPageHeader>{content.detailTitle}</ManagementPageHeader>
          <p className='max-w-3xl text-sm leading-6 text-muted-foreground'>
            {content.detailDescription}
          </p>
        </div>
      </div>

      <Card className='border-slate-200 shadow-sm'>
        <CardContent className='space-y-3 p-6'>
          <div className='flex flex-wrap items-center gap-2'>
            <h2 className='text-xl font-semibold text-slate-900'>
              {schema?.name ?? 'Configured schema'}
            </h2>
            {schemaData?.hasAssessmentData && (
              <span className='inline-flex items-center gap-1 rounded-full bg-amber-100 px-3 py-1 text-xs font-medium text-amber-800'>
                <Lock className='h-3.5 w-3.5' />
                Locked by submitted data
              </span>
            )}
          </div>
          <p className='text-sm leading-6 text-slate-600'>
            {schema?.description ||
              'This schema is assigned to the current phase. Review the categories and competencies below.'}
          </p>
          <p className='text-xs leading-5 text-slate-500'>
            Changes here affect the schema currently connected to this assessment type for the
            current course phase.
          </p>
        </CardContent>
      </Card>

      <CategoryList
        assessmentSchemaID={schemaId}
        assessmentType={assessmentType}
        hasAssessmentData={schemaData?.hasAssessmentData ?? false}
      />
    </div>
  )
}
