import { useState, useEffect } from 'react'
import { CoursePhaseConfig } from '../../../interfaces/coursePhaseConfig'
import { useCoursePhaseConfigStore } from '../../../zustand/useCoursePhaseConfigStore'
import { EvaluationOptions } from '../interfaces/EvaluationOption'

export const useEvaluationOptions = () => {
  const [selfEvaluationEnabled, setSelfEvaluationEnabled] = useState<boolean>(false)
  const [selfEvaluationSchema, setSelfEvaluationSchema] = useState<string>('')
  const [selfEvaluationStart, setSelfEvaluationStart] = useState<Date | undefined>(undefined)
  const [selfEvaluationDeadline, setSelfEvaluationDeadline] = useState<Date | undefined>(undefined)

  const [peerEvaluationEnabled, setPeerEvaluationEnabled] = useState<boolean>(false)
  const [peerEvaluationSchema, setPeerEvaluationSchema] = useState<string>('')
  const [peerEvaluationStart, setPeerEvaluationStart] = useState<Date | undefined>(undefined)
  const [peerEvaluationDeadline, setPeerEvaluationDeadline] = useState<Date | undefined>(undefined)

  const [tutorEvaluationEnabled, setTutorEvaluationEnabled] = useState<boolean>(false)
  const [tutorEvaluationSchema, setTutorEvaluationSchema] = useState<string>('')
  const [tutorEvaluationStart, setTutorEvaluationStart] = useState<Date | undefined>(undefined)
  const [tutorEvaluationDeadline, setTutorEvaluationDeadline] = useState<Date | undefined>(
    undefined,
  )

  const { coursePhaseConfig: config } = useCoursePhaseConfigStore()

  useEffect(() => {
    if (config) {
      // Self evaluation
      setSelfEvaluationEnabled(config.selfEvaluationEnabled || false)
      setSelfEvaluationSchema(config.selfEvaluationSchema || '')
      setSelfEvaluationStart(
        config.selfEvaluationStart ? new Date(config.selfEvaluationStart) : undefined,
      )
      setSelfEvaluationDeadline(
        config.selfEvaluationDeadline ? new Date(config.selfEvaluationDeadline) : undefined,
      )

      // Peer evaluation
      setPeerEvaluationEnabled(config.peerEvaluationEnabled || false)
      setPeerEvaluationSchema(config.peerEvaluationSchema || '')
      setPeerEvaluationStart(
        config.peerEvaluationStart ? new Date(config.peerEvaluationStart) : undefined,
      )
      setPeerEvaluationDeadline(
        config.peerEvaluationDeadline ? new Date(config.peerEvaluationDeadline) : undefined,
      )

      // Tutor evaluation
      setTutorEvaluationEnabled(config.tutorEvaluationEnabled || false)
      setTutorEvaluationSchema(config.tutorEvaluationSchema || '')
      setTutorEvaluationStart(
        config.tutorEvaluationStart ? new Date(config.tutorEvaluationStart) : undefined,
      )
      setTutorEvaluationDeadline(
        config.tutorEvaluationDeadline ? new Date(config.tutorEvaluationDeadline) : undefined,
      )
    }
  }, [config])

  const evaluationOptions: EvaluationOptions = {
    self: {
      enabled: selfEvaluationEnabled,
      schema: selfEvaluationSchema,
      start: selfEvaluationStart,
      deadline: selfEvaluationDeadline,
    },
    peer: {
      enabled: peerEvaluationEnabled,
      schema: peerEvaluationSchema,
      start: peerEvaluationStart,
      deadline: peerEvaluationDeadline,
    },
    tutor: {
      enabled: tutorEvaluationEnabled,
      schema: tutorEvaluationSchema,
      start: tutorEvaluationStart,
      deadline: tutorEvaluationDeadline,
    },
  }

  const hasEvaluationChanges = (originalConfig?: CoursePhaseConfig) => {
    if (!originalConfig) return false

    return (
      // Self evaluation changes
      selfEvaluationEnabled !== (originalConfig.selfEvaluationEnabled || false) ||
      selfEvaluationSchema !== (originalConfig.selfEvaluationSchema || '') ||
      selfEvaluationStart?.getTime() !==
        (originalConfig.selfEvaluationStart
          ? new Date(originalConfig.selfEvaluationStart).getTime()
          : undefined) ||
      selfEvaluationDeadline?.getTime() !==
        (originalConfig.selfEvaluationDeadline
          ? new Date(originalConfig.selfEvaluationDeadline).getTime()
          : undefined) ||
      // Peer evaluation changes
      peerEvaluationEnabled !== (originalConfig.peerEvaluationEnabled || false) ||
      peerEvaluationSchema !== (originalConfig.peerEvaluationSchema || '') ||
      peerEvaluationStart?.getTime() !==
        (originalConfig.peerEvaluationStart
          ? new Date(originalConfig.peerEvaluationStart).getTime()
          : undefined) ||
      peerEvaluationDeadline?.getTime() !==
        (originalConfig.peerEvaluationDeadline
          ? new Date(originalConfig.peerEvaluationDeadline).getTime()
          : undefined) ||
      // Tutor evaluation changes
      tutorEvaluationEnabled !== (originalConfig.tutorEvaluationEnabled || false) ||
      tutorEvaluationSchema !== (originalConfig.tutorEvaluationSchema || '') ||
      tutorEvaluationStart?.getTime() !==
        (originalConfig.tutorEvaluationStart
          ? new Date(originalConfig.tutorEvaluationStart).getTime()
          : undefined) ||
      tutorEvaluationDeadline?.getTime() !==
        (originalConfig.tutorEvaluationDeadline
          ? new Date(originalConfig.tutorEvaluationDeadline).getTime()
          : undefined)
    )
  }

  return {
    // Self evaluation
    selfEvaluationEnabled,
    setSelfEvaluationEnabled,
    selfEvaluationSchema,
    setSelfEvaluationSchema,
    selfEvaluationStart,
    setSelfEvaluationStart,
    selfEvaluationDeadline,
    setSelfEvaluationDeadline,

    // Peer evaluation
    peerEvaluationEnabled,
    setPeerEvaluationEnabled,
    peerEvaluationSchema,
    setPeerEvaluationSchema,
    peerEvaluationStart,
    setPeerEvaluationStart,
    peerEvaluationDeadline,
    setPeerEvaluationDeadline,

    // Tutor evaluation
    tutorEvaluationEnabled,
    setTutorEvaluationEnabled,
    tutorEvaluationSchema,
    setTutorEvaluationSchema,
    tutorEvaluationStart,
    setTutorEvaluationStart,
    tutorEvaluationDeadline,
    setTutorEvaluationDeadline,

    // Computed values
    evaluationOptions,
    hasEvaluationChanges,
  }
}
