import type React from 'react'
import { useEffect } from 'react'
import {
  Button,
  Input,
  Textarea,
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
import { templateFormSchema, type TemplateFormValues } from '@core/validations/template'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { CourseType, CourseTypeDetails } from '@tumaet/prompt-shared-state'
import { checkCourseNameExists } from '@core/network/queries/checkCourseNameExists'

interface AddTemplatePropertiesProps {
  onNext: (data: TemplateFormValues) => void
  onCancel: () => void
  initialValues?: Partial<TemplateFormValues>
}

export const AddTemplateProperties: React.FC<AddTemplatePropertiesProps> = ({
  onNext,
  onCancel,
  initialValues,
}) => {
  const form = useForm<TemplateFormValues>({
    resolver: zodResolver(templateFormSchema),
    defaultValues: {
      name: initialValues?.name || '',
      courseType: initialValues?.courseType || '',
      ects: initialValues?.ects ?? undefined,
      semesterTag: 'template',
      shortDescription: initialValues?.shortDescription || '',
      longDescription: initialValues?.longDescription || '',
    },
  })

  const onSubmit = async (data: TemplateFormValues) => {
    if (await checkTemplateExistsAndUpdateForm(data.name)) {
      return
    }
    onNext(data)
  }

  const selectedCourseType = form.watch('courseType')
  const isEctsDisabled = CourseTypeDetails[selectedCourseType]?.ects !== undefined

  const checkTemplateExistsAndUpdateForm = async (val) => {
    try {
      const exists = await checkCourseNameExists(val, 'template')
      if (exists) {
        form.setError('name', {
          type: 'manual',
          message: 'A template with this name already exists.',
        })
      }
      return exists
    } catch {
      return false
    }
  }

  useEffect(() => {
    const ectsValue = CourseTypeDetails[selectedCourseType]?.ects
    if (ectsValue !== undefined) {
      form.setValue('ects', ectsValue as number, { shouldValidate: true })
    }
  }, [selectedCourseType, form])

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
        <FormField
          control={form.control}
          name='name'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Template Name</FormLabel>
              <FormControl>
                <Input
                  placeholder='Enter a template name'
                  {...field}
                  className='w-full'
                  onChange={(e) => {
                    field.onChange(e)
                    form.clearErrors('name')
                  }}
                  onBlur={async () => {
                    field.onBlur()
                    const value = form.getValues('name')
                    if (!value) return
                    checkTemplateExistsAndUpdateForm(value)
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
          <FormField
            control={form.control}
            name='courseType'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Course Type</FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder='Select a course type' />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {Object.values(CourseType).map((type) => (
                      <SelectItem key={type} value={type}>
                        {CourseTypeDetails[type as CourseType].name}
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
            name='ects'
            render={({ field }) => (
              <FormItem>
                <FormLabel>ECTS</FormLabel>
                <FormControl>
                  <Input
                    type='number'
                    placeholder='Enter ECTS'
                    {...field}
                    onChange={(e) => {
                      const value = e.target.value
                      field.onChange(value === '' ? '' : Number(value))
                    }}
                    value={field.value ?? ''}
                    disabled={isEctsDisabled}
                    className='w-full'
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        <FormField
          control={form.control}
          name='shortDescription'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Short Description</FormLabel>
              <FormControl>
                <Input placeholder='One sentence summary' {...field} className='w-full' />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='longDescription'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Long Description</FormLabel>
              <FormControl>
                <Textarea
                  placeholder='Share more details about this template (optional)'
                  className='w-full'
                  rows={4}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className='flex justify-between space-x-4 pt-4'>
          <Button type='button' variant='outline' onClick={onCancel}>
            Cancel
          </Button>
          <Button type='submit'>Next</Button>
        </div>
      </form>
    </Form>
  )
}
