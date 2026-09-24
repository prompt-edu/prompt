import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import { Button, Input, Label, Skeleton, useToast } from '@tumaet/prompt-ui-components'
import { Save } from 'lucide-react'
import { useState } from 'react'
import { updateSetupConfig } from '../../network/mutations/updateSetupConfig'
import { getSetupConfig } from '../../network/queries/getSetupConfig'
import { describeError } from '../../utils/describeError'
import { SectionError } from '../SectionError'
import { ConfigurationSection } from './ConfigurationSection'

interface Props {
  courseId: string
  coursePhaseID: string
}

export const SemesterTagSection = ({ courseId, coursePhaseID }: Props) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const { courses } = useCourseStore()

  const course = courses.find((c) => c.id === courseId)

  // Null means "not edited yet", so the field shows the stored config. Seeding state
  // from the query instead would overwrite what the lecturer is typing on any refetch.
  const [editedTag, setEditedTag] = useState<string | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['setup-config', coursePhaseID],
    queryFn: () => getSetupConfig(coursePhaseID),
  })

  // The course's tag is only ever a suggestion. Showing it as the field's value made an
  // unsaved phase look configured, while the server kept resolving {{semesterTag}} to
  // nothing and naming a team's GitLab group "-ios-team-1".
  const storedTag = data?.semesterTag ?? ''
  const suggestedTag = course?.semesterTag ?? ''
  const semesterTag = editedTag ?? storedTag
  const isUnsaved = semesterTag.trim() !== storedTag

  const { mutate: save, isPending } = useMutation({
    mutationFn: () => updateSetupConfig(coursePhaseID, { semesterTag: semesterTag.trim() }),
    onSuccess: () => {
      // Hand the field back to the query: what it refetches is now what was saved.
      setEditedTag(null)
      queryClient.invalidateQueries({ queryKey: ['setup-config', coursePhaseID] })
      toast({ title: 'Semester tag saved' })
    },
    onError: (err: unknown) => {
      toast({
        title: 'Failed to save the semester tag',
        description: describeError(err),
        variant: 'destructive',
      })
    },
  })

  return (
    <ConfigurationSection
      id='semester-tag'
      step={1}
      title='Semester tag'
      description={
        <>
          Used as <code>{`{{semesterTag}}`}</code> in resource name templates, so each semester gets
          its own resources.
        </>
      }
    >
      {isLoading ? (
        <Skeleton className='h-10 w-full max-w-md' />
      ) : isError ? (
        <SectionError message='Failed to load the semester tag.' onRetry={() => refetch()} />
      ) : (
        <div className='max-w-md space-y-2'>
          <Label htmlFor='semesterTag'>Semester tag</Label>
          <div className='flex gap-2'>
            <Input
              id='semesterTag'
              value={semesterTag}
              onChange={(event) => setEditedTag(event.target.value)}
              placeholder={suggestedTag || 'ios26'}
            />
            <Button onClick={() => save()} disabled={isPending || !isUnsaved}>
              <Save className='mr-2 h-4 w-4' />
              {isPending ? 'Saving…' : 'Save'}
            </Button>
          </div>
          {storedTag === '' && (
            <div className='rounded-lg border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900'>
              <p>
                No semester tag is saved for this phase, so provisioning is refused for any template
                that uses <code>{`{{semesterTag}}`}</code>.
              </p>
              {suggestedTag && (
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  className='mt-2'
                  onClick={() => setEditedTag(suggestedTag)}
                >
                  Use the course tag ({suggestedTag})
                </Button>
              )}
            </div>
          )}
          {isUnsaved && semesterTag.trim() !== '' && (
            <p className='text-xs text-amber-700'>Not saved yet. Press Save to apply it.</p>
          )}
        </div>
      )}
    </ConfigurationSection>
  )
}
