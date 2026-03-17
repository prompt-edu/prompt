import type React from 'react'
import { useEffect } from 'react'
import {
  Button,
  Input,
  Textarea,
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  DatePickerWithRange,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { courseFormSchema, type CourseFormValues } from '@core/validations/course'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { CourseType, CourseTypeDetails } from '@tumaet/prompt-shared-state'
import { checkCourseNameExists } from '@core/network/queries/checkCourseNameExists'

interface AddCoursePropertiesProps {
  onNext: (data: CourseFormValues) => void
  onCancel: () => void
  initialValues?: Partial<CourseFormValues>
}

export const AddCourseProperties: React.FC<AddCoursePropertiesProps> = ({
  onNext,
  onCancel,
  initialValues,
}) => {
  const form = useForm<CourseFormValues>({
    resolver: zodResolver(courseFormSchema),
    defaultValues: {
      name: initialValues?.name || '',
      dateRange: initialValues?.dateRange,
      courseType: initialValues?.courseType || '',
      ects: initialValues?.ects ?? undefined,
      semesterTag: initialValues?.semesterTag || '',
      shortDescription: initialValues?.shortDescription || '',
      longDescription: initialValues?.longDescription || '',
    },
  })

  const onSubmit = (data: CourseFormValues) => {
    onNext(data)
  }

  const selectedCourseType = form.watch('courseType')
  const isEctsDisabled = CourseTypeDetails[selectedCourseType]?.ects !== undefined

  useEffect(() => {
    const ectsValue = CourseTypeDetails[selectedCourseType]?.ects
    if (ectsValue !== undefined) {
      form.setValue('ects', ectsValue as number, { shouldValidate: true })
    }
  }, [selectedCourseType, form])

  const handleSemesterTagChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const filteredValue = e.target.value.toLowerCase().replace(/[^a-z0-9]/g, '')
    form.setValue('semesterTag', filteredValue, { shouldValidate: true })
    form.clearErrors('name')
    form.clearErrors('semesterTag')
  }

  const checkNameAvailability = async () => {
    const name = form.getValues('name')
    const semesterTag = form.getValues('semesterTag')
    if (!name || !semesterTag) return
    try {
      const exists = await checkCourseNameExists(name, semesterTag)
      if (exists) {
        form.setError('name', {
          type: 'manual',
          message: `A course with name "${name}" and semester tag "${semesterTag}" already exists. Please
                choose a different name or semester tag.`,
        })
        form.setError('semesterTag', {
          type: 'manual',
          message: 'Part of a conflicting name + semester tag combination.',
        })
      }
    } catch {
      // silently ignore network errors — server-side 409 handles it
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
        <FormField
          control={form.control}
          name='name'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Course Name</FormLabel>
              <FormControl>
                <Input
                  placeholder='Enter a course name'
                  {...field}
                  className='w-full'
                  onChange={(e) => {
                    field.onChange(e)
                    form.clearErrors('name')
                    form.clearErrors('semesterTag')
                  }}
                  onBlur={async () => {
                    field.onBlur()
                    await checkNameAvailability()
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='dateRange'
          render={({ field }) => (
            <FormItem className='flex flex-col'>
              <FormLabel>Course Duration</FormLabel>
              <DatePickerWithRange date={field.value} setDate={field.onChange} />
              <FormDescription>Select the start and end dates for your course.</FormDescription>
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
          name='semesterTag'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Semester Tag</FormLabel>
              <FormControl>
                <Input
                  placeholder='Enter a semester tag'
                  value={field.value}
                  onChange={handleSemesterTagChange}
                  onBlur={async () => {
                    field.onBlur()
                    await checkNameAvailability()
                  }}
                  className='w-full'
                />
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
          name='shortDescription'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Short Description</FormLabel>
              <FormControl>
                <Input placeholder='One sentence summary' {...field} className='w-full' />
              </FormControl>
              <FormDescription>Shown in course listings (max. 255 characters).</FormDescription>
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
                  placeholder='Share more context about this course (optional)'
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
