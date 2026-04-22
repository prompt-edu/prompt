import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Loader2, Users, ArrowRight } from 'lucide-react'

import { getAllTeaseStudents } from '../../../network/queries/getAllTeaseStudents'
import type { TeaseStudent } from '../../../interfaces/tease/student'
import type { ValidationResult } from '../../../interfaces/validationResult'
import DataCompletionSummary from './DataCompletionSummary'
import { CheckItem } from './CheckItem'
import { checksConfig } from './ChecksConfig'

import SurveySubmissionOverview from './SurveySubmissionOverview'
import {
  Card,
  CardContent,
  Button,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  ErrorPage,
} from '@tumaet/prompt-ui-components'

export const StudentDataCheck = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const [checks, setChecks] = useState<ValidationResult[] | null>(null)

  const {
    data: students,
    isPending,
    isError,
    refetch,
  } = useQuery<TeaseStudent[]>({
    queryKey: ['tease_students', phaseId],
    queryFn: () => getAllTeaseStudents(phaseId ?? ''),
  })

  useEffect(() => {
    if (!students || students.length === 0) return

    const results: ValidationResult[] = checksConfig.map(
      ({
        label,
        extractor,
        isEmpty,
        missingMessage,
        problemDescription,
        details,
        category,
        icon,
        highLevelCategory,
      }) => {
        const validStudents = students.filter((s) => !isEmpty(extractor(s)))
        const completionRate =
          students.length > 0 ? Math.round((validStudents.length / students.length) * 100) : 0
        const allValid = completionRate === 100
        const noneValid = completionRate === 0

        return {
          label,
          isValid: allValid,
          category,
          highLevelCategory,
          completionRate,
          icon,
          details: allValid ? details : noneValid ? `${details}` : undefined,
          problemDescription: allValid
            ? undefined
            : noneValid
              ? problemDescription
              : `${validStudents.length} of ${students.length} students have provided ${missingMessage}.`,
        }
      },
    )

    setChecks(results)
  }, [students])

  if (isPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  if (isError) {
    return (
      <ErrorPage
        onRetry={() => {
          refetch()
        }}
      />
    )
  }

  if (!checks || checks.length === 0) {
    return (
      <Card>
        <CardContent className='p-6 text-center'>
          <Users className='h-12 w-12 text-muted-foreground mx-auto mb-4' />
          <h3 className='text-lg font-medium mb-2'>No student data available</h3>
          <p className='text-muted-foreground'>
            There are no students enrolled in this phase yet, or the data hasn`t been loaded.
          </p>
        </CardContent>
      </Card>
    )
  }

  // Group checks by category
  const deviceChecks = checks.filter((check) => check.category === 'devices')
  const commentChecks = checks.filter((check) => check.category === 'comments')
  const scoreChecks = checks.filter((check) => check.category === 'score')
  const languageChecks = checks.filter((check) => check.category === 'language')

  return (
    <div className='space-y-6'>
      <DataCompletionSummary
        checks={checks}
        students={students}
        isLoading={isPending}
        isError={isError}
      />

      <Tabs defaultValue='previous' className='w-full'>
        <TabsList className='grid grid-cols-2 mb-4'>
          <TabsTrigger value='previous'>Previous Phases</TabsTrigger>
          <TabsTrigger value='survey'>Survey Results</TabsTrigger>
        </TabsList>

        <TabsContent value='previous' className='space-y-4'>
          <div className='space-y-6'>
            {deviceChecks.map((check, index) => (
              <CheckItem key={index} check={check} />
            ))}
          </div>

          <div className='space-y-6'>
            {commentChecks.map((check, index) => (
              <CheckItem key={index} check={check} />
            ))}
          </div>

          <div className='space-y-6'>
            {scoreChecks.map((check, index) => (
              <CheckItem key={index} check={check} />
            ))}
          </div>

          <div className='space-y-6'>
            {languageChecks.map((check, index) => (
              <CheckItem key={index} check={check} />
            ))}
          </div>
        </TabsContent>

        <TabsContent value='survey' className='space-y-6'>
          <SurveySubmissionOverview students={students || []} />
        </TabsContent>
      </Tabs>

      {!checks?.every((c) => c.isValid) && (
        <p className='text-sm text-muted-foreground mt-2 text-left'>
          <span className='font-semibold'>
            Please ensure all student data fields are completed before proceeding to TEASE!{' '}
          </span>
        </p>
      )}
      <div className='mt-4 w-full'>
        <Button asChild className='gap-2 w-full'>
          <a href={`/tease?coursePhaseId=${phaseId ?? ''}`}>
            Launch Tease to Matchmake
            <ArrowRight className='ml-2 h-4 w-4' />
          </a>
        </Button>
      </div>
    </div>
  )
}

export default StudentDataCheck
