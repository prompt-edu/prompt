import { useCallback } from 'react'
import { Handle, Position, useReactFlow } from '@xyflow/react'
import { Users } from 'lucide-react'
import { EDGE_COLOR_BLUE } from '../edges/edgeColors'
import { useCourseConfigurationState } from '../../zustand/useCourseConfigurationStore'
import { Badge, Card, CardContent, CardHeader, Separator } from '@tumaet/prompt-ui-components'
import { IncomingDataHandle } from './components/IncomingDataHandle'
import { OutgoingDataHandle } from './components/OutgoingDataHandle'
import { NameEditingHeader } from './components/NameEditingHeader'

export function PhaseNode({ id, selected }: { id: string; selected?: boolean }) {
  // Retrieve phase and phase type data
  const { coursePhases, coursePhaseTypes } = useCourseConfigurationState()
  const coursePhase = coursePhases.find((phase) => phase.id === id)
  const phaseType = coursePhaseTypes.find((type) => type.id === coursePhase?.coursePhaseTypeID)

  const { setNodes } = useReactFlow()
  const onNodeClick = useCallback(() => {
    setNodes((nds) =>
      nds.map((node) => {
        node.selected = node.id === id
        return node
      }),
    )
  }, [id, setNodes])

  return (
    <Card
      className={`w-80 shadow-lg hover:shadow-xl transition-shadow duration-300 relative ${
        selected ? 'ring-2 ring-blue-500' : ''
      }`}
      onClick={onNodeClick}
    >
      {/* Card Header */}
      <CardHeader className='p-4'>
        <NameEditingHeader phaseID={id} />
        <Badge variant='secondary' className='mt-2 mr-auto'>
          {phaseType?.name}
        </Badge>
      </CardHeader>

      <Separator />

      {/* Card Content */}
      <CardContent className='p-4 relative'>
        {/* 1. Participants Row */}
        <div className='participants-row relative flex items-center justify-between bg-blue-100 p-2 rounded mb-4'>
          {!phaseType?.initialPhase && (
            <Handle
              type='target'
              position={Position.Left}
              id={`participants-in-${id}`}
              style={{ left: '-28px', background: EDGE_COLOR_BLUE }}
              className='w-3! h-3! rounded-full'
            />
          )}
          <div className='flex items-center justify-center w-full'>
            <Users className='w-6 h-6 text-blue-500 mr-2' />
            <span className='text-sm font-medium text-blue-500'>Participants</span>
          </div>
          <Handle
            type='source'
            position={Position.Right}
            id={`participants-out-${id}`}
            style={{ right: '-28px', background: EDGE_COLOR_BLUE }}
            className='w-3! h-3! rounded-full'
          />
        </div>

        <div className='flex flex-col gap-4'>
          {/* 2. Phase Data Inputs (required Input DTOs) */}
          {phaseType?.requiredPhaseInputDTOs && phaseType.requiredPhaseInputDTOs.length > 0 && (
            <div>
              <h4 className='text-sm font-semibold mb-2'>Phase Inputs:</h4>
              <div className='meta-data-inputs space-y-2 mr-16'>
                {phaseType.requiredPhaseInputDTOs.map((dto) => (
                  <IncomingDataHandle key={dto.id} phaseID={id} dto={dto} type={'phase-data'} />
                ))}
              </div>
            </div>
          )}

          {/* 3. Phase Data Outputs (provided Output DTOs) */}
          {phaseType?.providedPhaseOutputDTOs && phaseType.providedPhaseOutputDTOs.length > 0 && (
            <div>
              <h4 className='text-sm font-semibold mb-2'>Provided Phase Outputs:</h4>
              <div className='meta-data-outputs space-y-2 ml-16'>
                {phaseType.providedPhaseOutputDTOs.map((dto) => (
                  <OutgoingDataHandle key={dto.id} phaseID={id} dto={dto} type={'phase-data'} />
                ))}
              </div>
            </div>
          )}

          {/* 4. Participation Data Inputs (required Input DTOs) */}
          {phaseType?.requiredParticipationInputDTOs &&
            phaseType.requiredParticipationInputDTOs.length > 0 && (
              <div>
                <h4 className='text-sm font-semibold mb-2'>Participation Inputs:</h4>
                <div className='meta-data-inputs space-y-2 mr-16'>
                  {phaseType.requiredParticipationInputDTOs.map((dto) => (
                    <IncomingDataHandle
                      key={dto.id}
                      phaseID={id}
                      dto={dto}
                      type={'participation-data'}
                    />
                  ))}
                </div>
              </div>
            )}

          {/* 5. Participation Data Outputs (provided Output DTOs) */}
          {phaseType?.providedParticipationOutputDTOs &&
            phaseType.providedParticipationOutputDTOs.length > 0 && (
              <div>
                <h4 className='text-sm font-semibold mb-2'>Provided Participation Outputs:</h4>
                <div className='meta-data-outputs space-y-2 ml-16'>
                  {phaseType.providedParticipationOutputDTOs.map((dto) => (
                    <OutgoingDataHandle
                      key={dto.id}
                      phaseID={id}
                      dto={dto}
                      type={'participation-data'}
                    />
                  ))}
                </div>
              </div>
            )}
        </div>
      </CardContent>
    </Card>
  )
}
