import {
  Student,
  Gender,
  getGenderString,
  StudyDegree,
  getStudyDegreeString,
} from '@tumaet/prompt-shared-state'
import { forwardRef, useEffect, useImperativeHandle, useState } from 'react'
import { StudentComponentRef } from '../../utils/StudentComponentRef'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  Button,
  Input,
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
  cn,
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@tumaet/prompt-ui-components'
import { studentSchema, StudentFormValues } from '@core/validations/student'
import { translations } from '@tumaet/prompt-shared-state'
import { Check, ChevronDown } from 'lucide-react'
import { countriesArr } from '@tumaet/prompt-shared-state'

const studyPrograms = translations.university.studyPrograms.concat('Other')

interface StudentFormProps {
  student: Student
  isInstructorView?: boolean
  allowEditUniversityData: boolean
  onUpdate: (updatedStudent: Student) => void
}

export const StudentForm = forwardRef<StudentComponentRef, StudentFormProps>(function StudentForm(
  { student, isInstructorView = false, allowEditUniversityData, onUpdate },
  ref,
) {
  const hasUniversityAccount = student.hasUniversityAccount

  // this ugly setting is necessary due to typescript and two different validation schema
  const form = useForm<StudentFormValues>({
    resolver: zodResolver(studentSchema),
    defaultValues: hasUniversityAccount
      ? {
          matriculationNumber: student.matriculationNumber || '',
          universityLogin: student.universityLogin || '',
          firstName: student.firstName || '',
          lastName: student.lastName || '',
          email: student.email || '',
          gender: student.gender ?? undefined,
          nationality: student.nationality ?? '',
          studyDegree: student.studyDegree ?? undefined,
          studyProgram: student.studyProgram ?? '',
          currentSemester: student.currentSemester ?? undefined,
          hasUniversityAccount: true,
        }
      : {
          firstName: student.firstName || '',
          lastName: student.lastName || '',
          email: student.email || '',
          gender: student.gender ?? undefined,
          nationality: student.nationality ?? '',
          studyDegree: student.studyDegree ?? undefined,
          studyProgram: student.studyProgram ?? '',
          currentSemester: student.currentSemester ?? undefined,
          hasUniversityAccount: false,
        },

    mode: 'onChange',
  })

  useImperativeHandle(ref, () => ({
    async validate() {
      const valid = await form.trigger()
      return valid
    },
    rerender(updatedStudent: Student) {
      form.reset(
        updatedStudent.hasUniversityAccount
          ? {
              matriculationNumber: updatedStudent.matriculationNumber || '',
              universityLogin: updatedStudent.universityLogin || '',
              firstName: updatedStudent.firstName || '',
              lastName: updatedStudent.lastName || '',
              email: updatedStudent.email || '',
              gender: updatedStudent.gender ?? undefined,
              nationality: updatedStudent.nationality ?? '',
              studyDegree: updatedStudent.studyDegree ?? undefined,
              studyProgram: updatedStudent.studyProgram ?? '',
              currentSemester: updatedStudent.currentSemester ?? undefined,
              hasUniversityAccount: true,
            }
          : {
              firstName: updatedStudent.firstName || '',
              lastName: updatedStudent.lastName || '',
              email: updatedStudent.email || '',
              gender: updatedStudent.gender ?? undefined,
              nationality: updatedStudent.nationality ?? '',
              studyDegree: updatedStudent.studyDegree ?? undefined,
              studyProgram: updatedStudent.studyProgram ?? '',
              currentSemester: updatedStudent.currentSemester ?? undefined,
              hasUniversityAccount: false,
            },
      )
    },
  }))

  useEffect(() => {
    const subscription = form.watch((value) => {
      onUpdate({ ...student, ...value })
    })
    // Cleanup subscription on unmount
    return () => subscription.unsubscribe()
  }, [form.watch, student, onUpdate, form])

  const [otherStudyProgram, setOtherStudyProgram] = useState(false)
  const currStudyProgram = form.watch('studyProgram')

  const requiredStar = <span className='text-destructive'> *</span>

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit((data) => onUpdate({ ...student, ...data }))}
        className='space-y-6'
      >
        {hasUniversityAccount && (
          <div className='grid grid-cols-1 md:grid-cols-2 gap-6'>
            <FormField
              control={form.control}
              name='matriculationNumber'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Matriculation Number</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      disabled={
                        (!allowEditUniversityData && hasUniversityAccount) || isInstructorView
                      }
                    />
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
                  <FormLabel>
                    {translations.university['login-name']}
                    {requiredStar}
                  </FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      disabled={
                        (!allowEditUniversityData && hasUniversityAccount) || isInstructorView
                      }
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        )}
        <div className='grid grid-cols-1 md:grid-cols-2 gap-6'>
          <FormField
            control={form.control}
            name='firstName'
            render={({ field }) => (
              <FormItem>
                <FormLabel>First Name{requiredStar}</FormLabel>
                <FormControl>
                  <Input
                    {...field}
                    disabled={
                      (!allowEditUniversityData && hasUniversityAccount) || isInstructorView
                    }
                    maxLength={100}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='lastName'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Last Name{requiredStar}</FormLabel>
                <FormControl>
                  <Input
                    {...field}
                    disabled={
                      (!allowEditUniversityData && hasUniversityAccount) || isInstructorView
                    }
                    maxLength={100}
                  />
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
              <FormLabel>Email{requiredStar}</FormLabel>
              <FormControl>
                <Input
                  {...field}
                  type='email'
                  disabled={(!allowEditUniversityData && hasUniversityAccount) || isInstructorView}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className='grid grid-cols-1 md:grid-cols-2 gap-6'>
          <FormField
            control={form.control}
            name='gender'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Gender{requiredStar}</FormLabel>
                <Select
                  onValueChange={field.onChange}
                  defaultValue={field.value}
                  disabled={isInstructorView}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder='Select a gender' />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {Object.values(Gender).map((gender) => (
                      <SelectItem key={gender} value={gender}>
                        {getGenderString(gender)}
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
            name='nationality'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Nationality{requiredStar}</FormLabel>
                <Popover>
                  <PopoverTrigger asChild>
                    <FormControl>
                      <Button
                        variant='outline'
                        role='combobox'
                        className={cn(
                          'w-full justify-between',
                          !field.value && 'text-muted-foreground',
                        )}
                        disabled={isInstructorView}
                      >
                        {field.value
                          ? countriesArr.find((country) => country.value === field.value)?.label
                          : 'Select a nationality'}
                        <ChevronDown className='ml-2 h-4 w-4 shrink-0 opacity-50' />
                      </Button>
                    </FormControl>
                  </PopoverTrigger>
                  <PopoverContent className='w-full p-0'>
                    <Command>
                      <CommandInput placeholder='Search nationality...' />
                      <CommandList>
                        <CommandEmpty>No country found.</CommandEmpty>
                        <CommandGroup>
                          {countriesArr.map((country) => (
                            <CommandItem
                              value={country.label}
                              key={country.value}
                              onSelect={() => {
                                form.setValue('nationality', country.value)
                              }}
                            >
                              {country.label}
                              <Check
                                className={cn(
                                  'ml-auto',
                                  country.value === field.value ? 'opacity-100' : 'opacity-0',
                                )}
                              />
                            </CommandItem>
                          ))}
                        </CommandGroup>
                      </CommandList>
                    </Command>
                  </PopoverContent>
                </Popover>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <div className='grid grid-cols-1 md:grid-cols-3 gap-6'>
          <FormField
            control={form.control}
            name='studyDegree'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Study Degree{requiredStar}</FormLabel>
                <Select
                  onValueChange={field.onChange}
                  defaultValue={field.value}
                  disabled={isInstructorView}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder='Select your current study degree' />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {Object.values(StudyDegree).map((degree) => (
                      <SelectItem key={degree} value={degree}>
                        {getStudyDegreeString(degree)}
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
            name='studyProgram'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Study Program{requiredStar}</FormLabel>
                <Select
                  onValueChange={(value) => {
                    if (value === 'Other') {
                      setOtherStudyProgram(true)
                      form.setValue('studyProgram', '')
                    } else {
                      setOtherStudyProgram(false)
                      form.setValue('studyProgram', value)
                    }
                  }}
                  defaultValue={field.value}
                  disabled={isInstructorView}
                  value={
                    otherStudyProgram || (field.value != '' && !studyPrograms.includes(field.value))
                      ? 'Other'
                      : field.value
                  }
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder='Select your current study program' />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {studyPrograms.map((program) => (
                      <SelectItem key={program} value={program}>
                        {program}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {!otherStudyProgram && <FormMessage />}
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='currentSemester'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Current Semester{requiredStar}</FormLabel>
                <FormControl>
                  <Input
                    {...field}
                    disabled={isInstructorView}
                    type='number'
                    maxLength={2}
                    pattern='\d{1,2}'
                    placeholder='Bachelor (+ Master) Semesters'
                    onChange={(e) => {
                      const value = parseInt(e.target.value)
                      field.onChange(value)
                    }}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        {(otherStudyProgram ||
          (currStudyProgram != '' && !studyPrograms.includes(currStudyProgram))) && (
          <FormField
            control={form.control}
            name='studyProgram'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Specify Study Program{requiredStar}</FormLabel>
                <FormControl>
                  <Input
                    {...field}
                    disabled={isInstructorView}
                    placeholder='Please enter your other study program'
                    maxLength={100}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        )}
      </form>
    </Form>
  )
})
