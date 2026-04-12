import { useEffect, useMemo, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

import { AvailableMailPlaceholders } from '@/components/pages/Mailing/components/AvailableMailPlaceholders'
import { useGetMailingIsConfigured } from '@/hooks/useGetMailingIsConfigured'
import { useModifyCoursePhase } from '@/hooks/useModifyCoursePhase'
import { getCoursePhase } from '@/network/queries/getCoursePhase'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Badge,
  Card,
  useToast,
} from '@tumaet/prompt-ui-components'

import type { EvaluationCompletion } from '../../../interfaces/evaluationCompletion'
import type {
  AssessmentReminderMetaData,
  EvaluationReminderReport,
  EvaluationReminderType,
} from '../../../interfaces/evaluationReminder'
import { sendEvaluationReminder } from '../../../network/mutations/sendEvaluationReminder'
import { getAllEvaluationCompletionsInPhase } from '../../../network/queries/getAllEvaluationCompletionsInPhase'
import { useCoursePhaseConfigStore } from '../../../zustand/useCoursePhaseConfigStore'
import { useParticipationStore } from '../../../zustand/useParticipationStore'
import { useTeamStore } from '../../../zustand/useTeamStore'
import { ManualReminderSendingSection } from './AssessmentReminderCard/components/ManualReminderSendingSection'
import { ReminderSendConfirmationDialog } from './AssessmentReminderCard/components/ReminderSendConfirmationDialog'
import { ReminderTemplateEditor } from './AssessmentReminderCard/components/ReminderTemplateEditor'
import {
  ASSESSMENT_REMINDER_PLACEHOLDERS,
  deadlinePassed,
  EMPTY_REMINDER_META,
  formatDeadline,
  getReminderTypes,
  parseReminderMetaData,
} from './AssessmentReminderCard/utils'

interface ErrorResponse {
  error?: string
}

