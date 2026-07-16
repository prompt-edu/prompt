import { parseURL } from '@tumaet/prompt-shared-state'
import axios from 'axios'
import type {
  CategoryRequest,
  CreateSlotRequest,
  FeedbackAnswer,
  FeedbackCategory,
  FeedbackDocument,
  FeedbackEvent,
  MaterialDownload,
  MaterialUploadIntent,
  PresentationConfig,
  PresentationMaterial,
  PresentationSlot,
  PresentationSummary,
  PresentationTarget,
  TargetType,
} from './interfaces'

interface PresentationEnvironment {
  PRESENTATION_HOST?: string
}

const presentationHost =
  typeof window === 'undefined'
    ? ''
    : ((window as unknown as { env?: PresentationEnvironment }).env?.PRESENTATION_HOST ?? '')
const baseURL = parseURL(presentationHost)

export const presentationAxiosInstance = axios.create({ baseURL })

presentationAxiosInstance.interceptors.request.use((config) => {
  const token = localStorage.getItem('jwt_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

const phasePath = (coursePhaseId: string) => `presentation/api/course_phase/${coursePhaseId}`

export const presentationApi = {
  getConfig: async (coursePhaseId: string): Promise<PresentationConfig> =>
    (await presentationAxiosInstance.get<PresentationConfig>(`${phasePath(coursePhaseId)}/config`))
      .data,

  updateConfig: async (
    coursePhaseId: string,
    config: Pick<PresentationConfig, 'targetMode' | 'feedbackMode'>,
    resetExistingData = false,
  ): Promise<PresentationConfig> =>
    (
      await presentationAxiosInstance.put<PresentationConfig>(
        `${phasePath(coursePhaseId)}/config`,
        { ...config, resetExistingData },
      )
    ).data,

  getCategories: async (coursePhaseId: string): Promise<FeedbackCategory[]> =>
    (
      await presentationAxiosInstance.get<FeedbackCategory[]>(
        `${phasePath(coursePhaseId)}/categories`,
      )
    ).data,

  createCategory: async (
    coursePhaseId: string,
    request: CategoryRequest,
    resetExistingData = false,
  ): Promise<FeedbackCategory> =>
    (
      await presentationAxiosInstance.post<FeedbackCategory>(
        `${phasePath(coursePhaseId)}/categories`,
        request,
        { params: { resetExistingData } },
      )
    ).data,

  updateCategory: async (
    coursePhaseId: string,
    categoryId: string,
    request: CategoryRequest,
    resetExistingData = false,
  ): Promise<FeedbackCategory> =>
    (
      await presentationAxiosInstance.put<FeedbackCategory>(
        `${phasePath(coursePhaseId)}/categories/${categoryId}`,
        request,
        { params: { resetExistingData } },
      )
    ).data,

  deleteCategory: async (
    coursePhaseId: string,
    categoryId: string,
    resetExistingData = false,
  ): Promise<void> => {
    await presentationAxiosInstance.delete(`${phasePath(coursePhaseId)}/categories/${categoryId}`, {
      params: { resetExistingData },
    })
  },

  getPresentations: async (coursePhaseId: string): Promise<PresentationSummary[]> =>
    (
      await presentationAxiosInstance.get<PresentationSummary[]>(
        `${phasePath(coursePhaseId)}/presentations`,
      )
    ).data,

  getOwnPresentation: async (coursePhaseId: string): Promise<PresentationSummary | null> =>
    (
      await presentationAxiosInstance.get<PresentationSummary | null>(
        `${phasePath(coursePhaseId)}/presentations/me`,
      )
    ).data,

  getSlots: async (coursePhaseId: string): Promise<PresentationSlot[]> =>
    (await presentationAxiosInstance.get<PresentationSlot[]>(`${phasePath(coursePhaseId)}/slots`))
      .data,

  createSlot: async (
    coursePhaseId: string,
    request: CreateSlotRequest,
  ): Promise<PresentationSlot> =>
    (
      await presentationAxiosInstance.post<PresentationSlot>(
        `${phasePath(coursePhaseId)}/slots`,
        request,
      )
    ).data,

  updateSlot: async (
    coursePhaseId: string,
    slotId: string,
    request: CreateSlotRequest,
  ): Promise<PresentationSlot> =>
    (
      await presentationAxiosInstance.put<PresentationSlot>(
        `${phasePath(coursePhaseId)}/slots/${slotId}`,
        request,
      )
    ).data,

  deleteSlot: async (coursePhaseId: string, slotId: string): Promise<void> => {
    await presentationAxiosInstance.delete(`${phasePath(coursePhaseId)}/slots/${slotId}`)
  },

  getTargets: async (coursePhaseId: string): Promise<PresentationTarget[]> =>
    (
      await presentationAxiosInstance.get<PresentationTarget[]>(
        `${phasePath(coursePhaseId)}/targets`,
      )
    ).data,

  assignTarget: async (
    coursePhaseId: string,
    slotId: string,
    target: { targetId: string; targetName: string; targetType: TargetType },
  ): Promise<PresentationSummary> =>
    (
      await presentationAxiosInstance.put<PresentationSummary>(
        `${phasePath(coursePhaseId)}/slots/${slotId}/assignment`,
        target,
      )
    ).data,

  unassignTarget: async (coursePhaseId: string, slotId: string): Promise<void> => {
    await presentationAxiosInstance.delete(`${phasePath(coursePhaseId)}/slots/${slotId}/assignment`)
  },

  getMaterials: async (
    coursePhaseId: string,
    presentationId: string,
  ): Promise<PresentationMaterial[]> =>
    (
      await presentationAxiosInstance.get<PresentationMaterial[]>(
        `${phasePath(coursePhaseId)}/presentations/${presentationId}/materials`,
      )
    ).data,

  createUploadIntent: async (
    coursePhaseId: string,
    presentationId: string,
    file: File,
  ): Promise<MaterialUploadIntent> =>
    (
      await presentationAxiosInstance.post<MaterialUploadIntent>(
        `${phasePath(coursePhaseId)}/presentations/${presentationId}/materials/presign`,
        {
          fileName: file.name,
          contentType: file.type || 'application/octet-stream',
          sizeBytes: file.size,
        },
      )
    ).data,

  completeUpload: async (
    coursePhaseId: string,
    presentationId: string,
    uploadId: string,
  ): Promise<PresentationMaterial> =>
    (
      await presentationAxiosInstance.post<PresentationMaterial>(
        `${phasePath(coursePhaseId)}/presentations/${presentationId}/materials/${uploadId}/complete`,
      )
    ).data,

  getMaterialDownload: async (
    coursePhaseId: string,
    presentationId: string,
    materialId: string,
  ): Promise<MaterialDownload> =>
    (
      await presentationAxiosInstance.get<MaterialDownload>(
        `${phasePath(coursePhaseId)}/presentations/${presentationId}/materials/${materialId}/download`,
      )
    ).data,

  deleteMaterial: async (
    coursePhaseId: string,
    presentationId: string,
    materialId: string,
  ): Promise<void> => {
    await presentationAxiosInstance.delete(
      `${phasePath(coursePhaseId)}/presentations/${presentationId}/materials/${materialId}`,
    )
  },

  getFeedback: async (coursePhaseId: string, presentationId: string): Promise<FeedbackDocument> =>
    (
      await presentationAxiosInstance.get<FeedbackDocument>(
        `${phasePath(coursePhaseId)}/presentations/${presentationId}/feedback`,
      )
    ).data,

  updateAnswer: async (
    coursePhaseId: string,
    presentationId: string,
    categoryId: string,
    value: string,
    expectedRevision: number,
  ): Promise<FeedbackAnswer> =>
    (
      await presentationAxiosInstance.put<FeedbackAnswer>(
        `${phasePath(coursePhaseId)}/presentations/${presentationId}/feedback/answers/${categoryId}`,
        { value, expectedRevision },
      )
    ).data,

  submitFeedback: async (coursePhaseId: string, presentationId: string): Promise<void> => {
    await presentationAxiosInstance.post(
      `${phasePath(coursePhaseId)}/presentations/${presentationId}/feedback/submit`,
    )
  },

  reopenFeedback: async (coursePhaseId: string, presentationId: string): Promise<void> => {
    await presentationAxiosInstance.post(
      `${phasePath(coursePhaseId)}/presentations/${presentationId}/feedback/reopen`,
    )
  },

  deleteDraft: async (coursePhaseId: string, presentationId: string): Promise<void> => {
    await presentationAxiosInstance.delete(
      `${phasePath(coursePhaseId)}/presentations/${presentationId}/feedback/draft`,
    )
  },

  releaseFeedback: async (
    coursePhaseId: string,
    presentationId: string,
    releaseName: string,
  ): Promise<void> => {
    await presentationAxiosInstance.post(
      `${phasePath(coursePhaseId)}/presentations/${presentationId}/feedback/release`,
      { releaseName },
    )
  },

  unreleaseFeedback: async (coursePhaseId: string, presentationId: string): Promise<void> => {
    await presentationAxiosInstance.delete(
      `${phasePath(coursePhaseId)}/presentations/${presentationId}/feedback/release`,
    )
  },

  resetFeedback: async (coursePhaseId: string, presentationId: string): Promise<void> => {
    await presentationAxiosInstance.delete(
      `${phasePath(coursePhaseId)}/presentations/${presentationId}/feedback`,
    )
  },
}

export const uploadMaterial = async (
  coursePhaseId: string,
  presentationId: string,
  file: File,
): Promise<PresentationMaterial> => {
  const intent = await presentationApi.createUploadIntent(coursePhaseId, presentationId, file)
  const uploadResponse = await fetch(intent.uploadUrl, {
    method: 'PUT',
    headers: intent.headers,
    body: file,
  })
  if (!uploadResponse.ok) throw new Error(`Upload failed with status ${uploadResponse.status}`)
  return presentationApi.completeUpload(coursePhaseId, presentationId, intent.uploadId)
}

const eventsURL = (coursePhaseId: string, presentationId: string): string => {
  const root = String(baseURL).replace(/\/$/, '')
  return `${root}/${phasePath(coursePhaseId)}/presentations/${presentationId}/feedback/events`
}

const parseEventBlock = (block: string): FeedbackEvent | null => {
  const data = block
    .split('\n')
    .filter((line) => line.startsWith('data:'))
    .map((line) => line.slice(5).trimStart())
    .join('\n')
  if (!data) return null
  try {
    return JSON.parse(data) as FeedbackEvent
  } catch {
    return null
  }
}

export const streamFeedbackEvents = async (
  coursePhaseId: string,
  presentationId: string,
  onEvent: (event: FeedbackEvent) => void,
  signal: AbortSignal,
): Promise<void> => {
  const response = await fetch(eventsURL(coursePhaseId, presentationId), {
    headers: {
      Accept: 'text/event-stream',
      Authorization: `Bearer ${localStorage.getItem('jwt_token') ?? ''}`,
    },
    signal,
  })
  if (!response.ok || !response.body) throw new Error(`Feedback stream failed: ${response.status}`)

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  while (!signal.aborted) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true }).replace(/\r\n/g, '\n')
    let boundary = buffer.indexOf('\n\n')
    while (boundary >= 0) {
      const event = parseEventBlock(buffer.slice(0, boundary))
      if (event) onEvent(event)
      buffer = buffer.slice(boundary + 2)
      boundary = buffer.indexOf('\n\n')
    }
  }
}

export const openMaterialDownload = async (
  coursePhaseId: string,
  presentationId: string,
  materialId: string,
): Promise<void> => {
  const { downloadUrl, fileName } = await presentationApi.getMaterialDownload(
    coursePhaseId,
    presentationId,
    materialId,
  )
  const link = document.createElement('a')
  link.href = downloadUrl
  link.download = fileName
  link.rel = 'noopener'
  document.body.appendChild(link)
  link.click()
  link.remove()
}
