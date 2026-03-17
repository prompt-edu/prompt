import { useState, useEffect, useCallback } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm, useFieldArray } from 'react-hook-form'
import type * as z from 'zod'
import { Mail, Copy, EyeOff, Save, Plus, Trash2, Loader2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import {
  Button,
  CardContent,
  CardFooter,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Separator,
} from '@tumaet/prompt-ui-components'
import { useCourseStore, CourseMailingSettings } from '@tumaet/prompt-shared-state'
import { type CourseMailingFormValues, courseMailingSchema } from '@core/validations/courseMailing'
import { useSaveMailingData } from './hooks/useSaveMailingData'
import { SettingsCard } from '@/components/SettingsCard'

export const MailingConfigPage = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const { courses } = useCourseStore()
  const currentCourse = courses.find((course) => course.id === courseId)
  const applicationMailingMetaData = currentCourse?.restrictedData
    .mailingSettings as CourseMailingSettings
  const [isModified, setIsModified] = useState(false)

  const { mutate: mutateMailingData, isPending } = useSaveMailingData({
    onSuccess: () => setIsModified(false),
  })

  // Setup the form with default arrays for cc/bcc
  const form = useForm<CourseMailingFormValues>({
    resolver: zodResolver(courseMailingSchema),
    defaultValues: {
      replyToName: '',
      replyToEmail: '',
      ccAddresses: [],
      bccAddresses: [],
    },
  })

  // Use useFieldArray to manage dynamic fields for CC
  const {
    fields: ccFields,
    append: appendCC,
    remove: removeCC,
  } = useFieldArray({
    control: form.control,
    name: 'ccAddresses',
  })

  // Use useFieldArray to manage dynamic fields for BCC
  const {
    fields: bccFields,
    append: appendBCC,
    remove: removeBCC,
  } = useFieldArray({
    control: form.control,
    name: 'bccAddresses',
  })

  useEffect(() => {
    if (applicationMailingMetaData) {
      form.reset(applicationMailingMetaData)
      setIsModified(false)
    }
  }, [applicationMailingMetaData, form])

  const onSubmit = useCallback(
    (values: z.infer<typeof courseMailingSchema>) => {
      if (currentCourse) {
        const updatedCourse = {
          restrictedData: {
            mailingSettings: values,
          },
        }
        mutateMailingData(updatedCourse)
      }
    },
    [currentCourse, mutateMailingData],
  )

  // Handler to remove CC and submit form
  const handleRemoveCC = (index: number) => {
    removeCC(index)
    form.handleSubmit(onSubmit)()
  }

  // Handler to remove BCC and submit form
  const handleRemoveBCC = (index: number) => {
    removeBCC(index)
    form.handleSubmit(onSubmit)()
  }

  return (
    <SettingsCard
      icon={<Mail className='h-5 w-5 text-gray-600 dark:text-gray-400' />}
      title='Mailing Settings'
      description='Configure the mailing settings for this course.'
    >
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} onChange={() => setIsModified(true)}>
          <CardContent className='space-y-8'>
            {/* Reply-To Settings */}
            <div>
              <h3 className='text-lg font-semibold mb-2 flex items-center'>
                <Mail className='mr-2' size={20} />
                Replier Information
              </h3>
              <p className='text-sm text-muted-foreground mb-4'>
                The reply-to email and name will be used for the reply-to field in the mail header.
              </p>
              <div className='grid grid-cols-1 md:grid-cols-2 gap-6'>
                <FormField
                  control={form.control}
                  name='replyToEmail'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Reply To Email</FormLabel>
                      <FormControl>
                        <Input placeholder='i.e. course@management.de' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='replyToName'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Replier Name</FormLabel>
                      <FormControl>
                        <Input placeholder='i.e. Course Management' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </div>

            <Separator />

            {/* CC Settings */}
            <div>
              <h3 className='text-lg font-semibold mb-2 flex items-center'>
                <Copy className='mr-2' size={20} />
                CC Settings
              </h3>
              <p className='text-sm text-muted-foreground mb-4'>
                If a CC Email is set, every mail sent by the server (including confirmation and
                passed/failed mails) for EVERY STUDENT will also be sent to this email address.
              </p>

              <div className='space-y-4'>
                {ccFields.map((field, index) => (
                  <div key={field.id} className='flex flex-row gap-4 items-end'>
                    <FormField
                      control={form.control}
                      name={`ccAddresses.${index}.email`}
                      render={({ field: formField }) => (
                        <FormItem className='flex-1'>
                          <FormLabel>CC Email</FormLabel>
                          <FormControl>
                            <Input placeholder='i.e. cc@management.de' {...formField} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name={`ccAddresses.${index}.name`}
                      render={({ field: formField }) => (
                        <FormItem className='flex-1'>
                          <FormLabel>CC Name</FormLabel>
                          <FormControl>
                            <Input placeholder='i.e. CC Recipient' {...formField} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <Button
                      variant='destructive'
                      type='button'
                      onClick={() => handleRemoveCC(index)}
                    >
                      <Trash2 className='h-4 w-4' />
                    </Button>
                  </div>
                ))}

                {/* Add new CC entry button */}
                <Button
                  variant='outline'
                  type='button'
                  onClick={() => appendCC({ email: '', name: '' })}
                >
                  <Plus className='mr-2 h-4 w-4' />
                  Add CC
                </Button>
              </div>
            </div>

            <Separator />

            {/* BCC Settings */}
            <div>
              <h3 className='text-lg font-semibold mb-2 flex items-center'>
                <EyeOff className='mr-2' size={20} />
                BCC Settings
              </h3>
              <p className='text-sm text-muted-foreground mb-4'>
                BCC recipients will receive a copy of every email without other recipients knowing.
              </p>

              <div className='space-y-4'>
                {bccFields.map((field, index) => (
                  <div key={field.id} className='flex flex-row gap-4 items-end'>
                    <FormField
                      control={form.control}
                      name={`bccAddresses.${index}.email`}
                      render={({ field: formField }) => (
                        <FormItem className='flex-1'>
                          <FormLabel>BCC Email</FormLabel>
                          <FormControl>
                            <Input placeholder='i.e. bcc@management.de' {...formField} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name={`bccAddresses.${index}.name`}
                      render={({ field: formField }) => (
                        <FormItem className='flex-1'>
                          <FormLabel>BCC Name</FormLabel>
                          <FormControl>
                            <Input placeholder='i.e. BCC Recipient' {...formField} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <Button
                      variant='destructive'
                      type='button'
                      onClick={() => handleRemoveBCC(index)}
                    >
                      <Trash2 className='h-4 w-4' />
                    </Button>
                  </div>
                ))}

                {/* Add new BCC entry button */}
                <Button
                  variant='outline'
                  type='button'
                  onClick={() => appendBCC({ email: '', name: '' })}
                >
                  <Plus className='mr-2 h-4 w-4' />
                  Add BCC
                </Button>
              </div>
            </div>
          </CardContent>

          <CardFooter className='flex justify-end'>
            <Button type='submit' disabled={!isModified || isPending} className='w-full sm:w-auto'>
              {isPending ? (
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
