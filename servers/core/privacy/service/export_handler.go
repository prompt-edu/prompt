package service

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/gin-gonic/gin"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	authService "github.com/prompt-edu/prompt/servers/core/auth/service"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseType"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
)

type ExportRequest struct {
	Preparation Export
	Result      ExportResult
}

type Export struct {
	Record          privacyDTO.PrivacyExport
	Subject         sdk.SubjectIdentifiers
	CoreExport      ServiceExportRequest
	ExternalExports []ServiceExportRequest
}

func PrepareDataExport(c *gin.Context) (Export, error) {
	subjectIdentifiers, err := authService.GetSubjectIdentifiers(c)
	if err != nil {
		return Export{}, err
	}

	exportRecord, err := CreateExportRecord(c, subjectIdentifiers)
	if err != nil {
		return Export{}, err
	}

  // prepare core
	coreDoc, err := PrepareExportRecordDoc(c, exportRecord.ID, "Core", "")
	if err != nil {
		return Export{}, err
	}

  coursePhaseTypes, err := coursePhaseType.GetAllCoursePhaseTypes(c)
	if err != nil {
		return Export{}, err
	}

  externalExportDocs := make([]ServiceExportRequest, 0)

  // prepare External Exports
  for _, cpt := range coursePhaseTypes {
    _, err := url.ParseRequestURI(cpt.BaseUrl)
    if err != nil { continue }
    comparedoc, err := PrepareExportRecordDoc(c, exportRecord.ID, cpt.Name, cpt.BaseUrl+sdkTypes.PrivacyRouteDataExport)
    if err != nil { continue }
    externalExportDocs = append(externalExportDocs, comparedoc)
  }

	return Export{
		Record:       exportRecord,
    Subject:      subjectIdentifiers,
		CoreExport:      coreDoc,
		ExternalExports: externalExportDocs,
	}, nil
}

func RunDataExport(c *gin.Context, exportState Export) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	authHeader := c.GetHeader("Authorization")

	// core export
	wg.Go(func() {
		cCopy := c.Copy()

		err := AggregateSubjectDataFromCore(cCopy, exportState.CoreExport, exportState.Subject)
		UpdateExportDocStatus(err, cCopy, exportState.CoreExport.ExportDoc.ID)
		UpdateExportDocFileSize(cCopy, exportState.CoreExport.ExportDoc.ID)
		mu.Lock()
		updateExportStateForRequest(err, &exportState.CoreExport)
		mu.Unlock()
	})

	// external exports
	for i := range exportState.ExternalExports {
		i := i

		wg.Go(func() {
			cCopy := c.Copy()

			callErr := RequestExportFromCPM(exportState.ExternalExports[i], authHeader)

			UpdateExportDocStatus(callErr, cCopy, exportState.ExternalExports[i].ExportDoc.ID)
      UpdateExportDocFileSize(cCopy, exportState.ExternalExports[i].ExportDoc.ID)

			mu.Lock()
			updateExportStateForRequest(callErr, &exportState.ExternalExports[i])
			mu.Unlock()
		})
	}

	wg.Wait()

	updateExportState(c, &exportState)
}

func updateExportStateForRequest(callErr error, expReq *ServiceExportRequest) {
	if callErr != nil {
		expReq.Result = Failed
	} else {
		expReq.Result = Successful
	}
}

func updateExportState(c *gin.Context, e *Export) {
	failed := 0

	if e.CoreExport.Result == Failed || e.CoreExport.Result == Pending {
		failed++
	}

	for i := range e.ExternalExports {
		if e.ExternalExports[i].Result == Failed || e.ExternalExports[i].Result == Pending {
			failed++

			err := fmt.Errorf("state failed or still pending even though request finished")
			UpdateExportDocStatus(err, c, e.ExternalExports[i].ExportDoc.ID)
			updateExportStateForRequest(err, &e.ExternalExports[i])
		}
	}

	if failed > 0 {
    UpdateExportStatus(fmt.Errorf("at least one request failed"), c, e.Record.ID)
	} else {
    UpdateExportStatus(nil, c, e.Record.ID)
	}
}
