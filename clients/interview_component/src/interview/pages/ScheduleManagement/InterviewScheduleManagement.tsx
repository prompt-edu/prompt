import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type CoursePhaseParticipationsWithResolution,
  getCoursePhaseParticipations,
} from '@tumaet/prompt-shared-state'
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
  Input,
  Label,
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
import { isAxiosError } from 'axios'
import { format } from 'date-fns'
import {
  Calendar,
  Clock,
  Copy,
  MapPin,
  Pencil,
  Plus,
  Trash2,
  UserPlus,
  Users,
  X,
} from 'lucide-react'
import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { InterviewSlotWithAssignments } from '../../interfaces/InterviewSlots'
import { interviewAxiosInstance } from '../../network/interviewServerConfig'

interface SlotFormData {
  startTime: string
  endTime: string
  durationMinutes: number
  breakMinutes: number
  location: string
  capacity: number
}

interface ErrorResponse {
  error?: string
}

interface SlotTimeRange {
  start: Date
  end: Date
}

interface GeneratedSeries {
  slots: SlotTimeRange[]
  truncated: boolean
}

interface CreatedSlotResponse {
  id: string
}

const MAX_MULTIPLE_SLOTS = 100

// Only the series form uses these; the single-slot and edit forms keep them at their defaults.
const SERIES_DEFAULTS = { durationMinutes: 30, breakMinutes: 0 }

const EMPTY_SERIES: GeneratedSeries = { slots: [], truncated: false }

const getErrorMessage = (error: unknown, fallback: string) =>
  isAxiosError<ErrorResponse>(error) ? (error.response?.data?.error ?? fallback) : fallback

const formatSlotCount = (count: number, capitalize = false) => {
  const noun = capitalize ? 'Slot' : 'slot'
  return `${count} ${noun}${count === 1 ? '' : 's'}`
}

const emptySlotForm = (): SlotFormData => ({
  startTime: '',
  endTime: '',
  ...SERIES_DEFAULTS,
  location: '',
  capacity: 1,
})

const slotFormFromSlot = (slot: InterviewSlotWithAssignments): SlotFormData => ({
  startTime: format(new Date(slot.startTime), "yyyy-MM-dd'T'HH:mm"),
  endTime: format(new Date(slot.endTime), 'HH:mm'),
  ...SERIES_DEFAULTS,
  location: slot.location || '',
  capacity: slot.capacity ?? 1,
})

// endTime is a time-of-day (HH:mm); an end at or before the start means the slot runs past midnight.
const buildSlotTimes = (startDateTime: string, endTimeOfDay: string): SlotTimeRange => {
  const start = new Date(startDateTime)
  const end = new Date(`${startDateTime.slice(0, 10)}T${endTimeOfDay}`)
  if (end <= start) {
    end.setDate(end.getDate() + 1)
  }
  return { start, end }
}

const formatResolvedEnd = (range: SlotTimeRange) => {
  const spansNextDay = range.end.getDate() !== range.start.getDate()
  return `Ends ${format(range.end, 'EEE, MMM d')} at ${format(range.end, 'HH:mm')}${
    spansNextDay ? ' (next day)' : ''
  }`
}

const slotRequestBody = (data: SlotFormData) => {
  const { start, end } = buildSlotTimes(data.startTime, data.endTime)
  return {
    startTime: start.toISOString(),
    endTime: end.toISOString(),
    location: data.location || null,
    capacity: data.capacity,
  }
}

// Steps in wall-clock minutes rather than milliseconds, so a DST transition inside the range does
// not shift the generated slot times by an hour.
const addLocalMinutes = (base: Date, minutes: number) =>
  new Date(
    base.getFullYear(),
    base.getMonth(),
    base.getDate(),
    base.getHours(),
    base.getMinutes() + minutes,
  )

