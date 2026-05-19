import { MetaDataGraphItem } from '../../managementConsole/courseConfigurator/interfaces/courseMetaGraphItem'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const updateParticipationDataGraph = async (
  courseID: string,
  metaDataGraph: MetaDataGraphItem[],
): Promise<void> => {
  try {
    return await axiosInstance.put(
      `/api/courses/${courseID}/participation_data_graph`,
      metaDataGraph,
      {
        headers: {
          'Content-Type': 'application/json-path+json',
        },
      },
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
