import { useForm } from 'react-hook-form'
import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  Button,
  Form,
  FormControl,
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
  Input,
  Textarea,
  useToast,
  CardFooter,
  CardContent,
  LoadingPage,
  ErrorPage,
} from '@tumaet/prompt-ui-components'
import { CourseType, CourseTypeDetails, useCourseStore } from '@tumaet/prompt-shared-state'
import { EditCourseFormValues, editCourseSchema } from '@core/validations/editCourse'
import { updateCourseData } from '@core/network/mutations/updateCourseData'
import type { Course, UpdateCourseData } from '@tumaet/prompt-shared-state'
import {
  DEFAULT_COURSE_COLOR,
  DEFAULT_COURSE_ICON,
  courseAppearanceColors,
} from '@core/managementConsole/courseOverview/constants/courseAppearance'
import { IconSelector } from '@core/managementConsole/courseOverview/AddingCourse/components/IconSelector'
import { FileText, Loader2, Save } from 'lucide-react'
import { SettingsCard } from '@/components/SettingsCard'
import { getAllCourses } from '@core/network/queries/course'

export function CourseGeneralSettings() {
  const { courseId } = useParams<{ courseId: string }>()
  const { courses, setCourses } = useCourseStore()
  const { toast } = useToast()
  const queryClient = useQueryClient()

  // Fetch courses if not in store (handles direct page load)
  const {
    data: fetchedCourses,
    isPending: isLoadingCourses,
    isError: isCoursesError,
    refetch: refetchCourses,
  } = useQuery<Course[]>({
    queryKey: ['courses'],
    queryFn: () => getAllCourses(),
  })

  // Update store when data arrives
  useEffect(() => {
    if (fetchedCourses) {
      setCourses([...fetchedCourses])
    }
  }, [fetchedCourses, setCourses])

  // Use fetched data or store data
  const allCourses = fetchedCourses || courses
  const course = allCourses.find((c) => c.id === courseId) as Course | undefined

  const initialColor = (course?.studentReadableData?.['bg-color'] as string) || DEFAULT_COURSE_COLOR
  const initialIcon = (course?.studentReadableData?.['icon'] as string) || DEFAULT_COURSE_ICON

  const form = useForm<EditCourseFormValues>({
    resolver: zodResolver(editCourseSchema),
    defaultValues: {
      dateRange: {
        from: course?.startDate ? new Date(course.startDate) : new Date(),
        to: course?.endDate ? new Date(course.endDate) : new Date(),
      },
      courseType: course?.courseType,
      ects: course?.ects ?? 0,
      shortDescription: course?.shortDescription || '',
      longDescription: course?.longDescription || '',
      color: initialColor,
      icon: initialIcon,
    },
  })

  const selectedCourseType = form.watch('courseType')
  const isEctsDisabled = CourseTypeDetails[selectedCourseType]?.ects !== undefined

  // Reset form when course data becomes available
  useEffect(() => {
    if (course) {
      const newInitialColor =
        (course.studentReadableData?.['bg-color'] as string) || DEFAULT_COURSE_COLOR
      const newInitialIcon = (course.studentReadableData?.['icon'] as string) || DEFAULT_COURSE_ICON

      form.reset({
        dateRange: {
          from: course.startDate ? new Date(course.startDate) : new Date(),
          to: course.endDate ? new Date(course.endDate) : new Date(),
        },
        courseType: course.courseType,
        ects: course.ects ?? 0,
        shortDescription: course.shortDescription || '',
        longDescription: course.longDescription || '',
        color: newInitialColor,
        icon: newInitialIcon,
      })
    }
  }, [course, form])

  useEffect(() => {
    const ectsValue = CourseTypeDetails[selectedCourseType]?.ects
    if (ectsValue !== undefined) {
      form.setValue('ects', ectsValue as number, { shouldValidate: true })
    }
  }, [selectedCourseType, form])

  const { mutate: mutateCourse, isPending: isSaving } = useMutation({
    mutationFn: (courseData: UpdateCourseData) => {
      return updateCourseData(courseId ?? 'undefined', courseData)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['courses'] })
      toast({
        title: 'Successfully Updated Course',
      })
    },
    onError: () => {
      toast({
        title: 'Failed to Store Course Details',
        description: 'Please try again later!',
        variant: 'destructive',
      })
    },
  })

  // Track if form has been modified
  const isModified = form.formState.isDirty

  const onSubmit = (data: EditCourseFormValues) => {
    const updateData: UpdateCourseData = {
      startDate: data.dateRange.from,
      endDate: data.dateRange.to,
      courseType: data.courseType,
      ects: data.ects,
      shortDescription: data.shortDescription,
      longDescription: data.longDescription,
      studentReadableData: {
        icon: data.icon,
        'bg-color': data.color,
      },
    }
    mutateCourse(updateData)
  }

  // Handle loading state
  if (isLoadingCourses) {
    return <LoadingPage />
  }

  // Handle error state
  if (isCoursesError) {
    return (
      <ErrorPage
        onRetry={() => refetchCourses()}
        onLogout={() => {
          /* Add logout handler if needed */
        }}
      />
    )
  }

  // Handle course not found
  if (!course) {
    return (
      <SettingsCard
        icon={<FileText className='w-5 h-5' />}
        title='General Course Settings'
        description='Course not found.'
      >
        <CardContent>
          <p className='text-muted-foreground'>
            The course you are looking for does not exist or you do not have access to it.
          </p>
        </CardContent>
      </SettingsCard>
    )
  }

  return (
    <SettingsCard
      icon={<FileText className='w-5 h-5' />}
      title='General Course Settings'
      description='The course name and semester tag cannot be changed.'
    >
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
          {/* Course Basics Section */}
          <CardContent>
            <div className='space-y-6'>
              <div className='space-y-4'>
                <FormField
                  control={form.control}
                  name='dateRange'
                  render={({ field }) => (
                    <FormItem className='flex flex-col'>
                      <FormLabel className='text-sm font-medium'>Course Duration</FormLabel>
                      <DatePickerWithRange date={field.value} setDate={field.onChange} />
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className='grid grid-cols-2 gap-4'>
                  <FormField
                    control={form.control}
                    name='courseType'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-sm font-medium'>Course Type</FormLabel>
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
                        <FormLabel className='text-sm font-medium'>ECTS Credits</FormLabel>
                        <FormControl>
                          <Input
                            type='number'
                            placeholder='Enter ECTS'
                            {...field}
                            onChange={(e) => {
                              const value = e.target.value
                              field.onChange(value === '' ? '' : Number(value))
                            }}
                            value={field.value === 0 && !isEctsDisabled ? '' : field.value}
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
                      <FormLabel className='text-sm font-medium'>Short Description</FormLabel>
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
                      <FormLabel className='text-sm font-medium'>Long Description</FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder='Share more context about this course'
                          className='w-full min-h-[100px]'
                          rows={4}
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              {/* Appearance Section */}
              <div className='border-t pt-6 space-y-4'>
                <div>
                  <h3 className='text-sm font-semibold text-foreground'>Course Appearance</h3>
                  <p className='text-sm text-muted-foreground mt-1'>
                    Customize how this course appears to students
                  </p>
                </div>

                <div className='grid grid-cols-2 gap-4'>
                  <FormField
                    control={form.control}
                    name='color'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-sm font-medium'>Background Color</FormLabel>
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
                                  <div
                                    className={`w-5 h-5 rounded border border-border ${color}`}
                                  ></div>
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
                        <FormLabel className='text-sm font-medium'>Course Icon</FormLabel>
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
              </div>
            </div>
          </CardContent>

          <CardFooter className='flex justify-end'>
            <Button type='submit' disabled={!isModified || isSaving} className='w-full sm:w-auto'>
              {isSaving ? (
                <>
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  Saving...
                </>
              ) : (
                <>
                  <Save className='mr-2 h-4 w-4' />
                  Save Changes{' '}
                </>
              )}
            </Button>
          </CardFooter>
        </form>
      </Form>
    </SettingsCard>
  )
}
