import { Connection, Edge, MarkerType } from '@xyflow/react'

const EDGE_COLOR_GREEN = '#22c55e'

export const ParticipationDataEdgeProps = (params: Edge | Connection) => ({
  ...params,
  animated: false,
  style: { stroke: EDGE_COLOR_GREEN, strokeWidth: 2, strokeDasharray: '5,5' },
  type: 'selectableEdge',
  markerEnd: { type: MarkerType.ArrowClosed, color: EDGE_COLOR_GREEN },
})
