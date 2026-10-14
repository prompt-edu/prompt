import { keepPreviousData, useQuery } from '@tanstack/react-query'
import type { CoursePhaseWithType } from '@tumaet/prompt-shared-state'
import {
  Button,
  ErrorPage,
  LoadingPage,
  ManagementPageHeader,
  PromptTable,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { AICallDetailDialog } from '../components/AICallDetailDialog'
import { getAICallColumns } from '../components/aiCallColumns'
import { useCoursePhases } from '../hooks/useCoursePhases'
import type { AICallCursor } from '../interfaces/aiCallPage'
import { getAICalls } from '../network/queries/getAICalls'
import { currentCursor, hasNewerPage, popNewerPage, pushOlderPage } from '../utils/paging'

const PAGE_SIZE = 50

export const AICallsPage = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const phases = useCoursePhases(courseId)

  return (
    <div className='space-y-6'>
      <ManagementPageHeader>AI Calls</ManagementPageHeader>
      {phases.length === 0 ? (
        <p className='text-muted-foreground'>This course has no phases yet.</p>
      ) : (
        <AICallsBrowser key={courseId} phases={phases} />
      )}
    </div>
  )
}

const AICallsBrowser = ({ phases }: { phases: CoursePhaseWithType[] }) => {
  const [phaseId, setPhaseId] = useState(phases[0].id)
  const [cursorStack, setCursorStack] = useState<AICallCursor[]>([])
  const [openCallId, setOpenCallId] = useState<string>()
  const columns = useMemo(() => getAICallColumns(setOpenCallId), [])

  const cursor = currentCursor(cursorStack)
  const query = useQuery({
    queryKey: ['aiCalls', phaseId, PAGE_SIZE, cursor],
    queryFn: () => getAICalls(phaseId, PAGE_SIZE, cursor),
    placeholderData: keepPreviousData,
  })
  const page = query.data

  const selectPhase = (id: string) => {
    setPhaseId(id)
    setCursorStack([])
  }

  return (
    <>
      <Select value={phaseId} onValueChange={selectPhase}>
        <SelectTrigger className='w-72' aria-label='Course phase'>
          <SelectValue placeholder='Select a course phase' />
        </SelectTrigger>
        <SelectContent>
          {phases.map((phase) => (
            <SelectItem key={phase.id} value={phase.id}>
              {phase.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {query.isError ? (
        <ErrorPage message='Failed to load the AI calls.' />
      ) : !page ? (
        <LoadingPage />
      ) : (
        <>
          <PromptTable data={page.calls} columns={columns} pageSize={PAGE_SIZE} />
          {(hasNewerPage(cursorStack) || page.nextCursor) && (
            <div className='flex items-center justify-center gap-2'>
              <Button
                variant='outline'
                onClick={() => setCursorStack(popNewerPage)}
                disabled={!hasNewerPage(cursorStack) || query.isFetching}
              >
                <ChevronLeft className='mr-1 h-4 w-4' />
                Newer calls
              </Button>
              <Button
                variant='outline'
                onClick={() => setCursorStack((stack) => pushOlderPage(stack, page.nextCursor))}
                disabled={!page.nextCursor || query.isFetching}
              >
                Older calls
                <ChevronRight className='ml-1 h-4 w-4' />
              </Button>
            </div>
          )}
        </>
      )}

      <AICallDetailDialog
        phaseId={phaseId}
        callId={openCallId}
        onClose={() => setOpenCallId(undefined)}
      />
    </>
  )
}
