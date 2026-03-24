import { ManagementPageHeader, Card, CardContent } from '@tumaet/prompt-ui-components'
import { ChevronRight, Download, Trash2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

const privacyOptions = [
  {
    icon: <Download className='w-5 h-5 text-gray-700' />,
    title: 'Data Export',
    description:
      'Download a copy of all personal data stored about you in our systems, including your student record, course participation, and application data.',
    path: '/management/privacy/data-export',
  },
  {
    icon: <Trash2 className='w-5 h-5 text-gray-700' />,
    title: 'Data Deletion',
    description:
      'Request the deletion of your personal data from our systems. This action is irreversible and may affect your access to courses.',
    path: '/management/privacy/data-deletion',
  },
]

export function PrivacyOverviewPage() {
  const navigate = useNavigate()

  return (
    <div>
      <ManagementPageHeader>Privacy</ManagementPageHeader>
      <p className='mb-8 text-gray-600'>
        Manage your personal data stored in our systems. You have the right to access and request
        deletion of your data at any time.
      </p>

      <div className='space-y-4'>
        {privacyOptions.map((option) => (
          <Card
            key={option.path}
            className='border border-gray-200 cursor-pointer hover:border-gray-400 hover:shadow-sm transition-all'
            onClick={() => navigate(option.path)}
          >
            <CardContent className='p-5 flex items-center gap-4'>
              <div className='bg-gray-100 p-3 rounded-full shrink-0'>{option.icon}</div>
              <div className='flex-1 min-w-0'>
                <p className='font-semibold text-gray-900'>{option.title}</p>
                <p className='text-sm text-gray-500 mt-0.5'>{option.description}</p>
              </div>
              <ChevronRight className='w-5 h-5 text-gray-400 shrink-0' />
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
