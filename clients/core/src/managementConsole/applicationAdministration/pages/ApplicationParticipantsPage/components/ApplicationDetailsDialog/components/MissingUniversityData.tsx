import React from 'react'
import { useForm } from 'react-hook-form'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, XCircle } from 'lucide-react'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  Alert,
  AlertDescription,
  Button,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
} from '@tumaet/prompt-ui-components'
import {
  UniversityDataFormValues,
  formSchemaUniversityData,
} from '@core/validations/universityData'
import { Student } from '@tumaet/prompt-shared-state'
import { updateStudent } from '@core/network/mutations/updateStudent'
import translations from '@/lib/translations.json'

interface MissingUniversityDataProps {
  student: Student
}

export const MissingUniversityData = ({ student }: MissingUniversityDataProps) => {
  const [isAddingData, setIsAddingData] = React.useState(false)
  const queryClient = useQueryClient()

  const {
    mutate: mutateStudent,
    isError: isMutateError,
    reset: resetMutation,
  } = useMutation({
    mutationFn: (modifiedStudent: Student) => {
      return updateStudent(modifiedStudent)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['application'] })
      queryClient.invalidateQueries({ queryKey: ['application_participations'] })
      setIsAddingData(false)
    },
  })

  const form = useForm<UniversityDataFormValues>({
    resolver: zodResolver(formSchemaUniversityData),
    defaultValues: {
      universityLogin: student.universityLogin || '',
      matriculationNumber: student.matriculationNumber || '',
      email: student.email || '',
    },
  })

  function onSubmit(values: UniversityDataFormValues) {
    mutateStudent({ ...student, ...values, hasUniversityAccount: true })
  }

  const handleRetry = () => {
    resetMutation()
    form.reset()
  }

  return (
    <Alert variant={isMutateError ? 'destructive' : undefined}>
      <div className='flex items-center justify-between'>
        <AlertDescription className='flex items-center'>
          {isMutateError ? (
            <XCircle className='h-4 w-4 mr-2' />
          ) : (
            <AlertCircle className='h-4 w-4 mr-2' />
          )}
          {isMutateError
            ? 'Failed to update university data.'
            : 'The student has no university login.'}
        </AlertDescription>
        {!isAddingData && !isMutateError && (
          <Button onClick={() => setIsAddingData(true)} variant='outline' size='sm'>
            Add University Data
          </Button>
        )}
        {isMutateError && (
          <Button onClick={handleRetry} variant='outline' size='sm'>
            Retry
          </Button>
        )}
      </div>
      {isAddingData && !isMutateError && (
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              <FormField
                control={form.control}
                name='matriculationNumber'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Matriculation Number</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='universityLogin'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{translations.university['login-name']}</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <FormField
              control={form.control}
              name='email'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input {...field} type='email' />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className='flex justify-end space-x-2'>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                onClick={() => setIsAddingData(false)}
              >
                Cancel
              </Button>
              <Button type='submit' size='sm'>
                Submit
              </Button>
            </div>
          </form>
        </Form>
      )}
    </Alert>
  )
}
