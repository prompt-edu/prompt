import {
  useCreateStudyProgram,
  useDeleteStudyProgram,
  useStudyProgramStudentCounts,
  useStudyPrograms,
  useUpdateStudyProgram,
} from '@core/network/hooks/useStudyPrograms'
import {
  Button,
  DeleteConfirmation,
  ManagementPageHeader,
  PromptTable,
} from '@tumaet/prompt-ui-components'
import { Plus } from 'lucide-react'
import { useMemo, useState } from 'react'
import type { CreateStudyProgram } from '../../shared/interfaces/StudyProgram'
import { OTHER_STUDY_PROGRAM } from '../../shared/utils/otherStudyProgram'
import { StudyProgramFormDialog } from './components/StudyProgramFormDialog'
import { getStudyProgramTableActions } from './components/studyProgramTableActions'
import { studyProgramTableColumns } from './components/studyProgramTableColumns'
import {
  deleteWarning,
  type StudyProgramWithStudentCount,
  withStudentCounts,
} from './utils/studyProgramStudents'

export const StudyProgramsPage = () => {
  const studyPrograms = useStudyPrograms()
  const studentCounts = useStudyProgramStudentCounts()
  const createStudyProgram = useCreateStudyProgram()
  const updateStudyProgram = useUpdateStudyProgram()
  const deleteStudyPrograms = useDeleteStudyProgram()

  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<StudyProgramWithStudentCount | null>(null)
  const [deleting, setDeleting] = useState<StudyProgramWithStudentCount[]>([])

  const rows = useMemo(
    () => withStudentCounts(studyPrograms.data ?? [], studentCounts.data),
    [studyPrograms.data, studentCounts.data],
  )

  const actions = useMemo(
    () => getStudyProgramTableActions({ onEdit: setEditing, onDelete: setDeleting }),
    [],
  )

  const handleCreate = (values: CreateStudyProgram) => {
    createStudyProgram.mutate(values, { onSuccess: () => setCreateOpen(false) })
  }

  const handleUpdate = (values: CreateStudyProgram) => {
    if (!editing) return
    updateStudyProgram.mutate(
      { previous: editing, studyProgram: values },
      { onSuccess: () => setEditing(null) },
    )
  }

  const handleDelete = (confirmed: boolean) => {
    if (confirmed) {
      deleteStudyPrograms.mutate(deleting.map((studyProgram) => studyProgram.id))
    }
    setDeleting([])
  }

  return (
    <div className='w-full'>
      <div className='flex items-center justify-between'>
        <ManagementPageHeader>Study Programs</ManagementPageHeader>
        <Button onClick={() => setCreateOpen(true)} disabled={!studyPrograms.isSuccess}>
          <Plus className='h-4 w-4 mr-2' />
          Add Study Program
        </Button>
      </div>
      <p className='text-muted-foreground mb-6'>
        Applicants pick their study program from this list. Anyone else chooses &quot;
        {OTHER_STUDY_PROGRAM}&quot; and enters it as free text.
      </p>

      {studyPrograms.isPending && (
        <p className='text-muted-foreground text-sm'>Loading study programs…</p>
      )}
      {studyPrograms.isError && (
        <p className='text-destructive text-sm'>Failed to load study programs.</p>
      )}
      {studyPrograms.isSuccess && (
        <div className='flex flex-col gap-2'>
          {studentCounts.isError && (
            <p className='text-destructive text-sm'>
              Student counts could not be loaded, so they are not shown.
            </p>
          )}
          <PromptTable
            data={rows}
            columns={studyProgramTableColumns}
            actions={actions}
            onRowClick={setEditing}
          />
        </div>
      )}

      {createOpen && (
        <StudyProgramFormDialog
          title='Add Study Program'
          isPending={createStudyProgram.isPending}
          onClose={() => setCreateOpen(false)}
          onSubmit={handleCreate}
        />
      )}

      {editing && (
        <StudyProgramFormDialog
          title='Edit Study Program'
          editing={editing}
          isPending={updateStudyProgram.isPending}
          onClose={() => setEditing(null)}
          onSubmit={handleUpdate}
        />
      )}

      <DeleteConfirmation
        isOpen={deleting.length > 0}
        setOpen={(open) => !open && setDeleting([])}
        deleteMessage={
          deleting.length === 1
            ? `Are you sure you want to delete "${deleting[0].name}"?`
            : `Are you sure you want to delete ${deleting.length} study programs?`
        }
        customWarning={deleteWarning(deleting)}
        onClick={handleDelete}
      />
    </div>
  )
}
