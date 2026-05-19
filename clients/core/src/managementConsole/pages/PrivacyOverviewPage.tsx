import { ManagementPageHeader, Card, CardContent } from '@tumaet/prompt-ui-components'
import { ChevronRight, Download, Trash2 } from 'lucide-react'
import { Link, useNavigate } from 'react-router-dom'

const privacyOptions = [
  {
    icon: <Download className='w-5 h-5 text-foreground' />,
    title: 'Data Export',
    description: 'Download a copy of all personal data stored about you in our systems.',
    path: '/management/privacy/data-export',
  },
  {
    icon: <Trash2 className='w-5 h-5 text-foreground' />,
    title: 'Data Deletion',
    description: 'Request the deletion of your personal data from our systems.',
    path: '/management/privacy/data-deletion',
  },
]

export function PrivacyOverviewPage() {
  const navigate = useNavigate()

  return (
    <div>
      <ManagementPageHeader>Privacy</ManagementPageHeader>
      <p className='mb-8 text-muted-foreground'>Manage your personal data stored in our systems.</p>

      <div className='space-y-4'>
        {privacyOptions.map((option) => (
          <Card
            key={option.path}
            role='button'
            tabIndex={0}
            className='border border-border cursor-pointer hover:border-muted-foreground/40 hover:shadow-xs transition-all'
            onClick={() => navigate(option.path)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                navigate(option.path)
              }
            }}
          >
            <CardContent className='p-5 flex items-center gap-4'>
              <div className='bg-muted p-3 rounded-full shrink-0'>{option.icon}</div>
              <div className='flex-1 min-w-0'>
                <p className='font-semibold text-foreground'>{option.title}</p>
                <p className='text-sm text-muted-foreground mt-0.5'>{option.description}</p>
              </div>
              <ChevronRight className='w-5 h-5 text-muted-foreground shrink-0' />
            </CardContent>
          </Card>
        ))}
      </div>

      <p className='mt-8 text-muted-foreground'>
        Make sure to read{' '}
        <Link to='/privacy' className='underline'>
          Prompt&apos;s Privacy Policy
        </Link>
      </p>
      <p className='text-muted-foreground'>
        For any remaining questions, you can contact the Privacy contacts mentioned there.
      </p>
    </div>
  )
}
