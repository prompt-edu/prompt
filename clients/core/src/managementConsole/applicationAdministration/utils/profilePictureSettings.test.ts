import { PassStatus } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'
import {
  getApplicantPictureHiddenReason,
  HIDDEN_PICTURE_REASON,
  parseProfilePictureSettings,
} from './profilePictureSettings'

describe('parseProfilePictureSettings', () => {
  it('asks for no picture and hides pictures when nothing is configured', () => {
    expect(parseProfilePictureSettings(undefined)).toEqual({
      requirement: 'off',
      explanation: '',
      hideUntilAccepted: true,
    })
  })

  it('reads the configured values', () => {
    expect(
      parseProfilePictureSettings({
        profilePictureRequirement: 'required',
        profilePictureExplanation: 'For the team board',
        hideProfilePicturesUntilAccepted: false,
      }),
    ).toEqual({
      requirement: 'required',
      explanation: 'For the team board',
      hideUntilAccepted: false,
    })
  })

  it('treats unknown values as not configured', () => {
    expect(
      parseProfilePictureSettings({
        profilePictureRequirement: 'mandatory',
        profilePictureExplanation: 42,
        hideProfilePicturesUntilAccepted: 'no',
      }),
    ).toEqual({ requirement: 'off', explanation: '', hideUntilAccepted: true })
  })
})

describe('getApplicantPictureHiddenReason', () => {
  it('hides the pictures of applicants who are not accepted', () => {
    expect(getApplicantPictureHiddenReason(true, PassStatus.NOT_ASSESSED)).toBe(
      HIDDEN_PICTURE_REASON,
    )
    expect(getApplicantPictureHiddenReason(true, PassStatus.FAILED)).toBe(HIDDEN_PICTURE_REASON)
    expect(getApplicantPictureHiddenReason(true, undefined)).toBe(HIDDEN_PICTURE_REASON)
  })

  it('shows the picture once the applicant is accepted', () => {
    expect(getApplicantPictureHiddenReason(true, PassStatus.PASSED)).toBeUndefined()
  })

  it('shows every picture when hiding is turned off', () => {
    expect(getApplicantPictureHiddenReason(false, PassStatus.NOT_ASSESSED)).toBeUndefined()
  })
})
