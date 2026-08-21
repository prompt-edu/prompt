import type {
  MaterialType,
  PresentationMaterial,
  PresentationSummary,
  TargetMode,
} from './interfaces'
import { getMaterialTypeDefinition } from './materialTypes'

const SAMPLE_DAYS_AHEAD = 7
const SAMPLE_START_HOUR = 10
const SAMPLE_DURATION_MINUTES = 30
const SAMPLE_FILE_SIZE_BYTES = 2_411_724

// Obviously fictional stand-ins for the staff preview of the student page. Nothing here is
// ever sent to the API, so the ids only have to be stable within one render.
export const buildSamplePresentation = (
  coursePhaseId: string,
  targetMode: TargetMode,
): PresentationSummary => {
  const startTime = new Date()
  startTime.setDate(startTime.getDate() + SAMPLE_DAYS_AHEAD)
  startTime.setHours(SAMPLE_START_HOUR, 0, 0, 0)
  const endTime = new Date(startTime.getTime() + SAMPLE_DURATION_MINUTES * 60_000)

  return {
    id: 'sample-presentation',
    coursePhaseId,
    slotId: 'sample-slot',
    targetType: targetMode,
    targetName: targetMode === 'team' ? 'Sample Team' : 'Sample Student',
    targetId: 'sample-target',
    startTime: startTime.toISOString(),
    endTime: endTime.toISOString(),
    location: 'Room 101',
    materialCount: 1,
    feedbackCount: 0,
    submittedFeedbackCount: 0,
  }
}

// Fills the first requested upload only, so the preview shows both a handed-in and a still
// missing slot side by side.
export const buildSampleMaterials = (requiredTypes: MaterialType[]): PresentationMaterial[] => {
  const [firstType] = requiredTypes
  if (!firstType) return []
  const definition = getMaterialTypeDefinition(firstType)
  const extension = definition.extensions[0] ?? '.pdf'

  return [
    {
      id: 'sample-material',
      presentationId: 'sample-presentation',
      materialType: firstType,
      fileName: `sample-${firstType}${extension}`,
      contentType: 'application/octet-stream',
      sizeBytes: SAMPLE_FILE_SIZE_BYTES,
      uploadedByName: 'Sample Student',
      uploadedAt: new Date().toISOString(),
    },
  ]
}
