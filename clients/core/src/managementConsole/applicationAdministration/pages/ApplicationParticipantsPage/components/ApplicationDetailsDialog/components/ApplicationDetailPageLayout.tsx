import { Tabs, TabsContent, TabsList, TabsTrigger } from '@tumaet/prompt-ui-components'
import { ClipboardList, NotepadText } from 'lucide-react'
import { ReactNode } from 'react'

interface ApplicationDetailPageLayoutProps {
  left: ReactNode
  right: ReactNode
}

function SideBySideView({
  left,
  right,
  className = '',
}: ApplicationDetailPageLayoutProps & { className?: string }) {
  return (
    <div className={`grid grid-cols-3 gap-4 ${className}`}>
      <div className='col-span-2 flex flex-col gap-4'>{left}</div>
      <div className='col-span-1 flex flex-col gap-4'>{right}</div>
    </div>
  )
}

function TabView({
  left,
  right,
  className = '',
}: ApplicationDetailPageLayoutProps & { className?: string }) {
  return (
    <div className={className}>
      <Tabs defaultValue='application'>
        <TabsList className='w-full'>
          <TabsTrigger value='application' className='flex-1'>
            <div className='flex gap-1'>
              <ClipboardList className='w-5 h-5' />
              Application
            </div>
          </TabsTrigger>
          <TabsTrigger value='notes' className='flex-1'>
            <div className='flex gap-1'>
              <NotepadText className='w-5 h-5' />
              Notes & Enrollment
            </div>
          </TabsTrigger>
        </TabsList>
        <TabsContent value='application'>
          <div className='flex flex-col gap-4'>{left}</div>
        </TabsContent>
        <TabsContent value='notes'>
          <div className='flex flex-col gap-4'>{right}</div>
        </TabsContent>
      </Tabs>
    </div>
  )
}

export function ApplicationDetailPageLayout({ left, right }: ApplicationDetailPageLayoutProps) {
  return (
    <>
      <TabView className='lg:hidden' left={left} right={right} />
      <SideBySideView className='hidden lg:grid' left={left} right={right} />
    </>
  )
}
