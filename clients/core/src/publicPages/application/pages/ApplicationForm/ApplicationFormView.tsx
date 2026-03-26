import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Separator,
} from '@tumaet/prompt-ui-components'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import { ApplicationAnswerText } from '@core/interfaces/application/applicationAnswer/text/applicationAnswerText'
import { CreateApplicationAnswerText } from '@core/interfaces/application/applicationAnswer/text/createApplicationAnswerText'
import { ApplicationAnswerMultiSelect } from '@core/interfaces/application/applicationAnswer/multiSelect/applicationAnswerMultiSelect'
import { CreateApplicationAnswerMultiSelect } from '@core/interfaces/application/applicationAnswer/multiSelect/createApplicationAnswerMultiSelect'
import { ApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/applicationAnswerFileUpload'
import { CreateApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/createApplicationAnswerFileUpload'
import { Student } from '@tumaet/prompt-shared-state'
import { useEffect, useRef, useState } from 'react'
import { StudentForm } from './components/StudentForm/StudentForm'
import { ApplicationQuestionTextForm } from './components/TextForm/ApplicationQuestionTextForm'
import { ApplicationQuestionFileUploadForm } from './components/FileUploadForm/ApplicationQuestionFileUploadForm'
import { QuestionTextFormRef } from './utils/QuestionTextFormRef'
import { QuestionMultiSelectFormRef } from './utils/QuestionMultiSelectFormRef'
import { QuestionFileUploadFormRef } from './utils/QuestionFileUploadFormRef'
import { StudentComponentRef } from './utils/StudentComponentRef'
import { ApplicationQuestionMultiSelectForm } from './components/MultiSelectForm/ApplicationQuestionMultiSelectForm'

interface ApplicationFormProps {
  questionsText: ApplicationQuestionText[]
  questionsMultiSelect: ApplicationQuestionMultiSelect[]
  questionsFileUpload: ApplicationQuestionFileUpload[]
  initialAnswersText?: ApplicationAnswerText[]
  initialAnswersMultiSelect?: ApplicationAnswerMultiSelect[]
  initialAnswersFileUpload?: ApplicationAnswerFileUpload[]
  student?: Student
  isInstructorView?: boolean
  allowEditUniversityData?: boolean
  applicationId?: string
  coursePhaseId?: string
  onSubmit: (
    student: Student,
    answersText: CreateApplicationAnswerText[],
    answersMultiSelect: CreateApplicationAnswerMultiSelect[],
    answersFileUpload: CreateApplicationAnswerFileUpload[],
  ) => void
}

export const ApplicationFormView = ({
  questionsText,
  questionsMultiSelect,
  questionsFileUpload,
  initialAnswersMultiSelect,
  initialAnswersText,
  initialAnswersFileUpload,
  student,
  isInstructorView = false,
  allowEditUniversityData = false,
  applicationId,
  coursePhaseId,
  onSubmit,
}: ApplicationFormProps) => {
  const questions: (
    | ApplicationQuestionText
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionFileUpload
  )[] = [...questionsText, ...questionsMultiSelect, ...questionsFileUpload].sort(
    (a, b) => a.orderNum - b.orderNum,
  )

  const [studentData, setStudentData] = useState<Student>(student ?? ({} as Student))
  const studentRef = useRef<StudentComponentRef>(null)
  const questionTextRefs = useRef<Array<QuestionTextFormRef | null | undefined>>([])
  const questionMultiSelectRefs = useRef<Array<QuestionMultiSelectFormRef | null | undefined>>([])
  const questionFileUploadRefs = useRef<Array<QuestionFileUploadFormRef | null | undefined>>([])
  const [validationFailed, setValidationFailed] = useState(false)

  // correctly propagate student data changes
  const hasInitialized = useRef(false)

  useEffect(() => {
    if (!hasInitialized.current && student) {
      setStudentData(student)
      studentRef.current?.rerender(student)
      hasInitialized.current = true
    }
  }, [student])

  const handleSubmit = async () => {
    let allValid = true

    if (!studentRef.current) {
      console.log('Student ref is not set')
      return
    }
    const studentValid = await studentRef.current.validate()
    if (studentData && !studentValid) {
      allValid = false
    }

    // Loop over each child's ref, call validate()
    const answersText: CreateApplicationAnswerText[] = []
    for (const ref of questionTextRefs.current) {
      if (!ref) continue
      const isValid = await ref.validate()
      if (isValid) {
        answersText.push(ref.getValues())
      } else {
        allValid = false
      }
    }

    const answersMultiSelect: CreateApplicationAnswerMultiSelect[] = []
    for (const ref of questionMultiSelectRefs.current) {
      if (!ref) continue
      const isValid = await ref.validate()
      if (isValid) {
        answersMultiSelect.push(ref.getValues())
      } else {
        allValid = false
      }
    }

    const answersFileUpload: CreateApplicationAnswerFileUpload[] = []
    for (const ref of questionFileUploadRefs.current) {
      if (!ref) continue
      const isValid = await ref.validate()
      if (isValid) {
        const value = ref.getValues()
        if (value.fileID) {
          answersFileUpload.push(value)
        }
      } else {
        allValid = false
      }
    }

    if (!allValid) {
      setValidationFailed(true)
      return
    } else {
      setValidationFailed(false)
    }
    // call onSubmit
    onSubmit(studentData, answersText, answersMultiSelect, answersFileUpload)
  }

  return (
    <Card className={validationFailed ? 'border-red-500' : ''}>
      <CardHeader>
        <CardTitle>Application Form{isInstructorView}</CardTitle>
        {isInstructorView && (
          <div className='text-sm text-muted-foreground'>
            This form is in read-only mode because you are viewing this application as an
            instructor. Further, the input field descriptions are hidden for better readability.
          </div>
        )}
      </CardHeader>
      <CardContent>
        <div
          className={`${
            isInstructorView
              ? // On instructor view + large screen, we use grid with 2 columns
                'space-y-8 lg:grid lg:grid-cols-2 lg:gap-8 lg:space-y-0'
              : // Otherwise, maintain the original vertical spacing
                'space-y-8'
          }`}
        >
          <div>
            <h2 className='text-lg font-semibold mb-2'>Personal Information</h2>
            {!isInstructorView && (
              <p className='text-sm text-muted-foreground mb-4'>
                This information will be applied for all applications at this chair.
              </p>
            )}
            <StudentForm
              student={studentData}
              onUpdate={setStudentData}
              ref={studentRef}
              isInstructorView={isInstructorView}
              allowEditUniversityData={allowEditUniversityData}
            />
          </div>
          {!isInstructorView && <Separator />}

          <div>
            <h2 className='text-lg font-semibold mb-4'>Course Specific Questions</h2>
            {questions.map((question, index) => (
              <div key={question.id} className='mb-6'>
                {'options' in question ? (
                  <ApplicationQuestionMultiSelectForm
                    question={question}
                    initialAnswers={
                      initialAnswersMultiSelect?.find(
                        (a) => a.applicationQuestionID === question.id,
                      )?.answer ?? []
                    }
                    ref={(el) => {
                      questionMultiSelectRefs.current[index] = el
                    }}
                    isInstructorView={isInstructorView}
                  />
                ) : 'allowedFileTypes' in question ? (
                  <ApplicationQuestionFileUploadForm
                    question={question}
                    answer={initialAnswersFileUpload?.find(
                      (a) => a.applicationQuestionID === question.id,
                    )}
                    ref={(el) => {
                      questionFileUploadRefs.current[index] = el
                    }}
                    isInstructorView={isInstructorView}
                    applicationId={applicationId}
                    coursePhaseId={coursePhaseId}
                  />
                ) : (
                  <ApplicationQuestionTextForm
                    question={question}
                    initialAnswer={
                      initialAnswersText?.find((a) => a.applicationQuestionID === question.id)
                        ?.answer ?? ''
                    }
                    ref={(el) => {
                      questionTextRefs.current[index] = el
                    }}
                    isInstructorView={isInstructorView}
                  />
                )}
              </div>
            ))}
          </div>

          {!isInstructorView && (
            <div className='flex justify-end'>
              <Button onClick={handleSubmit} disabled={isInstructorView}>
                Submit
              </Button>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
