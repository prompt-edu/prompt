import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'
import { describeMailReport, isTemplateComplete, parseResultsReleasedMail } from './utils'

const coursePhaseWithMailingSettings = (mailingSettings: Record<string, unknown>) =>
  ({ restrictedData: { mailingSettings } }) as unknown as CoursePhaseWithMetaData

describe('parseResultsReleasedMail', () => {
  it('defaults to an empty, disabled mail', () => {
    expect(parseResultsReleasedMail(undefined)).toEqual({
      subject: '',
      content: '',
      sendOnRelease: false,
    })
  })

  it('reads the mail next to other mailing settings', () => {
    const coursePhase = coursePhaseWithMailingSettings({
      assessmentReminder: { subject: 'Reminder' },
      resultsReleasedMail: { subject: 'Results', content: '<p>Hi</p>', sendOnRelease: true },
    })

    expect(parseResultsReleasedMail(coursePhase)).toEqual({
      subject: 'Results',
      content: '<p>Hi</p>',
      sendOnRelease: true,
    })
  })
})

describe('isTemplateComplete', () => {
  it('requires a subject and content', () => {
    expect(isTemplateComplete({ subject: 'Results', content: ' ', sendOnRelease: true })).toBe(
      false,
    )
    expect(isTemplateComplete({ subject: 'Results', content: 'Hi', sendOnRelease: false })).toBe(
      true,
    )
  })
})

describe('describeMailReport', () => {
  it('counts sent and undelivered mails', () => {
    expect(
      describeMailReport({
        successfulEmails: ['a@example.com'],
        failedEmails: [],
        requestedRecipients: 1,
      }),
    ).toBe('1 notification mail was sent.')
    expect(
      describeMailReport({
        successfulEmails: ['a@example.com', 'b@example.com'],
        failedEmails: ['c@example.com'],
        requestedRecipients: 3,
      }),
    ).toBe(
      '2 notification mails were sent. 1 could not be delivered. To retry, unrelease and release the results again.',
    )
  })

  it('says when nobody was left to notify', () => {
    expect(
      describeMailReport({ successfulEmails: [], failedEmails: [], requestedRecipients: 0 }),
    ).toBe('All students were already notified.')
  })
})
