import { Tabs, TabsContent, TabsList, TabsTrigger } from '@tumaet/prompt-ui-components'
import { ClipboardList, NotepadText } from 'lucide-react'
import { ReactNode, useEffect, useState } from 'react'

interface ApplicationDetailPageLayoutProps {
  left: ReactNode
  right: ReactNode
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

function SideBySideView({ left, right }: ApplicationDetailPageLayoutProps) {
  return (
    <div className='grid grid-cols-3 gap-4'>
      <div className='col-span-2 flex flex-col gap-4'>{left}</div>
      <div className='col-span-1 flex flex-col gap-4'>{right}</div>
    </div>
  )
}

function TabView({ left, right }: ApplicationDetailPageLayoutProps) {
  return (
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
  )
}

export function ApplicationDetailPageLayout({ left, right }: ApplicationDetailPageLayoutProps) {
  const isLargeScreen = useIsLargeScreen()

  return isLargeScreen ? (
    <SideBySideView left={left} right={right} />
  ) : (
    <TabView left={left} right={right} />
  )
}
