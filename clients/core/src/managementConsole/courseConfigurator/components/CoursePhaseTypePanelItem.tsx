import { GripVertical } from 'lucide-react'
import { CoursePhaseType } from '../interfaces/coursePhaseType'
import { CoursePhaseTypeDescription } from './CoursePhaseTypeDescription'

interface CoursePhaseTypePanelItemProps {
  phase: CoursePhaseType
  isDraggable: boolean
  onDragStart?: () => void
}

export const CoursePhaseTypePanelItem = ({
  phase,
  isDraggable,
  onDragStart,
}: CoursePhaseTypePanelItemProps) => {
  const handleDragStart = (event: React.DragEvent<HTMLDivElement>, nodeType: string) => {
    event.dataTransfer.setData('application/@xyflow/react', nodeType)
    event.dataTransfer.effectAllowed = 'move'
    onDragStart?.()
  }

  return (
    <div
      key={phase.id}
      draggable={isDraggable}
      onDragStart={(event) => isDraggable && handleDragStart(event, phase.id)}
      className={`relative group flex items-center justify-between rounded-md border bg-card p-2 ${
        isDraggable ? 'cursor-move hover:bg-accent' : 'cursor-not-allowed opacity-50'
      }`}
    >
      <div className='flex items-center'>
        <GripVertical className='mr-2 h-4 w-4 text-muted-foreground' />
        <span className='text-sm font-medium'>{phase.name}</span>
      </div>
      <CoursePhaseTypeDescription phase={phase} />
    </div>
  )
}