export const AssessmentReminderCard = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { toast } = useToast()
  const queryClient = useQueryClient()

  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const { participations } = useParticipationStore()
  const { teams } = useTeamStore()
  const courseMailingIsConfigured = useGetMailingIsConfigured()

  const [subject, setSubject] = useState('')
  const [content, setContent] = useState('')
  const [initialMetaData, setInitialMetaData] = useState(EMPTY_REMINDER_META)
  const [confirmationType, setConfirmationType] = useState<EvaluationReminderType | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [lastReport, setLastReport] = useState<EvaluationReminderReport | null>(null)
  const contentTextareaRef = useRef<HTMLTextAreaElement>(null)
  const initializedPhaseIdRef = useRef<string | null>(null)

  const {
    data: coursePhase,
    isPending: isCoursePhasePending,
    isError: isCoursePhaseError,
  } = useQuery<CoursePhaseWithMetaData>({
    queryKey: ['course_phase', phaseId],
    queryFn: () => getCoursePhase(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const { data: evaluationCompletions, isPending: isEvaluationCompletionsPending } = useQuery<
    EvaluationCompletion[]
  >({
    queryKey: ['evaluationCompletions', phaseId],
    queryFn: () => getAllEvaluationCompletionsInPhase(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const { mutate: updateCoursePhase, isPending: isSavingTemplate } = useModifyCoursePhase(
    () => {
      toast({ title: 'Assessment reminder template updated' })
      queryClient.invalidateQueries({ queryKey: ['course_phase', phaseId] })
    },
    () => {
      toast({
        title: 'Failed to update reminder template',
        description: 'Please try again later.',
        variant: 'destructive',
      })
    },
  )

  const sendReminderMutation = useMutation({
    mutationFn: (type: EvaluationReminderType) =>
      sendEvaluationReminder(phaseId ?? '', { evaluationType: type }),
    onSuccess: (report) => {
      setLastReport(report)
      toast({
        title: `Reminder sent for ${report.evaluationType} evaluation`,
        description: `Successful: ${report.successfulEmails.length}, Failed: ${report.failedEmails.length}`,
      })
      queryClient.invalidateQueries({ queryKey: ['course_phase', phaseId] })
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      const serverError = error.response?.data?.error ?? 'Failed to send reminder emails.'
      toast({
        title: 'Reminder sending failed',
        description: serverError,
        variant: 'destructive',
      })
    },
    onSettled: () => {
      setDialogOpen(false)
      setConfirmationType(null)
    },
  })

  useEffect(() => {
    if (!coursePhase) return

    const currentPhaseId = phaseId ?? coursePhase.id
    if (initializedPhaseIdRef.current === currentPhaseId) return

    const parsed = parseReminderMetaData(coursePhase)
    setInitialMetaData(parsed)
    setSubject(parsed.subject)
    setContent(parsed.content)
    initializedPhaseIdRef.current = currentPhaseId
  }, [coursePhase, phaseId])

  useEffect(() => {
    if (!contentTextareaRef.current) return

    contentTextareaRef.current.style.height = 'auto'
    contentTextareaRef.current.style.height = `${contentTextareaRef.current.scrollHeight}px`
  }, [content])

  const currentReminderMetaData = useMemo(() => parseReminderMetaData(coursePhase), [coursePhase])
  const reminderTypes = useMemo(
    () => getReminderTypes(coursePhaseConfig, participations, teams, evaluationCompletions),
    [coursePhaseConfig, participations, teams, evaluationCompletions],
  )
  const confirmationReminderType = useMemo(
    () => reminderTypes.find((reminderType) => reminderType.type === confirmationType) ?? null,
    [confirmationType, reminderTypes],
  )

  const isModified = subject !== initialMetaData.subject || content !== initialMetaData.content
  const templateComplete = subject.trim() !== '' && content.trim() !== ''

  const handleSaveTemplate = () => {
    if (!phaseId || !coursePhase) return

    const mailingSettings =
      (coursePhase.restrictedData?.mailingSettings as Record<string, unknown>) ?? {}
    const updatedReminder: AssessmentReminderMetaData = {
      subject,
      content,
      lastSentAtByType: currentReminderMetaData.lastSentAtByType ?? {},
    }

    updateCoursePhase({
      id: coursePhase.id,
      name: coursePhase.name,
      studentReadableData: coursePhase.studentReadableData ?? {},
      restrictedData: {
        ...coursePhase.restrictedData,
        mailingSettings: {
          ...mailingSettings,
          assessmentReminder: updatedReminder,
        },
      },
    })
  }

  const openConfirmationDialog = (type: EvaluationReminderType) => {
    setConfirmationType(type)
    setDialogOpen(true)
  }

  const sendConfirmedReminder = () => {
    if (!confirmationType) return
    sendReminderMutation.mutate(confirmationType)
  }

  const handleDialogOpenChange = (open: boolean) => {
    setDialogOpen(open)
    if (!open) {
      setConfirmationType(null)
    }
  }

  const getDisableReason = (reminderType: (typeof reminderTypes)[number]): string | undefined => {
    if (isEvaluationCompletionsPending) {
      return 'Wait until recipient counts have finished loading.'
    }
    if (isModified) return 'Save template changes before sending reminders.'
    if (!courseMailingIsConfigured) return 'Configure course mailing reply-to settings first.'
    if (!templateComplete) return 'Reminder subject and content are required.'
    if (!deadlinePassed(reminderType.deadline))
      return `${reminderType.label} deadline has not passed yet (${formatDeadline(reminderType.deadline)}).`
    return undefined
  }

  if (reminderTypes.length === 0) {
    return (
      <Card className='border-border shadow-sm'>
        <div className='space-y-4 p-6'>
          <Alert>
            <AlertDescription>
              Enable at least one evaluation type to send reminder mails.
            </AlertDescription>
          </Alert>
        </div>
      </Card>
    )
  }

  return (
    <Card className='border-border shadow-sm'>
      <div className='space-y-6 p-6'>
        <div className='space-y-2'>
          <div className='flex items-center justify-between gap-2'>
            <h2 className='text-xl font-semibold text-foreground'>Evaluation Reminder Mailing</h2>
            {isModified && (
              <Badge variant='outline' className='border-yellow-300 bg-yellow-100 text-yellow-800'>
                Unsaved Changes
              </Badge>
            )}
          </div>
          <p className='max-w-3xl text-sm leading-6 text-muted-foreground'>
            Configure one shared reminder template and manually send reminders after each evaluation
            deadline.
          </p>
        </div>

        <AvailableMailPlaceholders
          customAdditionalPlaceholders={ASSESSMENT_REMINDER_PLACEHOLDERS}
        />

        {!courseMailingIsConfigured && (
          <Alert variant='destructive'>
            <AlertTitle>Course mailing not configured</AlertTitle>
            <AlertDescription>
              Configure the course reply-to mailing settings before sending reminders.
            </AlertDescription>
          </Alert>
        )}

        {isCoursePhaseError && (
          <Alert variant='destructive'>
            <AlertTitle>Failed to load phase metadata</AlertTitle>
            <AlertDescription>Cannot load existing reminder template settings.</AlertDescription>
          </Alert>
        )}

        {lastReport && (
          <Alert>
            <AlertTitle>Reminder send report</AlertTitle>
            <AlertDescription>
              Requested: {lastReport.requestedRecipients}, Successful:{' '}
              {lastReport.successfulEmails.length}, Failed: {lastReport.failedEmails.length}
            </AlertDescription>
          </Alert>
        )}

        <ReminderTemplateEditor
          subject={subject}
          content={content}
          onSubjectChange={setSubject}
          onContentChange={setContent}
          onSave={handleSaveTemplate}
          isPending={isCoursePhasePending}
          isSaving={isSavingTemplate}
          isModified={isModified}
          contentTextareaRef={contentTextareaRef}
        />

        <ManualReminderSendingSection
          reminderTypes={reminderTypes}
          lastSentAtByType={currentReminderMetaData.lastSentAtByType}
          getDisableReason={getDisableReason}
          isEvaluationCompletionsPending={isEvaluationCompletionsPending}
          isSending={sendReminderMutation.isPending}
          onSend={openConfirmationDialog}
        />
      </div>

      <ReminderSendConfirmationDialog
        open={dialogOpen}
        onOpenChange={handleDialogOpenChange}
        confirmationReminderType={confirmationReminderType}
        previousSentAt={
          confirmationType ? currentReminderMetaData.lastSentAtByType[confirmationType] : undefined
        }
        isSending={sendReminderMutation.isPending}
        onConfirm={sendConfirmedReminder}
      />
    </Card>
  )
}
