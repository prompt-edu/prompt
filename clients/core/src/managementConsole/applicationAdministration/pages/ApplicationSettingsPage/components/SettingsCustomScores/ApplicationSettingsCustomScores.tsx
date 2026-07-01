import { updateCoursePhase } from '@core/network/mutations/updateCoursePhase'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import { Card, CardContent, Label, Switch } from '@tumaet/prompt-ui-components'
import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { ApplicationMetaData } from '../../../../interfaces/applicationMetaData'

interface Props {
  initialData: ApplicationMetaData
}

export function ApplicationSettingsCustomScores({ initialData }: Props) {
  const queryClient = useQueryClient()
  const { phaseId } = useParams<{ phaseId: string }>()

  const [useCustomScores, setUseCustomScores] = useState(false)

  useEffect(() => {
    setUseCustomScores(initialData?.useCustomScores ?? false)
  }, [initialData])

  const {
    mutate: mutatePhase,
    isPending,
    isError,
  } = useMutation({
    mutationFn: (coursePhase: UpdateCoursePhase) => updateCoursePhase(coursePhase),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['course_phase', phaseId] }),
  })

  const handleToggle = (checked: boolean) => {
    setUseCustomScores(checked)

    const updatedPhase: UpdateCoursePhase = {
      id: phaseId ?? '',
      restrictedData: {
        useCustomScores: checked,
      },
    }

    mutatePhase(updatedPhase)
  }

  return (
    <Card className='w-full'>
      <CardContent>
        {isPending ? (
          <div className='flex items-center gap-2'>
            <span>Saving custom score setting...</span>
          </div>
        ) : isError ? (
          <div className='text-sm text-red-600'>Error saving setting</div>
        ) : (
          <div className='mb-2 mt-5'>
            <h3 className='text-lg font-semibold'>Custom Scores</h3>
            <p className='text-sm text-muted-foreground'>
              Enable custom scoring options for application assessment.
            </p>

            <div className='grid grid-cols-4 items-center gap-4 py-4'>
              <Label htmlFor='customScores' className='text-right'>
                Enable Custom Scores
              </Label>
              <div className='col-span-3 flex items-center gap-4'>
                <Switch
                  id='customScores'
                  checked={useCustomScores}
                  onCheckedChange={handleToggle}
                />
                <p className='text-sm text-muted-foreground'>
                  Enable custom scoring options for application assessment.
                </p>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default ApplicationSettingsCustomScores
