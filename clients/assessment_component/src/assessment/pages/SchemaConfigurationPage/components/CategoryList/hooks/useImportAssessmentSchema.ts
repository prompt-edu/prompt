import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import type { AssessmentType } from '../../../../../interfaces/assessmentType'
import { createCategory } from '../../../../../network/mutations/createCategory'
import { createCompetency } from '../../../../../network/mutations/createCompetency'
import { getAllCategoriesWithCompetencies } from '../../../../../network/queries/getAllCategoriesWithCompetencies'
import type { AssessmentSchemaTemplate } from '../../../assessmentSchemaTemplate'

export interface ImportSchemaResult {
  importedCategories: number
  importedCompetencies: number
  errors: string[]
}

export const useImportAssessmentSchema = (
  assessmentSchemaID: string,
  assessmentType: AssessmentType,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation<ImportSchemaResult, Error, AssessmentSchemaTemplate>({
    mutationFn: async (template) => {
      const coursePhaseID = phaseId ?? ''
      const errors: string[] = []
      let importedCategories = 0
      let importedCompetencies = 0

      for (const category of template.categories) {
        try {
          const before = await getAllCategoriesWithCompetencies(coursePhaseID, assessmentType)
          const knownIDs = new Set(before.map((existing) => existing.id))

          await createCategory(coursePhaseID, {
            name: category.name,
            shortName: category.shortName,
            description: category.description,
            weight: category.weight,
            assessmentSchemaID,
          })

          const after = await getAllCategoriesWithCompetencies(coursePhaseID, assessmentType)
          const created = after.find((existing) => !knownIDs.has(existing.id))
          if (!created) {
            errors.push(
              `Category "${category.name}" was created but could not be matched to add its competencies.`,
            )
            continue
          }
          importedCategories += 1

          for (const competency of category.competencies) {
            try {
              await createCompetency(coursePhaseID, { ...competency, categoryID: created.id })
              importedCompetencies += 1
            } catch {
              errors.push(
                `Competency "${competency.name}" in category "${category.name}" could not be imported.`,
              )
            }
          }
        } catch {
          errors.push(`Category "${category.name}" could not be imported.`)
        }
      }

      return { importedCategories, importedCompetencies, errors }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['categories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['selfEvaluationCategories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['peerEvaluationCategories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['tutorEvaluationCategories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['assessments'] })
    },
  })
}
