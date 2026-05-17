import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { Server } from 'lucide-react'

export const OverviewPage = () => {
  return (
    <Card className='w-full max-w-2xl mx-auto'>
      <CardHeader>
        <div className='flex items-center space-x-2'>
          <Server className='h-6 w-6 text-blue-500' />
          <CardTitle className='text-2xl'>Infrastructure Setup</CardTitle>
        </div>
        <CardDescription>
          Manage infrastructure providers, resource configurations, and execution
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className='p-4 border-2 border-dashed border-gray-300 rounded-lg text-gray-500'>
          Use the sidebar to navigate to Providers, Resource Configs, or Execution.
        </div>
      </CardContent>
    </Card>
  )
}

export default OverviewPage
