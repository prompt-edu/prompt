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
} from '@tumaet/prompt-ui-components'
import type { UseFormReturn } from 'react-hook-form'
import type { MakeTemplateCourseFormValues } from '@core/validations/makeTemplateCourse'

interface MakeTemplateCourseFormProps {
  form: UseFormReturn<MakeTemplateCourseFormValues>
  courseName: string
  onSubmit: (data: MakeTemplateCourseFormValues) => void
  onClose: () => void
}

export const MakeTemplateCourseForm = ({
  form,
  courseName,
  onSubmit,
  onClose,
}: MakeTemplateCourseFormProps) => {
  return (
    <>
      <DialogHeader>
        <DialogTitle>{`Create Template from Course: ${courseName}`}</DialogTitle>
        <DialogDescription>{'Create a new template based on this course.'}</DialogDescription>
      </DialogHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
          <FormField
            control={form.control}
            name='name'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Template Name</FormLabel>
                <FormControl>
                  <Input placeholder='Enter template name' {...field} className='w-full' />
                </FormControl>
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
                <FormDescription>
                  Brief summary of the template (max 255 characters).
                </FormDescription>
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
