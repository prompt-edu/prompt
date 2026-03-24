import { useState, useEffect } from 'react'
import { CoursePhaseConfig } from '../../../interfaces/coursePhaseConfig'
import { useCoursePhaseConfigStore } from '../../../zustand/useCoursePhaseConfigStore'

export interface MainConfigState {
  assessmentSchemaId: string
  start?: Date
  deadline?: Date
  evaluationResultsVisible: boolean
  gradeSuggestionVisible: boolean
  actionItemsVisible: boolean
  gradingSheetVisible: boolean
}

export const useCoursePhaseConfigForm = () => {
  const [assessmentSchemaId, setAssessmentSchemaId] = useState<string>('')
  const [start, setStart] = useState<Date | undefined>(undefined)
  const [deadline, setDeadline] = useState<Date | undefined>(undefined)
  const [evaluationResultsVisible, setEvaluationResultsVisible] = useState<boolean>(false)
  const [gradeSuggestionVisible, setGradeSuggestionVisible] = useState<boolean>(true)
  const [actionItemsVisible, setActionItemsVisible] = useState<boolean>(true)
  const [gradingSheetVisible, setGradingSheetVisible] = useState<boolean>(false)

  const { coursePhaseConfig: config } = useCoursePhaseConfigStore()

  useEffect(() => {
    if (config) {
      setAssessmentSchemaId(config.assessmentSchemaID || '')
      setStart(config.start ? new Date(config.start) : undefined)
      setDeadline(config.deadline ? new Date(config.deadline) : undefined)
      setEvaluationResultsVisible(config.evaluationResultsVisible || false)
      setGradeSuggestionVisible(config.gradeSuggestionVisible ?? true)
      setActionItemsVisible(config.actionItemsVisible ?? true)
      setGradingSheetVisible(config.gradingSheetVisible ?? false)
    }
  }, [config])

  const mainConfigState: MainConfigState = {
    assessmentSchemaId,
    start,
    deadline,
    evaluationResultsVisible,
    gradeSuggestionVisible,
    actionItemsVisible,
    gradingSheetVisible,
  }

  const hasMainConfigChanges = (originalConfig?: CoursePhaseConfig) => {
    if (!originalConfig) return true // Changed from false to true to enable saving when no config exists yet

    return (
      assessmentSchemaId !== (originalConfig.assessmentSchemaID || '') ||
      start?.getTime() !==
        (originalConfig.start ? new Date(originalConfig.start).getTime() : undefined) ||
      deadline?.getTime() !==
        (originalConfig.deadline ? new Date(originalConfig.deadline).getTime() : undefined) ||
      evaluationResultsVisible !== (originalConfig.evaluationResultsVisible || false) ||
      gradeSuggestionVisible !== (originalConfig.gradeSuggestionVisible ?? true) ||
      actionItemsVisible !== (originalConfig.actionItemsVisible ?? true) ||
      gradingSheetVisible !== (originalConfig.gradingSheetVisible ?? false)
    )
  }

  return {
    assessmentSchemaId,
    setAssessmentSchemaId,
    start,
    setStart,
    deadline,
    setDeadline,
    evaluationResultsVisible,
    setEvaluationResultsVisible,
    gradeSuggestionVisible,
    setGradeSuggestionVisible,
    actionItemsVisible,
    setActionItemsVisible,
    gradingSheetVisible,
    setGradingSheetVisible,
    mainConfigState,
    hasMainConfigChanges,
  }
}
