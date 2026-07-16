import type { UseMutateAsyncFunction } from '@tanstack/react-query'
import type { CreateCoursePhase, UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import type { Edge, Node } from '@xyflow/react'
import type { MetaDataGraphItem } from '../interfaces/courseMetaGraphItem'
import type { CoursePhaseGraphItem } from '../interfaces/coursePhaseGraphItem'
import type { CoursePhaseGraphUpdate } from '../interfaces/coursePhaseGraphUpdate'
import type { CoursePhaseWithPosition } from '../interfaces/coursePhaseWithPosition'
import { getMetaDataGraphFromEdges } from '../utils/getMetaDataGraphFromEdges'

interface HandleSaveProps {
  nodes: Node[]
  edges: Edge[]
  coursePhases: CoursePhaseWithPosition[]
  mutateDeletePhase: UseMutateAsyncFunction<string | undefined, Error, string, unknown>
  mutateAsyncPhases: (coursePhase: CreateCoursePhase) => Promise<string | undefined>
  mutateRenamePhase: UseMutateAsyncFunction<string | undefined, Error, UpdateCoursePhase, unknown>
  mutateCoursePhaseGraph: UseMutateAsyncFunction<void, Error, CoursePhaseGraphUpdate, unknown>
  mutateParticipationDataGraph: UseMutateAsyncFunction<void, Error, MetaDataGraphItem[], unknown>
  mutatePhaseDataGraph: UseMutateAsyncFunction<void, Error, MetaDataGraphItem[], unknown>
  queryClient: any
  setIsModified: (val: boolean) => void
}

// TODO: move this to the server side to enable transaction control!
export async function handleSave({
  nodes,
  edges,
  coursePhases,
  mutateDeletePhase,
  mutateAsyncPhases,
  mutateRenamePhase,
  mutateCoursePhaseGraph,
  mutateParticipationDataGraph,
  mutatePhaseDataGraph,
  queryClient,
  setIsModified,
}: HandleSaveProps) {
  const idReplacementMap: { [key: string]: string } = {}

  // 0.) Remove nodes that no longer exist in the graph
  const nodesToRemove = coursePhases.filter(
    (phase) => phase.id && !nodes.find((node) => node.id === phase.id),
  )
  for (const node of nodesToRemove) {
    try {
      await mutateDeletePhase(node.id as string)
    } catch (err) {
      console.error('Error deleting course phase', err)
      return
    }
  }

  // 1.) Add new phases
  const newPhases = coursePhases.filter((phase) => !phase.id || phase.id.startsWith('no-valid-id'))
  for (const phase of newPhases) {
    const createPhase: CreateCoursePhase = {
      courseID: phase.courseID,
      name: phase.name,
      coursePhaseTypeID: phase.coursePhaseTypeID,
      isInitialPhase: phase.isInitialPhase,
    }

    try {
      const newId = await mutateAsyncPhases(createPhase)
      if (phase.id) {
        idReplacementMap[phase.id] = newId as string
      }
    } catch (err) {
      console.error('Error saving course phase', err)
      return
    }
  }

  // 2.) Update the names of the phases if any
  const updatedPhases = coursePhases.filter((phase) => phase.isModified)
  for (const updatedPhase of updatedPhases) {
    try {
      await mutateRenamePhase({
        id: updatedPhase.id as string,
        name: updatedPhase.name,
      })
    } catch (err) {
      console.error('Error saving course phase', err)
      return
    }
  }

  // 3.) Update Course Graph Edges with replaced IDs if necessary
  const updatedPersonEdges = edges
    .filter((edge) => {
      return edge.id.startsWith('person-edge-')
    })
    .map((edge) => {
      const newSource = idReplacementMap[edge.source] || edge.source
      const newTarget = idReplacementMap[edge.target] || edge.target
      return { ...edge, source: newSource, target: newTarget }
    })

  const orderArray: CoursePhaseGraphItem[] = updatedPersonEdges.map((edge) => ({
    fromCoursePhaseID: edge.source,
    toCoursePhaseID: edge.target,
  }))

  let initialPhase = coursePhases.find((phase) => phase.isInitialPhase)?.id ?? 'undefined'
  if (initialPhase.startsWith('no-valid-id')) {
    if (idReplacementMap[initialPhase]) {
      initialPhase = idReplacementMap[initialPhase]
    } else {
      console.error('Initial phase has invalid ID')
    }
  }

  const graphUpdate: CoursePhaseGraphUpdate = {
    initialPhase: initialPhase,
    coursePhaseGraph: orderArray,
  }

  try {
    await mutateCoursePhaseGraph(graphUpdate)
  } catch (err) {
    console.error('Error saving course phase', err)
    return err
  }

  // 4.) Update the phase data graph with the new edges
  const updatedPhaseDataEdges = edges
    .filter((edge) => {
      return edge.id.startsWith('data-edge-from-phase')
    })
    .map((edge) => {
      const newSource = idReplacementMap[edge.source] || edge.source
      const newTarget = idReplacementMap[edge.target] || edge.target
      return { ...edge, source: newSource, target: newTarget }
    })

  console.log('Updated Phase Data Edges', updatedPhaseDataEdges)
  console.log('Edges', edges)
  const phaseDataGraph: MetaDataGraphItem[] = getMetaDataGraphFromEdges(updatedPhaseDataEdges)
  try {
    await mutatePhaseDataGraph(phaseDataGraph)
  } catch (err) {
    console.error('Error saving phase data graph', err)
    return err
  }

  // 5.) Update the participation graph with the new edges
  const updatedParticipationDataEdges = edges
    .filter((edge) => {
      return edge.id.startsWith('data-edge-from-participation')
    })
    .map((edge) => {
      const newSource = idReplacementMap[edge.source] || edge.source
      const newTarget = idReplacementMap[edge.target] || edge.target
      return { ...edge, source: newSource, target: newTarget }
    })

  const participationDataGraph: MetaDataGraphItem[] = getMetaDataGraphFromEdges(
    updatedParticipationDataEdges,
  )
  // TODO include the phase data graph

  try {
    await mutateParticipationDataGraph(participationDataGraph)
    queryClient.invalidateQueries({
      queryKey: [
        'courses',
        'participation_data_phase_graph',
        'phase_data_phase_graph',
        'course_phase_types',
        'course_phase_graph',
      ],
    })
    setIsModified(false)
    // Optionally reload if needed
    // window.location.reload()
  } catch (err) {
    console.error('Error saving course phase', err)
    return err
  }
}
