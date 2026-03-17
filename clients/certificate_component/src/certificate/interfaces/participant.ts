import { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'

export interface ParticipantWithDownloadStatus extends CoursePhaseParticipationWithStudent {
  hasDownloaded: boolean
  firstDownload?: string
  lastDownload?: string
  downloadCount: number
}

export interface CertificateStatus {
  available: boolean
  hasDownloaded: boolean
  lastDownload?: string
  downloadCount?: number
  message?: string
}

export interface CoursePhaseConfig {
  coursePhaseId: string
  templateContent?: string
  hasTemplate: boolean
  createdAt: string
  updatedAt: string
  updatedBy?: string
  releaseDate?: string
  hasDownloads: boolean
}
