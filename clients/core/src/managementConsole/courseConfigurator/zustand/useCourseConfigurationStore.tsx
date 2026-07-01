import { create } from 'zustand'
import { MetaDataGraphItem } from '../interfaces/courseMetaGraphItem'
import { CoursePhaseGraphItem } from '../interfaces/coursePhaseGraphItem'
import { CoursePhaseType } from '../interfaces/coursePhaseType'
import { CoursePhaseWithPosition } from '../interfaces/coursePhaseWithPosition'

interface CourseConfigurationState {
  coursePhaseTypes: CoursePhaseType[]
  coursePhaseGraph: CoursePhaseGraphItem[]
  coursePhases: CoursePhaseWithPosition[]
  participationDataGraph: MetaDataGraphItem[]
  phaseDataGraph: MetaDataGraphItem[]
  setCoursePhaseTypes: (coursePhaseTypes: CoursePhaseType[]) => void
  removeUnsavedCoursePhases: () => void
  appendCoursePhaseType: (coursePhaseType: CoursePhaseType) => void
  setCoursePhaseGraph: (coursePhaseGraph: CoursePhaseGraphItem[]) => void
  setCoursePhases: (coursePhases: CoursePhaseWithPosition[]) => void
  appendCoursePhase: (coursePhase: CoursePhaseWithPosition) => void
  setParticipationDataGraph: (participationDataGraph: MetaDataGraphItem[]) => void
  setPhaseDataGraph: (phaseDataGraph: MetaDataGraphItem[]) => void
}

export const useCourseConfigurationState = create<CourseConfigurationState>((set) => ({
  coursePhaseTypes: [],
  coursePhaseGraph: [],
  coursePhases: [],
  participationDataGraph: [],
  phaseDataGraph: [],
  setCoursePhaseTypes: (coursePhaseTypes) => set({ coursePhaseTypes }),
  removeUnsavedCoursePhases: () =>
    set((state) => ({
      coursePhases: state.coursePhases.filter(
        (phase) => phase.id && !phase.id.startsWith('no-valid-id'),
      ),
    })),
  appendCoursePhaseType: (coursePhaseType: CoursePhaseType) =>
    set((state) => ({
      coursePhaseTypes: state.coursePhaseTypes.concat(coursePhaseType),
    })),
  setCoursePhaseGraph: (coursePhaseGraph) => set({ coursePhaseGraph }),
  setCoursePhases: (coursePhases) => set({ coursePhases }),
  appendCoursePhase: (coursePhase) =>
    set((state) => ({ coursePhases: state.coursePhases.concat(coursePhase) })),
  setParticipationDataGraph: (participationDataGraph) => set({ participationDataGraph }),
  setPhaseDataGraph: (phaseDataGraph) => set({ phaseDataGraph }),
}))
