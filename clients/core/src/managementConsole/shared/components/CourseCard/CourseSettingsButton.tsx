import { Tooltip, TooltipTrigger, Button, TooltipContent } from '@tumaet/prompt-ui-components'
import { Settings } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

export function CourseSettingsButton({ courseID }: { courseID: string }) {
  const navigate = useNavigate()
  const settingsURL = `/management/course/${courseID}/settings`
  return (
    <Tooltip>
      <TooltipTrigger>
        <Button
          variant='outline'
          onClick={() => navigate(settingsURL)}
          className='shrink-0 focus-visible:ring-2 focus-visible:ring-offset-2'
        >
          <Settings className='text-gray-600 dark:text-gray-100' />
        </Button>
      </TooltipTrigger>
      <TooltipContent>Go to Course Settings</TooltipContent>
    </Tooltip>
  )
}
