import { updateCoursePhase } from '@core/network/mutations/updateCoursePhase'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import {
  Button,
  Card,
  CardContent,
  DescriptionMinimalTiptapEditor,
  TooltipProvider,
} from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { ApplicationMetaData } from '../../../../interfaces/applicationMetaData'

interface ApplicationSettingsWelcomeTextProps {
  initialData: ApplicationMetaData
}

// An emptied rich text editor still emits markup like `<p></p>`, which must count
// as no welcome text at all.
const normalizeHtml = (html: string) => (html.replace(/<[^>]*>/g, '').trim() === '' ? '' : html)

export function ApplicationSettingsWelcomeText({
  initialData,
}: ApplicationSettingsWelcomeTextProps) {
  const queryClient = useQueryClient()
  const { phaseId } = useParams<{ phaseId: string }>()

  const [welcomeText, setWelcomeText] = useState(initialData.welcomeText ?? '')
  const [savedWelcomeText, setSavedWelcomeText] = useState(initialData.welcomeText ?? '')

  // The metadata is parsed asynchronously, so the first render can precede it.
  useEffect(() => {
    setWelcomeText(initialData.welcomeText ?? '')
    setSavedWelcomeText(initialData.welcomeText ?? '')
  }, [initialData])

  const normalizedWelcomeText = normalizeHtml(welcomeText)
  const hasChanges = normalizedWelcomeText !== savedWelcomeText

  const {
    mutate: mutatePhase,
    isPending,
    isError,
  } = useMutation({
    mutationFn: (coursePhase: UpdateCoursePhase) => updateCoursePhase(coursePhase),
    onSuccess: (_, coursePhase) => {
      setSavedWelcomeText((coursePhase.restrictedData?.welcomeText as string | null) ?? '')
      queryClient.invalidateQueries({ queryKey: ['course_phase', phaseId] })
    },
  })

  const handleSave = () => {
    const updatedPhase: UpdateCoursePhase = {
      id: phaseId ?? '',
      restrictedData: {
        // null clears the key, so an emptied editor does not leave a blank string behind
        welcomeText: normalizedWelcomeText === '' ? null : normalizedWelcomeText,
      },
    }

    mutatePhase(updatedPhase)
  }

  return (
    <Card className='w-full' data-testid='application-welcome-text'>
      <CardContent>
        <div className='mb-2 mt-5 space-y-4'>
          <div>
            <h3 className='text-lg font-semibold'>Welcome Text</h3>
            <p className='text-sm text-muted-foreground'>
              An optional text shown to applicants above the application form, for example to help
              them confirm they are applying for the right course.
            </p>
          </div>

          <TooltipProvider>
            <DescriptionMinimalTiptapEditor
              // The editor reads its content once, on create, so it has to be
              // remounted whenever the stored text changes underneath it.
              key={savedWelcomeText}
              value={welcomeText}
              onChange={(value) => setWelcomeText(typeof value === 'string' ? value : '')}
              className='w-full'
              editorContentClassName='p-3'
              output='html'
              placeholder='Type your welcome text here...'
              autofocus={false}
              editable={true}
              editorClassName='focus:outline-hidden'
            />
          </TooltipProvider>

          {isError && <p className='text-sm text-red-600'>Error saving welcome text</p>}

          <div className='flex justify-end'>
            <Button
              data-testid='application-welcome-text-save'
              onClick={handleSave}
              disabled={isPending || !hasChanges}
            >
              {isPending && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
              Save Welcome Text
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
