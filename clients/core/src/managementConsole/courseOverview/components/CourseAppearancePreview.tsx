import React from 'react'
import { DynamicIcon } from '@tumaet/prompt-ui-components'
import {
  DEFAULT_COURSE_COLOR,
  DEFAULT_COURSE_ICON,
} from '@core/managementConsole/courseOverview/constants/courseAppearance'

interface CourseAppearancePreviewProps {
  color?: string
  icon?: string
  createTemplate?: boolean
}

export const CourseAppearancePreview = ({
  color,
  icon,
  createTemplate,
}: CourseAppearancePreviewProps) => {
  const appliedColor = color || DEFAULT_COURSE_COLOR
  const appliedIcon = icon || DEFAULT_COURSE_ICON

  return (
    <div className='space-y-3 p-4 bg-muted/30 rounded-lg border'>
      <h3 className='text-sm font-semibold text-foreground/90'>Preview</h3>
      <div className='flex items-center gap-4'>
        <div
          className={[
            'relative flex aspect-square size-14 items-center justify-center',
            'after:absolute after:inset-0 after:rounded-lg after:border-2 after:border-primary after:shadow-sm',
          ].join(' ')}
        >
          <div
            className={`flex aspect-square items-center justify-center rounded-lg text-gray-800 size-14 ${appliedColor}`}
          >
            <DynamicIcon name={appliedIcon} />
          </div>
        </div>
        <span className='text-xs text-muted-foreground leading-relaxed'>
          {createTemplate ? (
            'This is how your template icon will appear in the sidebar.'
          ) : (
            <>
              This is how your course icon will appear in the sidebar.
              <br />
              <span className='font-medium text-foreground/80'>Note:</span> Only displayed while the
              course is active.
            </>
          )}
        </span>
      </div>
    </div>
  )
}
