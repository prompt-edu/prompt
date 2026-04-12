import type { CoursePhaseWithMetaData, Team } from '@tumaet/prompt-shared-state'
import { format } from 'date-fns'

import type { AssessmentParticipationWithStudent } from '../../../../interfaces/assessmentParticipationWithStudent'
import type { CoursePhaseConfig } from '../../../../interfaces/coursePhaseConfig'
import type { EvaluationCompletion } from '../../../../interfaces/evaluationCompletion'
import { AssessmentType } from '../../../../interfaces/assessmentType'
import type {
  AssessmentReminderMetaData,
  EvaluationReminderType,
} from '../../../../interfaces/evaluationReminder'
import type { ReminderTypeConfig } from './interfaces/ReminderTypeConfig'

export const EMPTY_REMINDER_META: AssessmentReminderMetaData = {
  subject: '',
  content: '',
  lastSentAtByType: {},
}

export const ASSESSMENT_REMINDER_PLACEHOLDERS = [
  {
    placeholder: '{{evaluationType}}',
    description: 'Current evaluation type (Self Evaluation, Peer Evaluation, Tutor Evaluation)',
  },
  {
    placeholder: '{{evaluationDeadline}}',
    description: 'Deadline of the selected evaluation type',
  },
  {
    placeholder: '{{coursePhaseName}}',
    description: 'Name of the current course phase',
  },
]

export const parseReminderMetaData = (
  coursePhase: CoursePhaseWithMetaData | undefined,
): AssessmentReminderMetaData => {
  const mailingSettings = coursePhase?.restrictedData?.mailingSettings as
    | Record<string, unknown>
    | undefined
  const reminderData = mailingSettings?.assessmentReminder as
    | Partial<AssessmentReminderMetaData>
    | undefined

  if (!reminderData) {
    return EMPTY_REMINDER_META
  }

  return {
    subject: reminderData.subject ?? '',
    content: reminderData.content ?? '',
    lastSentAtByType: reminderData.lastSentAtByType ?? {},
  }
}

export const formatDeadline = (deadline?: Date) => {
  if (!deadline) return 'Not configured'
  return format(new Date(deadline), 'dd.MM.yyyy HH:mm')
}

export const formatSentAt = (sentAt?: string) => {
  if (!sentAt) return 'Never sent'
  return format(new Date(sentAt), 'dd.MM.yyyy HH:mm')
}

export const deadlinePassed = (deadline?: Date) => {
  if (!deadline) return false
  return new Date(deadline) < new Date()
}

export const mapReminderTypeToAssessmentType = (
  reminderType: EvaluationReminderType,
): AssessmentType => {
  switch (reminderType) {
    case 'self':
      return AssessmentType.SELF
    case 'peer':
      return AssessmentType.PEER
    case 'tutor':
      return AssessmentType.TUTOR
  }
}

const deduplicateIds = (ids: string[]) => [...new Set(ids)]

const getCompletedTargetsByAuthor = (
  evaluationCompletions: EvaluationCompletion[] | undefined,
  reminderType: EvaluationReminderType,
) => {
  const completedTargetsByAuthor = new Map<string, Set<string>>()

  evaluationCompletions?.forEach((completion) => {
    if (completion.type !== reminderType || !completion.completed) return

    const existingTargets = completedTargetsByAuthor.get(completion.authorCourseParticipationID)
    if (existingTargets) {
      existingTargets.add(completion.courseParticipationID)
      return
    }

    completedTargetsByAuthor.set(
      completion.authorCourseParticipationID,
      new Set([completion.courseParticipationID]),
    )
  })

  return completedTargetsByAuthor
}

const getExpectedTargets = (
  reminderType: EvaluationReminderType,
  courseParticipationID: string,
  teamID: string,
  teams: Team[],
) => {
  if (reminderType === 'self') {
    return [courseParticipationID]
  }

  const team = teams.find((currentTeam) => currentTeam.id === teamID)
  if (!team) return []

  if (reminderType === 'peer') {
    return deduplicateIds(
      (team.members ?? [])
        .map((member) => member.id)
        .filter((id): id is string => !!id && id !== courseParticipationID),
    )
  }

  return deduplicateIds(
    (team.tutors ?? []).map((tutor) => tutor.id).filter((id): id is string => !!id),
  )
}

export const getReminderTypes = (
  coursePhaseConfig: CoursePhaseConfig | undefined,
  participations: AssessmentParticipationWithStudent[],
  teams: Team[],
  evaluationCompletions: EvaluationCompletion[] | undefined,
): ReminderTypeConfig[] => {
  const getRecipientCount = (type: EvaluationReminderType) => {
    const completedTargetsByAuthor = getCompletedTargetsByAuthor(evaluationCompletions, type)

    return participations.filter((participation) => {
      const expectedTargets = getExpectedTargets(
        type,
        participation.courseParticipationID,
        participation.teamID,
        teams,
      )

      if (expectedTargets.length === 0) {
        return false
      }

      const authorCompletedTargets = completedTargetsByAuthor.get(
        participation.courseParticipationID,
      )

      return expectedTargets.some((targetID) => !authorCompletedTargets?.has(targetID))
    }).length
  }

  const activeReminderTypes: ReminderTypeConfig[] = []

  if (coursePhaseConfig?.selfEvaluationEnabled) {
    activeReminderTypes.push({
      type: 'self',
      label: 'Self Evaluation',
      deadline: coursePhaseConfig.selfEvaluationDeadline,
      recipientCount: getRecipientCount('self'),
    })
  }

  if (coursePhaseConfig?.peerEvaluationEnabled) {
    activeReminderTypes.push({
      type: 'peer',
      label: 'Peer Evaluation',
      deadline: coursePhaseConfig.peerEvaluationDeadline,
      recipientCount: getRecipientCount('peer'),
    })
  }

  if (coursePhaseConfig?.tutorEvaluationEnabled) {
    activeReminderTypes.push({
      type: 'tutor',
      label: 'Tutor Evaluation',
      deadline: coursePhaseConfig.tutorEvaluationDeadline,
      recipientCount: getRecipientCount('tutor'),
    })
  }

  return activeReminderTypes
}
