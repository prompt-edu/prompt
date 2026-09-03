import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  Button,
  ErrorPage,
  Input,
  Label,
  LoadingPage,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Info, Save, Settings } from 'lucide-react'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { updateSetupConfig } from '../network/mutations/updateSetupConfig'
import { getSetupConfig } from '../network/queries/getSetupConfig'
import { describeError } from '../utils/describeError'

export const SetupConfigPage = () => {
  const { courseId, phaseId: coursePhaseID } = useParams<{
    courseId: string
    phaseId: string
  }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const { courses } = useCourseStore()

  const course = courses.find((c) => c.id === courseId)

  // Null means "not edited yet", so the field shows the stored config. Seeding state
  // from the query instead would overwrite what the lecturer is typing on any refetch,
  // and React Query refetches on window focus.
  const [editedTag, setEditedTag] = useState<string | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['setup-config', coursePhaseID],
    queryFn: () => getSetupConfig(coursePhaseID!),
    enabled: !!coursePhaseID,
  })

  // The parent course's tag is the default for a phase that has none of its own.
  const storedTag = data?.semesterTag || (course?.semesterTag ?? '')
  const semesterTag = editedTag ?? storedTag

  const { mutate: save, isPending } = useMutation({
    mutationFn: () =>
      updateSetupConfig(coursePhaseID!, {
        semesterTag: semesterTag.trim(),
      }),
    onSuccess: () => {
      // Hand the field back to the query: what it refetches is now what was saved.
      setEditedTag(null)
      queryClient.invalidateQueries({ queryKey: ['setup-config', coursePhaseID] })
      toast({ title: 'Setup configuration saved' })
    },
    onError: (err: unknown) => {
      toast({
        title: 'Failed to save setup configuration',
        description: describeError(err),
        variant: 'destructive',
      })
    },
  })

  if (isLoading) {
    return <LoadingPage />
  }
  if (isError) {
    return <ErrorPage description='Failed to load setup configuration.' onRetry={() => refetch()} />
  }

  return (
    <div className='max-w-3xl space-y-4 p-6'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <Settings className='h-5 w-5 text-blue-500' />
          <h1 className='text-xl font-semibold'>Setup</h1>
        </div>
        <Button onClick={() => save()} disabled={isPending}>
          <Save className='mr-2 h-4 w-4' />
          {isPending ? 'Saving…' : 'Save'}
        </Button>
      </div>

      <Alert>
        <Info className='h-4 w-4' />
        <AlertDescription>
          Teams and participation data come from upstream phases wired in the course&apos;s phase
          configurator. Open the management console and connect a Team Allocation (or Self Team
          Allocation) phase to this Infrastructure Setup phase if you haven&apos;t already.
        </AlertDescription>
      </Alert>

      <div className='space-y-1'>
        <Label htmlFor='semesterTag'>Semester tag</Label>
        <Input
          id='semesterTag'
          value={semesterTag}
          onChange={(event) => setEditedTag(event.target.value)}
          placeholder='ios26'
        />
        <p className='text-xs text-muted-foreground'>
          Used as <code>{`{{semesterTag}}`}</code> in resource name templates.
        </p>
      </div>
    </div>
  )
}

export default SetupConfigPage
