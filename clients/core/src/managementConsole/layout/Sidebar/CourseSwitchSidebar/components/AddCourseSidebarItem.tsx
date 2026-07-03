import { CourseCreationChoiceDialog } from '@managementConsole/courseOverview/AddingCourse/components/CourseCreationChoiceDialog'
import { SidebarMenuButton, SidebarMenuItem } from '@tumaet/prompt-ui-components'
import { Plus } from 'lucide-react'

export const AddCourseButton = () => {
  return (
    <SidebarMenuItem>
      <CourseCreationChoiceDialog>
        <SidebarMenuButton
          size='lg'
          tooltip={{
            children: 'Add new course or template',
            hidden: false,
          }}
          className='min-w-12 min-h-12 p-0'
        >
          <div className='relative flex aspect-square size-12 items-center justify-center'>
            <div className='flex aspect-square size-10 items-center justify-center rounded-lg bg-gray-100 text-gray-800'>
              <Plus className='size-6' />
            </div>
          </div>
        </SidebarMenuButton>
      </CourseCreationChoiceDialog>
    </SidebarMenuItem>
  )
}
