import { useAuthStore } from '@tumaet/prompt-shared-state'
import { ErrorPage, LoadingPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { AddUserDialog } from '../../shared/components/AddUserDialog'
import { StaffMemberTable } from '../components/StaffMemberTable'
import { useAddCourseStaffMember } from '../hooks/useAddCourseStaffMember'
import { useCourseStaff } from '../hooks/useCourseStaff'
import { useRemoveCourseStaffMember } from '../hooks/useRemoveCourseStaffMember'
import type { CourseGroupName, StaffMember } from '../interfaces/StaffMember'

export const CourseUserManagementPage = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const { user } = useAuthStore()

  const { data, isLoading, isError, error, refetch } = useCourseStaff(courseId)
  const addMember = useAddCourseStaffMember(courseId ?? '')
  const removeMember = useRemoveCourseStaffMember(courseId ?? '')

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
    const serverMessage = (error as { response?: { data?: { error?: string } } })?.response?.data
      ?.error
    return (
      <ErrorPage
        onRetry={() => refetch()}
        description={serverMessage ?? 'Failed to load course staff.'}
      />
    )
  }

  const handleAdd = (groupName: CourseGroupName) => (member: StaffMember) => {
    addMember.mutate(
      { groupName, keycloakUserID: member.keycloakUserID },
      {
        onSuccess: () => setDialogGroup(null),
      },
    )
  }

  const handleRemove = (groupName: CourseGroupName) => (member: StaffMember) => {
    removeMember.mutate({ groupName, keycloakUserID: member.keycloakUserID })
  }

  const holdersOf = (groupName: CourseGroupName) =>
    groupName === 'Lecturer' ? lecturerIDs : editorIDs

  const membershipNote = (groupName: CourseGroupName, candidate: StaffMember): string | null => {
    if (holdersOf(groupName).has(candidate.keycloakUserID)) return `Already a ${groupName}`
    if (groupName === 'Lecturer' && editorIDs.has(candidate.keycloakUserID))
      return 'Currently an Editor'
    if (groupName === 'Editor' && lecturerIDs.has(candidate.keycloakUserID))
      return 'Currently a Lecturer'
    return null
  }

  return (
    <div className='space-y-6'>
      <ManagementPageHeader>User Management</ManagementPageHeader>
      <p className='text-sm text-muted-foreground'>
        Manage who has Lecturer or Editor access to this course. Lecturers can manage the course
        staff and all course settings; Editors can manage course content but not the staff.
      </p>

      <StaffMemberTable
        title='Lecturers'
        description='Full course administration, including staff management.'
        groupName='Lecturer'
        members={data.lecturers}
        currentUsername={user?.username}
        onAdd={() => setDialogGroup('Lecturer')}
        onRemove={handleRemove('Lecturer')}
        isRemoving={removeMember.isPending}
      />

      <StaffMemberTable
        title='Editors'
        description='Can edit course content but cannot manage the staff.'
        groupName='Editor'
        members={data.editors}
        currentUsername={user?.username}
        onAdd={() => setDialogGroup('Editor')}
        onRemove={handleRemove('Editor')}
        isRemoving={removeMember.isPending}
      />

      {dialogGroup && (
        <AddUserDialog
          open={true}
          onOpenChange={(open) => {
            if (!open) setDialogGroup(null)
          }}
          title={`Add ${dialogGroup}`}
          noteFor={(candidate) => membershipNote(dialogGroup, candidate)}
          isDisabled={(candidate) => holdersOf(dialogGroup).has(candidate.keycloakUserID)}
          onSelect={handleAdd(dialogGroup)}
          isAdding={addMember.isPending}
        />
      )}
    </div>
  )
}
