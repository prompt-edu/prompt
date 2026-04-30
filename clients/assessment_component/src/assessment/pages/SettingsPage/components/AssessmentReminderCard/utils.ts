import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import { format } from 'date-fns'

import type { CoursePhaseConfig } from '../../../../interfaces/coursePhaseConfig'
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

export const getReminderTypes = (
  coursePhaseConfig: CoursePhaseConfig | undefined,
): ReminderTypeConfig[] => {
  const activeReminderTypes: ReminderTypeConfig[] = []

  if (coursePhaseConfig?.selfEvaluationEnabled) {
    activeReminderTypes.push({
      type: 'self',
      label: 'Self Evaluation',
      deadline: coursePhaseConfig.selfEvaluationDeadline,
    })
  }

  if (coursePhaseConfig?.peerEvaluationEnabled) {
    activeReminderTypes.push({
      type: 'peer',
      label: 'Peer Evaluation',
      deadline: coursePhaseConfig.peerEvaluationDeadline,
    })
  }

  if (coursePhaseConfig?.tutorEvaluationEnabled) {
    activeReminderTypes.push({
      type: 'tutor',
      label: 'Tutor Evaluation',
      deadline: coursePhaseConfig.tutorEvaluationDeadline,
    })
  }

  return activeReminderTypes
}
