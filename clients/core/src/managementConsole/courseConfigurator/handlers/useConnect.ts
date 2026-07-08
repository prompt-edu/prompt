import { addEdge, type Connection, type Edge } from '@xyflow/react'
import { useCallback } from 'react'
import { ParticipantEdgeProps } from '../graphComponents/edges/ParticipantEdgeProps'
import { ParticipationDataEdgeProps } from '../graphComponents/edges/ParticipationDataEdgeProps'
import { PhaseDataEdgeProps } from '../graphComponents/edges/PhaseDataEdgeProps'

export const useConnect = (edges, nodes, setEdges, setIsModified) => {
  return useCallback(
    (params: Edge | Connection) => {
      if (params.sourceHandle && params.targetHandle) {
        const sourceHandle = params.sourceHandle as string
        const targetHandle = params.targetHandle as string

        if (sourceHandle.startsWith('participants') && targetHandle.startsWith('participants')) {
          const targetHasIncoming = edges.some(
            (edge) =>
              edge.target === params.target && edge.targetHandle?.startsWith('participants'),
          )
          const sourceHasOutgoing = edges.some(
            (edge) =>
              edge.source === params.source && edge.sourceHandle?.startsWith('participants'),
          )

          if (!targetHasIncoming && !sourceHasOutgoing) {
            const newEdge = ParticipantEdgeProps(params)
            setEdges((eds) =>
              addEdge({ ...newEdge, id: `person-edge-${newEdge.source}-${newEdge.target}` }, eds),
            )
            setIsModified(true)
          } else {
            console.log(
              'Participants connection not allowed: nodes can have at most one incoming and one outgoing participants edge.',
            )
          }
        } else if (
          (sourceHandle.startsWith('participation-data') &&
            targetHandle.startsWith('participation-data')) ||
          (sourceHandle.startsWith('phase-data') && targetHandle.startsWith('phase-data'))
        ) {
          const sourceNode = nodes.find((node) => node.id === params.source)
          const targetNode = nodes.find((node) => node.id === params.target)

          // enforce that only one incoming data edge per handle
          const targetHasIncoming = edges.some((edge) => {
            return edge.targetHandle === targetHandle
          })

          if (targetHasIncoming) {
            console.log(
              'Metadata connection not allowed: target node can only have one incoming data edge.',
            )
            return
          }

          if (sourceNode && targetNode) {
            const newEdge = sourceHandle.startsWith('participation-data')
              ? ParticipationDataEdgeProps(params)
              : PhaseDataEdgeProps(params)
            setEdges((eds) =>
              addEdge({ ...newEdge, id: `data-edge-from-${sourceHandle}-to-${targetHandle}` }, eds),
            )
            setIsModified(true)
          } else {
            console.log('Metadata connection not allowed: can only connect to subsequent phases.')
          }
        }
      }
      // TODO for phase data nodes
    },
    [edges, nodes, setEdges, setIsModified],
  )
}
