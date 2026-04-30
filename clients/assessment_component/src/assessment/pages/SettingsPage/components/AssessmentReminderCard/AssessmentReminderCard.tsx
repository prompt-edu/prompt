import { useEffect, useMemo, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import {
  getCoursePhase,
  type CoursePhaseWithMetaData,
  useGetMailingIsConfigured,
  useModifyCoursePhase,
} from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

import {
  Alert,
  AlertDescription,
  AlertTitle,
  Badge,
  Card,
  useToast,
} from '@tumaet/prompt-ui-components'

import type {
  AssessmentReminderMetaData,
  EvaluationReminderReport,
  EvaluationReminderType,
} from '../../../../interfaces/evaluationReminder'
import { sendEvaluationReminder } from '../../../../network/mutations/sendEvaluationReminder'
import { useCoursePhaseConfigStore } from '../../../../zustand/useCoursePhaseConfigStore'
import { ManualReminderSendingSection } from './components/ManualReminderSendingSection'
import { ReminderSendConfirmationDialog } from './components/ReminderSendConfirmationDialog'
import { ReminderTemplateEditor } from './components/ReminderTemplateEditor'
import {
  deadlinePassed,
  EMPTY_REMINDER_META,
  formatDeadline,
  getReminderTypes,
  parseReminderMetaData,
} from './utils'

interface ErrorResponse {
  error?: string
}

export const AssessmentReminderCard = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { toast } = useToast()
  const queryClient = useQueryClient()

  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const courseMailingIsConfigured = useGetMailingIsConfigured()

  const [subject, setSubject] = useState('')
  const [content, setContent] = useState('')
  const [initialMetaData, setInitialMetaData] = useState(EMPTY_REMINDER_META)
  const [confirmationType, setConfirmationType] = useState<EvaluationReminderType | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [lastReport, setLastReport] = useState<EvaluationReminderReport | null>(null)
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

  const { mutate: updateCoursePhase, isPending: isSavingTemplate } = useModifyCoursePhase(
    () => {
      setInitialMetaData({
        subject,
        content,
        lastSentAtByType: currentReminderMetaData.lastSentAtByType ?? {},
      })
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
        description: `${report.successfulEmails.length} ${
          report.successfulEmails.length === 1 ? 'email was' : 'emails were'
        } sent.`,
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

  const currentReminderMetaData = useMemo(() => parseReminderMetaData(coursePhase), [coursePhase])
  const reminderTypes = useMemo(() => getReminderTypes(coursePhaseConfig), [coursePhaseConfig])
  const confirmationReminderType = useMemo(
    () => reminderTypes.find((reminderType) => reminderType.type === confirmationType) ?? null,
    [confirmationType, reminderTypes],
  )

  const isModified = subject !== initialMetaData.subject || content !== initialMetaData.content
  const templateComplete = subject.trim() !== '' && content.trim() !== ''
  const templateReady = templateComplete && !isModified

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
    if (isModified) return 'Save template changes before sending reminders.'
    if (!courseMailingIsConfigured) return 'Configure course mailing reply-to settings first.'
    if (!templateComplete) return 'Reminder subject and content are required.'
    if (!deadlinePassed(reminderType.deadline))
      return `${reminderType.label} deadline must pass first (${formatDeadline(reminderType.deadline)}).`
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
            Write one reusable reminder and send it after each evaluation deadline. The system sends
            it only to students who still have incomplete evaluations.
          </p>
        </div>

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

        {!templateComplete && !isCoursePhasePending && (
          <Alert>
            <AlertTitle>Template incomplete</AlertTitle>
            <AlertDescription>
              Add a subject and message before reminders can be sent.
            </AlertDescription>
          </Alert>
        )}

        {isModified && templateComplete && (
          <Alert>
            <AlertTitle>Template not saved yet</AlertTitle>
            <AlertDescription>
              Save your template changes before sending new reminders.
            </AlertDescription>
          </Alert>
        )}

        <div className='grid gap-6 xl:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]'>
          <div className='space-y-4 rounded-lg border border-border bg-muted/20 p-4 sm:p-5'>
            <ReminderTemplateEditor
              subject={subject}
              content={content}
              onSubjectChange={setSubject}
              onContentChange={setContent}
              onSave={handleSaveTemplate}
              isPending={isCoursePhasePending}
              isSaving={isSavingTemplate}
              isModified={isModified}
            />

            <div className='space-y-3 border-t border-border pt-4'>
              <p className='text-xs leading-5 text-muted-foreground'>
                Available placeholders:{' '}
                <span className='font-mono text-foreground'>{'{{firstName}}'}</span>,{' '}
                <span className='font-mono text-foreground'>{'{{evaluationType}}'}</span>,{' '}
                <span className='font-mono text-foreground'>{'{{evaluationDeadline}}'}</span>
              </p>
            </div>
          </div>

          <div className='space-y-4 rounded-lg border border-border bg-muted/20 p-4 sm:p-5'>
            {lastReport && (
              <Alert>
                <AlertTitle>Reminder sent</AlertTitle>
                <AlertDescription>
                  {lastReport.successfulEmails.length} of {lastReport.requestedRecipients}{' '}
                  {lastReport.requestedRecipients === 1 ? 'email was' : 'emails were'} sent.
                  {lastReport.failedEmails.length > 0 &&
                    ` ${lastReport.failedEmails.length} could not be delivered.`}
                </AlertDescription>
              </Alert>
            )}

            {templateReady && courseMailingIsConfigured && (
              <p className='text-sm leading-6 text-muted-foreground'>
                Send a reminder when an evaluation deadline has passed. Recipient selection happens
                automatically at send time.
              </p>
            )}

            <ManualReminderSendingSection
              reminderTypes={reminderTypes}
              lastSentAtByType={currentReminderMetaData.lastSentAtByType}
              getDisableReason={getDisableReason}
              isSending={sendReminderMutation.isPending}
              onSend={openConfirmationDialog}
            />
          </div>
        </div>
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
