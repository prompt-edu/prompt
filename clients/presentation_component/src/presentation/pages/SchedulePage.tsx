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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  useToast,
} from '@tumaet/prompt-ui-components'
import { CalendarPlus, Loader2, MapPin, Save, Trash2, UserRoundCheck, X } from 'lucide-react'
import { useState } from 'react'
import { useCoursePhaseId } from '../hooks'
import type {
  CreateSlotRequest,
  PresentationSlot,
  PresentationSummary,
  PresentationTarget,
} from '../interfaces'
import { presentationApi } from '../network'
import { getErrorMessage, toDateTimeLocal } from '../utils'

interface SlotCardProps {
  slot: PresentationSlot
  presentation?: PresentationSummary
  targets: PresentationTarget[]
  overlaps: boolean
  isPending: boolean
  onSave: (slotId: string, request: CreateSlotRequest) => void
  onDelete: (slotId: string) => void
  onAssign: (slotId: string, target: PresentationTarget) => void
  onUnassign: (slotId: string) => void
}

const SlotCard = ({
  slot,
  presentation,
  targets,
  overlaps,
  isPending,
  onSave,
  onDelete,
  onAssign,
  onUnassign,
}: SlotCardProps) => {
  const [startTime, setStartTime] = useState(() => toDateTimeLocal(slot.startTime))
  const [endTime, setEndTime] = useState(() => toDateTimeLocal(slot.endTime))
  const [location, setLocation] = useState(slot.location ?? '')
  const [targetId, setTargetId] = useState(presentation?.targetId ?? '')
  const [deleteOpen, setDeleteOpen] = useState(false)
  const selectedTarget = targets.find((target) => target.id === targetId)
  const invalidTime =
    !startTime || !endTime || new Date(endTime).getTime() <= new Date(startTime).getTime()

  return (
    <Card>
      <CardHeader>
        <div className='flex flex-wrap items-start justify-between gap-3'>
          <div>
            <CardTitle className='text-base'>
              {presentation?.targetName ?? 'Unassigned presentation slot'}
            </CardTitle>
            <CardDescription>
              Slot times may overlap when presentations run in parallel.
            </CardDescription>
          </div>
          <div className='flex gap-2'>
            {overlaps ? <Badge variant='secondary'>Overlapping</Badge> : null}
            {presentation ? <Badge>Assigned</Badge> : <Badge variant='outline'>Open</Badge>}
          </div>
        </div>
      </CardHeader>
      <CardContent className='space-y-5'>
        <div className='grid gap-4 md:grid-cols-3'>
          <div className='space-y-2'>
            <Label htmlFor={`slot-start-${slot.id}`}>Start</Label>
            <Input
              id={`slot-start-${slot.id}`}
              type='datetime-local'
              value={startTime}
              onChange={(event) => setStartTime(event.target.value)}
            />
          </div>
          <div className='space-y-2'>
            <Label htmlFor={`slot-end-${slot.id}`}>End</Label>
            <Input
              id={`slot-end-${slot.id}`}
              type='datetime-local'
              value={endTime}
              onChange={(event) => setEndTime(event.target.value)}
            />
          </div>
          <div className='space-y-2'>
            <Label htmlFor={`slot-location-${slot.id}`}>Location (optional)</Label>
            <div className='relative'>
              <MapPin className='absolute left-3 top-2.5 h-4 w-4 text-muted-foreground' />
              <Input
                id={`slot-location-${slot.id}`}
                className='pl-9'
                value={location}
                onChange={(event) => setLocation(event.target.value)}
                placeholder='Room or meeting link'
              />
            </div>
          </div>
        </div>
        {invalidTime ? (
          <p className='text-sm text-destructive'>End time must be after the start time.</p>
        ) : null}
        <div className='flex flex-wrap items-end justify-between gap-3 border-t pt-4'>
          <div className='min-w-64 flex-1 space-y-2'>
            <Label>Presenter</Label>
            <Select value={targetId || undefined} onValueChange={setTargetId}>
              <SelectTrigger>
                <SelectValue placeholder='Select a team or student' />
              </SelectTrigger>
              <SelectContent>
                {targets.map((target) => {
                  const assignedElsewhere =
                    Boolean(target.assigned || target.assignedPresentationId) &&
                    target.id !== presentation?.targetId
                  return (
                    <SelectItem key={target.id} value={target.id} disabled={assignedElsewhere}>
                      {target.name}
                      {assignedElsewhere ? ' (already assigned)' : ''}
                    </SelectItem>
                  )
                })}
              </SelectContent>
            </Select>
          </div>
          <div className='flex flex-wrap gap-2'>
            <Button
              variant='outline'
              disabled={isPending || invalidTime}
              onClick={() =>
                onSave(slot.id, {
                  startTime: new Date(startTime).toISOString(),
                  endTime: new Date(endTime).toISOString(),
                  location: location.trim() || undefined,
                })
              }
            >
              <Save className='mr-2 h-4 w-4' />
              Save slot
            </Button>
            {presentation ? (
              <Button variant='outline' disabled={isPending} onClick={() => onUnassign(slot.id)}>
                <X className='mr-2 h-4 w-4' />
                Unassign
              </Button>
            ) : (
              <Button
                disabled={!selectedTarget || isPending}
                onClick={() => selectedTarget && onAssign(slot.id, selectedTarget)}
              >
                <UserRoundCheck className='mr-2 h-4 w-4' />
                Assign
              </Button>
            )}
            <Button
              variant='ghost'
              className='text-destructive hover:bg-destructive/10 hover:text-destructive'
              disabled={Boolean(presentation) || isPending}
              onClick={() => setDeleteOpen(true)}
            >
              <Trash2 className='h-4 w-4' />
              <span className='sr-only'>Delete slot</span>
            </Button>
          </div>
        </div>
      </CardContent>
      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete this presentation slot?</AlertDialogTitle>
            <AlertDialogDescription>
              The unassigned time slot will be permanently removed.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
              onClick={() => onDelete(slot.id)}
            >
              Delete slot
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  )
}

