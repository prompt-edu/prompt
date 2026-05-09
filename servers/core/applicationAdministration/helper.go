package applicationAdministration

import (
	"errors"

	"github.com/prompt-edu/prompt/servers/core/applicationAdministration/applicationDTO"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
)

func metaToScoresArray(metaData meta.MetaData) ([]applicationDTO.AdditionalScore, error) {
	var scores []applicationDTO.AdditionalScore

	oldScores, ok := metaData["additionalScores"]
	if ok && oldScores != nil {
		// Assert that oldScores is a slice of interface{}
		oldScoresArray, ok := oldScores.([]interface{})
		if !ok {
			log.Error("expected []interface{}, got: ", oldScores)
			return nil, errors.New("could not update additional scores")
		}

		// For each element in oldScoresArray, we expect a map with "key" and "name" fields
		for _, scoreItem := range oldScoresArray {
			scoreMap, ok := scoreItem.(map[string]interface{})
			if !ok {
				log.Error("expected map[string]interface{}, got: ", scoreItem)
				return nil, errors.New("could not update additional scores")
			}

			// Get "key" field
			keyVal, keyFound := scoreMap["key"]
			if !keyFound {
				log.Error("missing 'key' in score map: ", scoreMap)
				return nil, errors.New("could not update additional scores")
			}
			keyStr, ok := keyVal.(string)
			if !ok {
				log.Error("expected string for 'key', got: ", keyVal)
				return nil, errors.New("could not update additional scores")
			}

			// Get "name" field
			nameVal, nameFound := scoreMap["name"]
			if !nameFound {
				log.Error("missing 'name' in score map: ", scoreMap)
				return nil, errors.New("could not update additional scores")
			}
			nameStr, ok := nameVal.(string)
			if !ok {
				log.Error("expected string for 'name', got: ", nameVal)
				return nil, errors.New("could not update additional scores")
			}

			// Append to our scores slice
			scores = append(scores, applicationDTO.AdditionalScore{
				Key:  keyStr,
				Name: nameStr,
			})
		}
	}

	return scores, nil
}

func addScoreName(oldMetaData meta.MetaData, newName, newKey string) ([]byte, error) {
	var newScoreNamesArray []applicationDTO.AdditionalScore

	newScoreNamesArray, err := metaToScoresArray(oldMetaData)
	if err != nil {
		return nil, err
	}

	keyExists := false
	for i := range newScoreNamesArray {
		if newScoreNamesArray[i].Key == newKey {
			keyExists = true
			newScoreNamesArray[i].Name = newName
			break
		}
	}

	if !keyExists {
		newAdditionalScore := applicationDTO.AdditionalScore{
			Key:  newKey,
			Name: newName,
		}
		newScoreNamesArray = append(newScoreNamesArray, newAdditionalScore)
	}

	metaDataUpdate := meta.MetaData{
		"additionalScores": newScoreNamesArray,
	}

	byteArray, err := metaDataUpdate.GetDBModel()
	if err != nil {
		log.Error(err)
		return nil, errors.New("could not update additional scores")
	}

	return byteArray, nil
}
