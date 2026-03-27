package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
	log "github.com/sirupsen/logrus"
)


func RequestExportFromCPM(exportDoc ServiceExportRequest, subject sdk.SubjectIdentifiers) (error) {
  requestBody := sdk.PrivacyDataExportRequest{
    Subject: subject,
    PreSignedURL: exportDoc.PresignedUploadURL,
  }

  marshalledBody, err := json.Marshal(requestBody)
  if err != nil { return err }

  req, err := http.NewRequest("POST", exportDoc.APIURL, bytes.NewBuffer(marshalledBody))
  if err != nil { return err }
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}

  resp, err := client.Do(req)
  if err != nil {
    return err
  }
  defer func() {
    err := resp.Body.Close()
    if err != nil {
      log.Error(err)
    }
  }()

  fmt.Println("response Status:", resp.Status)
  fmt.Println("response Headers:", resp.Header)
  body, _ := io.ReadAll(resp.Body)
  fmt.Println("response Body:", string(body))

  // we need to standardize the result of the data export request 
  // and define a return state for export request succeeded but no data
  // then we can chick this here and return that state here, for now,
  // only (fail / no fail) -> This needs a prompt-skd change
  // until then we will keep it as it is here.

  return nil
}
