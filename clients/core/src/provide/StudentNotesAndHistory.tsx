import { StudentNotesAndHistory as StudentNotesAndHistoryPanel } from '@core/managementConsole/shared/components/StudentHistory/StudentNotesAndHistory'
import { TooltipProvider } from '@tumaet/prompt-ui-components'

export function StudentNotesAndHistory({ studentId }: { studentId: string }) {
  return (
    <TooltipProvider>
      <StudentNotesAndHistoryPanel studentId={studentId} />
    </TooltipProvider>
  )
}
