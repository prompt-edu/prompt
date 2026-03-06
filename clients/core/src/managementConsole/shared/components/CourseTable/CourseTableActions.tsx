import { Course } from '@tumaet/prompt-shared-state'
import { Archive, ArchiveRestore } from 'lucide-react'
import { archiveCourses, unarchiveCourses } from '@core/network/mutations/updateCourseArchiveStatus'
import { RowAction } from '@tumaet/prompt-ui-components'

export const CourseTableActions: RowAction<Course>[] = [
  {
    label: 'Archive',
    icon: <Archive />,
    onAction: async (cs) => {
      await archiveCourses(cs.filter((course) => !course.archived).map((course) => course.id))
    },
    hide: (rows) => rows.every((c) => c.archived),
    confirm: {
      title: 'Archive courses',
      description: (count) =>
        `Are you sure you want to archive ${count} course${count > 1 ? 's' : ''}?`,
      confirmLabel: 'Archive',
    },
  },

  {
    label: 'Unarchive',
    icon: <ArchiveRestore />,
    onAction: async (cs) => {
      await unarchiveCourses(cs.filter((course) => course.archived).map((course) => course.id))
    },
    hide: (rows) => rows.every((c) => !c.archived),
    confirm: {
      title: 'Unarchive courses',
      description: (count) =>
        `Are you sure you want to unarchive ${count} course${count > 1 ? 's' : ''}?`,
      confirmLabel: 'Unarchive',
    },
  },
]
