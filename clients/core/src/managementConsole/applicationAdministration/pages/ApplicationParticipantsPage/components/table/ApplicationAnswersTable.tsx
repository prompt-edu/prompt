import { ApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/applicationAnswerFileUpload'
import { ApplicationAnswerMultiSelect } from '@core/interfaces/application/applicationAnswer/multiSelect/applicationAnswerMultiSelect'
import { ApplicationAnswerText } from '@core/interfaces/application/applicationAnswer/text/applicationAnswerText'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { getApplicationFileDownloadUrl } from '@core/network/queries/applicationFileDownloadUrl'
import { formatFileSize, openFileDownload } from '@tumaet/prompt-shared-state'
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@tumaet/prompt-ui-components'
import { AlignLeft, CheckSquare, Download, Paperclip } from 'lucide-react'
import { useState } from 'react'

interface ApplicationAnswersTableProps {
  coursePhaseId: string
  questions: (
    | ApplicationQuestionText
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionFileUpload
  )[]
  answersMultiSelect: ApplicationAnswerMultiSelect[]
  answersText: ApplicationAnswerText[]
  answersFileUpload: ApplicationAnswerFileUpload[]
}

export const ApplicationAnswersTable = ({
  coursePhaseId,
  questions,
  answersMultiSelect,
  answersText,
  answersFileUpload,
}: ApplicationAnswersTableProps) => {
  const [downloadingFileId, setDownloadingFileId] = useState<string | null>(null)
  const sortedQuestions = [...questions].sort((a, b) => a.orderNum - b.orderNum)

  const handleDownload = async (fileAnswer: ApplicationAnswerFileUpload) => {
    if (!coursePhaseId || !fileAnswer.fileID) {
      return
    }

    try {
      setDownloadingFileId(fileAnswer.fileID)
      const downloadUrl = await getApplicationFileDownloadUrl(coursePhaseId, fileAnswer.fileID)
      await openFileDownload({
        downloadUrl,
        fileName: fileAnswer.fileName,
      })
    } catch (error) {
      console.error('Failed to download file:', error)
    } finally {
      setDownloadingFileId(null)
    }
  }

  return (
    <Card className='w-full'>
      <CardHeader>
        <CardTitle>Application Answers</CardTitle>
        <CardDescription>
          Review the applicant&apos;s responses to the application questions.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className='overflow-x-auto'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className='w-1/12'>Type</TableHead>
                <TableHead className='w-1/4'>Question</TableHead>
                <TableHead className='w-1/4'>Description</TableHead>
                <TableHead>Answer</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {sortedQuestions.map((question, index) => {
                const isFileUpload = 'allowedFileTypes' in question
                const isMultiSelect = 'options' in question
                const fileAnswer = isFileUpload
                  ? answersFileUpload.find((a) => a.applicationQuestionID === question.id)
                  : undefined
                const textAnswer = !isFileUpload
                  ? answersText.find((a) => a.applicationQuestionID === question.id)?.answer || ''
                  : ''
                const multiSelectAnswer = isMultiSelect
                  ? answersMultiSelect.find((a) => a.applicationQuestionID === question.id)
                      ?.answer || []
                  : []

                return (
                  <TableRow key={question.id} className={index % 2 === 0 ? 'bg-muted/50' : ''}>
                    <TableCell>
                      {isFileUpload ? (
                        <Paperclip className='h-4 w-4 text-muted-foreground' />
                      ) : isMultiSelect ? (
                        <CheckSquare className='h-4 w-4 text-muted-foreground' />
                      ) : (
                        <AlignLeft className='h-4 w-4 text-muted-foreground' />
                      )}
                    </TableCell>

                    <TableCell className='font-medium'>
                      <div className='flex items-center space-x-2'>
                        <span>{question.title}</span>
                      </div>
                    </TableCell>
                    <TableCell className='text-muted-foreground'>
                      <div className='flex items-center space-x-2'>
                        <span>{question.description}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      {isFileUpload ? (
                        fileAnswer ? (
                          <div className='flex flex-col gap-2'>
                            <div>
                              <div className='font-medium'>{fileAnswer.fileName}</div>
                              <div className='text-xs text-muted-foreground'>
                                {formatFileSize(fileAnswer.fileSize)}
                              </div>
                            </div>
                            <Button
                              variant='outline'
                              size='sm'
                              onClick={() => handleDownload(fileAnswer)}
                              disabled={
                                !coursePhaseId ||
                                !fileAnswer.fileID ||
                                downloadingFileId === fileAnswer.fileID
                              }
                            >
                              <Download className='mr-2 h-4 w-4' />
                              Download
                            </Button>
                          </div>
                        ) : (
                          <span className='text-muted-foreground'>No file uploaded</span>
                        )
                      ) : isMultiSelect ? (
                        <div>
                          {multiSelectAnswer.map((item, idx) => (
                            <Badge key={idx} className='mr-1'>
                              {item}
                            </Badge>
                          ))}
                        </div>
                      ) : (
                        <p className='whitespace-pre-wrap wrap-break-word'>{textAnswer}</p>
                      )}
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  )
}
