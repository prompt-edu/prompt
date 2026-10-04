import type { MetaDataGraphItem } from '@core/managementConsole/courseConfigurator/interfaces/courseMetaGraphItem'
import type { CoursePhaseGraphItem } from '@core/managementConsole/courseConfigurator/interfaces/coursePhaseGraphItem'
import type { CoursePhaseGraphUpdate } from '@core/managementConsole/courseConfigurator/interfaces/coursePhaseGraphUpdate'
import { API_PREFIX, coreRequest } from '../client'

const ofCourse = (courseID: string) => `${API_PREFIX}/courses/${courseID}`

export const courseGraphs = {
  phase: (courseID: string): Promise<CoursePhaseGraphItem[]> =>
    coreRequest.get(`${ofCourse(courseID)}/phase_graph`),

  phaseData: (courseID: string): Promise<MetaDataGraphItem[]> =>
    coreRequest.get(`${ofCourse(courseID)}/phase_data_graph`),

  participationData: (courseID: string): Promise<MetaDataGraphItem[]> =>
    coreRequest.get(`${ofCourse(courseID)}/participation_data_graph`),

  savePhase: (courseID: string, update: CoursePhaseGraphUpdate): Promise<void> =>
    coreRequest.put(`${ofCourse(courseID)}/phase_graph`, update),

  savePhaseData: (courseID: string, graph: MetaDataGraphItem[]): Promise<void> =>
    coreRequest.put(`${ofCourse(courseID)}/phase_data_graph`, graph),

  saveParticipationData: (courseID: string, graph: MetaDataGraphItem[]): Promise<void> =>
    coreRequest.put(`${ofCourse(courseID)}/participation_data_graph`, graph),
}
