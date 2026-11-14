import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import type { ApplicationMetaData } from '../interfaces/applicationMetaData'
import { parseProfilePictureSettings } from './profilePictureSettings'

export const parseApplicationMetaData = (
  restrictedData: CoursePhaseWithMetaData['restrictedData'] | undefined,
): ApplicationMetaData => {
  const externalStudentsAllowed = restrictedData?.externalStudentsAllowed
  const applicationStartDate = restrictedData?.applicationStartDate
  const applicationEndDate = restrictedData?.applicationEndDate
  const universityLoginAvailable = restrictedData?.universityLoginAvailable
  const autoAccept = restrictedData?.autoAccept
  const customScores = restrictedData?.useCustomScores
  const applicationMode = restrictedData?.applicationMode
  const welcomeText = restrictedData?.welcomeText
  const profilePicture = parseProfilePictureSettings(restrictedData)

  return {
    applicationStartDate: applicationStartDate ? new Date(applicationStartDate) : undefined,
    applicationEndDate: applicationEndDate ? new Date(applicationEndDate) : undefined,
    externalStudentsAllowed: externalStudentsAllowed ? externalStudentsAllowed : false,
    universityLoginAvailable: universityLoginAvailable ?? true,
    autoAccept: autoAccept ? autoAccept : false,
    useCustomScores: customScores ? customScores : false,
    applicationMode: applicationMode === 'import' ? 'import' : 'apply',
    welcomeText: typeof welcomeText === 'string' ? welcomeText : undefined,
    profilePictureRequirement: profilePicture.requirement,
    profilePictureExplanation: profilePicture.explanation,
    hideProfilePicturesUntilAccepted: profilePicture.hideUntilAccepted,
  }
}
