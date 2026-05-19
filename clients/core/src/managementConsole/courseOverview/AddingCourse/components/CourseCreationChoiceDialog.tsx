import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@tumaet/prompt-ui-components'
import { Plus, FileIcon as FileTemplate, LayoutTemplate } from 'lucide-react'
import { AddCourseDialog } from '@managementConsole/courseOverview/AddingCourse/AddCourseDialog'
import { TemplateSelectionDialog } from './TemplateSelectionDialog'
import { AddTemplateDialog } from '../AddTemplateDialog'

interface CourseCreationChoiceDialogProps {
  children: React.ReactNode
}

export const CourseCreationChoiceDialog = ({ children }: CourseCreationChoiceDialogProps) => {
  const [isOpen, setIsOpen] = useState(false)
  const [showAddCourse, setShowAddCourse] = useState(false)
  const [showTemplateSelection, setShowTemplateSelection] = useState(false)
  const [showAddTemplate, setShowAddTemplate] = useState(false)

  const handleNewCourse = () => {
    setIsOpen(false)
    setShowAddCourse(true)
  }

  const handleFromTemplate = () => {
    setIsOpen(false)
    setShowTemplateSelection(true)
  }

  const handleAddTemplate = () => {
    setIsOpen(false)
    setShowAddTemplate(true)
  }

  return (
    <>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger asChild>{children}</DialogTrigger>
        <DialogContent className='sm:max-w-[420px] p-6'>
          <DialogHeader className='text-center'>
            <DialogTitle className='text-2xl font-bold text-gray-900 dark:text-gray-50'>
              Select a Starting Point
            </DialogTitle>
            <p className='text-sm text-muted-foreground mt-2'>
              Choose how you would like to get started
            </p>
          </DialogHeader>
          <div className='space-y-4 py-6'>
            <div
              onClick={handleNewCourse}
              className={
                'group flex items-center gap-4 w-full p-5 rounded-xl border border-border bg-card shadow-xs cursor-pointer ' +
                'transition-colors duration-200 hover:bg-accent hover:text-accent-foreground'
              }
              role='button'
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  handleNewCourse()
                }
              }}
            >
              <div
                className={
                  'flex items-center justify-center w-12 h-12 rounded-full bg-primary/10 dark:bg-primary/20 shrink-0 ' +
                  'transition-colors duration-200 group-hover:text-accent-foreground'
                }
              >
                <Plus className='w-6 h-6 text-primary transition-colors duration-200 group-hover:text-accent-foreground' />
              </div>
              <div className='flex flex-col min-w-0 flex-1'>
                <span className='font-semibold text-base text-gray-800 dark:text-gray-100 transition-colors duration-200 group-hover:text-accent-foreground'>
                  Create New Course
                </span>
                <span className='text-sm text-muted-foreground mt-0.5'>
                  Start from scratch with a blank course
                </span>
              </div>
            </div>
            <div
              onClick={handleAddTemplate}
              className={
                'group flex items-center gap-4 w-full p-5 rounded-xl border border-border bg-card shadow-xs cursor-pointer ' +
                'transition-colors duration-200 hover:bg-accent hover:text-accent-foreground'
              }
              role='button'
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  handleAddTemplate()
                }
              }}
            >
              <div
                className={
                  'flex items-center justify-center w-12 h-12 rounded-full bg-primary/10 dark:bg-primary/20 shrink-0 ' +
                  'transition-colors duration-200 group-hover:text-accent-foreground'
                }
              >
                <LayoutTemplate className='w-6 h-6 text-primary transition-colors duration-200 group-hover:text-accent-foreground' />
              </div>
              <div className='flex flex-col min-w-0 flex-1'>
                <span className='font-semibold text-base text-gray-800 dark:text-gray-100 transition-colors duration-200 group-hover:text-accent-foreground'>
                  Create New Template
                </span>
                <span className='text-sm text-muted-foreground mt-0.5'>
                  Start from scratch with a blank template
                </span>
              </div>
            </div>
            <div
              onClick={handleFromTemplate}
              className={
                'group flex items-center gap-4 w-full p-5 rounded-xl border border-border bg-card shadow-xs cursor-pointer ' +
                'transition-colors duration-200 hover:bg-accent hover:text-accent-foreground'
              }
              role='button'
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  handleFromTemplate()
                }
              }}
            >
              <div
                className={
                  'flex items-center justify-center w-12 h-12 rounded-full bg-primary/10 dark:bg-primary/20 shrink-0 ' +
                  'transition-colors duration-200 group-hover:text-accent-foreground'
                }
              >
                <FileTemplate className='w-6 h-6 text-primary transition-colors duration-200 group-hover:text-accent-foreground' />
              </div>
              <div className='flex flex-col min-w-0 flex-1'>
                <span className='font-semibold text-base text-gray-800 dark:text-gray-100 transition-colors duration-200 group-hover:text-accent-foreground'>
                  Use Template
                </span>
                <span className='text-sm text-muted-foreground mt-0.5'>
                  Choose from your existing templates
                </span>
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
      <AddCourseDialog open={showAddCourse} onOpenChange={setShowAddCourse} />
      <TemplateSelectionDialog
        open={showTemplateSelection}
        onOpenChange={setShowTemplateSelection}
      />
      <AddTemplateDialog open={showAddTemplate} onOpenChange={setShowAddTemplate} />
    </>
  )
}
