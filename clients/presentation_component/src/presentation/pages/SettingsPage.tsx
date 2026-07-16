import { useQuery, useQueryClient } from '@tanstack/react-query'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Textarea,
  useToast,
} from '@tumaet/prompt-ui-components'
import { AlertTriangle, ListChecks, Loader2, Plus, Save, Settings2, Trash2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { DestructiveResetDialog } from '../components/DestructiveResetDialog'
import { useCoursePhaseId } from '../hooks'
import type { CategoryRequest, FeedbackCategory, FeedbackMode, TargetMode } from '../interfaces'
import { presentationApi } from '../network'
import { getApiError, getErrorMessage } from '../utils'

interface ResetAction {
  title: string
  description: string
  actionLabel: string
  run: () => Promise<void>
}

interface CategoryRowProps {
  category: FeedbackCategory
  isPending: boolean
  onSave: (request: CategoryRequest) => Promise<void>
  onDelete: () => Promise<void>
}

const CategoryRow = ({ category, isPending, onSave, onDelete }: CategoryRowProps) => {
  const [name, setName] = useState(category.name)
  const [description, setDescription] = useState(category.description)
  const [position, setPosition] = useState(String(category.position))
  const [deleteOpen, setDeleteOpen] = useState(false)
  const isValid = name.trim().length > 0 && Number(position) >= 0

  return (
    <div className='space-y-4 rounded-lg border p-4'>
      <div className='grid gap-4 md:grid-cols-[1fr_2fr_7rem]'>
        <div className='space-y-2'>
          <Label htmlFor={`category-name-${category.id}`}>Name</Label>
          <Input
            id={`category-name-${category.id}`}
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
        </div>
        <div className='space-y-2'>
          <Label htmlFor={`category-description-${category.id}`}>Guidance</Label>
          <Input
            id={`category-description-${category.id}`}
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            placeholder='What should instructors comment on?'
          />
        </div>
        <div className='space-y-2'>
          <Label htmlFor={`category-position-${category.id}`}>Position</Label>
          <Input
            id={`category-position-${category.id}`}
            type='number'
            min={0}
            value={position}
            onChange={(event) => setPosition(event.target.value)}
          />
        </div>
      </div>
      <div className='flex justify-end gap-2'>
        <Button
          variant='outline'
          disabled={!isValid || isPending}
          onClick={() =>
            void onSave({
              name: name.trim(),
              description: description.trim(),
              position: Number(position),
            })
          }
        >
          <Save className='mr-2 h-4 w-4' />
          Save
        </Button>
        <Button
          variant='ghost'
          className='text-destructive hover:bg-destructive/10 hover:text-destructive'
          disabled={isPending}
          onClick={() => setDeleteOpen(true)}
        >
          <Trash2 className='mr-2 h-4 w-4' />
          Delete
        </Button>
      </div>
      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete “{category.name}”?</AlertDialogTitle>
            <AlertDialogDescription>
              This removes the feedback category. If evaluations already exist, an explicit feedback
              reset will be required.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
              onClick={() => void onDelete()}
            >
              Delete category
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

const SettingsPage = () => {
  const coursePhaseId = useCoursePhaseId()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [targetMode, setTargetMode] = useState<TargetMode>('individual')
  const [feedbackMode, setFeedbackMode] = useState<FeedbackMode>('independent')
  const [newCategoryName, setNewCategoryName] = useState('')
  const [newCategoryDescription, setNewCategoryDescription] = useState('')
  const [isSaving, setIsSaving] = useState(false)
  const [resetAction, setResetAction] = useState<ResetAction>()

  const configQuery = useQuery({
    queryKey: ['presentation-config', coursePhaseId],
    queryFn: () => presentationApi.getConfig(coursePhaseId),
    enabled: Boolean(coursePhaseId),
  })
  const categoriesQuery = useQuery({
    queryKey: ['presentation-categories', coursePhaseId],
    queryFn: () => presentationApi.getCategories(coursePhaseId),
    enabled: Boolean(coursePhaseId),
  })

  useEffect(() => {
    if (!configQuery.data) return
    setTargetMode(configQuery.data.targetMode)
    setFeedbackMode(configQuery.data.feedbackMode)
  }, [configQuery.data])

  const invalidateSettings = () => {
    void queryClient.invalidateQueries({ queryKey: ['presentation-config', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentation-categories', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentations', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentation-slots', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentation-targets', coursePhaseId] })
  }

  const handleFailure = (
    error: unknown,
    fallback: string,
    reset: Omit<ResetAction, 'run'> & { run: () => Promise<void> },
  ) => {
    const apiError = getApiError(error)
    if (
      apiError.code === 'config_locked' ||
      apiError.code === 'feedback_locked' ||
      apiError.code === 'categories_locked'
    ) {
      setResetAction(reset)
      return
    }
    toast({
      title: fallback,
      description: getErrorMessage(error, 'Please try again.'),
      variant: 'destructive',
    })
  }

  const saveConfig = async (resetExistingData = false) => {
    setIsSaving(true)
    try {
      await presentationApi.updateConfig(
        coursePhaseId,
        { targetMode, feedbackMode },
        resetExistingData,
      )
      invalidateSettings()
      toast({
        title: 'Presentation settings saved',
        description: resetExistingData
          ? 'Existing presentation data was reset to apply the new modes.'
          : undefined,
      })
    } catch (error) {
      handleFailure(error, 'Could not save presentation settings', {
        title: 'Reset existing presentation data?',
        description:
          'Changing the target or feedback mode is locked because the phase already contains presentations or evaluations. Continuing permanently deletes the affected schedule assignments, uploaded materials, and feedback so the new mode can be applied.',
        actionLabel: 'Reset data and save',
        run: () => saveConfig(true),
      })
    } finally {
      setIsSaving(false)
    }
  }

  const createCategory = async (resetExistingData = false) => {
    const request: CategoryRequest = {
      name: newCategoryName.trim(),
      description: newCategoryDescription.trim(),
      position:
        Math.max(-1, ...(categoriesQuery.data ?? []).map((category) => category.position)) + 1,
    }
    setIsSaving(true)
    try {
      await presentationApi.createCategory(coursePhaseId, request, resetExistingData)
      invalidateSettings()
      setNewCategoryName('')
      setNewCategoryDescription('')
      toast({ title: 'Feedback category added' })
    } catch (error) {
      handleFailure(error, 'Could not add feedback category', {
        title: 'Reset feedback and add this category?',
        description:
          'The category structure is locked because feedback already exists. Continuing permanently deletes every draft and submitted evaluation in this phase before adding the category.',
        actionLabel: 'Reset feedback and add',
        run: () => createCategory(true),
      })
    } finally {
      setIsSaving(false)
    }
  }

  const updateCategory = async (
    categoryId: string,
    request: CategoryRequest,
    resetExistingData = false,
  ) => {
    setIsSaving(true)
    try {
      await presentationApi.updateCategory(coursePhaseId, categoryId, request, resetExistingData)
      invalidateSettings()
      toast({ title: 'Feedback category updated' })
    } catch (error) {
      handleFailure(error, 'Could not update feedback category', {
        title: 'Reset feedback and update this category?',
        description:
          'The category structure is locked because feedback already exists. Continuing permanently deletes every draft and submitted evaluation in this phase before saving the category.',
        actionLabel: 'Reset feedback and save',
        run: () => updateCategory(categoryId, request, true),
      })
    } finally {
      setIsSaving(false)
    }
  }

  const deleteCategory = async (categoryId: string, resetExistingData = false) => {
    setIsSaving(true)
    try {
      await presentationApi.deleteCategory(coursePhaseId, categoryId, resetExistingData)
      invalidateSettings()
      toast({ title: 'Feedback category deleted' })
    } catch (error) {
      handleFailure(error, 'Could not delete feedback category', {
        title: 'Reset feedback and delete this category?',
        description:
          'The category structure is locked because feedback already exists. Continuing permanently deletes every draft and submitted evaluation in this phase before deleting the category.',
        actionLabel: 'Reset feedback and delete',
        run: () => deleteCategory(categoryId, true),
      })
    } finally {
      setIsSaving(false)
    }
  }

  const confirmReset = async () => {
    if (!resetAction) return
    setIsSaving(true)
    try {
      await resetAction.run()
      setResetAction(undefined)
    } finally {
      setIsSaving(false)
    }
  }

  if (configQuery.isLoading || categoriesQuery.isLoading) return <LoadingPage />
  if (configQuery.isError || categoriesQuery.isError) {
    return (
      <ErrorPage
        message='Presentation settings could not be loaded.'
        onRetry={() => {
          void configQuery.refetch()
          void categoriesQuery.refetch()
        }}
      />
    )
  }

  const categories = categoriesQuery.data ?? []
  const configChanged =
    targetMode !== configQuery.data?.targetMode || feedbackMode !== configQuery.data?.feedbackMode

  return (
    <div className='space-y-6'>
      <div>
        <ManagementPageHeader>Presentation settings</ManagementPageHeader>
        <p className='text-muted-foreground'>
          Choose who presents, how instructors collaborate, and the written feedback structure.
        </p>
      </div>

      <Alert>
        <AlertTriangle className='h-4 w-4' />
        <AlertTitle>Configuration locks protect existing data</AlertTitle>
        <AlertDescription>
          Target mode, feedback mode, and categories can only change safely before presentation or
          feedback data exists. If locked, PROMPT requires a typed destructive reset confirmation.
        </AlertDescription>
      </Alert>

      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            <Settings2 className='h-5 w-5' />
            Presentation modes
          </CardTitle>
          <CardDescription>
            Course editors may evaluate presentations, while only lecturers and administrators can
            change these settings.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-5'>
          <div className='grid gap-4 md:grid-cols-2'>
            <div className='space-y-2'>
              <Label>Presentation target</Label>
              <Select
                value={targetMode}
                onValueChange={(value) => setTargetMode(value as TargetMode)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='individual'>Individual students</SelectItem>
                  <SelectItem value='team'>Teams</SelectItem>
                </SelectContent>
              </Select>
              <p className='text-xs text-muted-foreground'>
                Team access follows each student’s current team allocation.
              </p>
            </div>
            <div className='space-y-2'>
              <Label>Instructor feedback</Label>
              <Select
                value={feedbackMode}
                onValueChange={(value) => setFeedbackMode(value as FeedbackMode)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='independent'>Independent evaluations</SelectItem>
                  <SelectItem value='shared'>Shared live evaluation</SelectItem>
                </SelectContent>
              </Select>
              <p className='text-xs text-muted-foreground'>
                Independent drafts stay private to their author. Shared feedback is edited live with
                revision checks.
              </p>
            </div>
          </div>
          <Button disabled={!configChanged || isSaving} onClick={() => void saveConfig()}>
            {isSaving ? (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <Save className='mr-2 h-4 w-4' />
            )}
            Save modes
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            <ListChecks className='h-5 w-5' />
            Written feedback categories
          </CardTitle>
          <CardDescription>
            Instructors receive one written response field for each category in this order.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-5'>
          <div className='space-y-4 rounded-lg border border-dashed p-4'>
            <div className='grid gap-4 md:grid-cols-[1fr_2fr]'>
              <div className='space-y-2'>
                <Label htmlFor='new-category-name'>Category name</Label>
                <Input
                  id='new-category-name'
                  value={newCategoryName}
                  onChange={(event) => setNewCategoryName(event.target.value)}
                  placeholder='Delivery'
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='new-category-description'>Guidance (optional)</Label>
                <Textarea
                  id='new-category-description'
                  value={newCategoryDescription}
                  onChange={(event) => setNewCategoryDescription(event.target.value)}
                  placeholder='Comment on clarity, pacing, and audience engagement.'
                />
              </div>
            </div>
            <Button
              variant='outline'
              disabled={!newCategoryName.trim() || isSaving}
              onClick={() => void createCategory()}
            >
              <Plus className='mr-2 h-4 w-4' />
              Add category
            </Button>
          </div>

          {categories.length === 0 ? (
            <p className='text-sm text-muted-foreground'>
              Add at least one category before instructors evaluate presentations.
            </p>
          ) : null}
          {categories.map((category) => (
            <CategoryRow
              key={category.id}
              category={category}
              isPending={isSaving}
              onSave={(request) => updateCategory(category.id, request)}
              onDelete={() => deleteCategory(category.id)}
            />
          ))}
        </CardContent>
      </Card>

      <DestructiveResetDialog
        open={Boolean(resetAction)}
        title={resetAction?.title ?? 'Reset existing data?'}
        description={resetAction?.description ?? ''}
        actionLabel={resetAction?.actionLabel ?? 'Reset data'}
        isPending={isSaving}
        onOpenChange={(open) => {
          if (!open && !isSaving) setResetAction(undefined)
        }}
        onConfirm={confirmReset}
      />
    </div>
  )
}

export default SettingsPage
