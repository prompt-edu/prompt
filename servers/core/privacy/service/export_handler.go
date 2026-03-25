package service

import (
	"github.com/gin-gonic/gin"
	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
)

type ExportPreparation struct {
	Record       privacyDTO.PrivacyExport
	CoreDoc      PreparedExportDoc
	ExternalDocs []PreparedExportDoc
}

func PrepareStudentDataExport(c *gin.Context, subjectIdentifiers sdk.SubjectIdentifiers) (ExportPreparation, error) {
	exportRecord, err := CreateExportRecord(c, subjectIdentifiers)
	if err != nil {
		return ExportPreparation{}, err
	}

	coreDoc, err := PrepareExportRecordDoc(c, exportRecord.ID, "Core")
	if err != nil {
		return ExportPreparation{}, err
	}

	// TODO: prepare external microservice export docs here
	// this will come with a later PR
	return ExportPreparation{
		Record:       exportRecord,
		CoreDoc:      coreDoc,
		ExternalDocs: []PreparedExportDoc{},
	}, nil
}

func RunStudentDataExport(c *gin.Context, prep ExportPreparation, subjectIdentifiers sdk.SubjectIdentifiers) {
	cCopy := c.Copy()
	go func() {
		var goroutineErr error
		defer func() { UpdateExportStatus(goroutineErr, cCopy, prep.Record.ID) }()

		goroutineErr = AggregateSubjectDataFromCore(cCopy, prep.CoreDoc, subjectIdentifiers)

		// TODO: run external microservice exports concurrently here
    // this will come with a later PR
	}()
}
