import type { OpenApplicationDetails } from '@core/interfaces/application/openApplicationDetails'
import { getAllOpenApplications } from '@core/network/queries/openApplications'
import { useQuery } from '@tanstack/react-query'
import { Alert, AlertDescription, AlertTitle } from '@tumaet/prompt-ui-components'
import { AlertCircle, Loader2 } from 'lucide-react'
import { NonAuthenticatedPageWrapper } from '../shared/components/NonAuthenticatedPageWrapper'
import { CourseCard } from './components/CourseCard'

export function LandingPage() {
  const {
    data: openApplications,
    isPending,
    isError,
  } = useQuery<OpenApplicationDetails[]>({
    queryKey: ['open_applications'],
    queryFn: () => getAllOpenApplications(),
  })

  return (
    <NonAuthenticatedPageWrapper>
      <section className='text-center mb-16'>
        <h2 className='text-2xl font-bold text-foreground sm:text-3xl lg:text-4xl mb-4'>
          Course Application Portal
        </h2>
        <p className='text-xl text-muted-foreground max-w-4xl mx-auto mb-8'>
          Welcome! You’ll find all courses currently open for applications below—you can apply by
          selecting the course you are interested in.
        </p>
      </section>

      <section className='mb-16'>
        <h3 className='text-2xl font-semibold text-foreground mb-6'>
          Available Courses for Application
        </h3>
        {isPending ? (
          <div className='flex justify-center items-center h-64'>
            <Loader2 className='h-8 w-8 animate-spin text-primary' />
          </div>
        ) : isError ? (
          <Alert variant='destructive'>
            <AlertCircle className='h-4 w-4' />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>
              An error occurred while fetching courses. Please try again later.
            </AlertDescription>
          </Alert>
        ) : openApplications && openApplications.length > 0 ? (
          <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6'>
            {openApplications.map((courseDetails) => (
              <CourseCard key={courseDetails.id} courseDetails={courseDetails} />
            ))}
          </div>
        ) : (
          <Alert>
            <AlertCircle className='h-4 w-4' />
            <AlertTitle>No Open Applications</AlertTitle>
            <AlertDescription>
              No applications are currently open. Please check back later.
            </AlertDescription>
          </Alert>
        )}
      </section>
    </NonAuthenticatedPageWrapper>
  )
}
