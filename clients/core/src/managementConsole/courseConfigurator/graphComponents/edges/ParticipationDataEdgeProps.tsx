import { Connection, Edge, MarkerType } from '@xyflow/react'
import { EDGE_COLOR_GREEN } from './edgeColors'

export const ParticipationDataEdgeProps = (params: Edge | Connection) => ({
  ...params,
  animated: false,
  style: { stroke: EDGE_COLOR_GREEN, strokeWidth: 2, strokeDasharray: '5,5' },
  type: 'selectableEdge',
  markerEnd: { type: MarkerType.ArrowClosed, color: EDGE_COLOR_GREEN },
})
