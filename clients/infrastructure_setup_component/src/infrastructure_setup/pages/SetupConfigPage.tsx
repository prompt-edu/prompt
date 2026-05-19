import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import {
  Button,
  ErrorPage,
  Input,
  Label,
  LoadingPage,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Save, Settings } from 'lucide-react'
import { useCourseStore } from '@tumaet/prompt-shared-state'

import { getSetupConfig } from '../network/queries/getSetupConfig'
import { updateSetupConfig } from '../network/mutations/updateSetupConfig'

const UNSET = '__none__'

export const SetupConfigPage = () => {
  const { courseId, phaseId: coursePhaseID } = useParams<{
    courseId: string
    phaseId: string
  }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const { courses } = useCourseStore()

  const course = courses.find((c) => c.id === courseId)
  // Sibling phases that could supply team or student source data — exclude self.
  const siblingPhases = (course?.coursePhases ?? []).filter((p) => p.id !== coursePhaseID)

  const [teamSourceCoursePhaseID, setTeamSourceCoursePhaseID] = useState<string>('')
  const [studentSourceCoursePhaseID, setStudentSourceCoursePhaseID] = useState<string>('')
  const [semesterTag, setSemesterTag] = useState('')

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['setup-config', coursePhaseID],
    queryFn: () => getSetupConfig(coursePhaseID!),
    enabled: !!coursePhaseID,
  })

  useEffect(() => {
    if (!data) return
    setTeamSourceCoursePhaseID(data.teamSourceCoursePhaseId ?? '')
    setStudentSourceCoursePhaseID(data.studentSourceCoursePhaseId ?? '')
    setSemesterTag(data.semesterTag ?? '')
  }, [data])

  // Prefill semester tag from the parent course if not set yet.
  useEffect(() => {
    if (!course || semesterTag) return
    setSemesterTag(course.semesterTag ?? '')
  }, [course, semesterTag])

  const { mutate: save, isPending } = useMutation({
    mutationFn: () =>
      updateSetupConfig(coursePhaseID!, {
        teamSourceCoursePhaseId: teamSourceCoursePhaseID || null,
        studentSourceCoursePhaseId: studentSourceCoursePhaseID || null,
        semesterTag: semesterTag.trim(),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['setup-config', coursePhaseID] })
      toast({ title: 'Setup configuration saved' })
    },
    onError: (err: unknown) => {
      toast({
        title: 'Failed to save setup configuration',
        description: err instanceof Error ? err.message : 'Unknown error',
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

  const phaseSelect = (value: string, onChange: (value: string) => void, placeholder: string) => (
    <Select
      value={value === '' ? UNSET : value}
      onValueChange={(v) => onChange(v === UNSET ? '' : v)}
    >
      <SelectTrigger>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={UNSET}>— none —</SelectItem>
        {siblingPhases.map((phase) => (
          <SelectItem key={phase.id} value={phase.id}>
            {phase.name} ({phase.coursePhaseType})
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )

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

      <div className='space-y-4'>
        <div className='space-y-1'>
          <Label>Team source phase</Label>
          {phaseSelect(
            teamSourceCoursePhaseID,
            setTeamSourceCoursePhaseID,
            'Select a phase that provides teams',
          )}
          <p className='text-xs text-muted-foreground'>
            Phase whose teams are read for <code>per_team</code> resource configs.
          </p>
        </div>

        <div className='space-y-1'>
          <Label>Student source phase</Label>
          {phaseSelect(
            studentSourceCoursePhaseID,
            setStudentSourceCoursePhaseID,
            'Select a phase that provides students',
          )}
          <p className='text-xs text-muted-foreground'>
            Phase whose participants are read for <code>per_student</code> resource configs.
          </p>
        </div>

        <div className='space-y-1'>
          <Label htmlFor='semesterTag'>Semester tag</Label>
          <Input
            id='semesterTag'
            value={semesterTag}
            onChange={(event) => setSemesterTag(event.target.value)}
            placeholder='ios26'
          />
          <p className='text-xs text-muted-foreground'>
            Used as <code>{`{{semesterTag}}`}</code> in resource name templates.
          </p>
        </div>
      </div>
    </div>
  )
}

export default SetupConfigPage
