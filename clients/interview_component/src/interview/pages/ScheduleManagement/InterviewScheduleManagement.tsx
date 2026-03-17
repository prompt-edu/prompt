import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { interviewAxiosInstance } from '../../network/interviewServerConfig'
import { getCoursePhaseParticipations } from '@/network/queries/getCoursePhaseParticipations'
import { CoursePhaseParticipationsWithResolution } from '@tumaet/prompt-shared-state'
import {
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
  DialogTrigger,
  Input,
  Label,
  Badge,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  ManagementPageHeader,
  Alert,
  AlertDescription,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Calendar, Clock, MapPin, Users, Plus, Pencil, Trash2, X, UserPlus } from 'lucide-react'
import { format } from 'date-fns'
import { InterviewSlotWithAssignments } from '../../interfaces/InterviewSlots'

interface SlotFormData {
  startTime: string
  endTime: string
  location: string
  capacity: number
}

export const InterviewScheduleManagement = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
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
  const [formData, setFormData] = useState<SlotFormData>({
    startTime: '',
    endTime: '',
    location: '',
    capacity: 1,
  })

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
        {
          startTime: new Date(data.startTime).toISOString(),
          endTime: new Date(data.endTime).toISOString(),
          location: data.location || null,
          capacity: data.capacity,
        },
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
    onError: (error: any) => {
      toast({
        title: 'Creation failed',
        description: error?.response?.data?.error || 'Failed to create interview slot.',
        variant: 'destructive',
      })
    },
  })

  // Update slot mutation
  const updateSlotMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: SlotFormData }) => {
      const response = await interviewAxiosInstance.put(
        `interview/api/course_phase/${phaseId}/interview-slots/${id}`,
        {
          startTime: new Date(data.startTime).toISOString(),
          endTime: new Date(data.endTime).toISOString(),
          location: data.location || null,
          capacity: data.capacity,
        },
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
    onError: (error: any) => {
      toast({
        title: 'Update failed',
        description: error?.response?.data?.error || 'Failed to update interview slot.',
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
    onError: (error: any) => {
      toast({
        title: 'Deletion failed',
        description: error?.response?.data?.error || 'Failed to delete interview slot.',
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
    onError: (error: any) => {
      toast({
        title: 'Assignment failed',
        description: error.response?.data?.error || 'Failed to assign student to interview slot.',
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
    onError: (error: any) => {
      toast({
        title: 'Unassignment failed',
        description: error?.response?.data?.error || 'Failed to unassign student.',
        variant: 'destructive',
      })
    },
  })

  const resetForm = () => {
    setFormData({
      startTime: '',
      endTime: '',
      location: '',
      capacity: 1,
    })
  }

  const isTimeRangeValid =
    !!formData.startTime &&
    !!formData.endTime &&
    new Date(formData.startTime) < new Date(formData.endTime)

  const handleCreateSlot = () => {
    if (!isTimeRangeValid) return
    createSlotMutation.mutate(formData)
  }

  const handleUpdateSlot = () => {
    if (!isTimeRangeValid) return
    if (editingSlot) {
      updateSlotMutation.mutate({ id: editingSlot.id, data: formData })
    }
  }

  const handleEditClick = (slot: InterviewSlotWithAssignments) => {
    setEditingSlot(slot)
    setFormData({
      startTime: format(new Date(slot.startTime), "yyyy-MM-dd'T'HH:mm"),
      endTime: format(new Date(slot.endTime), "yyyy-MM-dd'T'HH:mm"),
      location: slot.location || '',
      capacity: slot.capacity ?? 1,
    })
    setIsEditDialogOpen(true)
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
              Create Slot
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Interview Slot</DialogTitle>
              <DialogDescription>Add a new time slot for student interviews</DialogDescription>
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
                  type='datetime-local'
                  value={formData.endTime}
                  onChange={(e) => setFormData({ ...formData, endTime: e.target.value })}
                />
                {formData.startTime && formData.endTime && !isTimeRangeValid && (
                  <p className='text-sm text-destructive'>End time must be after start time.</p>
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
                <Label htmlFor='capacity'>Capacity</Label>
                <Input
                  id='capacity'
                  type='number'
                  min='1'
                  value={formData.capacity}
                  onChange={(e) => {
                    const value = parseInt(e.target.value)
                    setFormData({ ...formData, capacity: value > 0 ? value : 1 })
                  }}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant='outline' onClick={() => setIsCreateDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreateSlot}
                disabled={!isTimeRangeValid || createSlotMutation.isPending}
              >
                {createSlotMutation.isPending ? 'Creating...' : 'Create'}
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
                type='datetime-local'
                value={formData.endTime}
                onChange={(e) => setFormData({ ...formData, endTime: e.target.value })}
              />
              {formData.startTime && formData.endTime && !isTimeRangeValid && (
                <p className='text-sm text-destructive'>End time must be after start time.</p>
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
                  const value = parseInt(e.target.value)
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
                            <MapPin className='h-4 w-4 text-muted-foreground flex-shrink-0' />
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
            No interview slots have been created yet. Click &quot;Create Slot&quot; to add your
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
