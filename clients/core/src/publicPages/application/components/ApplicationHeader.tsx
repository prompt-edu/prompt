import { OpenApplicationDetails } from '@core/interfaces/application/openApplicationDetails'
import { Button } from '@tumaet/prompt-ui-components'
import { format } from 'date-fns'
import { ArrowLeft, Calendar, Clock } from 'lucide-react'

interface ApplicationHeaderProps {
  applicationPhase: OpenApplicationDetails
  onBackClick: () => void
}

export const ApplicationHeader = ({ applicationPhase, onBackClick }: ApplicationHeaderProps) => {
  return (
    <div>
      <Button onClick={onBackClick} variant='ghost' className='mb-4'>
        <ArrowLeft className='mr-2 h-4 w-4' /> Back to Overview
      </Button>

      <div className='text-center space-y-4'>
        <h1 className='text-3xl font-bold'>{applicationPhase.courseName}</h1>
        <div className='flex flex-col sm:flex-row justify-center items-center space-y-2 sm:space-y-0 sm:space-x-4 text-muted-foreground'>
          <span className='flex items-center'>
            <Calendar className='mr-2 h-4 w-4' />
            {format(applicationPhase.startDate, 'MMM d, yyyy')} -{' '}
            {format(applicationPhase.endDate, 'MMM d, yyyy')}
          </span>
          <span className='hidden sm:inline'>•</span>
          <span className='flex items-center'>
            <Clock className='mr-2 h-4 w-4' />
            Apply by {format(applicationPhase.applicationDeadline, 'MMM d, yyyy')}
          </span>
        </div>
        {applicationPhase.longDescription && (
          <p className='text-base text-muted-foreground leading-relaxed max-w-3xl mx-auto'>
            {applicationPhase.longDescription}
          </p>
        )}
      </div>
    </div>
  )
}
