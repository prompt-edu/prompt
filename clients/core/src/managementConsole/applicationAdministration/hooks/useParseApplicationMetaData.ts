import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import { useEffect } from 'react'
import type { ApplicationMetaData } from '../interfaces/applicationMetaData'
import { parseProfilePictureSettings } from '../utils/profilePictureSettings'

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
      const applicationMode = coursePhase?.restrictedData?.applicationMode
      const welcomeText = coursePhase?.restrictedData?.welcomeText
      const profilePicture = parseProfilePictureSettings(coursePhase?.restrictedData)

      const parsedMetaData: ApplicationMetaData = {
        applicationStartDate: applicationStartDate ? new Date(applicationStartDate) : undefined,
        applicationEndDate: applicationEndDate ? new Date(applicationEndDate) : undefined,
        externalStudentsAllowed: externalStudentsAllowed ? externalStudentsAllowed : false,
        universityLoginAvailable: universityLoginAvailable ? universityLoginAvailable : false,
        autoAccept: autoAccept ? autoAccept : false,
        useCustomScores: customScores ? customScores : false,
        applicationMode: applicationMode === 'import' ? 'import' : 'apply',
        welcomeText: typeof welcomeText === 'string' ? welcomeText : undefined,
        profilePictureRequirement: profilePicture.requirement,
        profilePictureExplanation: profilePicture.explanation,
        hideProfilePicturesUntilAccepted: profilePicture.hideUntilAccepted,
      }
      setApplicationMetaData(parsedMetaData)
    }
  }, [coursePhase, setApplicationMetaData])
}