const SchedulePage = () => {
  const coursePhaseId = useCoursePhaseId()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const [location, setLocation] = useState('')

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
  const presentationsQuery = useQuery({
    queryKey: ['presentations', coursePhaseId, 'all'],
    queryFn: () => presentationApi.getPresentations(coursePhaseId),
    enabled: Boolean(coursePhaseId),
  })

  const invalidateSchedule = () => {
    void queryClient.invalidateQueries({ queryKey: ['presentation-slots', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentation-targets', coursePhaseId] })
    void queryClient.invalidateQueries({ queryKey: ['presentations', coursePhaseId] })
  }

  const mutation = useMutation({
    mutationFn: async (
      action:
        | { type: 'create'; request: CreateSlotRequest }
        | { type: 'update'; slotId: string; request: CreateSlotRequest }
        | { type: 'delete'; slotId: string }
        | { type: 'assign'; slotId: string; target: PresentationTarget }
        | { type: 'unassign'; slotId: string },
    ) => {
      if (action.type === 'create') {
        return presentationApi.createSlot(coursePhaseId, action.request)
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
    onSuccess: (_data, action) => {
      invalidateSchedule()
      if (action.type === 'create') {
        setStartTime('')
        setEndTime('')
        setLocation('')
      }
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

  if (slotsQuery.isLoading || targetsQuery.isLoading || presentationsQuery.isLoading) {
    return <LoadingPage />
  }
  if (slotsQuery.isError || targetsQuery.isError || presentationsQuery.isError) {
    return (
      <ErrorPage
        message='The presentation schedule could not be loaded.'
        onRetry={() => {
          void slotsQuery.refetch()
          void targetsQuery.refetch()
          void presentationsQuery.refetch()
        }}
      />
    )
  }

  const slots = slotsQuery.data ?? []
  const targets = targetsQuery.data ?? []
  const presentations = presentationsQuery.data ?? []
  const presentationBySlot = new Map(
    presentations.map((presentation) => [presentation.slotId, presentation]),
  )
  const createInvalid =
    !startTime || !endTime || new Date(endTime).getTime() <= new Date(startTime).getTime()

  const slotOverlaps = (slot: PresentationSlot): boolean =>
    slots.some(
      (candidate) =>
        candidate.id !== slot.id &&
        new Date(slot.startTime).getTime() < new Date(candidate.endTime).getTime() &&
        new Date(slot.endTime).getTime() > new Date(candidate.startTime).getTime(),
    )

  return (
    <div className='space-y-6'>
      <div>
        <ManagementPageHeader>Presentation schedule</ManagementPageHeader>
        <p className='text-muted-foreground'>
          Create staff-assigned slots for individual or team presentations. Overlapping slots are
          supported for parallel sessions.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            <CalendarPlus className='h-5 w-5' />
            Add presentation slot
          </CardTitle>
          <CardDescription>
            The location is optional and can be a room, stage, or meeting link.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='grid gap-4 md:grid-cols-3'>
            <div className='space-y-2'>
              <Label htmlFor='new-slot-start'>Start</Label>
              <Input
                id='new-slot-start'
                type='datetime-local'
                value={startTime}
                onChange={(event) => setStartTime(event.target.value)}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='new-slot-end'>End</Label>
              <Input
                id='new-slot-end'
                type='datetime-local'
                value={endTime}
                onChange={(event) => setEndTime(event.target.value)}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='new-slot-location'>Location (optional)</Label>
              <Input
                id='new-slot-location'
                value={location}
                onChange={(event) => setLocation(event.target.value)}
                placeholder='Room or meeting link'
              />
            </div>
          </div>
          {startTime && endTime && createInvalid ? (
            <Alert variant='destructive'>
              <AlertDescription>End time must be after the start time.</AlertDescription>
            </Alert>
          ) : null}
          <Button
            disabled={createInvalid || mutation.isPending}
            onClick={() =>
              mutation.mutate({
                type: 'create',
                request: {
                  startTime: new Date(startTime).toISOString(),
                  endTime: new Date(endTime).toISOString(),
                  location: location.trim() || undefined,
                },
              })
            }
          >
            {mutation.isPending ? (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <CalendarPlus className='mr-2 h-4 w-4' />
            )}
            Add slot
          </Button>
        </CardContent>
      </Card>

      <div className='space-y-4'>
        {slots.length === 0 ? (
          <Card>
            <CardContent className='py-10 text-center text-sm text-muted-foreground'>
              No presentation slots have been created.
            </CardContent>
          </Card>
        ) : null}
        {slots.map((slot) => (
          <SlotCard
            key={slot.id}
            slot={slot}
            presentation={slot.presentation ?? presentationBySlot.get(slot.id)}
            targets={targets}
            overlaps={slotOverlaps(slot)}
            isPending={mutation.isPending}
            onSave={(slotId, request) => mutation.mutate({ type: 'update', slotId, request })}
            onDelete={(slotId) => mutation.mutate({ type: 'delete', slotId })}
            onAssign={(slotId, target) => mutation.mutate({ type: 'assign', slotId, target })}
            onUnassign={(slotId) => mutation.mutate({ type: 'unassign', slotId })}
          />
        ))}
      </div>
    </div>
  )
}

export default SchedulePage
