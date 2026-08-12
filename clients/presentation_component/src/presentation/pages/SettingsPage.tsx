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
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Checkbox,
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
import {
  AlertTriangle,
  ListChecks,
  Loader2,
  Plus,
  Save,
  Settings2,
  Trash2,
  Upload,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { DestructiveResetDialog } from '../components/DestructiveResetDialog'
import { useCoursePhaseId } from '../hooks'
import type {
  CategoryRequest,
  FeedbackCategory,
  FeedbackMode,
  MaterialType,
  TargetMode,
} from '../interfaces'
import { MATERIAL_TYPE_CATALOG, sortMaterialTypes } from '../materialTypes'
import { presentationApi } from '../network'
import { getApiError, getErrorMessage } from '../utils'

type SettingsAction = { resetExistingData?: boolean } & (
  | { type: 'config' }
  | { type: 'materials' }
  | { type: 'create-category'; request: CategoryRequest }
  | { type: 'update-category'; categoryId: string; request: CategoryRequest }
  | { type: 'delete-category'; categoryId: string }
)

interface ResetPrompt {
  title: string
  description: string
  actionLabel: string
}

interface ResetAction extends ResetPrompt {
  run: () => SettingsAction
}

const CATEGORY_LOCK_DESCRIPTION =
  'The category structure is locked because feedback already exists. Continuing permanently deletes every draft and submitted evaluation in this phase'

const SETTINGS_ACTION_COPY: Record<
  SettingsAction['type'],
  { success: string; failure: string; resetDescription?: string; reset: ResetPrompt }
> = {
  config: {
    success: 'Presentation settings saved',
    failure: 'Could not save presentation settings',
    resetDescription: 'Existing presentation data was reset to apply the new modes.',
    reset: {
      title: 'Reset existing presentation data?',
      description:
        'Changing the target or feedback mode is locked because the phase already contains presentations or evaluations. Continuing permanently deletes the affected schedule assignments, uploaded materials, and feedback so the new mode can be applied.',
      actionLabel: 'Reset data and save',
    },
  },
  materials: {
    success: 'Requested uploads saved',
    failure: 'Could not save the requested uploads',
    // Requirements never invalidate existing data, so the server does not lock this change.
    // The prompt only exists because every action shares one confirmation path.
    reset: {
      title: 'Reset existing presentation data?',
      description:
        'This phase already contains presentations or evaluations that block the change. Continuing permanently deletes the affected schedule assignments, uploaded materials, and feedback.',
      actionLabel: 'Reset data and save',
    },
  },
  'create-category': {
    success: 'Feedback category added',
    failure: 'Could not add feedback category',
    reset: {
      title: 'Reset feedback and add this category?',
      description: `${CATEGORY_LOCK_DESCRIPTION} before adding the category.`,
      actionLabel: 'Reset feedback and add',
    },
  },
  'update-category': {
    success: 'Feedback category updated',
    failure: 'Could not update feedback category',
    reset: {
      title: 'Reset feedback and update this category?',
      description: `${CATEGORY_LOCK_DESCRIPTION} before saving the category.`,
      actionLabel: 'Reset feedback and save',
    },
  },
  'delete-category': {
    success: 'Feedback category deleted',
    failure: 'Could not delete feedback category',
    reset: {
      title: 'Reset feedback and delete this category?',
      description: `${CATEGORY_LOCK_DESCRIPTION} before deleting the category.`,
      actionLabel: 'Reset feedback and delete',
    },
  },
}

interface CategoryRowProps {
  category: FeedbackCategory
  isPending: boolean
  onSave: (request: CategoryRequest) => void
  onDelete: () => void
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
  const [requiredMaterialTypes, setRequiredMaterialTypes] = useState<MaterialType[]>([])
  const [newCategoryName, setNewCategoryName] = useState('')
  const [newCategoryDescription, setNewCategoryDescription] = useState('')
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
    setRequiredMaterialTypes(sortMaterialTypes(configQuery.data.requiredMaterialTypes ?? []))
  }, [configQuery.data])

  const invalidateSettings = () => {
    void queryClient.invalidateQueries({ queryKey: ['presentation-config', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentation-categories', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentations', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentation-slots', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentation-targets', coursePhaseId] })
  }

  // The modes and the requested uploads have their own save button, so each one submits its
  // own edits together with the values already stored for the other.
  const configPayload = (action: SettingsAction) => ({
    targetMode:
      action.type === 'materials' ? (configQuery.data?.targetMode ?? targetMode) : targetMode,
    feedbackMode:
      action.type === 'materials' ? (configQuery.data?.feedbackMode ?? feedbackMode) : feedbackMode,
    requiredMaterialTypes:
      action.type === 'materials'
        ? requiredMaterialTypes
        : (configQuery.data?.requiredMaterialTypes ?? requiredMaterialTypes),
  })

  const mutation = useMutation({
    mutationFn: async (action: SettingsAction) => {
      const reset = action.resetExistingData ?? false
      if (action.type === 'config' || action.type === 'materials') {
        await presentationApi.updateConfig(coursePhaseId, configPayload(action), reset)
      } else if (action.type === 'create-category') {
        await presentationApi.createCategory(coursePhaseId, action.request, reset)
      } else if (action.type === 'update-category') {
        await presentationApi.updateCategory(
          coursePhaseId,
          action.categoryId,
          action.request,
          reset,
        )
      } else {
        await presentationApi.deleteCategory(coursePhaseId, action.categoryId, reset)
      }
    },
    onSuccess: (_data, action) => {
      invalidateSettings()
      setResetAction(undefined)
      if (action.type === 'create-category') {
        setNewCategoryName('')
        setNewCategoryDescription('')
      }
      const copy = SETTINGS_ACTION_COPY[action.type]
      toast({
        title: copy.success,
        description: action.resetExistingData ? copy.resetDescription : undefined,
      })
    },
    onError: (error, action) => {
      const copy = SETTINGS_ACTION_COPY[action.type]
      const code = getApiError(error).code
      // The lock is not a failure: it asks the lecturer to confirm a destructive reset,
      // which replays the same action with resetExistingData set.
      if (code === 'config_locked' || code === 'categories_locked') {
        setResetAction({ ...copy.reset, run: () => ({ ...action, resetExistingData: true }) })
        return
      }
      toast({
        title: copy.failure,
        description: getErrorMessage(error, 'Please try again.'),
        variant: 'destructive',
      })
    },
  })

  const isSaving = mutation.isPending

  const newCategoryAction = (): SettingsAction => ({
    type: 'create-category',
    request: {
      name: newCategoryName.trim(),
      description: newCategoryDescription.trim(),
      position:
        Math.max(-1, ...(categoriesQuery.data ?? []).map((category) => category.position)) + 1,
    },
  })

  const confirmReset = () => {
    if (!resetAction) return
    mutation.mutate(resetAction.run())
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
  const savedMaterialTypes = sortMaterialTypes(configQuery.data?.requiredMaterialTypes ?? [])
  const materialTypesChanged =
    savedMaterialTypes.length !== requiredMaterialTypes.length ||
    savedMaterialTypes.some((type, index) => type !== requiredMaterialTypes[index])

  const toggleMaterialType = (type: MaterialType, enabled: boolean) =>
    setRequiredMaterialTypes((current) =>
      sortMaterialTypes(
        enabled ? [...current, type] : current.filter((candidate) => candidate !== type),
      ),
    )

  return (
    <div className='space-y-6'>
      <div>
        <ManagementPageHeader>Presentation settings</ManagementPageHeader>
        <p className='text-muted-foreground'>
          Choose who presents, what they hand in, how instructors collaborate, and the written
          feedback structure.
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
          <Button
            disabled={!configChanged || isSaving}
            onClick={() => mutation.mutate({ type: 'config' })}
          >
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
            <Upload className='h-5 w-5' />
            Requested student uploads
          </CardTitle>
          <CardDescription>
            Presenters get one upload slot per selected item, and only the listed file formats are
            accepted. Changing this never deletes files that were already handed in.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-5'>
          <div className='space-y-3'>
            {MATERIAL_TYPE_CATALOG.map((definition) => (
              <div key={definition.type} className='flex items-start gap-3 rounded-lg border p-3'>
                <Checkbox
                  id={`material-type-${definition.type}`}
                  className='mt-0.5'
                  checked={requiredMaterialTypes.includes(definition.type)}
                  onCheckedChange={(checked) =>
                    toggleMaterialType(definition.type, checked === true)
                  }
                />
                <div className='space-y-1'>
                  <Label htmlFor={`material-type-${definition.type}`} className='cursor-pointer'>
                    {definition.label}
                  </Label>
                  <p className='text-xs text-muted-foreground'>
                    {definition.description} {definition.formats}
                    {definition.note ? ` · ${definition.note}` : ''}
                  </p>
                </div>
              </div>
            ))}
          </div>
          {requiredMaterialTypes.length === 0 ? (
            <p className='text-sm text-muted-foreground'>
              Presenters are not asked for any uploads and cannot attach files.
            </p>
          ) : null}
          <Button
            disabled={!materialTypesChanged || isSaving}
            onClick={() => mutation.mutate({ type: 'materials' })}
          >
            {isSaving ? (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <Save className='mr-2 h-4 w-4' />
            )}
            Save requested uploads
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
              onClick={() => mutation.mutate(newCategoryAction())}
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
              onSave={(request) =>
                mutation.mutate({ type: 'update-category', categoryId: category.id, request })
              }
              onDelete={() => mutation.mutate({ type: 'delete-category', categoryId: category.id })}
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
