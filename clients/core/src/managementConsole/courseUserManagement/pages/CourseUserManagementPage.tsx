import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { ErrorPage, LoadingPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import { useCourseTeam } from '../hooks/useCourseTeam'
import { useAddCourseTeamMember } from '../hooks/useAddCourseTeamMember'
import { useRemoveCourseTeamMember } from '../hooks/useRemoveCourseTeamMember'
import { TeamMemberTable } from '../components/TeamMemberTable'
import { AddUserDialog } from '../components/AddUserDialog'
import { CourseGroupName, TeamMember } from '../interfaces/TeamMember'

export const CourseUserManagementPage = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const { user } = useAuthStore()

  const { data, isLoading, isError, refetch } = useCourseTeam(courseId)
  const addMember = useAddCourseTeamMember(courseId ?? '')
  const removeMember = useRemoveCourseTeamMember(courseId ?? '')

  const [dialogGroup, setDialogGroup] = useState<CourseGroupName | null>(null)

  const lecturerIDs = useMemo(
    () => new Set((data?.lecturers ?? []).map((m) => m.keycloakUserID)),
    [data],
  )
  const editorIDs = useMemo(
    () => new Set((data?.editors ?? []).map((m) => m.keycloakUserID)),
    [data],
  )

  if (!courseId) return <ErrorPage onRetry={() => undefined} description='No course selected.' />
  if (isLoading) return <LoadingPage />
  if (isError || !data) {
    return <ErrorPage onRetry={() => refetch()} description='Failed to load course team.' />
  }

  const handleAdd = (groupName: CourseGroupName) => (member: TeamMember) => {
    addMember.mutate(
      { groupName, keycloakUserID: member.keycloakUserID },
      {
        onSuccess: () => setDialogGroup(null),
      },
    )
  }

  const handleRemove = (groupName: CourseGroupName) => (member: TeamMember) => {
    removeMember.mutate({ groupName, keycloakUserID: member.keycloakUserID })
  }

  return (
    <div className='space-y-6'>
      <ManagementPageHeader>User Management</ManagementPageHeader>
      <p className='text-sm text-muted-foreground'>
        Manage who has Lecturer or Editor access to this course. Lecturers can manage the course
        team and all course settings; Editors can manage course content but not the team.
      </p>

      <TeamMemberTable
        title='Lecturers'
        description='Full course administration, including team management.'
        groupName='Lecturer'
        members={data.lecturers}
        currentUsername={user?.username}
        onAdd={() => setDialogGroup('Lecturer')}
        onRemove={handleRemove('Lecturer')}
        isRemoving={removeMember.isPending}
      />

      <TeamMemberTable
        title='Editors'
        description='Can edit course content but cannot manage the team.'
        groupName='Editor'
        members={data.editors}
        currentUsername={user?.username}
        onAdd={() => setDialogGroup('Editor')}
        onRemove={handleRemove('Editor')}
        isRemoving={removeMember.isPending}
      />

      {dialogGroup && (
        <AddUserDialog
          open={dialogGroup !== null}
          onOpenChange={(open) => {
            if (!open) setDialogGroup(null)
          }}
          groupName={dialogGroup}
          existingLecturerIDs={lecturerIDs}
          existingEditorIDs={editorIDs}
          onSelect={handleAdd(dialogGroup)}
          isAdding={addMember.isPending}
        />
      )}
    </div>
  )
}
