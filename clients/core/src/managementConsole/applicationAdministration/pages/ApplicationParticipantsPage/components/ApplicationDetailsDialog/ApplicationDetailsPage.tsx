import { useQuery } from '@tanstack/react-query'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Loader2 } from 'lucide-react'
import { Button, ErrorPage } from '@tumaet/prompt-ui-components'
import { GetApplication } from '@core/interfaces/application/getApplication'
import { getApplicationAssessment } from '@core/network/queries/applicationAssessment'
import { ApplicationForm } from '../../../../interfaces/form/applicationForm'
import { getApplicationForm } from '@core/network/queries/applicationForm'
import { MissingUniversityData } from './components/MissingUniversityData'
import { AssessmentCard } from './components/AssessmentCard'
import { ApplicationAnswersTable } from '../table/ApplicationAnswersTable'
import { StudentProfile } from '@/components/StudentProfile'
import { useApplicationStore } from '@core/managementConsole/applicationAdministration/zustand/useApplicationStore'
import { useStudentEnrollments } from '@core/network/hooks/useStudentEnrollments'
import { CourseEnrollmentSummary } from '@core/managementConsole/shared/components/StudentDetail/CourseEnrollmentSummary'
import { StudentCourseEnrollment } from '@core/managementConsole/shared/components/StudentDetail/StudentCourseEnrollment'
import { CourseEnrollment } from '@core/network/queries/getStudentEnrollments'
import { ApplicationDetailPageLayout } from './components/ApplicationDetailPageLayout'
import { useMemo } from 'react'
import { getApplicationNavigationButtonColorClass } from '../table/getApplicationStatusBadge'

interface ApplicationDetailsLocationState {
  filteredApplicationIds?: string[]
}

