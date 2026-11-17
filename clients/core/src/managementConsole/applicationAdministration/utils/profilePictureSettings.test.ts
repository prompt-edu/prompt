import { PassStatus } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'
import {
  getApplicantPictureHiddenReason,
  getPreviewProfilePicture,
  HIDDEN_PICTURE_REASON,
  parseProfilePictureSettings,
  rebaseProfilePictureSettings,
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

describe('getPreviewProfilePicture', () => {
  it('shows the stored settings in the form shape', () => {
    expect(
      getPreviewProfilePicture({
        profilePictureRequirement: 'optional',
        profilePictureExplanation: 'For the team board',
        hideProfilePicturesUntilAccepted: false,
      }),
    ).toEqual({
      requirement: 'optional',
      explanation: 'For the team board',
      hiddenUntilAccepted: false,
    })
  })
})

describe('rebaseProfilePictureSettings', () => {
  const stored = { requirement: 'off', explanation: '', hideUntilAccepted: true } as const

  it('follows the stored settings while nothing is edited', () => {
    const refreshed = {
      requirement: 'required',
      explanation: 'For the team board',
      hideUntilAccepted: false,
    } as const
    expect(rebaseProfilePictureSettings(stored, stored, refreshed)).toEqual(refreshed)
  })

  it('keeps edited fields when another setting of the phase is saved', () => {
    const draft = { ...stored, requirement: 'optional', explanation: 'Draft' } as const
    expect(rebaseProfilePictureSettings(draft, stored, stored)).toEqual(draft)
  })

  it('keeps an edit made while the previous one was being saved', () => {
    const submitted = { ...stored, requirement: 'optional' } as const
    const draft = { ...stored, requirement: 'required' } as const
    expect(rebaseProfilePictureSettings(draft, stored, submitted)).toEqual(draft)
  })

  it('updates the fields that were not edited', () => {
    const draft = { ...stored, requirement: 'optional' } as const
    const refreshed = { ...stored, hideUntilAccepted: false } as const
    expect(rebaseProfilePictureSettings(draft, stored, refreshed)).toEqual({
      requirement: 'optional',
      explanation: '',
      hideUntilAccepted: false,
    })
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
