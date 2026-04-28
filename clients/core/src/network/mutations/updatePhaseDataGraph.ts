import { MetaDataGraphItem } from '../../managementConsole/courseConfigurator/interfaces/courseMetaGraphItem'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const updatePhaseDataGraph = async (
  courseID: string,
  metaDataGraph: MetaDataGraphItem[],
): Promise<void> => {
  try {
    return await axiosInstance.put(`/api/courses/${courseID}/phase_data_graph`, metaDataGraph, {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    })
  } catch (err) {
    console.error(err)
    throw err
  }
}
