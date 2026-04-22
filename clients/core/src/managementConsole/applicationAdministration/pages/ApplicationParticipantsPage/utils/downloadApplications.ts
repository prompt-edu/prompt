import { AdditionalScore } from '../../../interfaces/additionalScore/additionalScore'
import { ApplicationParticipation } from '../../../interfaces/applicationParticipation'
import { ApplicationForm } from '../../../interfaces/form/applicationForm'
import {
  ApplicationCsvExportSettings,
  shouldExportQuestionToCsv,
} from '../../../utils/applicationCsvExportSettings'
import { GetApplication } from '@core/interfaces/application/getApplication'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import { saveAs } from 'file-saver'

type ApplicationQuestion =
  | ApplicationQuestionText
  | ApplicationQuestionMultiSelect
  | ApplicationQuestionFileUpload

interface CsvColumn {
  header: string
  getValue: (row: ApplicationParticipation) => unknown
}

export type ApplicationDetailsByParticipationID = Record<string, GetApplication>

const getApplicationQuestions = (applicationForm?: ApplicationForm): ApplicationQuestion[] => {
  return [
    ...(applicationForm?.questionsMultiSelect ?? []),
    ...(applicationForm?.questionsText ?? []),
    ...(applicationForm?.questionsFileUpload ?? []),
  ].sort((a, b) => a.orderNum - b.orderNum)
}

const getQuestionAnswerValue = (
  question: ApplicationQuestion,
  application: GetApplication | undefined,
): string => {
  if (!application) {
    return ''
  }

  if ('allowedFileTypes' in question) {
    return (
      application.answersFileUpload.find((answer) => answer.applicationQuestionID === question.id)
        ?.fileName ?? ''
    )
  }

  if ('options' in question) {
    return (
      application.answersMultiSelect
        .find((answer) => answer.applicationQuestionID === question.id)
        ?.answer.join(', ') ?? ''
    )
  }

  return (
    application.answersText.find((answer) => answer.applicationQuestionID === question.id)
      ?.answer ?? ''
  )
}

export const buildApplicationCsvContent = (
  data: ApplicationParticipation[],
  additionalScores: AdditionalScore[] = [],
  applicationForm?: ApplicationForm,
  applicationsByParticipationID: ApplicationDetailsByParticipationID = {},
  csvExportSettings: ApplicationCsvExportSettings = {},
): string => {
  const baseColumns: CsvColumn[] = [
    { header: 'firstName', getValue: (row) => row.student?.firstName },
    { header: 'lastName', getValue: (row) => row.student?.lastName },
    { header: 'email', getValue: (row) => row.student?.email },
    { header: 'matriculationNumber', getValue: (row) => row.student?.matriculationNumber },
    { header: 'universityLogin', getValue: (row) => row.student?.universityLogin },
    { header: 'hasUniversityAccount', getValue: (row) => row.student?.hasUniversityAccount },
    { header: 'gender', getValue: (row) => row.student?.gender },
    { header: 'passStatus', getValue: (row) => row.passStatus },
    { header: 'score', getValue: (row) => row.score },
  ]

  const additionalScoreColumns: CsvColumn[] = additionalScores.map((score) => ({
    header: score.key,
    getValue: (row) => row.restrictedData?.[score.key],
  }))

  const questionColumns: CsvColumn[] = getApplicationQuestions(applicationForm)
    .filter((question) => shouldExportQuestionToCsv(csvExportSettings, question.id))
    .map((question) => ({
      header: question.title,
      getValue: (row) =>
        getQuestionAnswerValue(question, applicationsByParticipationID[row.courseParticipationID]),
    }))

  const csvColumns = [...baseColumns, ...additionalScoreColumns, ...questionColumns]
  const stringifiedHeaders = csvColumns.map((column) => JSON.stringify(column.header))
  const csvRows = data.map((row) =>
    csvColumns.map((column) => JSON.stringify(column.getValue(row) ?? '')).join(';'),
  )

  return [stringifiedHeaders.join(';'), ...csvRows].join('\n')
}

export const downloadApplications = (
  data: ApplicationParticipation[],
  additionalScores: AdditionalScore[] = [],
  filename = 'application-export.csv',
  applicationForm?: ApplicationForm,
  applicationsByParticipationID: ApplicationDetailsByParticipationID = {},
  csvExportSettings: ApplicationCsvExportSettings = {},
) => {
  if (!data || data.length === 0) {
    console.error('No data available to download.')
    return
  }

  const csvContent = buildApplicationCsvContent(
    data,
    additionalScores,
    applicationForm,
    applicationsByParticipationID,
    csvExportSettings,
  )

  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  saveAs(blob, filename)
}
