import { PassStatus, useCourseStore, useGetMailingIsConfigured } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  AvailableMailPlaceholders,
  availablePlaceholders,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  EmailTemplateEditor,
  ErrorPage,
  Input,
  Label,
  MultiSelect,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Eye, EyeOff, Loader2, MailWarning, Save, Send, TestTube } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { MailItemListEditor } from './components/MailItemListEditor'
import { RecipientListPanel } from './components/RecipientListPanel'
import { SendConfirmationDialog } from './components/SendConfirmationDialog'
import { useCanSendMailCampaign } from './hooks/useCanSendMailCampaign'
import {
  useCreateMailCampaign,
  useGetMailCampaign,
  useGetRecipientPreview,
  useSendMailCampaign,
  useTestSendMailCampaign,
  useUpdateMailCampaign,
} from './hooks/useCourseMailingCampaigns'
import type { MailCampaignRequest, MailItem } from './interfaces/mailCampaign'

const statusOptions = [
  { value: 'all', label: 'All participants' },
  { value: PassStatus.PASSED, label: 'Passed' },
  { value: PassStatus.FAILED, label: 'Failed' },
  { value: PassStatus.NOT_ASSESSED, label: 'Not assessed' },
]

export const CourseMailingComposePage = () => {
  const { courseId, campaignId } = useParams<{ courseId: string; campaignId: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const { courses } = useCourseStore()

  const sortedPhases = useMemo(() => {
    const course = courses.find((c) => c.id === courseId)
    return (
      course?.coursePhases
        .filter((phase) => phase.sequenceOrder !== -1)
        .sort((a, b) => a.sequenceOrder - b.sequenceOrder) ?? []
    )
  }, [courses, courseId])

  const mailingConfigured = useGetMailingIsConfigured()
  const {
    data: existing,
    isLoading,
    isError,
    refetch: refetchCampaign,
  } = useGetMailCampaign(courseId ?? '', campaignId)

  const [ready, setReady] = useState(!campaignId)
  const [name, setName] = useState('')
  const [subject, setSubject] = useState('')
  const [body, setBody] = useState('')
  const [targetCoursePhaseID, setTargetCoursePhaseID] = useState('')
  const [targetPassStatuses, setTargetPassStatuses] = useState<string[]>([])
  const [replyToEmail, setReplyToEmail] = useState('')
  const [replyToName, setReplyToName] = useState('')
  const [ccOverride, setCcOverride] = useState<MailItem[]>([])
  const [bccOverride, setBccOverride] = useState<MailItem[]>([])
  const [isDirty, setIsDirty] = useState(false)
  const [showRecipients, setShowRecipients] = useState(false)
  const [sendDialogOpen, setSendDialogOpen] = useState(false)

  const { mutate: createCampaign, isPending: isCreating } = useCreateMailCampaign(courseId ?? '')
  const { mutate: updateCampaign, isPending: isUpdating } = useUpdateMailCampaign(
    courseId ?? '',
    campaignId ?? '',
  )
  const { mutate: sendCampaign, isPending: isSending } = useSendMailCampaign(courseId ?? '')
  const { mutate: testSend, isPending: isTesting } = useTestSendMailCampaign(courseId ?? '')

  const canSendRole = useCanSendMailCampaign(courseId)

  const canPreview = !!campaignId && !isDirty
  const {
    data: recipientPreview,
    isLoading: isPreviewLoading,
    isFetching: isPreviewFetching,
    isError: isPreviewError,
  } = useGetRecipientPreview(
    courseId ?? '',
    campaignId,
    canPreview && (showRecipients || sendDialogOpen),
  )

  useEffect(() => {
    if (existing) {
      setName(existing.name)
      setSubject(existing.subject)
      setBody(existing.body)
      setTargetCoursePhaseID(existing.targetCoursePhaseID ?? '')
      setTargetPassStatuses(existing.targetPassStatuses ?? [])
      setReplyToEmail(existing.replyToOverride?.email ?? '')
      setReplyToName(existing.replyToOverride?.name ?? '')
      setCcOverride(existing.ccOverride ?? [])
      setBccOverride(existing.bccOverride ?? [])
      setIsDirty(false)
      setReady(true)
    }
  }, [existing])

  const markDirty = () => setIsDirty(true)

  const buildRequest = (): MailCampaignRequest => {
    const replyToOverride: MailItem | null = replyToEmail
      ? { name: replyToName, email: replyToEmail }
      : null
    return {
      name,
      subject,
      body,
      targetCoursePhaseID: targetCoursePhaseID || null,
      targetPassStatuses,
      replyToOverride,
      ccOverride: ccOverride.filter((item) => item.email.trim() !== ''),
      bccOverride: bccOverride.filter((item) => item.email.trim() !== ''),
    }
  }

  const handleEditorChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name: fieldName, value } = e.target
    if (fieldName === 'subject') setSubject(value)
    else if (fieldName === 'body') setBody(value)
    markDirty()
  }

  const handleSave = () => {
    if (!name.trim()) {
      toast({ title: 'A campaign name is required', variant: 'destructive' })
      return
    }
    if (campaignId) {
      updateCampaign(buildRequest(), {
        onSuccess: () => {
          setIsDirty(false)
          toast({ title: 'Campaign saved' })
        },
        onError: () => toast({ title: 'Failed to save campaign', variant: 'destructive' }),
      })
    } else {
      createCampaign(buildRequest(), {
        onSuccess: (created) => {
          toast({ title: 'Draft created' })
          navigate(`/management/course/${courseId}/mailing/${created.id}`)
        },
        onError: () => toast({ title: 'Failed to create campaign', variant: 'destructive' }),
      })
    }
  }

  const handleTestSend = () => {
    if (!campaignId) return
    testSend(campaignId, {
      onSuccess: () => toast({ title: 'Test mail sent to your address' }),
      onError: () => toast({ title: 'Failed to send test mail', variant: 'destructive' }),
    })
  }

  const handleSend = () => {
    if (!campaignId) return
    sendCampaign(campaignId, {
      onSuccess: (data) => {
        setSendDialogOpen(false)
        toast({ title: `Sending to ${data.recipientCount} recipients` })
        navigate(`/management/course/${courseId}/mailing`)
      },
      onError: () => {
        setSendDialogOpen(false)
        toast({ title: 'Failed to send campaign', variant: 'destructive' })
      },
    })
  }

  const isSaving = isCreating || isUpdating
  // A per-campaign reply-to override is sufficient even when the course-level
  // mailing config is unset.
  const hasReplyToOverride = replyToEmail.trim() !== ''
  const canSend = mailingConfigured || hasReplyToOverride
  // Course editors may edit drafts but not send; only lecturers/admins can send.
  const sendDisabled = !campaignId || isDirty || !canSend || !canSendRole
  // A cached preview stays readable while a targeting change refetches it, so gate
  // on isFetching rather than isLoading to avoid confirming a stale recipient count.
  const isRecipientCountKnown =
    canPreview && !isPreviewFetching && !isPreviewError && !!recipientPreview

  if (campaignId && isError) {
    return <ErrorPage message='Failed to load the campaign' onRetry={refetchCampaign} />
  }

  if (campaignId && isLoading && !ready) {
    return <p className='text-muted-foreground'>Loading campaign...</p>
  }

  return (
    <div className='space-y-6'>
      <div className='flex items-center justify-between'>
        <h1 className='text-3xl font-bold'>{campaignId ? 'Edit Campaign' : 'New Campaign'}</h1>
        <Button
          variant='outline'
          onClick={() => navigate(`/management/course/${courseId}/mailing`)}
        >
          Back to overview
        </Button>
      </div>

      {!canSend && (
        <Alert>
          <MailWarning className='h-4 w-4' />
          <AlertTitle>Course mailing is not configured</AlertTitle>
          <AlertDescription>
            Set the reply-to address in the course settings, or add a per-campaign reply-to override
            below, before sending.
          </AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Campaign details</CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div>
            <Label htmlFor='campaign-name'>Name</Label>
            <Input
              id='campaign-name'
              value={name}
              onChange={(e) => {
                setName(e.target.value)
                markDirty()
              }}
              placeholder='Internal name for this campaign'
              className='mt-1'
            />
          </div>

          <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
            <div>
              <Label>Course phase</Label>
              <Select
                value={targetCoursePhaseID}
                onValueChange={(value) => {
                  setTargetCoursePhaseID(value)
                  markDirty()
                }}
              >
                <SelectTrigger className='mt-1'>
                  <SelectValue placeholder='Select a phase' />
                </SelectTrigger>
                <SelectContent>
                  {sortedPhases.map((phase) => (
                    <SelectItem key={phase.id} value={phase.id}>
                      {phase.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label>Student statuses</Label>
              {ready && (
                <MultiSelect
                  className='mt-1'
                  options={statusOptions}
                  defaultValue={targetPassStatuses}
                  onValueChange={(value) => {
                    setTargetPassStatuses(value)
                    markDirty()
                  }}
                  placeholder='Select statuses'
                />
              )}
            </div>
          </div>

          <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
            <div>
              <Label htmlFor='reply-to-email'>Reply-to override (optional)</Label>
              <Input
                id='reply-to-email'
                value={replyToEmail}
                onChange={(e) => {
                  setReplyToEmail(e.target.value)
                  markDirty()
                }}
                placeholder='Overrides the course reply-to'
                className='mt-1'
              />
            </div>
            <div>
              <Label htmlFor='reply-to-name'>Reply-to name (optional)</Label>
              <Input
                id='reply-to-name'
                value={replyToName}
                onChange={(e) => {
                  setReplyToName(e.target.value)
                  markDirty()
                }}
                placeholder='Reply-to display name'
                className='mt-1'
              />
            </div>
          </div>

          <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
            <MailItemListEditor
              label='CC'
              items={ccOverride}
              onChange={(items) => {
                setCcOverride(items)
                markDirty()
              }}
            />
            <MailItemListEditor
              label='BCC'
              items={bccOverride}
              onChange={(items) => {
                setBccOverride(items)
                markDirty()
              }}
            />
          </div>
        </CardContent>
      </Card>

      <AvailableMailPlaceholders />

      {ready && (
        <EmailTemplateEditor
          subject={subject}
          content={body}
          onInputChange={handleEditorChange}
          label='Campaign'
          subjectHTMLLabel='subject'
          contentHTMLLabel='body'
          placeholders={availablePlaceholders.map((placeholder) => placeholder.placeholder)}
        />
      )}

      <Card>
        <CardHeader>
          <CardTitle className='flex items-center justify-between'>
            Recipients
            <Button
              variant='outline'
              size='sm'
              disabled={!canPreview}
              onClick={() => setShowRecipients((prev) => !prev)}
            >
              {showRecipients ? (
                <>
                  <EyeOff className='mr-2 h-4 w-4' />
                  Hide
                </>
              ) : (
                <>
                  <Eye className='mr-2 h-4 w-4' />
                  Show recipients
                </>
              )}
            </Button>
          </CardTitle>
        </CardHeader>
        {showRecipients && (
          <CardContent>
            {canPreview ? (
              <RecipientListPanel
                recipients={recipientPreview?.recipients ?? []}
                isLoading={isPreviewLoading}
              />
            ) : (
              <p className='text-sm text-muted-foreground'>
                Save your changes to preview the recipient list.
              </p>
            )}
          </CardContent>
        )}
      </Card>

      <div className='flex flex-wrap justify-end gap-2'>
        <Button onClick={handleSave} disabled={isSaving}>
          {isSaving ? (
            <Loader2 className='mr-2 h-4 w-4 animate-spin' />
          ) : (
            <Save className='mr-2 h-4 w-4' />
          )}
          Save
        </Button>
        <Button variant='outline' onClick={handleTestSend} disabled={sendDisabled || isTesting}>
          {isTesting ? (
            <Loader2 className='mr-2 h-4 w-4 animate-spin' />
          ) : (
            <TestTube className='mr-2 h-4 w-4' />
          )}
          Test send
        </Button>
        <Button onClick={() => setSendDialogOpen(true)} disabled={sendDisabled}>
          <Send className='mr-2 h-4 w-4' />
          Send
        </Button>
      </div>

      <SendConfirmationDialog
        isOpen={sendDialogOpen}
        onClose={() => setSendDialogOpen(false)}
        onConfirm={handleSend}
        recipientCount={recipientPreview?.count ?? 0}
        isCountKnown={isRecipientCountKnown}
        isPending={isSending}
      />
    </div>
  )
}
