import type { CoursePhaseGraphItem } from './coursePhaseGraphItem'

export interface CoursePhaseGraphUpdate {
  initialPhase: string
  coursePhaseGraph: CoursePhaseGraphItem[]
}
