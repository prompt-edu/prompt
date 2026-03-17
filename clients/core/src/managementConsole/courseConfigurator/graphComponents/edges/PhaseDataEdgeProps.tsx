import { Connection, Edge, MarkerType } from '@xyflow/react'

const EDGE_COLOR_PURPLE = '#a855f7'

export const PhaseDataEdgeProps = (params: Edge | Connection) => ({
  ...params,
  animated: false,
  style: { stroke: EDGE_COLOR_PURPLE, strokeWidth: 2, strokeDasharray: '5,5' },
  type: 'selectableEdge',
  markerEnd: { type: MarkerType.ArrowClosed, color: EDGE_COLOR_PURPLE },
})
