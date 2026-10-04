import type { ApplicationFormWithDetails } from '@core/interfaces/application/applicationFormWithDetails'
import type { GetApplication } from '@core/interfaces/application/getApplication'
import type { OpenApplicationDetails } from '@core/interfaces/application/openApplicationDetails'
import type { PostApplication } from '@core/interfaces/application/postApplication'
import type { PostApplicationResponse } from '@core/publicPages/application/interfaces/postApplicationConfirmation'
import { API_PREFIX, coreRequest, publicRequest } from '../client'

const path = `${API_PREFIX}/apply`
const authenticated = `${path}/authenticated`

/**
 * The student-facing side of an application phase.
 *
 * `listOpen`, `form` and `submitExternal` are reached before there is a token, so they go through
 * the unauthenticated instance; their authenticated twins do not.
 */
export const apply = {
  listOpen: (): Promise<OpenApplicationDetails[]> => publicRequest.get(path),

  form: (coursePhaseID: string): Promise<ApplicationFormWithDetails> =>
    publicRequest.get(`${path}/${coursePhaseID}`),

  submitExternal: (
    coursePhaseID: string,
    application: PostApplication,
  ): Promise<PostApplicationResponse> =>
    publicRequest.post(`${path}/${coursePhaseID}`, application),

  mine: (coursePhaseID: string): Promise<GetApplication> =>
    coreRequest.get(`${authenticated}/${coursePhaseID}`),

  submitAuthenticated: (
    coursePhaseID: string,
    application: PostApplication,
  ): Promise<PostApplicationResponse> =>
    coreRequest.post(`${authenticated}/${coursePhaseID}`, application),
}
