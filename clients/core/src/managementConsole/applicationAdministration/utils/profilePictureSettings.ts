import type { ApplicationProfilePictureConfig } from '@core/interfaces/application/openApplicationDetails'
import { PassStatus } from '@tumaet/prompt-shared-state'
import type { ProfilePictureRequirement } from '@tumaet/prompt-ui-components'

export interface ApplicationProfilePictureSettings {
  requirement: ProfilePictureRequirement
  explanation: string
  hideUntilAccepted: boolean
}

const REQUIREMENTS: readonly ProfilePictureRequirement[] = ['off', 'optional', 'required']

export const HIDDEN_PICTURE_REASON =
  'Hidden until the applicant is accepted. You can change this in the application settings.'

export const parseProfilePictureSettings = (
  restrictedData: Record<string, unknown> | undefined,
): ApplicationProfilePictureSettings => {
  const requirement = restrictedData?.profilePictureRequirement
  const explanation = restrictedData?.profilePictureExplanation
  return {
    requirement: REQUIREMENTS.includes(requirement as ProfilePictureRequirement)
      ? (requirement as ProfilePictureRequirement)
      : 'off',
    explanation: typeof explanation === 'string' ? explanation : '',
    hideUntilAccepted: restrictedData?.hideProfilePicturesUntilAccepted !== false,
  }
}

export const getPreviewProfilePicture = (
  restrictedData: Record<string, unknown> | undefined,
): ApplicationProfilePictureConfig => {
  const { requirement, explanation, hideUntilAccepted } =
    parseProfilePictureSettings(restrictedData)
  return { requirement, explanation, hiddenUntilAccepted: hideUntilAccepted }
}

export const rebaseProfilePictureSettings = (
  draft: ApplicationProfilePictureSettings,
  previouslyStored: ApplicationProfilePictureSettings,
  stored: ApplicationProfilePictureSettings,
): ApplicationProfilePictureSettings => ({
  requirement:
    draft.requirement === previouslyStored.requirement ? stored.requirement : draft.requirement,
  explanation:
    draft.explanation === previouslyStored.explanation ? stored.explanation : draft.explanation,
  hideUntilAccepted:
    draft.hideUntilAccepted === previouslyStored.hideUntilAccepted
      ? stored.hideUntilAccepted
      : draft.hideUntilAccepted,
})

export const getApplicantPictureHiddenReason = (
  hideUntilAccepted: boolean,
  passStatus: PassStatus | undefined,
): string | undefined =>
  hideUntilAccepted && passStatus !== PassStatus.PASSED ? HIDDEN_PICTURE_REASON : undefined
