import {
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
  Input,
  Textarea,
  Button,
  DatePickerWithRange,
} from '@tumaet/prompt-ui-components'
import type { UseFormReturn } from 'react-hook-form'
import type { CopyCourseFormValues } from '@core/validations/copyCourse'

interface CopyCourseFormProps {
  form: UseFormReturn<CopyCourseFormValues>
  courseName: string
  onSubmit: (data: CopyCourseFormValues) => void
  onClose: () => void
  useTemplateCopy?: boolean
}

export const CopyCourseForm = ({
  form,
  courseName,
  onSubmit,
  onClose,
  useTemplateCopy,
}: CopyCourseFormProps) => {
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {useTemplateCopy ? `Create Course from Template: ${courseName}` : `Copy: ${courseName}`}
        </DialogTitle>
        <DialogDescription>
          {useTemplateCopy
            ? 'Create a new course based on this template.'
            : 'Create a complete copy of this course.'}
        </DialogDescription>
      </DialogHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
          <FormField
            control={form.control}
            name='name'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Course Name</FormLabel>
                <FormControl>
                  <Input placeholder='Enter course name' {...field} className='w-full' />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='semesterTag'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Semester Tag</FormLabel>
                <FormControl>
                  <Input placeholder='Enter semester tag' {...field} className='w-full' />
                </FormControl>
                <FormDescription>
                  e.g. ios2425 or ws2425 (lowercase letters and numbers only)
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='dateRange'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Course Duration</FormLabel>
                <DatePickerWithRange
                  date={
                    field.value && ('from' in field.value || 'to' in field.value)
                      ? { from: field.value.from ?? undefined, to: field.value.to ?? undefined }
                      : undefined
                  }
                  setDate={field.onChange}
                  className='w-full'
                />
                <FormDescription>Select the start and end dates for your course.</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='shortDescription'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Short Description</FormLabel>
                <FormControl>
                  <Input placeholder='Enter a short description' {...field} className='w-full' />
                </FormControl>
                <FormDescription>Brief summary of the course (max 255 characters).</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='longDescription'
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  Long Description{' '}
                  <span className='text-muted-foreground font-normal'>(optional)</span>
                </FormLabel>
                <FormControl>
                  <Textarea
                    placeholder='Enter a detailed description'
                    className='w-full resize-none'
                    rows={4}
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <DialogFooter>
            <Button type='button' variant='outline' onClick={onClose}>
              Cancel
            </Button>
            <Button type='submit'>Continue</Button>
          </DialogFooter>
        </form>
      </Form>
    </>
  )
}
