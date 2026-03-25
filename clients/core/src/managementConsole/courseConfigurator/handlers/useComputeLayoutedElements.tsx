import { useCourseConfigurationState } from '../zustand/useCourseConfigurationStore'
import { getLayoutedElements } from '../utils/getLayoutedElements'
import { ParticipantEdgeProps } from '../graphComponents/edges/ParticipantEdgeProps'
import { ParticipationDataEdgeProps } from '../graphComponents/edges/ParticipationDataEdgeProps'
import { PhaseDataEdgeProps } from '../graphComponents/edges/PhaseDataEdgeProps'

export const useComputeLayoutedElements = () => {
  const { coursePhases, coursePhaseGraph, participationDataGraph, phaseDataGraph } =
    useCourseConfigurationState()

  const initialNodes = coursePhases.map((phase) => ({
    id: phase.id || `no-valid-id-${Date.now()}`,
    type: 'phaseNode',
    position: phase.position,
    data: {},
  }))

  const initialPersonEdges = coursePhaseGraph.map((item) => {
    return {
      id: 'person-edge-' + item.fromCoursePhaseID + '-' + item.toCoursePhaseID,
      source: item.fromCoursePhaseID,
      target: item.toCoursePhaseID,
      type: 'selectableEdge',
    }
  })

  const initialParticipationDataEdges = participationDataGraph.map((item) => {
    return {
      id:
        'data-edge-from-participation-data-out-phase-' +
        item.fromCoursePhaseID +
        '-dto-' +
        item.fromCoursePhaseDtoID +
        '-to-participation-data-in-phase-' +
        item.toCoursePhaseID +
        '-dto-' +
        item.toCoursePhaseDtoID,
      source: item.fromCoursePhaseID,
      sourceHandle: `participation-data-out-phase-${item.fromCoursePhaseID}-dto-${item.fromCoursePhaseDtoID}`,
      targetHandle: `participation-data-in-phase-${item.toCoursePhaseID}-dto-${item.toCoursePhaseDtoID}`,
      target: item.toCoursePhaseID,
      type: 'selectableEdge',
    }
  })

  const initialPhaseDataEdges = phaseDataGraph.map((item) => {
    return {
      id:
        'data-edge-from-phase-data-out-phase-' +
        item.fromCoursePhaseID +
        '-dto-' +
        item.fromCoursePhaseDtoID +
        '-to-phase-data-in-phase-' +
        item.toCoursePhaseID +
        '-dto-' +
        item.toCoursePhaseDtoID,
      source: item.fromCoursePhaseID,
      sourceHandle: `phase-data-out-phase-${item.fromCoursePhaseID}-dto-${item.fromCoursePhaseDtoID}`,
      targetHandle: `phase-data-in-phase-${item.toCoursePhaseID}-dto-${item.toCoursePhaseDtoID}`,
      target: item.toCoursePhaseID,
      type: 'selectableEdge',
    }
  })

  const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements({
    nodes: initialNodes,
    edges: [...initialPersonEdges, ...initialParticipationDataEdges, ...initialPhaseDataEdges],
  })

  const designedPersonEdges = layoutedEdges
    .filter((edge) => edge.id.startsWith('person-edge'))
    .map((edge) => {
      const participantEdge = ParticipantEdgeProps(edge)
      return {
        ...participantEdge,
        id: edge.id,
        sourceHandle: `participants-out-${edge.source}`,
        targetHandle: `participants-in-${edge.target}`,
      }
    })

  const designedMetaDataEdges = layoutedEdges
    .filter((edge) => edge.id.startsWith('data-edge'))
    .map((edge) => {
      const metaDataEdge = edge.id.startsWith('data-edge-from-participation-data')
        ? ParticipationDataEdgeProps(edge)
        : PhaseDataEdgeProps(edge)
      return {
        ...metaDataEdge,
        id: edge.id,
        sourceHandle: metaDataEdge.sourceHandle,
        targetHandle: metaDataEdge.targetHandle,
      }
    })

  const designedEdges = [...designedPersonEdges, ...designedMetaDataEdges]

  return { nodes: layoutedNodes, edges: designedEdges }
}
