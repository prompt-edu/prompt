import React from 'react'
import {
  Button,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'

import { useForm } from 'react-hook-form'
import {
  CourseAppearanceFormValues,
  courseAppearanceFormSchema,
} from '@core/validations/courseAppearance'
import { zodResolver } from '@hookform/resolvers/zod'
import { IconSelector } from './components/IconSelector'
import { CourseAppearancePreview } from '../components/CourseAppearancePreview'
import { courseAppearanceColors } from '@core/managementConsole/courseOverview/constants/courseAppearance'

interface AddCourseAppearanceProps {
  onBack: () => void
  onSubmit: (data: CourseAppearanceFormValues) => void
  createTemplate?: boolean
}

export const AddCourseAppearance: React.FC<AddCourseAppearanceProps> = ({
  onBack,
  onSubmit,
  createTemplate,
}) => {
  const form = useForm<CourseAppearanceFormValues>({
    resolver: zodResolver(courseAppearanceFormSchema),
    defaultValues: {
      color: '',
      icon: '',
    },
  })

  const selectedColor = form.watch('color')
  const selectedIcon = form.watch('icon')

  return (
    <Form {...form}>
      <div className='space-y-5'>
        <CourseAppearancePreview
          color={selectedColor}
          icon={selectedIcon}
          createTemplate={createTemplate}
        />

        <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-5'>
          <div className='grid grid-cols-2 gap-4'>
            <FormField
              control={form.control}
              name='color'
              render={({ field }) => (
                <FormItem>
                  <FormLabel className='text-sm font-medium'>Color</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder='Select a color' />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {courseAppearanceColors.map((color) => (
                        <SelectItem key={color} value={color}>
                          <div className='flex items-center gap-2'>
                            <div className={`w-5 h-5 rounded-sm border border-border ${color}`}></div>
                            <span className='capitalize'>{color.split('-')[1]}</span>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='icon'
              render={({ field }) => (
                <FormItem>
                  <FormLabel className='text-sm font-medium'>Icon</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder='Select an icon' />
                      </SelectTrigger>
                    </FormControl>
                    <IconSelector />
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className='flex justify-between gap-3 pt-2'>
            <Button type='button' variant='outline' onClick={onBack}>
              Back
            </Button>
            <Button type='submit'>{createTemplate ? 'Add Template' : 'Add Course'}</Button>
          </div>
        </form>
      </div>
    </Form>
  )
}
