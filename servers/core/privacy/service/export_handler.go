package service

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	authService "github.com/prompt-edu/prompt/servers/core/auth/service"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseType"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	log "github.com/sirupsen/logrus"
)

const ExportRunTimeout = 30 * time.Minute

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

func (s *PrivacyService) PrepareDataExport(c *gin.Context) (Export, error) {
	subjectIdentifiers, err := authService.GetSubjectIdentifiers(c)
	if err != nil {
		return Export{}, err
	}

	exportRecord, err := s.CreateExportRecord(c, subjectIdentifiers)
	if err != nil {
		return Export{}, err
	}

	// prepare core
	coreDoc, err := s.PrepareExportRecordDoc(c, exportRecord.ID, "Core", "")
	if err != nil {
		return Export{}, err
	}

	coursePhaseTypes, err := coursePhaseType.GetCoursePhaseTypesForStudentCourses(c, subjectIdentifiers.StudentID)
	if err != nil {
		return Export{}, fmt.Errorf("failed to load involved course phase types: %w", err)
	}

	externalExportDocs := make([]ServiceExportRequest, 0)

	// prepare External Exports
	for _, cpt := range coursePhaseTypes {
		_, err := url.ParseRequestURI(cpt.BaseUrl)
		if err != nil {
			continue
		}
		comparedoc, err := s.PrepareExportRecordDoc(c, exportRecord.ID, cpt.Name, cpt.BaseUrl+sdkTypes.PrivacyRouteDataExport)
		if err != nil {
			continue
		}
		externalExportDocs = append(externalExportDocs, comparedoc)
	}

	return Export{
		Record:          exportRecord,
		Subject:         subjectIdentifiers,
		CoreExport:      coreDoc,
		ExternalExports: externalExportDocs,
	}, nil
}

func (s *PrivacyService) RunDataExport(ctx context.Context, authHeader string, exportState Export) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Go(func() {
		err := s.AggregateSubjectDataFromCore(ctx, exportState.CoreExport, exportState.Subject)
		s.UpdateExportDocFileSize(ctx, exportState.CoreExport.ExportDoc.ID)
		mu.Lock()
		updateExportStateForRequest(err, &exportState.CoreExport)
		mu.Unlock()
	})

	for i := range exportState.ExternalExports {
		i := i

		wg.Go(func() {
			result := RequestExportFromCPM(exportState.ExternalExports[i], authHeader)

			if setErr := s.SetExportDocStatus(context.WithoutCancel(ctx), exportState.ExternalExports[i].ExportDoc.ID, exportResultToDBStatus(result)); setErr != nil {
				log.WithError(setErr).Error("failed to set export doc status")
			}
			if result == Successful {
				s.UpdateExportDocFileSize(ctx, exportState.ExternalExports[i].ExportDoc.ID)
			}

			mu.Lock()
			exportState.ExternalExports[i].Result = result
			mu.Unlock()
		})
	}

	wg.Wait()

	s.updateExportState(ctx, &exportState)
}

func updateExportStateForRequest(callErr error, expReq *ServiceExportRequest) {
	if callErr != nil {
		expReq.Result = Failed
	} else {
		expReq.Result = Successful
	}
}

func (s *PrivacyService) updateExportState(ctx context.Context, e *Export) {
	statusCtx := context.WithoutCancel(ctx)
	failed := 0

	if e.CoreExport.Result == Failed || e.CoreExport.Result == Pending {
		failed++
	}

	for i := range e.ExternalExports {
		if e.ExternalExports[i].Result == Failed || e.ExternalExports[i].Result == Pending {
			failed++
			if e.ExternalExports[i].Result == Pending {
				log.Errorf("export doc %s still pending after request finished", e.ExternalExports[i].ExportDoc.ID)
				if setErr := s.SetExportDocStatus(statusCtx, e.ExternalExports[i].ExportDoc.ID, exportResultToDBStatus(Failed)); setErr != nil {
					log.WithError(setErr).Error("failed to mark pending export doc as failed")
				}
			}
		}
	}

	if failed > 0 {
		s.UpdateExportStatus(fmt.Errorf("at least one request failed"), statusCtx, e.Record.ID)
	} else {
		s.UpdateExportStatus(nil, statusCtx, e.Record.ID)
	}
}
