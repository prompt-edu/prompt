import { axiosInstance } from '@tumaet/prompt-shared-state'
import { MetaDataGraphItem } from '../../managementConsole/courseConfigurator/interfaces/courseMetaGraphItem'

export const getPhaseDataGraph = async (courseID: string): Promise<MetaDataGraphItem[]> => {
  try {
    return (await axiosInstance.get(`/api/courses/${courseID}/phase_data_graph`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
