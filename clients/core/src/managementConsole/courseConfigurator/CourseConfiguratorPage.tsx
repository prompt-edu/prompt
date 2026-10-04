import { Card, ErrorPage, LoadingPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Canvas } from './Canvas'
import { HelpDialog } from './components/HelpDialog'
import { useCourseConfiguratorDataSetup } from './handlers/useCourseConfiguratorDataSetup'

export default function CourseConfiguratorPage() {
  const { isError, isPending, error, finishedSetup, refetchAll } = useCourseConfiguratorDataSetup()

  return (
    <div className='h-full flex flex-col min-h-0'>
      <div className='flex items-center justify-between mb-4'>
        <div className='-mb-6'>
          <ManagementPageHeader>Course Configurator</ManagementPageHeader>
        </div>
        <HelpDialog />
      </div>

      <Card className='grow min-h-0 flex flex-col overflow-hidden'>
        {isError ? (
          <ErrorPage
            title='Error'
            description='Failed to fetch course phase types'
            message={error?.message}
            onRetry={() => refetchAll()}
          />
        ) : isPending || !finishedSetup ? (
          <LoadingPage />
        ) : (
          <Canvas />
        )}
      </Card>
    </div>
  )
}
