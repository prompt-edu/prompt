import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Alert,
  AlertDescription,
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertTitle,
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  ErrorPage,
  Input,
  Label,
  LoadingPage,
  ManagementPageHeader,
  Separator,
  Textarea,
  useToast,
} from '@tumaet/prompt-ui-components'
import {
  AlertTriangle,
  Check,
  CloudOff,
  Loader2,
  LockKeyhole,
  MessageSquareText,
  RefreshCw,
  RotateCcw,
  Send,
  Trash2,
  UnlockKeyhole,
  Users,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { DestructiveResetDialog } from '../components/DestructiveResetDialog'
import { useCoursePhaseId, usePresentationAccess } from '../hooks'
import type { ActiveEditor, FeedbackAnswer, FeedbackDocument, FeedbackForm } from '../interfaces'
import { presentationApi, streamFeedbackEvents } from '../network'
import { formatDateTime, getApiError, getErrorMessage } from '../utils'

type SaveStatus = 'idle' | 'saving' | 'saved' | 'conflict' | 'error'
type StreamStatus = 'connecting' | 'live' | 'offline'
type FeedbackActionType = 'submit' | 'reopen' | 'delete-draft' | 'release' | 'unrelease' | 'reset'

interface FeedbackAction {
  type: FeedbackActionType
  releaseName?: string
}

interface SubmittedFormProps {
  form: FeedbackForm
  document: FeedbackDocument
}

const SubmittedForm = ({ form, document }: SubmittedFormProps) => {
  const answers = new Map(form.answers.map((answer) => [answer.categoryId, answer]))
  return (
    <Card>
      <CardHeader>
        <div className='flex flex-wrap items-start justify-between gap-3'>
          <div>
            <CardTitle className='text-base'>{form.evaluatorName ?? 'Instructor'}</CardTitle>
            <CardDescription>
              {form.submittedAt ? `Submitted ${formatDateTime(form.submittedAt)}` : 'Evaluation'}
            </CardDescription>
          </div>
          <Badge variant={form.status === 'submitted' ? 'default' : 'secondary'}>
            {form.status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className='space-y-5'>
        {document.categories.map((category) => (
          <div key={category.id} className='space-y-1'>
            <p className='text-sm font-medium'>{category.name}</p>
            {category.description ? (
              <p className='text-xs text-muted-foreground'>{category.description}</p>
            ) : null}
            <p className='whitespace-pre-wrap rounded-md bg-muted/50 p-3 text-sm'>
              {answers.get(category.id)?.value || 'No comment provided.'}
            </p>
          </div>
        ))}
        {form.contributors.length > 0 ? (
          <>
            <Separator />
            <p className='text-xs text-muted-foreground'>
              Contributors: {form.contributors.map((contributor) => contributor.name).join(', ')}
            </p>
          </>
        ) : null}
      </CardContent>
    </Card>
  )
}

const FeedbackWorkspacePage = () => {
  const coursePhaseId = useCoursePhaseId()
  const { presentationId = '' } = useParams<{ presentationId: string }>()
  const { isLecturer, isStaff } = usePresentationAccess()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [values, setValues] = useState<Record<string, string>>({})
  const [revisions, setRevisions] = useState<Record<string, number>>({})
  const [dirty, setDirty] = useState<Record<string, boolean>>({})
  const [saveStatuses, setSaveStatuses] = useState<Record<string, SaveStatus>>({})
  const [conflicts, setConflicts] = useState<Record<string, FeedbackAnswer | undefined>>({})
  const [activeEditors, setActiveEditors] = useState<ActiveEditor[]>([])
  const [streamStatus, setStreamStatus] = useState<StreamStatus>('connecting')
  const [deleteDraftOpen, setDeleteDraftOpen] = useState(false)
  const [releaseOpen, setReleaseOpen] = useState(false)
  const [releaseName, setReleaseName] = useState('')
  const [resetOpen, setResetOpen] = useState(false)
  const dirtyRef = useRef(dirty)

  useEffect(() => {
    dirtyRef.current = dirty
  }, [dirty])

  const feedbackQuery = useQuery({
    queryKey: ['presentation-feedback', coursePhaseId, presentationId],
    queryFn: () => presentationApi.getFeedback(coursePhaseId, presentationId),
    enabled: Boolean(coursePhaseId && presentationId),
  })

  const feedback = feedbackQuery.data
  const editableForm = feedback?.ownForm

  useEffect(() => {
    if (!feedback) return
    const nextValues: Record<string, string> = {}
    const nextRevisions: Record<string, number> = {}
    for (const answer of editableForm?.answers ?? []) {
      if (!dirtyRef.current[answer.categoryId]) nextValues[answer.categoryId] = answer.value
      nextRevisions[answer.categoryId] = answer.revision
    }
    setValues((current) => ({ ...current, ...nextValues }))
    setRevisions((current) => ({ ...current, ...nextRevisions }))
    setActiveEditors(feedback.activeEditors ?? [])
  }, [editableForm?.answers, feedback])

  const applyLiveAnswer = useCallback((answer: FeedbackAnswer) => {
    setRevisions((current) => ({ ...current, [answer.categoryId]: answer.revision }))
    if (!dirtyRef.current[answer.categoryId]) {
      setValues((current) => ({ ...current, [answer.categoryId]: answer.value }))
      setSaveStatuses((current) => ({ ...current, [answer.categoryId]: 'saved' }))
    }
  }, [])

  useEffect(() => {
    if (feedback?.mode !== 'shared' || feedback.presentation.feedbackReleasedAt || !isStaff) {
      setStreamStatus('offline')
      return
    }

    const controller = new AbortController()
    let retryTimer: ReturnType<typeof setTimeout> | undefined

    const connect = async () => {
      while (!controller.signal.aborted) {
        setStreamStatus('connecting')
        try {
          await streamFeedbackEvents(
            coursePhaseId,
            presentationId,
            (event) => {
              setStreamStatus('live')
              if (event.answer) applyLiveAnswer(event.answer)
              if (event.activeEditors) setActiveEditors(event.activeEditors)
              if (
                event.type === 'snapshot' ||
                event.type === 'form.status.changed' ||
                event.type === 'released'
              ) {
                void queryClient.invalidateQueries({
                  queryKey: ['presentation-feedback', coursePhaseId, presentationId],
                })
              }
            },
            controller.signal,
          )
        } catch {
          if (!controller.signal.aborted) setStreamStatus('offline')
        }
        if (!controller.signal.aborted) {
          await new Promise<void>((resolve) => {
            retryTimer = setTimeout(resolve, 1500)
          })
        }
      }
    }

    void connect()
    return () => {
      controller.abort()
      if (retryTimer) clearTimeout(retryTimer)
    }
  }, [
    applyLiveAnswer,
    coursePhaseId,
    feedback?.mode,
    feedback?.presentation.feedbackReleasedAt,
    isStaff,
    presentationId,
    queryClient,
  ])

  const invalidateFeedback = () => {
    void queryClient.invalidateQueries({
      queryKey: ['presentation-feedback', coursePhaseId, presentationId],
    })
    void queryClient.invalidateQueries({ queryKey: ['presentations', coursePhaseId] })
  }

  const actionMutation = useMutation({
    mutationFn: async (action: FeedbackAction) => {
      if (action.type === 'submit') {
        return presentationApi.submitFeedback(coursePhaseId, presentationId)
      }
      if (action.type === 'reopen') {
        return presentationApi.reopenFeedback(coursePhaseId, presentationId)
      }
      if (action.type === 'delete-draft') {
        return presentationApi.deleteDraft(coursePhaseId, presentationId)
      }
      if (action.type === 'release') {
        return presentationApi.releaseFeedback(
          coursePhaseId,
          presentationId,
          action.releaseName ?? '',
        )
      }
      if (action.type === 'unrelease') {
        return presentationApi.unreleaseFeedback(coursePhaseId, presentationId)
      }
      return presentationApi.resetFeedback(coursePhaseId, presentationId)
    },
    onSuccess: (_data, action) => {
      invalidateFeedback()
      setDeleteDraftOpen(false)
      setReleaseOpen(false)
      setReleaseName('')
      setResetOpen(false)
      if (action.type === 'delete-draft' || action.type === 'reset') {
        setValues({})
        setRevisions({})
        setDirty({})
      }
      const messages: Record<FeedbackActionType, string> = {
        submit: 'Feedback submitted',
        reopen: 'Feedback reopened',
        'delete-draft': 'Draft deleted',
        release: 'Feedback released to presenters',
        unrelease: 'Feedback hidden from presenters',
        reset: 'All feedback reset',
      }
      toast({ title: messages[action.type] })
    },
    onError: (error) => {
      toast({
        title: 'Feedback action failed',
        description: getErrorMessage(error, 'Please try again.'),
        variant: 'destructive',
      })
    },
  })

  const saveAnswer = async (categoryId: string, expectedRevision?: number) => {
    if (!feedback?.canEdit) return
    const value = values[categoryId] ?? ''
    setSaveStatuses((current) => ({ ...current, [categoryId]: 'saving' }))
    try {
      const answer = await presentationApi.updateAnswer(
        coursePhaseId,
        presentationId,
        categoryId,
        value,
        expectedRevision ?? revisions[categoryId] ?? 0,
      )
      setRevisions((current) => ({ ...current, [categoryId]: answer.revision }))
      setDirty((current) => ({ ...current, [categoryId]: false }))
      setConflicts((current) => ({ ...current, [categoryId]: undefined }))
      setSaveStatuses((current) => ({ ...current, [categoryId]: 'saved' }))
      void queryClient.invalidateQueries({
        queryKey: ['presentation-feedback', coursePhaseId, presentationId],
      })
    } catch (error) {
      const apiError = getApiError(error)
      const conflictAnswer =
        apiError.answer ??
        (typeof apiError.details === 'object' && apiError.details
          ? (apiError.details as FeedbackAnswer)
          : undefined)
      if (apiError.code === 'feedback_conflict' && conflictAnswer) {
        setConflicts((current) => ({ ...current, [categoryId]: conflictAnswer }))
        setSaveStatuses((current) => ({ ...current, [categoryId]: 'conflict' }))
        return
      }
      setSaveStatuses((current) => ({ ...current, [categoryId]: 'error' }))
      toast({
        title: 'Feedback could not be saved',
        description: getErrorMessage(error, 'Please try again.'),
        variant: 'destructive',
      })
    }
  }

  const formsToDisplay = useMemo(() => {
    if (!feedback) return []
    return feedback.forms.filter((form) => form.id !== editableForm?.id)
  }, [editableForm?.id, feedback])

  if (feedbackQuery.isLoading) return <LoadingPage />
  if (feedbackQuery.isError || !feedback) {
    return (
      <ErrorPage
        message='Feedback could not be loaded.'
        onRetry={() => void feedbackQuery.refetch()}
      />
    )
  }

  const released = feedback.released ?? Boolean(feedback.presentation.feedbackReleasedAt)
  const formSubmitted = editableForm?.status === 'submitted'
  const canWrite = feedback.canEdit && !formSubmitted && !released
  const hasUnsavedAnswers =
    Object.values(dirty).some(Boolean) ||
    Object.values(saveStatuses).some((status) => status === 'saving' || status === 'conflict')
  const submittedCount = feedback.presentation.submittedFeedbackCount ?? 0
  const draftCount = Math.max(0, (feedback.presentation.feedbackCount ?? 0) - submittedCount)

  return (
    <div className='space-y-6'>
      <div className='flex flex-wrap items-start justify-between gap-4'>
        <div>
          <ManagementPageHeader>{feedback.presentation.targetName}</ManagementPageHeader>
          <p className='text-muted-foreground'>
            Structured instructor feedback ·{' '}
            {feedback.mode === 'shared' ? 'shared live' : 'independent'}
          </p>
        </div>
        <div className='flex flex-wrap items-center gap-2'>
          <Badge variant={released ? 'default' : 'outline'}>
            {released ? 'Released' : 'Not released'}
          </Badge>
          {released && feedback.presentation.feedbackReleaseName ? (
            <Badge variant='secondary'>{feedback.presentation.feedbackReleaseName}</Badge>
          ) : null}
          {feedback.mode === 'shared' && isStaff ? (
            <Badge variant={streamStatus === 'live' ? 'secondary' : 'outline'}>
              {streamStatus === 'live' ? (
                <Users className='mr-1 h-3 w-3' />
              ) : streamStatus === 'connecting' ? (
                <Loader2 className='mr-1 h-3 w-3 animate-spin' />
              ) : (
                <CloudOff className='mr-1 h-3 w-3' />
              )}
              {streamStatus === 'live'
                ? 'Live sync'
                : streamStatus === 'connecting'
                  ? 'Connecting'
                  : 'Reconnecting'}
            </Badge>
          ) : null}
        </div>
      </div>

      {feedback.mode === 'independent' && isStaff ? (
        <Alert>
          <LockKeyhole className='h-4 w-4' />
          <AlertTitle>Independent evaluator drafts</AlertTitle>
          <AlertDescription>
            Your draft is private until submitted. Other instructors’ drafts are not visible here.
          </AlertDescription>
        </Alert>
      ) : null}

      {feedback.mode === 'shared' && isStaff ? (
        <Alert>
          <Users className='h-4 w-4' />
          <AlertTitle>
            {activeEditors.length > 0
              ? `${activeEditors.length} active editor${activeEditors.length === 1 ? '' : 's'}`
              : 'Shared evaluation'}
          </AlertTitle>
          <AlertDescription>
            {activeEditors.length > 0
              ? activeEditors.map((editor) => editor.name).join(', ')
              : 'Edits synchronize live. Each category uses a revision check so simultaneous changes are never silently overwritten.'}
          </AlertDescription>
        </Alert>
      ) : null}

      {isStaff && (editableForm || feedback.canEdit) ? (
        <Card>
          <CardHeader>
            <div className='flex flex-wrap items-start justify-between gap-3'>
              <div>
                <CardTitle className='flex items-center gap-2'>
                  <MessageSquareText className='h-5 w-5' />
                  {feedback.mode === 'independent' ? 'My evaluation' : 'Shared evaluation'}
                </CardTitle>
                <CardDescription>
                  Responses save when you leave a field. Submitted feedback must be reopened before
                  editing.
                </CardDescription>
              </div>
              <Badge variant={formSubmitted ? 'default' : 'secondary'}>
                {editableForm?.status ?? 'draft'}
              </Badge>
            </div>
          </CardHeader>
          <CardContent className='space-y-6'>
            {feedback.categories.length === 0 ? (
              <Alert variant='destructive'>
                <AlertDescription>
                  No feedback categories are configured. A lecturer must add categories in settings.
                </AlertDescription>
              </Alert>
            ) : null}
            {feedback.categories.map((category) => {
              const status = saveStatuses[category.id] ?? 'idle'
              const conflict = conflicts[category.id]
              return (
                <div key={category.id} className='space-y-2'>
                  <div className='flex flex-wrap items-center justify-between gap-2'>
                    <div>
                      <p className='font-medium'>{category.name}</p>
                      {category.description ? (
                        <p className='text-sm text-muted-foreground'>{category.description}</p>
                      ) : null}
                    </div>
                    <span className='flex items-center gap-1 text-xs text-muted-foreground'>
                      {status === 'saving' ? (
                        <>
                          <Loader2 className='h-3 w-3 animate-spin' />
                          Saving
                        </>
                      ) : null}
                      {status === 'saved' ? (
                        <>
                          <Check className='h-3 w-3' />
                          Saved
                        </>
                      ) : null}
                      {status === 'error' ? 'Save failed' : null}
                    </span>
                  </div>
                  <Textarea
                    value={values[category.id] ?? ''}
                    disabled={!canWrite}
                    rows={5}
                    onChange={(event) => {
                      setValues((current) => ({
                        ...current,
                        [category.id]: event.target.value,
                      }))
                      setDirty((current) => ({ ...current, [category.id]: true }))
                      setSaveStatuses((current) => ({ ...current, [category.id]: 'idle' }))
                    }}
                    onBlur={() => {
                      if (dirty[category.id]) void saveAnswer(category.id)
                    }}
                  />
                  {conflict ? (
                    <Alert variant='destructive'>
                      <AlertTriangle className='h-4 w-4' />
                      <AlertTitle>Another instructor saved this category first</AlertTitle>
                      <AlertDescription className='space-y-3'>
                        <p>
                          Latest server value by {conflict.updatedByName ?? 'another instructor'}:
                        </p>
                        <p className='whitespace-pre-wrap rounded bg-background/70 p-2 text-foreground'>
                          {conflict.value || 'No comment'}
                        </p>
                        <div className='flex flex-wrap gap-2'>
                          <Button
                            size='sm'
                            variant='outline'
                            onClick={() => {
                              setValues((current) => ({
                                ...current,
                                [category.id]: conflict.value,
                              }))
                              setRevisions((current) => ({
                                ...current,
                                [category.id]: conflict.revision,
                              }))
                              setDirty((current) => ({ ...current, [category.id]: false }))
                              setConflicts((current) => ({
                                ...current,
                                [category.id]: undefined,
                              }))
                              setSaveStatuses((current) => ({
                                ...current,
                                [category.id]: 'saved',
                              }))
                            }}
                          >
                            Use server value
                          </Button>
                          <Button
                            size='sm'
                            onClick={() => void saveAnswer(category.id, conflict.revision)}
                          >
                            Save my value over latest
                          </Button>
                        </div>
                      </AlertDescription>
                    </Alert>
                  ) : null}
                </div>
              )
            })}
            <Separator />
            <div className='flex flex-wrap gap-2'>
              {!formSubmitted ? (
                <Button
                  disabled={
                    actionMutation.isPending ||
                    feedback.categories.length === 0 ||
                    hasUnsavedAnswers
                  }
                  onClick={() => actionMutation.mutate({ type: 'submit' })}
                >
                  <Send className='mr-2 h-4 w-4' />
                  Submit evaluation
                </Button>
              ) : (
                <Button
                  variant='outline'
                  disabled={actionMutation.isPending || released}
                  onClick={() => actionMutation.mutate({ type: 'reopen' })}
                >
                  <RotateCcw className='mr-2 h-4 w-4' />
                  Reopen evaluation
                </Button>
              )}
              {!formSubmitted ? (
                <Button
                  variant='ghost'
                  className='text-destructive hover:bg-destructive/10 hover:text-destructive'
                  disabled={actionMutation.isPending}
                  onClick={() => setDeleteDraftOpen(true)}
                >
                  <Trash2 className='mr-2 h-4 w-4' />
                  Delete my draft
                </Button>
              ) : null}
            </div>
          </CardContent>
        </Card>
      ) : null}

      {isStaff && (feedback.canRelease || (isLecturer && released)) ? (
        <Card>
          <CardHeader>
            <CardTitle>Release feedback</CardTitle>
            <CardDescription>
              Release is manual and named. Presenters only see submitted evaluations after release.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            {draftCount > 0 ? (
              <Alert variant='destructive'>
                <AlertTriangle className='h-4 w-4' />
                <AlertTitle>{draftCount} draft evaluation(s) remain</AlertTitle>
                <AlertDescription>
                  Feedback cannot be released while an evaluation is still a draft.
                </AlertDescription>
              </Alert>
            ) : null}
            {draftCount === 0 && submittedCount === 0 ? (
              <Alert>
                <AlertDescription>
                  At least one instructor evaluation must be submitted before feedback can be
                  released.
                </AlertDescription>
              </Alert>
            ) : null}
            <div className='flex flex-wrap gap-2'>
              {released ? (
                <Button
                  variant='outline'
                  disabled={actionMutation.isPending}
                  onClick={() => actionMutation.mutate({ type: 'unrelease' })}
                >
                  <UnlockKeyhole className='mr-2 h-4 w-4' />
                  Unrelease for revision
                </Button>
              ) : (
                <Button
                  disabled={actionMutation.isPending || draftCount > 0 || submittedCount === 0}
                  onClick={() => setReleaseOpen(true)}
                >
                  <LockKeyhole className='mr-2 h-4 w-4' />
                  Release feedback
                </Button>
              )}
              <Button
                variant='ghost'
                className='text-destructive hover:bg-destructive/10 hover:text-destructive'
                disabled={actionMutation.isPending}
                onClick={() => setResetOpen(true)}
              >
                <RefreshCw className='mr-2 h-4 w-4' />
                Reset all feedback
              </Button>
            </div>
            {feedback.presentation.feedbackReleasedByName ? (
              <p className='text-sm text-muted-foreground'>
                {feedback.presentation.feedbackReleaseName
                  ? `“${feedback.presentation.feedbackReleaseName}” · `
                  : ''}
                Released by {feedback.presentation.feedbackReleasedByName}
                {feedback.presentation.feedbackReleasedAt
                  ? ` on ${formatDateTime(feedback.presentation.feedbackReleasedAt)}`
                  : ''}
              </p>
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      <section className='space-y-4'>
        <div>
          <h2 className='text-lg font-semibold'>
            {isStaff ? 'Submitted evaluations' : 'Instructor feedback'}
          </h2>
          <p className='text-sm text-muted-foreground'>
            {isStaff
              ? 'Submitted independent evaluations remain separate and attributable.'
              : 'These evaluations were released by your course instructors.'}
          </p>
        </div>
        {!isStaff && !released ? (
          <Alert>
            <LockKeyhole className='h-4 w-4' />
            <AlertTitle>Feedback has not been released</AlertTitle>
            <AlertDescription>
              Your instructors are still preparing the presentation feedback.
            </AlertDescription>
          </Alert>
        ) : null}
        {!isStaff && released && feedback.presentation.feedbackReleaseName ? (
          <Alert>
            <MessageSquareText className='h-4 w-4' />
            <AlertTitle>{feedback.presentation.feedbackReleaseName}</AlertTitle>
            <AlertDescription>
              Released
              {feedback.presentation.feedbackReleasedAt
                ? ` on ${formatDateTime(feedback.presentation.feedbackReleasedAt)}`
                : ''}
              {feedback.presentation.feedbackReleasedByName
                ? ` by ${feedback.presentation.feedbackReleasedByName}`
                : ''}
              .
            </AlertDescription>
          </Alert>
        ) : null}
        {(isStaff || released) && formsToDisplay.length === 0 ? (
          <Card>
            <CardContent className='py-8 text-center text-sm text-muted-foreground'>
              No submitted evaluations are available.
            </CardContent>
          </Card>
        ) : null}
        {isStaff || released
          ? formsToDisplay.map((form) => (
              <SubmittedForm key={form.id} form={form} document={feedback} />
            ))
          : null}
      </section>

      <AlertDialog open={deleteDraftOpen} onOpenChange={setDeleteDraftOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete your feedback draft?</AlertDialogTitle>
            <AlertDialogDescription>
              Your independent draft and all of its category answers will be permanently deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
              onClick={() => actionMutation.mutate({ type: 'delete-draft' })}
            >
              Delete draft
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={releaseOpen} onOpenChange={setReleaseOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Name and release this feedback</AlertDialogTitle>
            <AlertDialogDescription>
              The release name is shown to the presenters and distinguishes this published feedback
              from later revised releases.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className='space-y-2'>
            <Label htmlFor='feedback-release-name'>Release name</Label>
            <Input
              id='feedback-release-name'
              value={releaseName}
              onChange={(event) => setReleaseName(event.target.value)}
              placeholder='Final presentation feedback'
              maxLength={120}
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              disabled={!releaseName.trim() || actionMutation.isPending}
              onClick={() =>
                actionMutation.mutate({ type: 'release', releaseName: releaseName.trim() })
              }
            >
              Release feedback
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <DestructiveResetDialog
        open={resetOpen}
        title='Reset all presentation feedback?'
        description='This permanently deletes every instructor draft, submitted evaluation, written category answer, contributor record, and current release for this presentation.'
        actionLabel='Reset all feedback'
        isPending={actionMutation.isPending}
        onOpenChange={setResetOpen}
        onConfirm={() => actionMutation.mutateAsync({ type: 'reset' }).then(() => undefined)}
      />
    </div>
  )
}

export default FeedbackWorkspacePage
