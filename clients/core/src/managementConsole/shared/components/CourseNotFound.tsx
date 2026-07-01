import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { useNavigate } from 'react-router-dom'

interface CourseNotFoundProps {
  courseId: string
}

export default function CourseNotFound({ courseId }: CourseNotFoundProps) {
  const navigate = useNavigate()
  return (
    <div className='flex items-center justify-center min-h-screen bg-gray-100'>
      <Card className='w-full max-w-md'>
        <CardHeader>
          <CardTitle className='text-2xl font-bold text-center'>Course Not Found</CardTitle>
          <CardDescription className='text-center'>
            We couldn&apos;t find the course you&apos;re looking for.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className='text-center mb-4'>
            The course with ID <span className='font-semibold'>{courseId}</span> does not exist or
            may have been removed.
          </p>
        </CardContent>
        <CardFooter className='flex justify-center'>
          <Button onClick={() => navigate('/management/courses')}>Return to Overview</Button>
        </CardFooter>
      </Card>
    </div>
  )
}
