import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { useParams } from 'react-router-dom'
import { format } from 'date-fns'

import { getCoursePhase } from '@/network/queries/getCoursePhase'
import { useModifyCoursePhase } from '@/hooks/useModifyCoursePhase'
import { useGetMailingIsConfigured } from '@/hooks/useGetMailingIsConfigured'
import { AvailableMailPlaceholders } from '@/components/pages/Mailing/components/AvailableMailPlaceholders'
import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
  Textarea,
  useToast,
} from '@tumaet/prompt-ui-components'

import { useCoursePhaseConfigStore } from '../../../zustand/useCoursePhaseConfigStore'
import {
  AssessmentReminderMetaData,
  EvaluationReminderReport,
  EvaluationReminderType,
} from '../../../interfaces/evaluationReminder'
import { sendEvaluationReminder } from '../../../network/mutations/sendEvaluationReminder'

const EMPTY_REMINDER_META: AssessmentReminderMetaData = {
  subject: '',
  content: '',
  lastSentAtByType: {},
}

type ReminderTypeConfig = {
  type: EvaluationReminderType
  label: string
  enabled: boolean
  deadline?: Date
}

interface ErrorResponse {
  error?: string
}

const parseReminderMetaData = (
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

const formatDeadline = (deadline?: Date) => {
  if (!deadline) return 'Not configured'
  return format(new Date(deadline), 'dd.MM.yyyy HH:mm')
}

const formatSentAt = (sentAt?: string) => {
  if (!sentAt) return 'Never sent'
  return format(new Date(sentAt), 'dd.MM.yyyy HH:mm')
}

const deadlinePassed = (deadline?: Date) => {
  if (!deadline) return false
  return new Date(deadline) < new Date()
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
    const parsed = parseReminderMetaData(coursePhase)
    setInitialMetaData(parsed)
    setSubject(parsed.subject)
    setContent(parsed.content)
  }, [coursePhase])

  const currentReminderMetaData = useMemo(() => parseReminderMetaData(coursePhase), [coursePhase])

  const isModified = subject !== initialMetaData.subject || content !== initialMetaData.content
  const templateComplete = subject.trim() !== '' && content.trim() !== ''

  const reminderTypes: ReminderTypeConfig[] = useMemo(
    () => [
      {
        type: 'self',
        label: 'Self Evaluation',
        enabled: coursePhaseConfig?.selfEvaluationEnabled ?? false,
        deadline: coursePhaseConfig?.selfEvaluationDeadline,
      },
      {
        type: 'peer',
        label: 'Peer Evaluation',
        enabled: coursePhaseConfig?.peerEvaluationEnabled ?? false,
        deadline: coursePhaseConfig?.peerEvaluationDeadline,
      },
      {
        type: 'tutor',
        label: 'Tutor Evaluation',
        enabled: coursePhaseConfig?.tutorEvaluationEnabled ?? false,
        deadline: coursePhaseConfig?.tutorEvaluationDeadline,
      },
    ],
    [coursePhaseConfig],
  )

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

  const getDisableReason = (config: ReminderTypeConfig): string | undefined => {
    if (isModified) return 'Save template changes before sending reminders.'
    if (!courseMailingIsConfigured) return 'Configure course mailing reply-to settings first.'
    if (!templateComplete) return 'Reminder subject and content are required.'
    if (!config.enabled) return `${config.label} is disabled in the assessment configuration.`
    if (!deadlinePassed(config.deadline))
      return `${config.label} deadline has not passed yet (${formatDeadline(config.deadline)}).`
    return undefined
  }

  return (
    <Card>
      <CardHeader>
        <div className='flex items-center justify-between gap-2'>
          <CardTitle>Evaluation Reminder Mailing</CardTitle>
          {isModified && (
            <Badge variant='outline' className='bg-yellow-100 text-yellow-800 border-yellow-300'>
              Unsaved Changes
            </Badge>
          )}
        </div>
        <CardDescription>
          Configure one shared reminder template and manually send reminders after each evaluation
          deadline.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-6'>
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

        <div className='space-y-3'>
          <Label htmlFor='assessment-reminder-subject'>Reminder Subject</Label>
          <Input
            id='assessment-reminder-subject'
            placeholder='Assessment reminder for {{evaluationType}}'
            value={subject}
            onChange={(event) => setSubject(event.target.value)}
            disabled={isCoursePhasePending}
          />
        </div>

        <div className='space-y-3'>
          <Label htmlFor='assessment-reminder-content'>Reminder Message</Label>
          <Textarea
            id='assessment-reminder-content'
            placeholder='Hi {{firstName}}, please complete your {{evaluationType}} before {{evaluationDeadline}}.'
            className='min-h-[130px]'
            value={content}
            onChange={(event) => setContent(event.target.value)}
            disabled={isCoursePhasePending}
          />
        </div>

        <AvailableMailPlaceholders
          customAdditionalPlaceholders={[
            {
              placeholder: '{{evaluationType}}',
              description:
                'Current evaluation type (Self Evaluation, Peer Evaluation, Tutor Evaluation)',
            },
            {
              placeholder: '{{evaluationDeadline}}',
              description: 'Deadline of the selected evaluation type',
            },
            {
              placeholder: '{{coursePhaseName}}',
              description: 'Name of the current course phase',
            },
          ]}
        />

        <div className='flex justify-end'>
          <Button
            onClick={handleSaveTemplate}
            disabled={!isModified || isSavingTemplate || isCoursePhasePending}
          >
            {isSavingTemplate ? 'Saving...' : 'Save Template'}
          </Button>
        </div>

        <div className='space-y-3'>
          <h3 className='text-sm font-semibold text-gray-700 dark:text-gray-300'>
            Manual Reminder Sending
          </h3>
          {reminderTypes.map((reminderType) => {
            const lastSent = currentReminderMetaData.lastSentAtByType[reminderType.type]
            const disableReason = getDisableReason(reminderType)
            const disabled = !!disableReason || sendReminderMutation.isPending

            return (
              <div
                key={reminderType.type}
                className='border rounded-md p-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'
              >
                <div className='space-y-1'>
                  <p className='font-medium'>{reminderType.label}</p>
                  <p className='text-sm text-muted-foreground'>
                    Deadline: {formatDeadline(reminderType.deadline)}
                  </p>
                  <p className='text-sm text-muted-foreground'>
                    Last sent: {formatSentAt(lastSent)}
                  </p>
                  {disableReason && <p className='text-sm text-red-600'>{disableReason}</p>}
                </div>
                <Button
                  onClick={() => openConfirmationDialog(reminderType.type)}
                  disabled={disabled}
                  className='sm:w-auto'
                >
                  Send Reminder
                </Button>
              </div>
            )
          })}
        </div>
      </CardContent>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Reminder Send</DialogTitle>
            <DialogDescription>
              {confirmationType &&
                `Send reminder emails for ${confirmationType} evaluations to all currently incomplete students?`}
            </DialogDescription>
            {confirmationType && currentReminderMetaData.lastSentAtByType[confirmationType] && (
              <DialogDescription>
                A reminder for this type was already sent on{' '}
                {formatSentAt(currentReminderMetaData.lastSentAtByType[confirmationType])}. Sending
                again will resend to currently incomplete students.
              </DialogDescription>
            )}
          </DialogHeader>
          <DialogFooter>
            <Button variant='outline' onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={sendConfirmedReminder} disabled={sendReminderMutation.isPending}>
              {sendReminderMutation.isPending ? 'Sending...' : 'Confirm Send'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  )
}
