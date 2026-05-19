import { Button } from '@tumaet/prompt-ui-components'
import { PassStatus, Role, useAuthStore } from '@tumaet/prompt-shared-state'

import { useUpdateCoursePhaseParticipation } from '@tumaet/prompt-shared-state'

import { useParticipationStore } from '../../../zustand/useParticipationStore'

interface PassStatusControlsProps {
  courseParticipationID?: string
}

export const PassStatusControls = ({ courseParticipationID }: PassStatusControlsProps) => {
  const { permissions } = useAuthStore()
  const { participations } = useParticipationStore()
  const participant = participations.find(
    (participation) => participation.courseParticipationID === courseParticipationID,
  )

  const { mutate: updateParticipation, isPending } = useUpdateCoursePhaseParticipation()

  const isLecturerOrHigher = permissions.some(
    (permission) =>
      permission === Role.PROMPT_ADMIN ||
      permission === Role.PROMPT_LECTURER ||
      permission.endsWith(Role.COURSE_LECTURER),
  )

  if (!isLecturerOrHigher || !participant || !courseParticipationID) {
    return null
  }

  const handleUpdatePassStatus = (status: PassStatus) => {
    updateParticipation({
      coursePhaseID: participant.coursePhaseID,
      courseParticipationID,
      passStatus: status,
      restrictedData: participant.restrictedData ?? {},
      studentReadableData: participant.studentReadableData ?? {},
    })
  }

  return (
    <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between border-t pt-4'>
      <Button
        variant='outline'
        disabled={isPending || participant.passStatus === PassStatus.FAILED}
        className='border-red-500 text-red-600 hover:bg-red-50 hover:text-red-700'
        onClick={() => handleUpdatePassStatus(PassStatus.FAILED)}
      >
        Set Failed
      </Button>

      <Button
        variant='default'
        disabled={isPending || participant.passStatus === PassStatus.PASSED}
        className='bg-green-500 hover:bg-green-600 text-white'
        onClick={() => handleUpdatePassStatus(PassStatus.PASSED)}
      >
        Set Passed
      </Button>
    </div>
  )
}
