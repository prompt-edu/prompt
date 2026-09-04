import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { Play, Server, Settings, ShieldCheck } from 'lucide-react'
import { Link, useResolvedPath } from 'react-router-dom'

const sections = [
  {
    title: 'Setup',
    description: 'Set the semester tag used in resource name templates.',
    icon: Settings,
    to: 'setup',
  },
  {
    title: 'Providers',
    description: 'Configure credentials and validate them per provider.',
    icon: ShieldCheck,
    to: 'providers',
  },
  {
    title: 'Resource configs',
    description: 'Define what to provision per team or per student.',
    icon: Server,
    to: 'resource-configs',
  },
  {
    title: 'Execution',
    description: 'Trigger provisioning, monitor instances, retry failures.',
    icon: Play,
    to: 'execution',
  },
]

export const OverviewPage = () => {
  const base = useResolvedPath('').pathname.replace(/\/$/, '')

  return (
    <div className='mx-auto max-w-3xl space-y-6 p-6'>
      <div className='flex items-center gap-2'>
        <Server className='h-6 w-6 text-blue-500' />
        <h1 className='text-2xl font-semibold'>Infrastructure Setup</h1>
      </div>
      <p className='text-muted-foreground'>
        Manage infrastructure providers, resource configurations, and execution.
      </p>
      <div className='grid gap-4 sm:grid-cols-2'>
        {sections.map(({ title, description, icon: Icon, to }) => (
          <Link key={to} to={`${base}/${to}`} className='block'>
            <Card className='transition hover:border-blue-300 hover:shadow-sm'>
              <CardHeader>
                <div className='flex items-center gap-2'>
                  <Icon className='h-5 w-5 text-blue-500' />
                  <CardTitle className='text-lg'>{title}</CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                <CardDescription>{description}</CardDescription>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  )
}

export default OverviewPage
