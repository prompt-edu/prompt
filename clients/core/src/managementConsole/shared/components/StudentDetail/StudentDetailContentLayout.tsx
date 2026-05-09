import { Card, Tabs, TabsContent, TabsList, TabsTrigger } from '@tumaet/prompt-ui-components'
import { GalleryVerticalEnd, NotepadText } from 'lucide-react'
import { ReactNode } from 'react'

interface StudentDetailContentProps {
  courseEnrollment: ReactNode
  instructorNotes: ReactNode
  defaultTab?: 'courseEnrollment' | 'instructorNotes'
}

interface SideBySideViewProps extends StudentDetailContentProps {
  className?: string
}

interface TabViewProps extends StudentDetailContentProps {
  className?: string
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

function SideBySideView({
  courseEnrollment,
  instructorNotes,
  className = '',
}: SideBySideViewProps) {
  return (
    <div className={`grid grid-cols-2 mt-4 ${className}`}>
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
  className = '',
}: TabViewProps) {
  return (
    <div className={className}>
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
    </div>
  )
}

export function StudentDetailContentLayout({
  courseEnrollment,
  instructorNotes,
  defaultTab,
}: StudentDetailContentProps) {
  return (
    <>
      <style>{`
        .student-detail-tab-view { display: block }
        .student-detail-side-by-side { display: none }
        @media (min-width: 1024px) {
          .student-detail-tab-view { display: none }
          .student-detail-side-by-side { display: grid }
        }
      `}</style>
      <TabView
        className='student-detail-tab-view'
        courseEnrollment={courseEnrollment}
        instructorNotes={instructorNotes}
        defaultTab={defaultTab}
      />
      <SideBySideView
        className='student-detail-side-by-side'
        courseEnrollment={courseEnrollment}
        instructorNotes={instructorNotes}
      />
    </>
  )
}
