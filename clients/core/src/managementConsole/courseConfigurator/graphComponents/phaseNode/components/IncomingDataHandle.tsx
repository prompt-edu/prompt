import type { RequiredInputDTO } from '@core/managementConsole/courseConfigurator/interfaces/requiredInputDto'
import { Handle, Position, useHandleConnections } from '@xyflow/react'
import { useEffect, useState } from 'react'
import { schemaFulfills } from './utils/compareSchema'
import { useCourseConfigurationState } from '@core/managementConsole/courseConfigurator/zustand/useCourseConfigurationStore'
import { CircleCheckBig, OctagonX, TriangleAlert } from 'lucide-react'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'
import { camelToTitle } from './utils/camelToTitle'
import { EDGE_COLOR_GREEN, EDGE_COLOR_PURPLE } from '../../edges/edgeColors'

interface IncomingDataHandleProps {
  phaseID: string
  dto: RequiredInputDTO
  type: 'participation-data' | 'phase-data'
}

export const IncomingDataHandle = ({ phaseID, dto, type }: IncomingDataHandleProps) => {
  const isParticipationEdge = type === 'participation-data'
  const handleName = isParticipationEdge
    ? `participation-data-in-phase-${phaseID}-dto-${dto.id}`
    : `phase-data-in-phase-${phaseID}-dto-${dto.id}`

  const { coursePhaseTypes } = useCourseConfigurationState()
  const [matches, setMatches] = useState(false)
  const incomingEdge = useHandleConnections({
    type: 'target',
    id: handleName,
  })

  const incomingDTOs = incomingEdge
    .map((edge) => {
      if (edge?.sourceHandle && edge?.sourceHandle.split('dto-').length === 2) {
        const dtoID = edge?.sourceHandle.split('dto-')[1]
        return dtoID
      } else {
        return null
      }
    })
    .map((dtoID) =>
      coursePhaseTypes
        .flatMap((phase) =>
          type === 'participation-data'
            ? phase.providedParticipationOutputDTOs
            : phase.providedPhaseOutputDTOs,
        )
        .filter((reqDTO) => reqDTO !== null)
        .find((reqDTO) => reqDTO.id === dtoID),
    )
    .filter((reqDTO) => reqDTO !== null && reqDTO !== undefined) as RequiredInputDTO[]

  useEffect(() => {
    if (incomingDTOs.length > 1) {
      console.log('ERROR: Multiple incoming connections to a DTO')
      setMatches(false)
    } else if (incomingDTOs.length === 1) {
      const incomingDTO = incomingDTOs[0]
      setMatches(schemaFulfills(incomingDTO.specification, dto.specification))
    } else {
      setMatches(false)
    }
  }, [dto.specification, incomingDTOs])

  const status = matches ? 'success' : incomingDTOs.length === 1 ? 'warning' : 'error'

  const statusConfig = {
    success: {
      bgColor: isParticipationEdge ? 'bg-green-50' : 'bg-purple-50',
      textColor: isParticipationEdge ? 'text-green-700' : 'text-purple-700',
      icon: (
        <CircleCheckBig
          className={`w-5 h-5 ${isParticipationEdge ? ' text-green-500' : 'text-purple-500'}`}
        />
      ),
      tooltipText: 'Incoming Data Object Matches',
    },
    warning: {
      bgColor: 'bg-yellow-50',
      textColor: 'text-yellow-700',
      icon: <TriangleAlert className='w-5 h-5 text-yellow-500' />,
      tooltipText: 'Incoming Data Object Does Not Match',
    },
    error: {
      bgColor: 'bg-red-50',
      textColor: 'text-red-700',
      icon: <OctagonX className='w-5 h-5 text-red-500' />,
      tooltipText: 'No Incoming Data Object',
    },
  }

  const { bgColor, textColor, icon, tooltipText } = statusConfig[status]

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div
            className={`flex items-center p-2 rounded-md ${bgColor} ${textColor} relative shadow-sm transition-all duration-200 hover:shadow-md`}
          >
            <Handle
              type='target'
              position={Position.Left}
              id={handleName}
              style={{
                left: '-28px',
                top: '50%',
                background: isParticipationEdge ? EDGE_COLOR_GREEN : EDGE_COLOR_PURPLE,
              }}
              className='!w-3 !h-3 rounded-full'
            />
            <div className='flex items-center space-x-2'>
              {icon}
              <span className='text-sm font-medium'>{camelToTitle(dto.dtoName)}</span>
            </div>
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>{tooltipText}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
