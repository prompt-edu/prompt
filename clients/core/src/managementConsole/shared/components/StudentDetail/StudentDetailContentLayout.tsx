import { Card, Tabs, TabsContent, TabsList, TabsTrigger } from '@tumaet/prompt-ui-components'
import { GalleryVerticalEnd, NotepadText } from 'lucide-react'
import { ReactNode, useEffect, useState } from 'react'

interface StudentDetailContentProps {
  courseEnrollment: ReactNode
  instructorNotes: ReactNode
  defaultTab?: 'courseEnrollment' | 'instructorNotes'
}

function useIsLargeScreen() {
  const [isLarge, setIsLarge] = useState(() => window.matchMedia('(min-width: 1024px)').matches)

  useEffect(() => {
    const mq = window.matchMedia('(min-width: 1024px)')
    const handler = (e: MediaQueryListEvent) => setIsLarge(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])

  return isLarge
}

function CourseEnrollmentDescriptor() {
  return (
    <div className='flex gap-1'>
      <GalleryVerticalEnd className='w-5 h-5' />
      Course Progression
    </div>
  )
}

function InstructorNoteDescriptor() {
  return (
    <div className='flex gap-1'>
      <NotepadText className='w-5 h-5' />
      Instructor Notes
    </div>
  )
}

function SideBySideView({ courseEnrollment, instructorNotes }: StudentDetailContentProps) {
  return (
    <div className='grid grid-cols-2 mt-4'>
      <div className='lg:pr-2 xl:pr-4'>
        <Card className='p-3'>{courseEnrollment}</Card>
      </div>
      <div>
        <Card className='p-3'>{instructorNotes}</Card>
      </div>
    </div>
  )
}

function TabView({
  courseEnrollment,
  instructorNotes,
  defaultTab = 'courseEnrollment',
}: StudentDetailContentProps) {
  return (
    <Tabs defaultValue={defaultTab}>
      <TabsList className='w-full'>
        <TabsTrigger value='courseEnrollment' className='flex-1'>
          <CourseEnrollmentDescriptor />
        </TabsTrigger>
        <TabsTrigger value='instructorNotes' className='flex-1'>
          <InstructorNoteDescriptor />
        </TabsTrigger>
      </TabsList>
      <TabsContent value='courseEnrollment'>
        <Card className='p-3'>{courseEnrollment}</Card>
      </TabsContent>
      <TabsContent value='instructorNotes'>
        <Card className='p-3'>{instructorNotes}</Card>
      </TabsContent>
    </Tabs>
  )
}

export function StudentDetailContentLayout({
  courseEnrollment,
  instructorNotes,
  defaultTab,
}: StudentDetailContentProps) {
  const isLargeScreen = useIsLargeScreen()

  return isLargeScreen ? (
    <SideBySideView courseEnrollment={courseEnrollment} instructorNotes={instructorNotes} />
  ) : (
    <TabView
      courseEnrollment={courseEnrollment}
      instructorNotes={instructorNotes}
      defaultTab={defaultTab}
    />
  )
}
