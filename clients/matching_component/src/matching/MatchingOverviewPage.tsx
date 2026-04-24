import { useEffect } from 'react'
import { ClipboardList, FileUp, Loader2, UserRoundCheck } from 'lucide-react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { CoursePhaseParticipationsWithResolution } from '@tumaet/prompt-shared-state'
import { getCoursePhaseParticipations } from '@tumaet/prompt-shared-state'
import { useMatchingStore } from './zustand/useMatchingStore'
import { UploadButton } from './components/UploadButton'
import { useUploadAndParseXLSX } from './hooks/useUploadAndParseXLSX'
import { useUploadAndParseCSV } from './hooks/useUploadAndParseCSV'
import {
  ManagementPageHeader,
  ErrorPage,
  Separator,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'

export const MatchingOverviewPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const path = useLocation().pathname

  const { setParticipations } = useMatchingStore()
  const { parseFileXLSX } = useUploadAndParseXLSX()
  const { parseFileCSV } = useUploadAndParseCSV()

  const {
    data: coursePhaseParticipations,
    isPending: isCoursePhaseParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
  } = useQuery<CoursePhaseParticipationsWithResolution>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
  })

  useEffect(() => {
    if (coursePhaseParticipations) {
      setParticipations(coursePhaseParticipations.participations)
    }
  }, [coursePhaseParticipations, setParticipations])

  return (
    <div>
      <ManagementPageHeader>Matching Data Export and Import</ManagementPageHeader>
      {isParticipationsError ? (
        <ErrorPage onRetry={refetchCoursePhaseParticipations} />
      ) : isCoursePhaseParticipationsPending ? (
        <div className='flex justify-center items-center h-64'>
          <Loader2 className='h-12 w-12 animate-spin text-primary' />
        </div>
      ) : (
        <div>
          <div className='grid md:grid-cols-2 gap-8'>
            <section className='space-y-4'>
              <h2 className='text-2xl font-bold flex items-center'>
                <span className='bg-primary text-primary-foreground rounded-full w-8 h-8 flex items-center justify-center mr-2'>
                  1
                </span>
                Data Export
              </h2>
              <UploadButton
                title='Export Data for TUM Matching'
                description='Upload the file which you have received from TUM Matching to enter the ranks.'
                icon={<FileUp className='h-6 w-6 mr-2' />}
                onUploadFinish={() => {
                  navigate(`${path}/export`)
                }}
                onUploadFunction={parseFileXLSX}
                acceptedFileTypes={['.xlsx', '.xls']}
              />
            </section>
            <section className='space-y-4'>
              <h2 className='text-2xl font-bold flex items-center'>
                <span className='bg-primary text-primary-foreground rounded-full w-8 h-8 flex items-center justify-center mr-2'>
                  2
                </span>
                Data Re-Import
              </h2>
              <UploadButton
                title='Re-Import Assigned Students'
                description='Upload the students that the TUM Matching System has assigned to this course.'
                icon={<UserRoundCheck className='h-6 w-6 mr-2' />}
                onUploadFinish={() => {
                  navigate(`${path}/re-import`)
                }}
                onUploadFunction={parseFileCSV}
                acceptedFileTypes={['.csv']}
              />
            </section>
          </div>
          <Separator className='mt-16 mb-16' />
          <div>
            <Card className='w-[350px]'>
              <CardHeader>
                <CardTitle className='text-2xl font-bold flex items-center'>
                  <span className='bg-primary text-primary-foreground rounded-full w-8 h-8 flex items-center justify-center mr-2'>
                    3
                  </span>
                  Manual Data Inspection
                </CardTitle>
                <CardDescription>
                  Review and manage participants manually. This option allows you to inspect, pass,
                  or drop students outside of the automated import/export process.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Button onClick={() => navigate(`${path}/participants`)}>
                  <ClipboardList className='h-5 w-5 mr-2' />
                  Manual Inspection
                </Button>
              </CardContent>
            </Card>
          </div>
        </div>
      )}
    </div>
  )
}
