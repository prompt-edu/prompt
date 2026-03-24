package service

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
)

type ExportPreparation struct {
	Record       privacyDTO.PrivacyExport
	CoreDoc      PreparedExportDoc
	ExternalDocs []PreparedExportDoc
}

var externalServices = []string{"Assessment", "Team Allocation", "Interview"}

func PrepareStudentDataExport(c *gin.Context, subjectIdentifiers sdk.SubjectIdentifiers) (ExportPreparation, error) {
	exportRecord, err := CreateExportRecord(c, subjectIdentifiers)
	if err != nil {
		return ExportPreparation{}, err
	}

	coreDoc, err := PrepareExportRecordDoc(c, exportRecord.ID, "Core")
	if err != nil {
		return ExportPreparation{}, err
	}

	externalDocs := make([]PreparedExportDoc, 0, len(externalServices))
	for _, svc := range externalServices {
		doc, err := PrepareExportRecordDoc(c, exportRecord.ID, svc)
		if err != nil {
			return ExportPreparation{}, err
		}
		externalDocs = append(externalDocs, doc)
	}

	return ExportPreparation{
		Record:       exportRecord,
		CoreDoc:      coreDoc,
		ExternalDocs: externalDocs,
	}, nil
}

func RunStudentDataExport(c *gin.Context, prep ExportPreparation, subjectIdentifiers sdk.SubjectIdentifiers) {
	cCopy := c.Copy()
	go func() {
		var goroutineErr error
		defer func() { UpdateExportStatus(goroutineErr, cCopy, prep.Record.ID) }()

		var wg sync.WaitGroup
		wg.Add(1 + len(prep.ExternalDocs))

		go func() {
			defer wg.Done()
			_ = AggregateSubjectDataFromCore(cCopy, prep.CoreDoc, subjectIdentifiers, 5*time.Second)
		}()

		for _, doc := range prep.ExternalDocs {
			go func(d PreparedExportDoc) {
				defer wg.Done()
				// TODO: replace with actual API call to the external microservice
				MockExternalServiceExport(cCopy, d)
			}(doc)
		}

		wg.Wait()
	}()
}
