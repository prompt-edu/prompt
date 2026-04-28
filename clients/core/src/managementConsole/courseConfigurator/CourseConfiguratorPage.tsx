import { Canvas } from './Canvas'
import { Loader2 } from 'lucide-react'
import { useCourseConfiguratorDataSetup } from './handlers/useCourseConfiguratorDataSetup'
import { ManagementPageHeader, ErrorPage, Card } from '@tumaet/prompt-ui-components'
import { HelpDialog } from './components/HelpDialog'

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
          <div className='flex justify-center items-center h-64'>
            <Loader2 className='h-12 w-12 animate-spin text-primary' />
          </div>
        ) : (
          <Canvas />
        )}
      </Card>
    </div>
  )
}
