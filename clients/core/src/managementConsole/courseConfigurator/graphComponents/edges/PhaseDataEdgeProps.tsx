import { Connection, Edge, MarkerType } from '@xyflow/react'
import { EDGE_COLOR_PURPLE } from './edgeColors'

export const PhaseDataEdgeProps = (params: Edge | Connection) => ({
  ...params,
  animated: false,
  style: { stroke: EDGE_COLOR_PURPLE, strokeWidth: 2, strokeDasharray: '5,5' },
  type: 'selectableEdge',
  markerEnd: { type: MarkerType.ArrowClosed, color: EDGE_COLOR_PURPLE },
})
