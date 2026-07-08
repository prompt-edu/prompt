import type { StudentWithCourses } from '@core/network/queries/getStudentsWithCourses'
import {
  DropdownMenuCheckboxItem,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  type TableFilter,
} from '@tumaet/prompt-ui-components'
import type { NoteTagColor } from '../../interfaces/InstructorNote'
import { InstructorNoteTag } from '../InstructorNote/InstructorNoteTag'
import { StudentCoursePreview } from './components/StudentCoursePreview'

export function getStudentTableFilters(studentsWithCourses: StudentWithCourses[]): TableFilter[] {
  const tagOptions = Array.from(
    new Map(studentsWithCourses.flatMap((s) => s.noteTags).map((t) => [t.id, t])).values(),
  )

  return [
    {
      type: 'custom',
      id: 'courses',
      label: 'Course',
      render: ({ column }) => {
        const selected = (column.getFilterValue() as string[]) ?? []

        const courseOptions = Array.from(
          new Map(
            studentsWithCourses
              .flatMap((s) => s.courses)
              .filter((c) => c.courseName)
              .map((c) => [c.courseName, c]),
          ).values(),
        )

        return (
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>Course</DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
              {courseOptions.map((course) => (
                <DropdownMenuCheckboxItem
                  key={course.courseId}
                  checked={selected.includes(course.courseName)}
                  onCheckedChange={(checked) => {
                    const next = checked
                      ? [...selected, course.courseName]
                      : selected.filter((c) => c !== course.courseName)
                    column.setFilterValue(next.length === 0 ? undefined : next)
                  }}
                >
                  <StudentCoursePreview studentCourseParticipation={course} />
                </DropdownMenuCheckboxItem>
              ))}
            </DropdownMenuSubContent>
          </DropdownMenuSub>
        )
      },
    },
    {
      type: 'custom',
      id: 'noteTags',
      label: 'Tag',
      badge: {
        label: 'Tag',
        displayValue: (filtervalue: unknown) =>
          tagOptions.find((t) => t.id === filtervalue)?.name ?? String(filtervalue),
      },
      render: ({ column }) => {
        const selected = (column.getFilterValue() as string[]) ?? []

        return (
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>Tag</DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
              {tagOptions.length === 0 && (
                <p className='text-muted-foreground text-sm px-1'>No tags available</p>
              )}
              {tagOptions.map((tag) => (
                <DropdownMenuCheckboxItem
                  key={tag.id}
                  checked={selected.includes(tag.id)}
                  onCheckedChange={(checked) => {
                    const next = checked
                      ? [...selected, tag.id]
                      : selected.filter((id) => id !== tag.id)
                    column.setFilterValue(next.length === 0 ? undefined : next)
                  }}
                >
                  <InstructorNoteTag
                    tag={{ id: tag.id, name: tag.name, color: tag.color as NoteTagColor }}
                  />
                </DropdownMenuCheckboxItem>
              ))}
            </DropdownMenuSubContent>
          </DropdownMenuSub>
        )
      },
    },
    {
      type: 'custom',
      id: 'lastModified',
      label: 'Last Modified',
      badge: {
        label: 'Not modified in',
        displayValue: (filterValue: unknown) =>
          typeof filterValue === 'number' ? `${filterValue} years` : '',
      },
      render: ({ column }) => {
        const selected = column.getFilterValue() as number | undefined
        return (
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>Last Modified</DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
              {[3, 5, 7].map((years) => (
                <DropdownMenuCheckboxItem
                  key={years}
                  checked={selected === years}
                  onCheckedChange={(checked) => column.setFilterValue(checked ? years : undefined)}
                >
                  Not modified in {years} years
                </DropdownMenuCheckboxItem>
              ))}
            </DropdownMenuSubContent>
          </DropdownMenuSub>
        )
      },
    },
  ]
}
