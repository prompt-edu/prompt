import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import type { AssessmentType } from '../../../../../interfaces/assessmentType'
import { createCategory } from '../../../../../network/mutations/createCategory'
import { createCompetency } from '../../../../../network/mutations/createCompetency'
import { getAllCategoriesWithCompetencies } from '../../../../../network/queries/getAllCategoriesWithCompetencies'
import type { AssessmentSchemaTemplate } from '../../../utils/assessmentSchemaTemplate'

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

      const existing = await getAllCategoriesWithCompetencies(coursePhaseID, assessmentType)
      const existingNames = new Set(existing.map((category) => category.name))

      for (const category of template.categories) {
        if (existingNames.has(category.name)) {
          errors.push(`Category "${category.name}" already exists and was skipped.`)
          continue
        }

        let created: Awaited<ReturnType<typeof createCategory>>
        try {
          created = await createCategory(coursePhaseID, {
            name: category.name,
            shortName: category.shortName,
            description: category.description,
            weight: category.weight,
            assessmentSchemaID,
          })
        } catch {
          errors.push(`Category "${category.name}" could not be imported.`)
          continue
        }
        existingNames.add(category.name)
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
