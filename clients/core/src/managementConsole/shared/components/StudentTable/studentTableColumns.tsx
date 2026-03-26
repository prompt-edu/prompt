import {
  StudentCourseParticipation,
  StudentNoteTag,
  StudentWithCourses,
} from '@core/network/queries/getStudentsWithCourses'
import { StudentCoursePreview } from './components/StudentCoursePreview'
import { InstructorNoteTag } from '../InstructorNote/InstructorNoteTag'
import { NoteTagColor } from '../../interfaces/InstructorNote'
import { ColumnDef, Row } from '@tanstack/react-table'
import { ProfilePicture } from '@/components/StudentProfilePicture'

export const studentTableColumns: ColumnDef<StudentWithCourses>[] = [
  {
    id: 'profilepicture',
    header: '',
    cell: ({ row }) => (
      <ProfilePicture
        email={row.original.email}
        firstName={row.original.firstName}
        lastName={row.original.lastName}
        size='sm'
      />
    ),
  },
  {
    accessorKey: 'firstName',
    header: 'First Name',
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: 'lastName',
    header: 'Last Name',
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: 'email',
    header: 'Email',
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: 'currentSemester',
    header: 'Semester',
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: 'studyProgram',
    header: 'Program',
    cell: (info) => info.getValue(),
  },
  {
    id: 'courses',
    header: 'Courses',
    enableSorting: false,

    accessorFn: (row: StudentWithCourses) => row.courses.map((c) => c.courseName).join(' '),

    cell: ({ row }) => (
      <div className='flex flex-col gap-2'>
        {row.original.courses.map((scp: StudentCourseParticipation) => (
          <StudentCoursePreview studentCourseParticipation={scp} key={scp.courseId + row.id} />
        ))}
      </div>
    ),

    filterFn: (row: Row<StudentWithCourses>, _columnId, filterValue: string[]) => {
      if (!filterValue?.length) return true

      return row.original.courses.some((course) => filterValue.includes(course.courseName))
    },
  },
  {
    id: 'noteTags',
    header: 'Tags',
    enableSorting: false,

    accessorFn: (row: StudentWithCourses) => row.noteTags.map((t) => t.name).join(' '),

    cell: ({ row }) => (
      <div className='flex flex-wrap gap-1'>
        {row.original.noteTags.map((tag: StudentNoteTag) => (
          <InstructorNoteTag
            key={tag.id}
            tag={{ id: tag.id, name: tag.name, color: tag.color as NoteTagColor }}
          />
        ))}
      </div>
    ),

    filterFn: (row: Row<StudentWithCourses>, _columnId, filterValue: string[]) => {
      if (!filterValue?.length) return true

      return row.original.noteTags.some((tag) => filterValue.includes(tag.id))
    },
  },
]
