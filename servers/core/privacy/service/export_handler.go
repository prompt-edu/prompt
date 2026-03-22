package service

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
)

func HandleStudentDataExportRequest(c *gin.Context, subjectIdentifiers sdk.SubjectIdentifiers) (privacyDTO.PrivacyExport, error) {

  exportRecord, err := CreateExportRecord(c, subjectIdentifiers)
  if err != nil {
    return privacyDTO.PrivacyExport{}, err
  }

  cCopy := c.Copy()
  go func() {
    var goroutineErr error
    defer func() { UpdateExportStatus(goroutineErr, cCopy, exportRecord.ID) }()

    var wg sync.WaitGroup
    wg.Add(4)

    go func() { defer wg.Done(); AggregateSubjectDataFromCore(cCopy, exportRecord.ID, subjectIdentifiers, "Core", 0*time.Second) }()
    go func() { defer wg.Done(); AggregateSubjectDataFromCore(cCopy, exportRecord.ID, subjectIdentifiers, "Assessment", 18*time.Second) }()
    go func() { defer wg.Done(); AggregateSubjectDataFromCore(cCopy, exportRecord.ID, subjectIdentifiers, "Team Allocation", 5*time.Second) }()
    go func() { defer wg.Done(); AggregateSubjectDataFromCore(cCopy, exportRecord.ID, subjectIdentifiers, "Interview", 9*time.Second) }()

    wg.Wait()

    // TODO: aggregate from each microservice
  }()

  return exportRecord, nil
}

