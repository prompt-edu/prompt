import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import { useEffect } from 'react'
import type { ApplicationMetaData } from '../interfaces/applicationMetaData'
import { parseApplicationMetaData } from '../utils/parseApplicationMetaData'

export const useParseApplicationMetaData = (
  coursePhase: CoursePhaseWithMetaData | undefined,
  setApplicationMetaData: (restrictedData: ApplicationMetaData) => void,
) => {
  return useEffect(() => {
    if (coursePhase) {
      setApplicationMetaData(parseApplicationMetaData(coursePhase.restrictedData))
    }
  }, [coursePhase, setApplicationMetaData])
}