const generateMultipleSlots = (data: SlotFormData): GeneratedSeries => {
  const slots: SlotTimeRange[] = []
  if (!data.startTime || !data.endTime || data.durationMinutes <= 0) return EMPTY_SERIES

  const { start, end } = buildSlotTimes(data.startTime, data.endTime)
  const stepMinutes = data.durationMinutes + Math.max(0, data.breakMinutes)

  for (let offset = 0; ; offset += stepMinutes) {
    const slotEnd = addLocalMinutes(start, offset + data.durationMinutes)
    if (slotEnd > end) break
    if (slots.length === MAX_MULTIPLE_SLOTS) return { slots, truncated: true }
    slots.push({ start: addLocalMinutes(start, offset), end: slotEnd })
  }

  return { slots, truncated: false }
}

export const InterviewScheduleManagement = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [createMultipleSlots, setCreateMultipleSlots] = useState(false)
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
  const [isAssignDialogOpen, setIsAssignDialogOpen] = useState(false)
  const [isUnassignDialogOpen, setIsUnassignDialogOpen] = useState(false)
  const [assigningSlot, setAssigningSlot] = useState<InterviewSlotWithAssignments | null>(null)
  const [unassigningInfo, setUnassigningInfo] = useState<{
    assignmentId: string
    studentName: string
  } | null>(null)
  const [selectedParticipationId, setSelectedParticipationId] = useState<string>('')
  const [editingSlot, setEditingSlot] = useState<InterviewSlotWithAssignments | null>(null)
  const [formData, setFormData] = useState<SlotFormData>(emptySlotForm)

  // Fetch all participants
  const { data: participations } = useQuery<CoursePhaseParticipationsWithResolution>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
    enabled: !!phaseId,
  })

  // Fetch all slots
  const {
    data: slots,
    isLoading,
    isError,
  } = useQuery<InterviewSlotWithAssignments[]>({
    queryKey: ['interviewSlotsWithAssignments', phaseId],
    queryFn: async () => {
      const response = await interviewAxiosInstance.get(
        `interview/api/course_phase/${phaseId}/interview-slots`,
      )
      return response.data
    },
    enabled: !!phaseId,
  })

  // Create slot mutation
  const createSlotMutation = useMutation({
    mutationFn: async (data: SlotFormData) => {
      const response = await interviewAxiosInstance.post(
        `interview/api/course_phase/${phaseId}/interview-slots`,
        slotRequestBody(data),
      )
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interviewSlotsWithAssignments', phaseId] })
      setIsCreateDialogOpen(false)
      resetForm()
      toast({
        title: 'Slot created',
        description: 'Interview slot has been created successfully.',
      })
    },
    onError: (error: unknown) => {
      toast({
        title: 'Creation failed',
        description: getErrorMessage(error, 'Failed to create interview slot.'),
        variant: 'destructive',
      })
    },
  })

  // Update slot mutation
  const updateSlotMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: SlotFormData }) => {
      const response = await interviewAxiosInstance.put(
        `interview/api/course_phase/${phaseId}/interview-slots/${id}`,
        slotRequestBody(data),
      )
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interviewSlotsWithAssignments', phaseId] })
      setIsEditDialogOpen(false)
      setEditingSlot(null)
      resetForm()
      toast({
        title: 'Slot updated',
        description: 'Interview slot has been updated successfully.',
      })
    },
    onError: (error: unknown) => {
      toast({
        title: 'Update failed',
        description: getErrorMessage(error, 'Failed to update interview slot.'),
        variant: 'destructive',
      })
    },
  })

  // Create multiple slots mutation - the series is generated client-side and created in one
  // transactional request, so a rejected slot never leaves a partial series behind
  const createMultipleSlotsMutation = useMutation({
    mutationFn: async (data: SlotFormData) => {
      const { slots: generatedSlots } = generateMultipleSlots(data)
      const response = await interviewAxiosInstance.post<CreatedSlotResponse[]>(
        `interview/api/course_phase/${phaseId}/interview-slots/batch`,
        {
          slots: generatedSlots.map((slot) => ({
            startTime: slot.start.toISOString(),
            endTime: slot.end.toISOString(),
            location: data.location || null,
            capacity: data.capacity,
          })),
        },
      )
      return response.data
    },
    onSuccess: (createdSlots) => {
      queryClient.invalidateQueries({ queryKey: ['interviewSlotsWithAssignments', phaseId] })
      setIsCreateDialogOpen(false)
      resetForm()
      toast({
        title: 'Slots created',
        description: `${formatSlotCount(createdSlots.length)} created successfully.`,
      })
    },
    onError: (error: unknown) => {
      toast({
        title: 'Slot creation failed',
        description: getErrorMessage(error, 'Failed to create interview slots.'),
        variant: 'destructive',
      })
    },
  })

  // Delete slot mutation
  const deleteSlotMutation = useMutation({
    mutationFn: async (slotId: string) => {
      await interviewAxiosInstance.delete(
        `interview/api/course_phase/${phaseId}/interview-slots/${slotId}`,
      )
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interviewSlotsWithAssignments', phaseId] })
      toast({
        title: 'Slot deleted',
        description: 'Interview slot has been deleted successfully.',
      })
    },
    onError: (error: unknown) => {
      toast({
        title: 'Deletion failed',
        description: getErrorMessage(error, 'Failed to delete interview slot.'),
        variant: 'destructive',
      })
    },
  })

  // Manual assignment mutation (admin)
  const assignStudentMutation = useMutation({
    mutationFn: async ({
      slotId,
      participationId,
    }: {
      slotId: string
      participationId: string
    }) => {
      const response = await interviewAxiosInstance.post(
        `interview/api/course_phase/${phaseId}/interview-assignments/admin`,
        {
          interview_slot_id: slotId,
          course_participation_id: participationId,
        },
      )
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interviewSlotsWithAssignments', phaseId] })
      setIsAssignDialogOpen(false)
      setSelectedParticipationId('')
      setAssigningSlot(null)
      toast({
        title: 'Student assigned',
        description: 'Student has been assigned to the interview slot successfully.',
      })
    },
    onError: (error: unknown) => {
      toast({
        title: 'Assignment failed',
        description: getErrorMessage(error, 'Failed to assign student to interview slot.'),
        variant: 'destructive',
      })
    },
  })

  // Unassign student mutation
  const unassignStudentMutation = useMutation({
    mutationFn: async (assignmentId: string) => {
      await interviewAxiosInstance.delete(
        `interview/api/course_phase/${phaseId}/interview-assignments/${assignmentId}`,
      )
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interviewSlotsWithAssignments', phaseId] })
      setIsUnassignDialogOpen(false)
      setUnassigningInfo(null)
      toast({
        title: 'Student unassigned',
        description: 'Student has been removed from the interview slot.',
      })
    },
    onError: (error: unknown) => {
      toast({
        title: 'Unassignment failed',
        description: getErrorMessage(error, 'Failed to unassign student.'),
        variant: 'destructive',
      })
    },
  })

  const resetForm = () => {
    setCreateMultipleSlots(false)
    setFormData(emptySlotForm())
  }

  const slotTimes =
    formData.startTime && formData.endTime
      ? buildSlotTimes(formData.startTime, formData.endTime)
      : null
  const isTimeRangeValid = !!slotTimes && slotTimes.start < slotTimes.end

  const seriesPreview = useMemo(
    () => (createMultipleSlots ? generateMultipleSlots(formData) : EMPTY_SERIES),
    [createMultipleSlots, formData],
  )

  const handleCreateSlots = () => {
    if (!isTimeRangeValid) return
    if (createMultipleSlots) {
      if (seriesPreview.slots.length === 0) return
      createMultipleSlotsMutation.mutate(formData)
    } else {
      createSlotMutation.mutate(formData)
    }
  }

  const handleUpdateSlot = () => {
    if (!isTimeRangeValid) return
    if (editingSlot) {
      updateSlotMutation.mutate({ id: editingSlot.id, data: formData })
    }
  }

  const handleEditClick = (slot: InterviewSlotWithAssignments) => {
    setEditingSlot(slot)
    setFormData(slotFormFromSlot(slot))
    setIsEditDialogOpen(true)
  }

  const handleCloneClick = (slot: InterviewSlotWithAssignments) => {
    setCreateMultipleSlots(false)
    setFormData(slotFormFromSlot(slot))
    setIsCreateDialogOpen(true)
  }

  const handleDeleteClick = (slotId: string) => {
    if (
      confirm(
        'Are you sure you want to delete this interview slot? All assignments will be removed.',
      )
    ) {
      deleteSlotMutation.mutate(slotId)
    }
  }

  const handleAssignClick = (slot: InterviewSlotWithAssignments) => {
    setAssigningSlot(slot)
    setSelectedParticipationId('')
    setIsAssignDialogOpen(true)
  }

  const handleAssignStudent = () => {
    if (assigningSlot && selectedParticipationId) {
      assignStudentMutation.mutate({
        slotId: assigningSlot.id,
        participationId: selectedParticipationId,
      })
    }
  }

  const handleUnassignClick = (assignmentId: string, studentName: string) => {
    setUnassigningInfo({ assignmentId, studentName })
    setIsUnassignDialogOpen(true)
  }

  const handleConfirmUnassign = () => {
    if (unassigningInfo) {
      unassignStudentMutation.mutate(unassigningInfo.assignmentId)
    }
  }

  // Get list of already assigned participation IDs
  const assignedParticipationIds = new Set(
    slots?.flatMap((slot) => slot.assignments.map((a) => a.courseParticipationId)) || [],
  )

  // Filter unassigned students
  const unassignedStudents =
    participations?.participations.filter(
      (p) => !assignedParticipationIds.has(p.courseParticipationID),
    ) || []

  if (isLoading) {
    return (
      <div className='flex items-center justify-center h-64'>
        <div className='text-muted-foreground'>Loading...</div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className='container mx-auto py-8 px-4'>
        <ManagementPageHeader>Interview Schedule Management</ManagementPageHeader>
        <Alert variant='destructive' className='mt-4'>
          <AlertDescription>
            Failed to load interview slots. Please try again later.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className='container mx-auto py-8 px-4'>
      <ManagementPageHeader>Interview Schedule Management</ManagementPageHeader>

      <div className='flex justify-between items-center mb-6'>
        <p className='text-muted-foreground'>Create and manage interview time slots for students</p>
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button onClick={resetForm}>
              <Plus className='mr-2 h-4 w-4' />
              Create Slots
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Interview Slots</DialogTitle>
              <DialogDescription>
                Add one interview slot or divide the time range into multiple slots.
              </DialogDescription>
            </DialogHeader>
            <div className='space-y-4 py-4'>
              <div className='space-y-2'>
                <Label htmlFor='startTime'>Start Time</Label>
                <Input
                  id='startTime'
                  type='datetime-local'
                  value={formData.startTime}
                  onChange={(e) => setFormData({ ...formData, startTime: e.target.value })}
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='endTime'>End Time</Label>
                <Input
                  id='endTime'
                  type='time'
                  value={formData.endTime}
                  onChange={(e) => setFormData({ ...formData, endTime: e.target.value })}
                />
                {slotTimes && (
                  <p className='text-sm text-muted-foreground'>{formatResolvedEnd(slotTimes)}</p>
                )}
              </div>
              <div className='space-y-2'>
                <Label htmlFor='location'>Location (Optional)</Label>
                <Input
                  id='location'
                  placeholder='e.g., Room 101, Building A'
                  value={formData.location}
                  onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='capacity'>Capacity per Slot</Label>
                <Input
                  id='capacity'
                  type='number'
                  min='1'
                  value={formData.capacity}
                  onChange={(e) => {
                    const value = parseInt(e.target.value, 10)
                    setFormData({ ...formData, capacity: value > 0 ? value : 1 })
                  }}
                />
              </div>
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
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            durationMinutes: parseInt(e.target.value, 10) || 0,
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
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            breakMinutes: Math.max(0, parseInt(e.target.value, 10) || 0),
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
                      The time range fits more slots than the limit of {MAX_MULTIPLE_SLOTS}. Only
                      the first {MAX_MULTIPLE_SLOTS} will be created — shorten the range or create
                      the rest separately.
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
                  createSlotMutation.isPending ||
                  createMultipleSlotsMutation.isPending
                }
              >
                {createSlotMutation.isPending || createMultipleSlotsMutation.isPending
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
            <DialogTitle>Edit Interview Slot</DialogTitle>
            <DialogDescription>Update the interview slot details</DialogDescription>
          </DialogHeader>
          <div className='space-y-4 py-4'>
            <div className='space-y-2'>
              <Label htmlFor='edit_startTime'>Start Time</Label>
              <Input
                id='edit_startTime'
                type='datetime-local'
                value={formData.startTime}
                onChange={(e) => setFormData({ ...formData, startTime: e.target.value })}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='edit_endTime'>End Time</Label>
              <Input
                id='edit_endTime'
                type='time'
                value={formData.endTime}
                onChange={(e) => setFormData({ ...formData, endTime: e.target.value })}
              />
              {slotTimes && (
                <p className='text-sm text-muted-foreground'>{formatResolvedEnd(slotTimes)}</p>
              )}
            </div>
            <div className='space-y-2'>
              <Label htmlFor='edit_location'>Location (Optional)</Label>
              <Input
                id='edit_location'
                placeholder='e.g., Room 101, Building A'
                value={formData.location}
                onChange={(e) => setFormData({ ...formData, location: e.target.value })}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='edit_capacity'>Capacity</Label>
              <Input
                id='edit_capacity'
                type='number'
                min='1'
                value={formData.capacity}
                onChange={(e) => {
                  const value = parseInt(e.target.value, 10)
                  setFormData({ ...formData, capacity: value > 0 ? value : 1 })
                }}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setIsEditDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleUpdateSlot}
              disabled={!isTimeRangeValid || updateSlotMutation.isPending}
            >
              {updateSlotMutation.isPending ? 'Updating...' : 'Update'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Slots Table */}
      {slots && slots.length > 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>Interview Slots</CardTitle>
            <CardDescription>Manage all scheduled interview time slots</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date & Time</TableHead>
                  <TableHead>Location</TableHead>
                  <TableHead>Capacity</TableHead>
                  <TableHead>Assigned Students</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className='text-right'>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {slots.map((slot) => {
                  const isFull = (slot.assignedCount ?? 0) >= (slot.capacity ?? 1)
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
                            <MapPin className='h-4 w-4 text-muted-foreground shrink-0' />
                            {slot.location.match(/^https?:\/\//) ? (
                              <a
                                href={slot.location}
                                target='_blank'
                                rel='noopener noreferrer'
                                className='text-blue-600 hover:underline truncate min-w-0'
                              >
                                {slot.location}
                              </a>
                            ) : (
                              <span className='truncate min-w-0'>{slot.location}</span>
                            )}
                          </div>
                        ) : (
                          <span className='text-muted-foreground'>—</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className='flex items-center gap-2'>
                          <Users className='h-4 w-4 text-muted-foreground' />
                          <span>
                            {slot.assignedCount ?? 0} / {slot.capacity ?? 1}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        {slot.assignments && slot.assignments.length > 0 ? (
                          <div className='flex flex-wrap gap-1'>
                            {slot.assignments.map((assignment) => {
                              const studentName = assignment.student
                                ? `${assignment.student.firstName} ${assignment.student.lastName}`
                                : assignment.courseParticipationId
                              return (
                                <Badge
                                  key={assignment.id}
                                  variant='outline'
                                  className='cursor-pointer hover:bg-destructive hover:text-destructive-foreground transition-colors group pr-1'
                                  onClick={() => handleUnassignClick(assignment.id, studentName)}
                                  title={`Click to remove ${studentName}`}
                                >
                                  {studentName}
                                  <X className='ml-1 h-3 w-3 opacity-50 group-hover:opacity-100' />
                                </Badge>
                              )
                            })}
                          </div>
                        ) : (
                          <span className='text-muted-foreground text-sm'>No bookings yet</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {isPast ? (
                          <Badge variant='secondary'>Past</Badge>
                        ) : isFull ? (
                          <Badge variant='destructive'>Full</Badge>
                        ) : (
                          <Badge variant='secondary' className='bg-green-100 text-green-800'>
                            Available
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell className='text-right'>
                        <div className='flex justify-end gap-2'>
                          <Button
                            variant='ghost'
                            size='sm'
                            onClick={() => handleAssignClick(slot)}
                            disabled={isFull || isPast}
                            aria-label='Assign student'
                            title='Assign student to slot'
                          >
                            <UserPlus className='h-4 w-4' />
                          </Button>
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
                            onClick={() => handleDeleteClick(slot.id)}
                            disabled={deleteSlotMutation.isPending}
                            aria-label='Delete slot'
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
            No interview slots have been created yet. Click &quot;Create Slots&quot; to add your
            first slot.
          </AlertDescription>
        </Alert>
      )}

      {/* Assign Student Dialog */}
      <Dialog open={isAssignDialogOpen} onOpenChange={setIsAssignDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Assign Student to Interview Slot</DialogTitle>
            <DialogDescription>
              {assigningSlot && (
                <>
                  Manually assign a student to the slot on{' '}
                  {new Date(assigningSlot.startTime).toLocaleString('en-US', {
                    dateStyle: 'medium',
                    timeStyle: 'short',
                  })}
                  . This slot has {assigningSlot.assignedCount ?? 0}/{assigningSlot.capacity ?? 1}{' '}
                  students assigned.
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-4 py-4'>
            {unassignedStudents.length > 0 ? (
              <div className='space-y-2'>
                <Label htmlFor='student-select'>Select Student</Label>
                <Select value={selectedParticipationId} onValueChange={setSelectedParticipationId}>
                  <SelectTrigger id='student-select'>
                    <SelectValue placeholder='Choose a student...' />
                  </SelectTrigger>
                  <SelectContent>
                    {unassignedStudents.map((participation) => (
                      <SelectItem
                        key={participation.courseParticipationID}
                        value={participation.courseParticipationID}
                      >
                        {participation.student.firstName} {participation.student.lastName}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className='text-sm text-muted-foreground'>
                  {unassignedStudents.length} unassigned student
                  {unassignedStudents.length !== 1 && 's'} available
                </p>
              </div>
            ) : (
              <Alert>
                <AlertDescription>
                  All students have been assigned to interview slots. No unassigned students are
                  available.
                </AlertDescription>
              </Alert>
            )}
          </div>
          <DialogFooter>
            <Button
              variant='outline'
              onClick={() => {
                setIsAssignDialogOpen(false)
                setSelectedParticipationId('')
                setAssigningSlot(null)
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleAssignStudent}
              disabled={
                !selectedParticipationId ||
                assignStudentMutation.isPending ||
                unassignedStudents.length === 0
              }
            >
              {assignStudentMutation.isPending ? 'Assigning...' : 'Assign Student'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Unassign Confirmation Dialog */}
      <Dialog open={isUnassignDialogOpen} onOpenChange={setIsUnassignDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Unassign Student</DialogTitle>
            <DialogDescription>
              {unassigningInfo && (
                <>
                  Are you sure you want to unassign <strong>{unassigningInfo.studentName}</strong>{' '}
                  from this interview slot? The student will need to book a new slot if needed.
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant='outline'
              onClick={() => {
                setIsUnassignDialogOpen(false)
                setUnassigningInfo(null)
              }}
            >
              Cancel
            </Button>
            <Button
              variant='destructive'
              onClick={handleConfirmUnassign}
              disabled={unassignStudentMutation.isPending}
            >
              {unassignStudentMutation.isPending ? 'Unassigning...' : 'Unassign'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
