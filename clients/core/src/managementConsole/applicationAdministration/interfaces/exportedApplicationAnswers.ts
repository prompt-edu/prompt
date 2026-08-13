export interface ExportedAnswerColumn {
  questionID: string
  key: string
  title: string
  orderNum: number
  type: 'text' | 'multiselect'
}

export interface ExportedAnswer {
  questionID: string
  answer: string
}

export interface ParticipationExportedAnswers {
  courseParticipationID: string
  answers: ExportedAnswer[]
}

export interface ExportedApplicationAnswersResponse {
  columns: ExportedAnswerColumn[]
  answers: ParticipationExportedAnswers[]
}
