import { HelpCircle } from 'lucide-react'
import { Card, CardContent } from '@tumaet/prompt-ui-components'
import { useState, useRef } from 'react'
import { createPortal } from 'react-dom'

interface CoursePhaseTypeDescriptionProps {
  title: string
  description: string
}

const TOOLTIP_WIDTH = 260
const TOOLTIP_OFFSET = 8

export const CoursePhaseTypeDescription = ({
  title,
  description,
}: CoursePhaseTypeDescriptionProps) => {
  const [open, setOpen] = useState(false)
  const [coords, setCoords] = useState<{ top: number; left: number } | null>(null)
  const iconRef = useRef<HTMLDivElement>(null)

  const handleMouseEnter = () => {
    if (iconRef.current) {
      const rect = iconRef.current.getBoundingClientRect()
      const viewportWidth = window.innerWidth

      const top = rect.bottom + TOOLTIP_OFFSET
      const left =
        rect.right + TOOLTIP_WIDTH > viewportWidth ? rect.left - TOOLTIP_WIDTH : rect.left

      setCoords({ top, left })
    }
    setOpen(true)
  }

  return (
    <div
      ref={iconRef}
      className='inline-block'
      onMouseEnter={handleMouseEnter}
      onMouseLeave={() => setOpen(false)}
    >
      <HelpCircle className='h-4 w-4 cursor-pointer' />

      {open &&
        coords &&
        createPortal(
          <div className='fixed z-[9999]' style={{ top: coords.top, left: coords.left }}>
            <Card className='w-64 shadow-lg'>
              <CardContent className='p-4'>
                <p className='font-semibold'>{title}</p>
                <p className='text-sm text-muted-foreground'>{description}</p>
              </CardContent>
            </Card>
          </div>,
          document.body,
        )}
    </div>
  )
}
