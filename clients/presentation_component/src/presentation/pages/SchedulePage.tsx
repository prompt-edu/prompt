import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Alert,
  AlertDescription,
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Checkbox,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  useToast,
} from '@tumaet/prompt-ui-components'
import { format } from 'date-fns'
import {
  Calendar,
  Clock,
  Copy,
  MapPin,
  MessageSquareText,
  Paperclip,
  Pencil,
  Plus,
  Trash2,
  UserPlus,
  X,
} from 'lucide-react'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCoursePhaseId } from '../hooks'
import type { CreateSlotRequest, PresentationSlot, PresentationTarget } from '../interfaces'
import { presentationApi } from '../network'
import {
  buildSlotTimes,
  EMPTY_SERIES,
  emptySlotForm,
  formatResolvedEnd,
  formatSlotCount,
  generateSlotSeries,
  MAX_SERIES_SLOTS,
  type SlotFormData,
  slotFormFromTimes,
} from '../slotSeries'
import { getErrorMessage } from '../utils'

type ScheduleAction =
  | { type: 'create'; request: CreateSlotRequest }
  | { type: 'create-series'; requests: CreateSlotRequest[] }
  | { type: 'update'; slotId: string; request: CreateSlotRequest }
  | { type: 'delete'; slotId: string }
  | { type: 'assign'; slotId: string; target: PresentationTarget }
  | { type: 'unassign'; slotId: string }

const slotRequest = (data: SlotFormData): CreateSlotRequest => {
  const { start, end } = buildSlotTimes(data.startTime, data.endTime)
  return {
    startTime: start.toISOString(),
    endTime: end.toISOString(),
    location: data.location.trim() || undefined,
  }
}

const seriesRequests = (data: SlotFormData): CreateSlotRequest[] =>
  generateSlotSeries(data).slots.map((slot) => ({
    startTime: slot.start.toISOString(),
    endTime: slot.end.toISOString(),
    location: data.location.trim() || undefined,
  }))

interface SlotFormFieldsProps {
  idPrefix: string
  formData: SlotFormData
  onChange: (data: SlotFormData) => void
}

const SlotFormFields = ({ idPrefix, formData, onChange }: SlotFormFieldsProps) => {
  const slotTimes =
    formData.startTime && formData.endTime
      ? buildSlotTimes(formData.startTime, formData.endTime)
      : null

  return (
    <>
      <div className='space-y-2'>
        <Label htmlFor={`${idPrefix}-start`}>Start Time</Label>
        <Input
          id={`${idPrefix}-start`}
          type='datetime-local'
          value={formData.startTime}
          onChange={(event) => onChange({ ...formData, startTime: event.target.value })}
        />
      </div>
      <div className='space-y-2'>
        <Label htmlFor={`${idPrefix}-end`}>End Time</Label>
        <Input
          id={`${idPrefix}-end`}
          type='time'
          value={formData.endTime}
          onChange={(event) => onChange({ ...formData, endTime: event.target.value })}
        />
        {slotTimes && (
          <p className='text-sm text-muted-foreground'>{formatResolvedEnd(slotTimes)}</p>
        )}
      </div>
      <div className='space-y-2'>
        <Label htmlFor={`${idPrefix}-location`}>Location (Optional)</Label>
        <Input
          id={`${idPrefix}-location`}
          placeholder='e.g., Room 101, Building A'
          value={formData.location}
          onChange={(event) => onChange({ ...formData, location: event.target.value })}
        />
      </div>
    </>
  )
}

