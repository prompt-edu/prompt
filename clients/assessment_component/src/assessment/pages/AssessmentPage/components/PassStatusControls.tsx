import {
  getPermissionString,
  PassStatus,
  Role,
  useAuthStore,
  useCourseStore,
  useUpdateCoursePhaseParticipation,
} from '@tumaet/prompt-shared-state'
import { Button } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

import { useGetCoursePhaseParticipations } from '../../hooks/useGetCoursePhaseParticipations'

interface PassStatusControlsProps {
  courseParticipationID?: string
  disabled?: boolean
}

export const PassStatusControls = ({
  courseParticipationID,
  disabled = false,
}: PassStatusControlsProps) => {
  const { courseId } = useParams<{ courseId: string }>()
  const { permissions } = useAuthStore()
  const { courses } = useCourseStore()
  const { data: participations } = useGetCoursePhaseParticipations()
  const participant = participations.find(
    (participation) => participation.courseParticipationID === courseParticipationID,
  )

  const { mutate: updateParticipation, isPending } = useUpdateCoursePhaseParticipation()

  const course = courses.find((c) => c.id === courseId)
  const canSetPassStatus =
    permissions.includes(getPermissionString(Role.PROMPT_ADMIN)) ||
    permissions.includes(getPermissionString(Role.PROMPT_LECTURER)) ||
    permissions.includes(
      getPermissionString(Role.COURSE_LECTURER, course?.name, course?.semesterTag),
    )

  if (!canSetPassStatus || !participant || !courseParticipationID) {
    return null
  }

  const handleUpdatePassStatus = (status: PassStatus) => {
    if (disabled) return

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
        disabled={disabled || isPending || participant.passStatus === PassStatus.FAILED}
        className='border-red-500 text-red-600 hover:bg-red-50 hover:text-red-700'
        onClick={() => handleUpdatePassStatus(PassStatus.FAILED)}
      >
        Set Failed
      </Button>

      <Button
        variant='default'
        disabled={disabled || isPending || participant.passStatus === PassStatus.PASSED}
        className='bg-green-500 hover:bg-green-600 text-white'
        onClick={() => handleUpdatePassStatus(PassStatus.PASSED)}
      >
        Set Passed
      </Button>
    </div>
  )
}
