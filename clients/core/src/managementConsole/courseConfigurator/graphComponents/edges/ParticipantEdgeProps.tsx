import { Connection, Edge, MarkerType } from '@xyflow/react'

const EDGE_COLOR_BLUE = '#3b82f6'

export const ParticipantEdgeProps = (params: Edge | Connection) => ({
  ...params,
  animated: false,
  style: { stroke: EDGE_COLOR_BLUE, strokeWidth: 2, strokeDasharray: '5,5' },
  type: 'selectableEdge',
  markerEnd: { type: MarkerType.ArrowClosed, color: EDGE_COLOR_BLUE },
})