const SchedulePage = () => {
  const coursePhaseId = useCoursePhaseId()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const { toast } = useToast()
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [createMultipleSlots, setCreateMultipleSlots] = useState(false)
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
  const [isAssignDialogOpen, setIsAssignDialogOpen] = useState(false)
  const [editingSlot, setEditingSlot] = useState<PresentationSlot>()
  const [assigningSlot, setAssigningSlot] = useState<PresentationSlot>()
  const [selectedTargetId, setSelectedTargetId] = useState('')
  const [unassigningSlot, setUnassigningSlot] = useState<PresentationSlot>()
  const [deletingSlot, setDeletingSlot] = useState<PresentationSlot>()
  const [formData, setFormData] = useState<SlotFormData>(emptySlotForm)

  const slotsQuery = useQuery({
    queryKey: ['presentation-slots', coursePhaseId],
    queryFn: () => presentationApi.getSlots(coursePhaseId),
    enabled: Boolean(coursePhaseId),
  })
  const targetsQuery = useQuery({
    queryKey: ['presentation-targets', coursePhaseId],
    queryFn: () => presentationApi.getTargets(coursePhaseId),
    enabled: Boolean(coursePhaseId),
  })

  const slots = slotsQuery.data ?? []
  const targets = targetsQuery.data ?? []
  const unassignedTargets = targets.filter(
    (target) => !target.assigned && !target.assignedPresentationId,
  )

  const invalidateSchedule = () => {
    void queryClient.invalidateQueries({ queryKey: ['presentation-slots', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentation-targets', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentations', coursePhaseId] })
  }

  const resetForm = () => {
    setCreateMultipleSlots(false)
    setFormData(emptySlotForm())
  }

  const mutation = useMutation({
    mutationFn: async (action: ScheduleAction) => {
      if (action.type === 'create') {
        return presentationApi.createSlot(coursePhaseId, action.request)
      }
      if (action.type === 'create-series') {
        return presentationApi.createSlots(coursePhaseId, action.requests)
      }
      if (action.type === 'update') {
        return presentationApi.updateSlot(coursePhaseId, action.slotId, action.request)
      }
      if (action.type === 'delete') {
        return presentationApi.deleteSlot(coursePhaseId, action.slotId)
      }
      if (action.type === 'assign') {
        return presentationApi.assignTarget(coursePhaseId, action.slotId, {
          targetId: action.target.id,
          targetName: action.target.name,
          targetType: action.target.type,
        })
      }
      return presentationApi.unassignTarget(coursePhaseId, action.slotId)
    },
    onSuccess: (result, action) => {
      invalidateSchedule()
      if (action.type === 'create' || action.type === 'create-series') {
        setIsCreateDialogOpen(false)
        resetForm()
        toast({
          title: action.type === 'create' ? 'Slot created' : 'Slots created',
          description:
            action.type === 'create'
              ? 'The presentation slot has been created.'
              : `${formatSlotCount(Array.isArray(result) ? result.length : 0)} created successfully.`,
        })
        return
      }
      if (action.type === 'update') {
        setIsEditDialogOpen(false)
        setEditingSlot(undefined)
        resetForm()
      }
      if (action.type === 'assign') {
        setIsAssignDialogOpen(false)
        setAssigningSlot(undefined)
        setSelectedTargetId('')
      }
      if (action.type === 'unassign') setUnassigningSlot(undefined)
      if (action.type === 'delete') setDeletingSlot(undefined)
      toast({ title: 'Schedule updated' })
    },
    onError: (error) => {
      toast({
        title: 'Could not update schedule',
        description: getErrorMessage(error, 'Please try again.'),
        variant: 'destructive',
      })
    },
  })

  const slotTimes =
    formData.startTime && formData.endTime
      ? buildSlotTimes(formData.startTime, formData.endTime)
      : null
  const isTimeRangeValid = !!slotTimes && slotTimes.start < slotTimes.end
  const seriesPreview = useMemo(
    () => (createMultipleSlots ? generateSlotSeries(formData) : EMPTY_SERIES),
    [createMultipleSlots, formData],
  )

  const handleCreateSlots = () => {
    if (!isTimeRangeValid) return
    if (createMultipleSlots) {
      if (seriesPreview.slots.length === 0) return
      mutation.mutate({ type: 'create-series', requests: seriesRequests(formData) })
    } else {
      mutation.mutate({ type: 'create', request: slotRequest(formData) })
    }
  }

  const handleEditClick = (slot: PresentationSlot) => {
    setEditingSlot(slot)
    setFormData(slotFormFromTimes(slot.startTime, slot.endTime, slot.location ?? ''))
    setIsEditDialogOpen(true)
  }

  const handleCloneClick = (slot: PresentationSlot) => {
    setCreateMultipleSlots(false)
    setFormData(slotFormFromTimes(slot.startTime, slot.endTime, slot.location ?? ''))
    setIsCreateDialogOpen(true)
  }

  const handleAssignClick = (slot: PresentationSlot) => {
    setAssigningSlot(slot)
    setSelectedTargetId('')
    setIsAssignDialogOpen(true)
  }

  const handleAssign = () => {
    const target = targets.find((candidate) => candidate.id === selectedTargetId)
    if (!assigningSlot || !target) return
    mutation.mutate({ type: 'assign', slotId: assigningSlot.id, target })
  }

  if (slotsQuery.isLoading || targetsQuery.isLoading) return <LoadingPage />
  if (slotsQuery.isError || targetsQuery.isError) {
    return (
      <ErrorPage
        message='The presentation schedule could not be loaded.'
        onRetry={() => {
          void slotsQuery.refetch()
          void targetsQuery.refetch()
        }}
      />
    )
  }

  return (
    <div className='container mx-auto px-4 py-8'>
      <ManagementPageHeader>Presentation Schedule Management</ManagementPageHeader>

      <div className='mb-6 flex items-center justify-between gap-4'>
        <p className='text-muted-foreground'>
          Create presentation slots and assign presenters. Slots may overlap for parallel sessions.
        </p>
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button onClick={resetForm}>
              <Plus className='mr-2 h-4 w-4' />
              Create Slots
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Presentation Slots</DialogTitle>
              <DialogDescription>
                Add one presentation slot or divide the time range into multiple slots.
              </DialogDescription>
            </DialogHeader>
            <div className='space-y-4 py-4'>
              <SlotFormFields idPrefix='create-slot' formData={formData} onChange={setFormData} />
              <div className='flex items-center gap-2'>
                <Checkbox
                  id='createMultipleSlots'
                  checked={createMultipleSlots}
                  onCheckedChange={(checked) => setCreateMultipleSlots(checked === true)}
                />
                <Label htmlFor='createMultipleSlots'>Create multiple slots</Label>
              </div>
              {createMultipleSlots && (
                <>
                  <div className='grid grid-cols-2 gap-4'>
                    <div className='space-y-2'>
                      <Label htmlFor='durationMinutes'>Slot Duration (min)</Label>
                      <Input
                        id='durationMinutes'
                        type='number'
                        min='1'
                        value={formData.durationMinutes}
                        onChange={(event) =>
                          setFormData({
                            ...formData,
                            durationMinutes: parseInt(event.target.value, 10) || 0,
                          })
                        }
                      />
                    </div>
                    <div className='space-y-2'>
                      <Label htmlFor='breakMinutes'>Break Between (min)</Label>
                      <Input
                        id='breakMinutes'
                        type='number'
                        min='0'
                        value={formData.breakMinutes}
                        onChange={(event) =>
                          setFormData({
                            ...formData,
                            breakMinutes: Math.max(0, parseInt(event.target.value, 10) || 0),
                          })
                        }
                      />
                    </div>
                  </div>
                  <p className='text-sm text-muted-foreground'>
                    {seriesPreview.slots.length > 0
                      ? `This will create ${formatSlotCount(seriesPreview.slots.length)}.`
                      : 'Choose a valid time range and duration to preview the slots.'}
                  </p>
                  {seriesPreview.truncated && (
                    <p className='text-sm text-destructive'>
                      The time range fits more slots than the limit of {MAX_SERIES_SLOTS}. Only the
                      first {MAX_SERIES_SLOTS} will be created. Shorten the range or create the rest
                      separately.
                    </p>
                  )}
                </>
              )}
            </div>
            <DialogFooter>
              <Button variant='outline' onClick={() => setIsCreateDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreateSlots}
                disabled={
                  !isTimeRangeValid ||
                  (createMultipleSlots && seriesPreview.slots.length === 0) ||
                  mutation.isPending
                }
              >
                {mutation.isPending
                  ? 'Creating...'
                  : createMultipleSlots
                    ? `Create ${formatSlotCount(seriesPreview.slots.length, true)}`
                    : 'Create Slot'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Presentation Slot</DialogTitle>
            <DialogDescription>Update the presentation slot details</DialogDescription>
          </DialogHeader>
          <div className='space-y-4 py-4'>
            <SlotFormFields idPrefix='edit-slot' formData={formData} onChange={setFormData} />
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setIsEditDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={() =>
                editingSlot &&
                isTimeRangeValid &&
                mutation.mutate({
                  type: 'update',
                  slotId: editingSlot.id,
                  request: slotRequest(formData),
                })
              }
              disabled={!isTimeRangeValid || mutation.isPending}
            >
              {mutation.isPending ? 'Updating...' : 'Update'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Slots Table */}
      {slots.length > 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>Presentation Slots</CardTitle>
            <CardDescription>Manage all scheduled presentation slots</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date &amp; Time</TableHead>
                  <TableHead>Location</TableHead>
                  <TableHead>Presenter</TableHead>
                  <TableHead>Materials</TableHead>
                  <TableHead>Feedback</TableHead>
                  <TableHead className='text-right'>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {slots.map((slot) => {
                  const presentation = slot.presentation
                  const isPast = new Date(slot.endTime) < new Date()

                  return (
                    <TableRow key={slot.id}>
                      <TableCell>
                        <div className='space-y-1'>
                          <div className='flex items-center gap-2'>
                            <Calendar className='h-4 w-4 text-muted-foreground' />
                            <span className='font-medium'>
                              {format(new Date(slot.startTime), 'EEE, MMM d, yyyy')}
                            </span>
                            {isPast ? <Badge variant='secondary'>Past</Badge> : null}
                          </div>
                          <div className='flex items-center gap-2 text-sm text-muted-foreground'>
                            <Clock className='h-4 w-4' />
                            <span>
                              {format(new Date(slot.startTime), 'HH:mm')} -{' '}
                              {format(new Date(slot.endTime), 'HH:mm')}
                            </span>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        {slot.location ? (
                          <div className='flex items-center gap-2'>
                            <MapPin className='h-4 w-4 shrink-0 text-muted-foreground' />
                            {slot.location.match(/^https?:\/\//) ? (
                              <a
                                href={slot.location}
                                target='_blank'
                                rel='noopener noreferrer'
                                className='min-w-0 truncate text-blue-600 hover:underline'
                              >
                                {slot.location}
                              </a>
                            ) : (
                              <span className='min-w-0 truncate'>{slot.location}</span>
                            )}
                          </div>
                        ) : (
                          <span className='text-muted-foreground'>—</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {presentation ? (
                          <Badge
                            variant='outline'
                            className='group cursor-pointer pr-1 transition-colors hover:bg-destructive hover:text-destructive-foreground'
                            onClick={() => setUnassigningSlot(slot)}
                            title={`Click to unassign ${presentation.targetName}`}
                          >
                            {presentation.targetName}
                            <X className='ml-1 h-3 w-3 opacity-50 group-hover:opacity-100' />
                          </Badge>
                        ) : (
                          <span className='text-sm text-muted-foreground'>Unassigned</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className='flex items-center gap-2'>
                          <Paperclip className='h-4 w-4 text-muted-foreground' />
                          <span>{presentation?.materialCount ?? 0}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        {presentation?.feedbackReleasedAt ? (
                          <Badge>Released</Badge>
                        ) : (presentation?.submittedFeedbackCount ?? 0) > 0 ? (
                          <Badge variant='secondary'>
                            {presentation?.submittedFeedbackCount} submitted
                          </Badge>
                        ) : (
                          <span className='text-sm text-muted-foreground'>No evaluations</span>
                        )}
                      </TableCell>
                      <TableCell className='text-right'>
                        <div className='flex justify-end gap-2'>
                          {presentation ? (
                            <Button
                              variant='ghost'
                              size='sm'
                              onClick={() => navigate(`../presentations/${presentation.id}`)}
                              aria-label='Open feedback'
                              title='Open feedback'
                            >
                              <MessageSquareText className='h-4 w-4' />
                            </Button>
                          ) : (
                            <Button
                              variant='ghost'
                              size='sm'
                              onClick={() => handleAssignClick(slot)}
                              disabled={unassignedTargets.length === 0}
                              aria-label='Assign presenter'
                              title='Assign presenter to slot'
                            >
                              <UserPlus className='h-4 w-4' />
                            </Button>
                          )}
                          <Button
                            variant='ghost'
                            size='sm'
                            onClick={() => handleCloneClick(slot)}
                            aria-label='Clone slot'
                            title='Clone slot'
                          >
                            <Copy className='h-4 w-4' />
                          </Button>
                          <Button
                            variant='ghost'
                            size='sm'
                            onClick={() => handleEditClick(slot)}
                            aria-label='Edit slot'
                          >
                            <Pencil className='h-4 w-4' />
                          </Button>
                          <Button
                            variant='ghost'
                            size='sm'
                            onClick={() => setDeletingSlot(slot)}
                            disabled={Boolean(presentation) || mutation.isPending}
                            aria-label='Delete slot'
                            title={
                              presentation
                                ? 'Unassign the presenter before deleting the slot'
                                : 'Delete slot'
                            }
                          >
                            <Trash2 className='h-4 w-4' />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      ) : (
        <Alert>
          <AlertDescription>
            No presentation slots have been created yet. Click &quot;Create Slots&quot; to add your
            first slot.
          </AlertDescription>
        </Alert>
      )}

      {/* Assign Presenter Dialog */}
      <Dialog open={isAssignDialogOpen} onOpenChange={setIsAssignDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Assign Presenter to Slot</DialogTitle>
            <DialogDescription>
              {assigningSlot && (
                <>
                  Assign a presenter to the slot on{' '}
                  {new Date(assigningSlot.startTime).toLocaleString('en-US', {
                    dateStyle: 'medium',
                    timeStyle: 'short',
                  })}
                  . Students cannot pick their own slot.
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-4 py-4'>
            {unassignedTargets.length > 0 ? (
              <div className='space-y-2'>
                <Label htmlFor='presenter-select'>Select Presenter</Label>
                <Select value={selectedTargetId} onValueChange={setSelectedTargetId}>
                  <SelectTrigger id='presenter-select'>
                    <SelectValue placeholder='Choose a team or student...' />
                  </SelectTrigger>
                  <SelectContent>
                    {unassignedTargets.map((target) => (
                      <SelectItem key={target.id} value={target.id}>
                        {target.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className='text-sm text-muted-foreground'>
                  {unassignedTargets.length} unassigned presenter
                  {unassignedTargets.length !== 1 && 's'} available
                </p>
              </div>
            ) : (
              <Alert>
                <AlertDescription>
                  Every presenter already has a slot. No unassigned presenters are available.
                </AlertDescription>
              </Alert>
            )}
          </div>
          <DialogFooter>
            <Button
              variant='outline'
              onClick={() => {
                setIsAssignDialogOpen(false)
                setAssigningSlot(undefined)
                setSelectedTargetId('')
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleAssign}
              disabled={!selectedTargetId || mutation.isPending || unassignedTargets.length === 0}
            >
              {mutation.isPending ? 'Assigning...' : 'Assign Presenter'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Unassign Confirmation Dialog */}
      <Dialog
        open={Boolean(unassigningSlot)}
        onOpenChange={(open) => !open && setUnassigningSlot(undefined)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Unassign Presenter</DialogTitle>
            <DialogDescription>
              {unassigningSlot?.presentation && (
                <>
                  Are you sure you want to unassign{' '}
                  <strong>{unassigningSlot.presentation.targetName}</strong> from this slot?
                  Uploaded materials and feedback have to be deleted first.
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant='outline' onClick={() => setUnassigningSlot(undefined)}>
              Cancel
            </Button>
            <Button
              variant='destructive'
              onClick={() =>
                unassigningSlot && mutation.mutate({ type: 'unassign', slotId: unassigningSlot.id })
              }
              disabled={mutation.isPending}
            >
              {mutation.isPending ? 'Unassigning...' : 'Unassign'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog
        open={Boolean(deletingSlot)}
        onOpenChange={(open) => !open && setDeletingSlot(undefined)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Presentation Slot</DialogTitle>
            <DialogDescription>
              The unassigned time slot will be permanently removed.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant='outline' onClick={() => setDeletingSlot(undefined)}>
              Cancel
            </Button>
            <Button
              variant='destructive'
              onClick={() =>
                deletingSlot && mutation.mutate({ type: 'delete', slotId: deletingSlot.id })
              }
              disabled={mutation.isPending}
            >
              {mutation.isPending ? 'Deleting...' : 'Delete Slot'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default SchedulePage
