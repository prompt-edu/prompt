import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { Save, Settings } from 'lucide-react'
import { useEffect, useState } from 'react'
import { infrastructureSetupAxiosInstance } from '../network/infrastructureSetupServerConfig'

interface SetupConfig {
  coursePhaseId: string
  teamSourceCoursePhaseId?: string
  studentSourceCoursePhaseId?: string
  semesterTag: string
}

const emptyToNull = (value: string): string | null => {
  const trimmed = value.trim()
  return trimmed === '' ? null : trimmed
}

const fetchSetupConfig = async (coursePhaseID: string): Promise<SetupConfig> => {
  const response = await infrastructureSetupAxiosInstance.get(
    `infrastructure-setup/api/course_phase/${coursePhaseID}/setup-config`,
  )
  return response.data
}

const updateSetupConfig = async (
  coursePhaseID: string,
  config: Omit<SetupConfig, 'coursePhaseId'>,
): Promise<SetupConfig> => {
  const response = await infrastructureSetupAxiosInstance.put(
    `infrastructure-setup/api/course_phase/${coursePhaseID}/setup-config`,
    {
      teamSourceCoursePhaseId: emptyToNull(config.teamSourceCoursePhaseId ?? ''),
      studentSourceCoursePhaseId: emptyToNull(config.studentSourceCoursePhaseId ?? ''),
      semesterTag: config.semesterTag.trim(),
    },
  )
  return response.data
}

export const SetupConfigPage = () => {
  const { coursePhaseID } = useParams<{ coursePhaseID: string }>()
  const queryClient = useQueryClient()
  const [teamSourceCoursePhaseID, setTeamSourceCoursePhaseID] = useState('')
  const [studentSourceCoursePhaseID, setStudentSourceCoursePhaseID] = useState('')
  const [semesterTag, setSemesterTag] = useState('')

  const { data, isLoading, isError } = useQuery({
    queryKey: ['setup-config', coursePhaseID],
    queryFn: () => fetchSetupConfig(coursePhaseID!),
    enabled: !!coursePhaseID,
  })

  useEffect(() => {
    if (!data) return
    setTeamSourceCoursePhaseID(data.teamSourceCoursePhaseId ?? '')
    setStudentSourceCoursePhaseID(data.studentSourceCoursePhaseId ?? '')
    setSemesterTag(data.semesterTag ?? '')
  }, [data])

  const { mutate: saveConfig, isPending } = useMutation({
    mutationFn: () =>
      updateSetupConfig(coursePhaseID!, {
        teamSourceCoursePhaseId: teamSourceCoursePhaseID,
        studentSourceCoursePhaseId: studentSourceCoursePhaseID,
        semesterTag,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['setup-config', coursePhaseID] })
    },
  })

  if (isLoading) {
    return <div className='p-4 text-gray-500'>Loading setup configuration...</div>
  }

  if (isError) {
    return <div className='p-4 text-red-500'>Failed to load setup configuration.</div>
  }

  return (
    <div className='p-6 space-y-4 max-w-3xl'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center space-x-2'>
          <Settings className='h-5 w-5 text-blue-500' />
          <h1 className='text-xl font-semibold'>Setup</h1>
        </div>
        <button
          onClick={() => saveConfig()}
          disabled={isPending}
          className='px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm disabled:opacity-50 flex items-center space-x-1'
        >
          <Save className='h-4 w-4' />
          <span>{isPending ? 'Saving...' : 'Save'}</span>
        </button>
      </div>

      <div className='space-y-4'>
        <label className='block space-y-1'>
          <span className='text-sm font-medium text-gray-700'>Team source course phase ID</span>
          <input
            value={teamSourceCoursePhaseID}
            onChange={(event) => setTeamSourceCoursePhaseID(event.target.value)}
            className='w-full rounded border border-gray-300 px-3 py-2 text-sm font-mono'
            placeholder='00000000-0000-0000-0000-000000000000'
          />
        </label>

        <label className='block space-y-1'>
          <span className='text-sm font-medium text-gray-700'>Student source course phase ID</span>
          <input
            value={studentSourceCoursePhaseID}
            onChange={(event) => setStudentSourceCoursePhaseID(event.target.value)}
            className='w-full rounded border border-gray-300 px-3 py-2 text-sm font-mono'
            placeholder='00000000-0000-0000-0000-000000000000'
          />
        </label>

        <label className='block space-y-1'>
          <span className='text-sm font-medium text-gray-700'>Semester tag</span>
          <input
            value={semesterTag}
            onChange={(event) => setSemesterTag(event.target.value)}
            className='w-full rounded border border-gray-300 px-3 py-2 text-sm'
            placeholder='ios26'
          />
        </label>
      </div>
    </div>
  )
}

export default SetupConfigPage