export const ApplicationDetailsPage = () => {
  const { phaseId, participationId } = useParams<{ phaseId: string; participationId: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const { participations } = useApplicationStore()
  const navigationState = (location.state as ApplicationDetailsLocationState | null) ?? null
  const filteredApplicationIds = navigationState?.filteredApplicationIds

  const participation = participations.find((p) => p.courseParticipationID === participationId)
  const status = participation?.passStatus
  const score = participation?.score ?? null
  const restrictedData = participation?.restrictedData ?? {}
  const participationById = useMemo(
    () => new Map(participations.map((p) => [p.courseParticipationID, p])),
    [participations],
  )
  const navigationOrder = useMemo(() => {
    const candidateIds = filteredApplicationIds ?? []
    const validUniqueIds = candidateIds.filter(
      (id, index) => Boolean(id) && participationById.has(id) && candidateIds.indexOf(id) === index,
    )

    if (validUniqueIds.length > 0) {
      return validUniqueIds
    }

    return participations.map((p) => p.courseParticipationID)
  }, [filteredApplicationIds, participationById, participations])
  const currentIndex = navigationOrder.findIndex((id) => id === participationId)
  const previousParticipation =
    currentIndex === -1 || navigationOrder.length <= 1
      ? undefined
      : participationById.get(
          navigationOrder[(currentIndex - 1 + navigationOrder.length) % navigationOrder.length],
        )
  const nextParticipation =
    currentIndex === -1 || navigationOrder.length <= 1
      ? undefined
      : participationById.get(navigationOrder[(currentIndex + 1) % navigationOrder.length])

  const navigateToParticipation = (nextParticipationId: string) => {
    navigate(`../${nextParticipationId}`, {
      relative: 'path',
      state: { filteredApplicationIds: navigationOrder },
    })
  }

  const {
    data: fetchedApplication,
    isPending: isFetchingApplication,
    isError: isApplicationError,
    refetch: refetchApplication,
  } = useQuery<GetApplication>({
    queryKey: ['application', participationId],
    queryFn: () => getApplicationAssessment(phaseId ?? '', participationId ?? ''),
    enabled: !!phaseId && !!participationId,
  })

  const {
    data: fetchedApplicationForm,
    isPending: isFetchingApplicationForm,
    isError: isApplicationFormError,
    refetch: refetchApplicationForm,
  } = useQuery<ApplicationForm>({
    queryKey: ['application_form', phaseId],
    queryFn: () => getApplicationForm(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const studentId = fetchedApplication?.student?.id
  const enrollments = useStudentEnrollments(studentId)

  if (isApplicationError || isApplicationFormError) {
    return (
      <ErrorPage
        onRetry={() => {
          refetchApplication()
          refetchApplicationForm()
        }}
      />
    )
  }

  if (isFetchingApplication || isFetchingApplicationForm) {
    return (
      <div className='flex justify-center items-center min-h-screen'>
        <div className='flex flex-col items-center gap-3'>
          <Loader2 className='h-12 w-12 animate-spin text-primary' />
          <span className='text-sm text-muted-foreground'>Loading application details...</span>
        </div>
      </div>
    )
  }

  if (!participation) {
    return (
      <div className='flex justify-center items-center min-h-screen'>
        <p className='text-muted-foreground'>Application not found</p>
      </div>
    )
  }

  return (
    <div className='space-y-6'>
      {previousParticipation && nextParticipation && (
        <div className='flex justify-between'>
          <Button
            variant='outline'
            className={`gap-2 ${getApplicationNavigationButtonColorClass(previousParticipation.passStatus)}`}
            onClick={() => navigateToParticipation(previousParticipation.courseParticipationID)}
          >
            <ChevronLeft className='h-4 w-4' />
            {previousParticipation.student.firstName} {previousParticipation.student.lastName}
          </Button>

          <Button
            variant='outline'
            className={`gap-2 ${getApplicationNavigationButtonColorClass(nextParticipation.passStatus)}`}
            onClick={() => navigateToParticipation(nextParticipation.courseParticipationID)}
          >
            {nextParticipation.student.firstName} {nextParticipation.student.lastName}
            <ChevronRight className='h-4 w-4' />
          </Button>
        </div>
      )}

      {fetchedApplication?.student && status && (
        <StudentProfile student={fetchedApplication.student} status={status} />
      )}

      {fetchedApplication?.student && !fetchedApplication.student.hasUniversityAccount && (
        <MissingUniversityData student={fetchedApplication.student} />
      )}

      <ApplicationDetailPageLayout
        applicationViews={
          <div className='flex flex-col gap-2'>
            {fetchedApplication && fetchedApplicationForm && (
              <ApplicationAnswersTable
                coursePhaseId={phaseId ?? ''}
                questions={[
                  ...fetchedApplicationForm.questionsMultiSelect,
                  ...fetchedApplicationForm.questionsText,
                  ...fetchedApplicationForm.questionsFileUpload,
                ]}
                answersMultiSelect={fetchedApplication.answersMultiSelect}
                answersText={fetchedApplication.answersText}
                answersFileUpload={fetchedApplication.answersFileUpload}
              />
            )}

            {status && (
              <AssessmentCard
                score={score}
                restrictedData={restrictedData}
                acceptanceStatus={status}
                courseParticipationID={participationId ?? ''}
              />
            )}
          </div>
        }
        courseEnrollment={
          <div className='flex flex-col gap-5 border rounded-lg p-5'>
            {enrollments.isLoading && <Loader2 className='h-4 w-4 animate-spin' />}
            {enrollments.isError && <p className='text-destructive'>Failed to load enrollments</p>}
            {enrollments.isSuccess && studentId && (
              <>
                {enrollments.data?.courses.map((ce: CourseEnrollment) => (
                  <div className='flex gap-4 w-full' key={ce.courseId}>
                    <StudentCourseEnrollment courseEnrollment={ce} studentId={studentId} />
                  </div>
                ))}
                <CourseEnrollmentSummary enrollments={enrollments.data?.courses || []} />
              </>
            )}
          </div>
        }
      />
    </div>
  )
}
