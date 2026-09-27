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

/**
 * Reads the profile picture settings from the application phase's restricted data. Pictures are
 * hidden unless a lecturer turned it off, which also covers phases created before the setting.
 */
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

/** Why an applicant's picture is hidden in the review, or undefined when it may be shown. */
export const getApplicantPictureHiddenReason = (
  hideUntilAccepted: boolean,
  passStatus: PassStatus | undefined,
): string | undefined =>
  hideUntilAccepted && passStatus !== PassStatus.PASSED ? HIDDEN_PICTURE_REASON : undefined
