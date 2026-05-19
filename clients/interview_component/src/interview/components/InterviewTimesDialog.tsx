import { useEffect, useRef, useState } from 'react'
import { Clock, Plus, Trash } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Button,
  Separator,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  ScrollArea,
} from '@tumaet/prompt-ui-components'
import { useCoursePhaseStore } from '../zustand/useCoursePhaseStore'
import { useParticipationStore } from '../zustand/useParticipationStore'
import { InterviewSlot } from '../interfaces/InterviewSlots'
import { useUpdateCoursePhaseMetaData } from '@tumaet/prompt-shared-state'

export const InterviewTimesDialog = () => {
  const { coursePhase } = useCoursePhaseStore()
  const [isOpen, setIsOpen] = useState(false)
  const { participations } = useParticipationStore()
  const [interviewSlots, setInterviewSlots] = useState<InterviewSlot[]>([])
  const scrollRef = useRef<HTMLDivElement>(null)

  const { mutate } = useUpdateCoursePhaseMetaData()

  const scrollToBottom = () => {
    scrollRef.current?.scrollIntoView(false)
  }

  useEffect(() => {
    if (isOpen) {
      setInterviewSlots(coursePhase?.restrictedData?.interviewSlots ?? [])
    }
  }, [isOpen, coursePhase])

  const addSlot = () => {
    setInterviewSlots((prevSlots) => {
      const prevStartTime = prevSlots.length > 0 ? prevSlots[prevSlots.length - 1].startTime : ''
      const prevEndTime = prevSlots.length > 0 ? prevSlots[prevSlots.length - 1].endTime : ''

      let newEndTime = ''

      if (prevStartTime && prevEndTime) {
        // Convert existing times to milliseconds
        const startTimeMs = new Date(`1970-01-01T${prevStartTime}`).getTime()
        const endTimeMs = new Date(`1970-01-01T${prevEndTime}`).getTime()

        // Calculate the difference and apply it to prevEndTime
        const diff = endTimeMs - startTimeMs
        const newTimeMs = endTimeMs + diff

        // Convert the resulting milliseconds back to "HH:MM"
        const newDate = new Date(newTimeMs)
        const hours = String(newDate.getHours()).padStart(2, '0')
        const minutes = String(newDate.getMinutes()).padStart(2, '0')
        newEndTime = `${hours}:${minutes}`
      }

      // Return new array of slots with the newly added slot
      return [
        ...prevSlots,
        {
          id: Date.now().toString(),
          startTime: prevEndTime,
          endTime: newEndTime,
        },
      ]
    })

    requestAnimationFrame(() => {
      scrollToBottom()
    })
  }

  const removeSlot = (id: string) => {
    setInterviewSlots(interviewSlots.filter((slot) => slot.id !== id))
  }

  const updateSlot = (id: string, field: keyof InterviewSlot, value: string) => {
    setInterviewSlots(
      interviewSlots.map((slot) => (slot.id === id ? { ...slot, [field]: value } : slot)),
    )
  }

  const saveSlots = () => {
    if (coursePhase) {
      mutate({
        id: coursePhase.id,
        restrictedData: {
          interviewSlots: interviewSlots,
        },
      })
    }
    setIsOpen(false)
  }

  return (
    <>
      <Button variant='outline' onClick={() => setIsOpen(true)} className='gap-2'>
        <Clock className='h-4 w-4' />
        Set Interview Times
      </Button>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className='max-w-2xl'>
          <DialogHeader>
            <DialogTitle>Manage Interview Times</DialogTitle>
            <DialogDescription>
              Set the times when the students have their booked interview.
            </DialogDescription>
          </DialogHeader>
          <Separator />

          <div className='space-y-4 overflow-hidden'>
            <ScrollArea className='h-[300px]'>
              {interviewSlots.map((slot, index) => (
                <div key={slot.id} className='flex flex-row items-center space-x-2 p-2 pr-4'>
                  <span className='font-medium w-1/12'>{index + 1}.</span>
                  <Input
                    type='time'
                    value={slot.startTime || ''}
                    onChange={(e) => updateSlot(slot.id, 'startTime', e.target.value)}
                    placeholder='Start Time'
                    className='w-1/6'
                  />
                  <Input
                    type='time'
                    value={slot.endTime || ''}
                    onChange={(e) => updateSlot(slot.id, 'endTime', e.target.value)}
                    placeholder='End Time'
                    className='w-1/6'
                  />
                  <Select
                    value={slot.courseParticipationID}
                    onValueChange={(value) => updateSlot(slot.id, 'courseParticipationID', value)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder='Select participant' />
                    </SelectTrigger>
                    <SelectContent>
                      {participations.map((participant) => (
                        <SelectItem
                          key={participant.courseParticipationID}
                          value={participant.courseParticipationID}
                        >
                          {participant.student.firstName} {participant.student.lastName}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <Button variant='ghost' size='icon' onClick={() => removeSlot(slot.id)}>
                    <Trash className='h-4 w-4' />
                  </Button>
                </div>
              ))}
            </ScrollArea>
          </div>

          <Button onClick={addSlot} variant='outline' className='w-full pt-2'>
            <Plus className='h-4 w-4 mr-2' />
            Add Slot
          </Button>

          <DialogFooter className='sm:justify-between'>
            <Button variant='outline' onClick={() => setIsOpen(false)}>
              Cancel
            </Button>
            <Button onClick={saveSlots}>Save Changes</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
