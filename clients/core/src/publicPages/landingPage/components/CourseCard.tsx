import {
  Badge,
  Button,
  Card,
  CardHeader,
  CardTitle,
  CardContent,
} from '@tumaet/prompt-ui-components'
import { DeadlineInfo } from './DeadlineInfo'
import { BookOpen, ArrowRight, Calendar } from 'lucide-react'
import { format } from 'date-fns'
import type { OpenApplicationDetails } from '@core/interfaces/application/openApplicationDetails'
import { CourseTypeDetails } from '@tumaet/prompt-shared-state'
import { useNavigate } from 'react-router-dom'

interface CourseCardProps {
  courseDetails: OpenApplicationDetails
}

export const CourseCard = ({ courseDetails }: CourseCardProps) => {
  const navigate = useNavigate()

  return (
    <Card
      key={courseDetails.id}
      className='group relative overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1 bg-card flex flex-col'
    >
      <CardHeader className='relative pb-3'>
        <div className='flex items-start justify-between gap-3'>
          <CardTitle className='text-xl font-bold leading-tight transition-colors duration-200'>
            {courseDetails.courseName}
          </CardTitle>
        </div>
      </CardHeader>

      <CardContent className='relative space-y-4 flex flex-col justify-between flex-grow'>
        <div>
          {/* Course Type and ECTS Badges */}
          <div className='flex items-center justify-between gap-3'>
            <Badge variant='secondary' className='text-sm font-medium'>
              {CourseTypeDetails[courseDetails.courseType].name}
            </Badge>
            <div className='flex items-center gap-2 text-sm font-medium px-3 py-1 rounded-full border'>
              <BookOpen className='h-4 w-4' />
              {courseDetails.ects} ECTS
            </div>
          </div>

          <div className='p-3 space-y-4'>
            <div className='flex items-center space-x-2'>
              <Calendar className='h-4 w-4 flex-shrink-0' />
              <div className='flex gap-4 flex-1 text-sm font-medium text-gray-600'>
                <div className='flex-1'>
                  <span className='font-semibold text-slate-600'>Course Start</span>
                  <div className='text-slate-900'>
                    {format(courseDetails.startDate, 'MMM d, yyyy')}
                  </div>
                </div>
                <div className='flex-1'>
                  <span className='font-semibold text-slate-600'>Course End</span>
                  <div className='text-slate-900'>
                    {format(courseDetails.endDate, 'MMM d, yyyy')}
                  </div>
                </div>
              </div>
            </div>
            <DeadlineInfo deadline={courseDetails.applicationDeadline} />
          </div>

          {courseDetails.shortDescription && (
            <p className='text-sm text-muted-foreground leading-relaxed whitespace-pre-line'>
              {courseDetails.shortDescription}
            </p>
          )}
        </div>

        {/* Apply Button */}
        <Button
          className={`w-full bg-blue-800 hover:bg-blue-800 text-white font-semibold py-3 transition-all 
            duration-200 transform hover:scale-[1.02] active:scale-[0.98] group/button mt-2`}
          onClick={() => navigate(`/apply/${courseDetails.id}`)}
        >
          <span className='flex items-center justify-center gap-2'>
            Apply Now
            <ArrowRight className='h-4 w-4 transition-transform duration-200 group-hover/button:translate-x-1' />
          </span>
        </Button>
      </CardContent>
    </Card>
  )
}
