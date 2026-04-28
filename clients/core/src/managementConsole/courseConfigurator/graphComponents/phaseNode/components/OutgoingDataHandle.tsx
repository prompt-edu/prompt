import { ProvidedOutputDTO } from '@core/managementConsole/courseConfigurator/interfaces/providedOutputDto'
import { Handle, Position } from '@xyflow/react'
import { camelToTitle } from './utils/camelToTitle'
import { EDGE_COLOR_GREEN, EDGE_COLOR_PURPLE } from '../../edges/edgeColors'

interface OutgoingDataHandleProps {
  phaseID: string
  dto: ProvidedOutputDTO
  type: 'participation-data' | 'phase-data'
}

export const OutgoingDataHandle = ({ phaseID, dto, type }: OutgoingDataHandleProps) => {
  const isParticipationEdge = type === 'participation-data'
  const handleName = isParticipationEdge
    ? `participation-data-out-phase-${phaseID}-dto-${dto.id}`
    : `phase-data-out-phase-${phaseID}-dto-${dto.id}`
  return (
    <div
      className={`flex items-center justify-end p-2 rounded-md 
        ${isParticipationEdge ? 'bg-green-50 text-green-700' : 'bg-purple-50 text-purple-700'} 
        relative shadow-xs transition-all duration-200`}
    >
      <span className='mr-2 text-sm'>{camelToTitle(dto.dtoName)}</span>
      <Handle
        type='source'
        position={Position.Right}
        id={handleName}
        style={{
          right: '-28px',
          top: '50%',
          background: isParticipationEdge ? EDGE_COLOR_GREEN : EDGE_COLOR_PURPLE,
        }}
        className='w-3! h-3! rounded-full'
      />
    </div>
  )
}
