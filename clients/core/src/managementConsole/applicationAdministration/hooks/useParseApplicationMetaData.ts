import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import { useEffect } from 'react'
import type { ApplicationMetaData } from '../interfaces/applicationMetaData'

export const useParseApplicationMetaData = (
  coursePhase: CoursePhaseWithMetaData | undefined,
  setApplicationMetaData: (restrictedData: ApplicationMetaData) => void,
) => {
  return useEffect(() => {
    if (coursePhase) {
      const externalStudentsAllowed = coursePhase?.restrictedData?.externalStudentsAllowed
      const applicationStartDate = coursePhase?.restrictedData?.applicationStartDate
      const applicationEndDate = coursePhase?.restrictedData?.applicationEndDate
      const universityLoginAvailable = coursePhase?.restrictedData?.universityLoginAvailable
      const autoAccept = coursePhase?.restrictedData?.autoAccept
      const customScores = coursePhase?.restrictedData?.useCustomScores

      const parsedMetaData: ApplicationMetaData = {
        applicationStartDate: applicationStartDate ? new Date(applicationStartDate) : undefined,
        applicationEndDate: applicationEndDate ? new Date(applicationEndDate) : undefined,
        externalStudentsAllowed: externalStudentsAllowed ? externalStudentsAllowed : false,
        universityLoginAvailable: universityLoginAvailable ? universityLoginAvailable : false,
        autoAccept: autoAccept ? autoAccept : false,
        useCustomScores: customScores ? customScores : false,
      }
      setApplicationMetaData(parsedMetaData)
    }
  }, [coursePhase, setApplicationMetaData])
}
