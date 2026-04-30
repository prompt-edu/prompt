import { useToast } from '@tumaet/prompt-ui-components'
import { useModifyCoursePhase } from '@tumaet/prompt-shared-state'
import { UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

export const useHideMailingWarning = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { toast } = useToast()

  const { mutate: mutateCoursePhase } = useModifyCoursePhase(
    () => {
      toast({
        title: 'The warning has been hidden',
      })
    },
    () => {
      toast({
        title: 'Error on hiding the warning',
        description: 'Please try again later',
        variant: 'destructive',
      })
    },
  )

  const hideMailingWarning = () => {
    const updatedCoursePhase: UpdateCoursePhase = {
      id: phaseId ?? '',
      restrictedData: {
        hideMailingWarning: true,
      },
    }
    mutateCoursePhase(updatedCoursePhase)
  }

  return { hideMailingWarning }
}
