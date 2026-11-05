import { format } from 'date-fns'
import type { ApplicationMetaData, ApplicationMode } from '../interfaces/applicationMetaData'

export interface GeneralSettingsValues {
  startDate?: Date
  endDate?: Date
  startTime: string
  endTime: string
  externalStudentsAllowed: boolean
  universityLoginAvailable: boolean
  autoAccept: boolean
  mode: ApplicationMode
}

const getTimeString = (date: Date) =>
  `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`

const getDayString = (date?: Date) => (date ? format(date, 'yyyy-MM-dd') : '')

export const getGeneralSettingsValues = (data: ApplicationMetaData): GeneralSettingsValues => ({
  startDate: data.applicationStartDate ? new Date(data.applicationStartDate) : undefined,
  endDate: data.applicationEndDate ? new Date(data.applicationEndDate) : undefined,
  startTime: data.applicationStartDate
    ? getTimeString(new Date(data.applicationStartDate))
    : '00:00',
  endTime: data.applicationEndDate ? getTimeString(new Date(data.applicationEndDate)) : '23:59',
  externalStudentsAllowed: data.externalStudentsAllowed ?? false,
  universityLoginAvailable: data.universityLoginAvailable ?? false,
  autoAccept: data.autoAccept ?? false,
  mode: data.applicationMode === 'import' ? 'import' : 'apply',
})

export const hasGeneralSettingsChanges = (
  values: GeneralSettingsValues,
  stored: GeneralSettingsValues,
): boolean =>
  getDayString(values.startDate) !== getDayString(stored.startDate) ||
  getDayString(values.endDate) !== getDayString(stored.endDate) ||
  values.startTime !== stored.startTime ||
  values.endTime !== stored.endTime ||
  values.externalStudentsAllowed !== stored.externalStudentsAllowed ||
  values.universityLoginAvailable !== stored.universityLoginAvailable ||
  values.autoAccept !== stored.autoAccept ||
  values.mode !== stored.mode
