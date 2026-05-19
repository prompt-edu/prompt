import { HelpCircle } from 'lucide-react'
import { Card, CardContent } from '@tumaet/prompt-ui-components'
import { useState, useRef } from 'react'
import { createPortal } from 'react-dom'
import { CoursePhaseType } from '../interfaces/coursePhaseType'
import {
  EDGE_COLOR_BLUE,
  EDGE_COLOR_GREEN,
  EDGE_COLOR_PURPLE,
} from '../graphComponents/edges/edgeColors'
import { camelToTitle } from '../graphComponents/phaseNode/components/utils/camelToTitle'

interface CoursePhaseTypeDescriptionProps {
  phase: CoursePhaseType
}

const TOOLTIP_WIDTH = 260
const TOOLTIP_OFFSET = 8

const LeftDotLabel = ({ color, name }: { color: string; name: string }) => (
  <div className='flex items-start gap-1.5'>
    <span
      className='inline-block w-2 h-2 mt-[3px] rounded-full shrink-0'
      style={{ background: color }}
    />
    <span className='text-xs leading-tight text-muted-foreground'>{name}</span>
  </div>
)

const RightDotLabel = ({ color, name }: { color: string; name: string }) => (
  <div className='flex items-start justify-end gap-1.5'>
    <span className='text-xs leading-tight text-muted-foreground text-right'>{name}</span>
    <span
      className='inline-block w-2 h-2 mt-[3px] rounded-full shrink-0'
      style={{ background: color }}
    />
  </div>
)

export const CoursePhaseTypeDescription = ({ phase }: CoursePhaseTypeDescriptionProps) => {
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

  const participationInputs = (phase.requiredParticipationInputDTOs ?? []).map((dto) =>
    camelToTitle(dto.dtoName),
  )
  const phaseInputs = (phase.requiredPhaseInputDTOs ?? []).map((dto) => camelToTitle(dto.dtoName))
  const participationOutputs = (phase.providedParticipationOutputDTOs ?? []).map((dto) =>
    camelToTitle(dto.dtoName),
  )
  const phaseOutputs = (phase.providedPhaseOutputDTOs ?? []).map((dto) => camelToTitle(dto.dtoName))

  const inputs = [
    ...(!phase.initialPhase ? [{ color: EDGE_COLOR_BLUE, name: 'Participants' }] : []),
    ...phaseInputs.map((name) => ({ color: EDGE_COLOR_PURPLE, name })),
    ...participationInputs.map((name) => ({ color: EDGE_COLOR_GREEN, name })),
  ]

  const outputs = [
    { color: EDGE_COLOR_BLUE, name: 'Participants' },
    ...phaseOutputs.map((name) => ({ color: EDGE_COLOR_PURPLE, name })),
    ...participationOutputs.map((name) => ({ color: EDGE_COLOR_GREEN, name })),
  ]

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
          <div className='fixed z-99' style={{ top: coords.top, left: coords.left }}>
            <Card className='w-64 shadow-lg'>
              <CardContent className='p-4'>
                <p className='font-semibold'>{phase.name}</p>
                <p className='text-sm text-muted-foreground mb-3'>{phase.description}</p>

                <div className='flex'>
                  <div className='flex-1 space-y-1.5'>
                    <p className='text-xs font-semibold'>Inputs</p>
                    {inputs.map((item, i) => (
                      <LeftDotLabel key={i} {...item} />
                    ))}
                  </div>
                  <div className='w-px bg-border shrink-0' />
                  <div className='flex-1 space-y-1.5'>
                    <p className='text-xs font-semibold text-right'>Outputs</p>
                    {outputs.map((item, i) => (
                      <RightDotLabel key={i} {...item} />
                    ))}
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>,
          document.body,
        )}
    </div>
  )
}
