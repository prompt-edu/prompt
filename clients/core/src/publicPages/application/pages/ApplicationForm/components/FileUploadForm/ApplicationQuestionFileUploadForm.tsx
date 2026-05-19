import { forwardRef, useEffect, useImperativeHandle, useState } from 'react'
import { FileUpload } from '@tumaet/prompt-ui-components'
import { FileList } from '@tumaet/prompt-ui-components'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import { ApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/applicationAnswerFileUpload'
import { CreateApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/createApplicationAnswerFileUpload'
import { FileResponse } from '@tumaet/prompt-shared-state'
import { deleteApplicationFile } from '@tumaet/prompt-shared-state'
import { Alert, AlertDescription, CardTitle } from '@tumaet/prompt-ui-components'
import { FormDescriptionHTML } from '../FormDescriptionHTML'

export interface QuestionFileUploadFormRef {
  validate: () => Promise<boolean>
  getValues: () => CreateApplicationAnswerFileUpload
  rerender: (answer?: ApplicationAnswerFileUpload) => void
}

interface ApplicationQuestionFileUploadFormProps {
  question: ApplicationQuestionFileUpload
  answer?: ApplicationAnswerFileUpload
  isInstructorView?: boolean
  applicationId?: string
  coursePhaseId?: string
}

export const ApplicationQuestionFileUploadForm = forwardRef<
  QuestionFileUploadFormRef,
  ApplicationQuestionFileUploadFormProps
>(({ question, answer, isInstructorView = false, applicationId, coursePhaseId }, ref) => {
  const [uploadedFile, setUploadedFile] = useState<FileResponse | null>(null)
  const [existingAnswer, setExistingAnswer] = useState<ApplicationAnswerFileUpload | null>(
    answer ?? null,
  )
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setExistingAnswer(answer ?? null)
    // After a successful submit, the backend answer arrives via props.
    // Clear local upload state to avoid rendering the same file twice.
    if (answer) {
      setUploadedFile(null)
    }
  }, [answer])

  const canDeleteFiles = !isInstructorView && !!coursePhaseId && !!applicationId
  const shouldShowUploader =
    !isInstructorView && !existingAnswer && (!uploadedFile || !canDeleteFiles)

  useImperativeHandle(ref, () => ({
    validate: async () => {
      if (question.isRequired && !uploadedFile && !existingAnswer) {
        setError('This file upload is required')
        return false
      }
      setError(null)
      return true
    },
    getValues: () => ({
      applicationQuestionID: question.id,
      fileID: uploadedFile?.id || existingAnswer?.fileID || '',
    }),
    rerender: () => {
      // If we have a new answer, we don't need to do anything as the file is already uploaded
      // The FileList component will show the uploaded files
    },
  }))

  const handleUploadSuccess = (file: FileResponse) => {
    setUploadedFile(file)
    setError(null)
  }

  const handleUploadError = (err: Error) => {
    setError(err.message)
  }

  const handleDelete = async (fileId: string) => {
    if (!coursePhaseId) {
      setError('Course phase ID missing, cannot delete file.')
      return
    }

    try {
      await deleteApplicationFile(coursePhaseId, fileId)
      setUploadedFile(null)
      setExistingAnswer(null)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete file.'
      setError(message)
    }
  }

  return (
    <div className='space-y-4'>
      <div>
        <CardTitle className='text-lg'>{question.title}</CardTitle>
        {!isInstructorView && question.description && (
          <FormDescriptionHTML htmlCode={question.description} />
        )}
        {question.isRequired && <span className='text-red-500 ml-1'>*</span>}
      </div>

      {canDeleteFiles && (uploadedFile || existingAnswer) && (
        <p className='text-sm text-muted-foreground'>
          Only one file can be uploaded. Delete the current file before uploading a new one.
        </p>
      )}

      {shouldShowUploader && (
        <FileUpload
          coursePhaseId={coursePhaseId}
          accept={question.allowedFileTypes}
          maxSizeMB={question.maxFileSizeMB || 50}
          onSuccess={handleUploadSuccess}
          onError={handleUploadError}
        />
      )}

      {uploadedFile && (
        <div className='mt-4'>
          <p className='text-sm font-medium mb-2'>Uploaded file:</p>
          <FileList files={[uploadedFile]} allowDelete={canDeleteFiles} onDelete={handleDelete} />
        </div>
      )}

      {existingAnswer && (
        <div className='mt-4'>
          <p className='text-sm font-medium mb-2'>Previously uploaded file:</p>
          <FileList
            files={[
              {
                id: existingAnswer.fileID,
                filename: existingAnswer.fileName,
                originalFilename: existingAnswer.fileName,
                contentType: '',
                sizeBytes: existingAnswer.fileSize,
                storageKey: '',
                downloadUrl: existingAnswer.downloadUrl,
                uploadedByUserId: '',
                createdAt: existingAnswer.uploadedAt,
                updatedAt: existingAnswer.uploadedAt,
              },
            ]}
            allowDelete={canDeleteFiles}
            onDelete={handleDelete}
          />
        </div>
      )}

      {error && (
        <Alert variant='destructive'>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
    </div>
  )
})

ApplicationQuestionFileUploadForm.displayName = 'ApplicationQuestionFileUploadForm'
