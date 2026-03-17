import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { interviewAxiosInstance } from '../../network/interviewServerConfig'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Badge,
  Alert,
  AlertDescription,
  AlertTitle,
  cn,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Clock, MapPin, Users, AlertCircle, TriangleAlert } from 'lucide-react'
import { format } from 'date-fns'

interface InterviewSlot {
  id: string
  coursePhaseId: string
  startTime: string
  endTime: string
  location: string | null
  capacity: number
  assignedCount: number
  createdAt: string
  updatedAt: string
}

interface InterviewAssignment {
  id: string
  interview_slot_id: string
  course_participation_id: string
  assigned_at: string
  slot_details?: InterviewSlot
}

export const StudentInterviewPage = () => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')
  const queryClient = useQueryClient()
  const [selectedSlotId, setSelectedSlotId] = useState<string | null>(null)
  const { toast } = useToast()

  const { data: slots, isLoading: slotsLoading } = useQuery<InterviewSlot[]>({
    queryKey: ['interviewSlotsWithAssignments', phaseId],
    queryFn: async () => {
      const response = await interviewAxiosInstance.get(
        `interview/api/course_phase/${phaseId}/interview-slots`,
      )
      return response.data
    },
    enabled: !!phaseId,
  })

  const { data: myAssignment, isLoading: assignmentLoading } = useQuery<InterviewAssignment | null>(
    {
      queryKey: ['myInterviewAssignment', phaseId],
      queryFn: async () => {
        try {
          const response = await interviewAxiosInstance.get(
            `interview/api/course_phase/${phaseId}/interview-assignments/my-assignment`,
          )
          return response.data
        } catch (error: any) {
          // 404 means no assignment exists, which is valid
          if (error.response?.status === 404) {
            return null
          }
          throw error
        }
      },
      enabled: !!phaseId,
      retry: false, // Don't retry - 404 is a valid response for users without assignments
    },
  )

  const bookSlotMutation = useMutation({
    mutationFn: async (slotId: string) => {
      const response = await interviewAxiosInstance.post(
        `interview/api/course_phase/${phaseId}/interview-assignments`,
        { interview_slot_id: slotId },
      )
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['myInterviewAssignment', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['interviewSlotsWithAssignments', phaseId] })
      setSelectedSlotId(null)
      toast({
        title: 'Slot booked successfully',
        description: 'Your interview slot has been confirmed.',
      })
    },
    onError: (error: any) => {
      toast({
        title: 'Booking failed',
        description:
          error?.response?.data?.error || 'Failed to book interview slot. Please try again.',
        variant: 'destructive',
      })
    },
  })

  const cancelBookingMutation = useMutation({
    mutationFn: async (assignmentId: string) => {
      await interviewAxiosInstance.delete(
        `interview/api/course_phase/${phaseId}/interview-assignments/${assignmentId}`,
      )
    },
    onSuccess: async () => {
      queryClient.setQueryData(['myInterviewAssignment', phaseId], null)
      await Promise.all([
        queryClient.refetchQueries({ queryKey: ['myInterviewAssignment', phaseId] }),
        queryClient.refetchQueries({ queryKey: ['interviewSlotsWithAssignments', phaseId] }),
      ])
      toast({
        title: 'Booking cancelled',
        description: 'Your interview slot booking has been cancelled.',
      })
    },
    onError: (error: any) => {
      toast({
        title: 'Cancellation failed',
        description: error?.response?.data?.error || 'Failed to cancel booking. Please try again.',
        variant: 'destructive',
      })
    },
  })

  const handleBookSlot = () => {
    if (selectedSlotId) {
      bookSlotMutation.mutate(selectedSlotId)
    }
  }

  const handleCancelBooking = () => {
    if (myAssignment?.id) {
      cancelBookingMutation.mutate(myAssignment.id)
    }
  }

  const isSlotFull = (slot: InterviewSlot) => slot.assignedCount >= slot.capacity
  const isSlotPast = (slot: InterviewSlot) => new Date(slot.startTime) < new Date()

  if (slotsLoading || assignmentLoading) {
    return (
      <div className='flex items-center justify-center h-64'>
        <div className='text-muted-foreground'>Loading interview slots...</div>
      </div>
    )
  }

  return (
    <div className='container mx-auto pb-8 px-4 max-w-6xl'>
      {/* Non-student Disclaimer */}
      {!isStudent && (
        <Alert className='mb-6'>
          <TriangleAlert className='h-4 w-4' />
          <AlertTitle>You are not a student of this course.</AlertTitle>
          <AlertDescription>
            You can view all interview slots below, but booking is disabled because you are not a
            student of this course. For configuring interview slots, please refer to the Schedule
            Management page.
          </AlertDescription>
        </Alert>
      )}

      <div className={!isStudent ? 'mb-8' : 'mb-8 mt-8'}>
        <h1 className='text-3xl font-bold mb-2'>Interview Scheduling</h1>
        <p className='text-muted-foreground'>
          {myAssignment
            ? 'Your booked interview slot is highlighted below'
            : 'Select an available time slot for your interview'}
        </p>
      </div>

      {/* Slots Grid */}
      {slots && slots.length > 0 ? (
        <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-3'>
          {slots.map((slot) => {
            const isFull = isSlotFull(slot)
            const isPast = isSlotPast(slot)
            const isBooked = myAssignment?.interview_slot_id === slot.id
            const isSelected = selectedSlotId === slot.id
            const hasBooking = !!myAssignment
            const isDisabled = !isBooked && (isFull || isPast || !isStudent || hasBooking)

            return (
              <Card
                key={slot.id}
                className={cn(
                  'transition-all',
                  isBooked && 'ring-2 ring-green-500 bg-green-50 dark:bg-green-950/20',
                  !isBooked && isSelected && 'ring-2 ring-primary',
                  !isBooked &&
                    isStudent &&
                    !isFull &&
                    !isPast &&
                    !hasBooking &&
                    'cursor-pointer hover:shadow-md',
                  isDisabled && 'opacity-50 cursor-not-allowed',
                )}
                onClick={() =>
                  !isBooked &&
                  isStudent &&
                  !isFull &&
                  !isPast &&
                  !hasBooking &&
                  setSelectedSlotId(slot.id)
                }
              >
                <CardHeader>
                  <div className='flex justify-between items-start'>
                    <CardTitle className='text-lg'>
                      {format(new Date(slot.startTime), 'EEE, MMM d')}
                    </CardTitle>
                    {isBooked ? (
                      <Badge className='bg-green-600'>Booked</Badge>
                    ) : isFull ? (
                      <Badge variant='destructive'>Full</Badge>
                    ) : isPast ? (
                      <Badge variant='secondary'>Past</Badge>
                    ) : (
                      <Badge variant='secondary' className='bg-green-100 text-green-800'>
                        Available
                      </Badge>
                    )}
                  </div>
                  <CardDescription>
                    <div className='flex items-center gap-1 mt-1'>
                      <Clock className='h-3 w-3' />
                      <span className='text-sm'>
                        {format(new Date(slot.startTime), 'HH:mm')} -{' '}
                        {format(new Date(slot.endTime), 'HH:mm')}
                      </span>
                    </div>
                  </CardDescription>
                </CardHeader>
                <CardContent className='space-y-3'>
                  {slot.location && (
                    <div className='flex items-center gap-2 text-sm text-muted-foreground'>
                      <MapPin className='h-4 w-4 flex-shrink-0' />
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
                  )}
                  <div className='flex items-center gap-2 text-sm text-muted-foreground'>
                    <Users className='h-4 w-4' />
                    <span>
                      {slot.assignedCount} / {slot.capacity} booked
                    </span>
                  </div>

                  {/* Action Button */}
                  {isStudent && isBooked && (
                    <Button
                      variant='destructive'
                      size='sm'
                      onClick={(e) => {
                        e.stopPropagation()
                        handleCancelBooking()
                      }}
                      disabled={cancelBookingMutation.isPending}
                      className='w-full mt-2'
                    >
                      {cancelBookingMutation.isPending ? 'Canceling...' : 'Cancel Booking'}
                    </Button>
                  )}
                  {isStudent && !isBooked && !hasBooking && isSelected && !isFull && !isPast && (
                    <Button
                      size='sm'
                      onClick={(e) => {
                        e.stopPropagation()
                        handleBookSlot()
                      }}
                      disabled={bookSlotMutation.isPending}
                      className='w-full mt-2'
                    >
                      {bookSlotMutation.isPending ? 'Booking...' : 'Confirm Booking'}
                    </Button>
                  )}
                </CardContent>
              </Card>
            )
          })}
        </div>
      ) : (
        <Alert>
          <AlertCircle className='h-4 w-4' />
          <AlertDescription className='ml-2'>
            No interview slots are currently available. Please check back later.
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}
