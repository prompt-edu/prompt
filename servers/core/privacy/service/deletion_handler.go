package service

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	authService "github.com/prompt-edu/prompt/servers/core/auth/service"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseType"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	log "github.com/sirupsen/logrus"
)

const DeletionRunTimeout = 30 * time.Minute

type ServiceDeletionRequest struct {
	APIURL     string
	Subrequest privacyDTO.PrivacyDeletionSubrequest
	Result     db.PrivacyDeletionSubrequestStatus
}

type Deletion struct {
	Record            privacyDTO.PrivacyDeletionRequest
	Subject           sdk.SubjectIdentifiers
	CoreDeletion      ServiceDeletionRequest
	ExternalDeletions []ServiceDeletionRequest
}

func PrepareDataDeletion(c context.Context, record privacyDTO.PrivacyDeletionRequest) (Deletion, error) {
	studentID := uuid.Nil
	if record.StudentID != nil {
		studentID = *record.StudentID
	}
	coursePhaseTypes, err := coursePhaseType.GetCoursePhaseTypesForStudentCourses(c, studentID)
	if err != nil {
		return Deletion{}, fmt.Errorf("failed to load involved course phase types: %w", err)
	}

	tx, err := PrivacyServiceSingleton.conn.Begin(c)
	if err != nil {
		return Deletion{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(c) }()
	txQueries := PrivacyServiceSingleton.queries.WithTx(tx)

	coreReq, err := CreateDeletionSubrequest(c, txQueries, record.ID, "Core", "")
	if err != nil {
		return Deletion{}, err
	}

	externalDeletions := make([]ServiceDeletionRequest, 0)
	for _, cpt := range coursePhaseTypes {
		if _, err := url.ParseRequestURI(cpt.BaseUrl); err != nil {
			continue
		}
		sub, err := CreateDeletionSubrequest(c, txQueries, record.ID, cpt.Name, cpt.BaseUrl+sdkTypes.PrivacyRouteDataDeletion)
		if err != nil {
			return Deletion{}, err
		}
		externalDeletions = append(externalDeletions, sub)
	}

	if err := tx.Commit(c); err != nil {
		return Deletion{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	userID := uuid.Nil
	if record.UserID != nil {
		userID = *record.UserID
	}
	subject, err := authService.AssembleSubjectIdentifiers(c, userID, record.StudentID)
	if err != nil {
		return Deletion{}, fmt.Errorf("failed to assemble subject identifiers: %w", err)
	}

	return Deletion{
		Record:            record,
		Subject:           subject,
		CoreDeletion:      coreReq,
		ExternalDeletions: externalDeletions,
	}, nil
}

func RunDataDeletion(ctx context.Context, authHeader string, state Deletion) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Go(func() {
		err := ExecuteCoreDeletion(ctx, state.Subject)
		updateSubrequestStatus(context.WithoutCancel(ctx), state.CoreDeletion.Subrequest.ID, err)
		mu.Lock()
		state.CoreDeletion.Result = subrequestStatusFromError(err)
		mu.Unlock()
	})

	for i := range state.ExternalDeletions {
		i := i
		wg.Go(func() {
			err := RequestDeletionFromCPM(ctx, state.ExternalDeletions[i].APIURL, state.Subject, authHeader)
			updateSubrequestStatus(context.WithoutCancel(ctx), state.ExternalDeletions[i].Subrequest.ID, err)
			mu.Lock()
			state.ExternalDeletions[i].Result = subrequestStatusFromError(err)
			mu.Unlock()
		})
	}

	wg.Wait()

	finalStatus := UpdateDeletionRequestStatus(ctx, &state)
	sendDeletionConfirmationMail(context.WithoutCancel(ctx), state.Record.ID, state.Record.RecipientEmail, finalStatus)
}

func updateSubrequestStatus(ctx context.Context, subrequestID uuid.UUID, callErr error) {
	errMsg := ""
	if callErr != nil {
		errMsg = callErr.Error()
	}
	if _, err := PrivacyServiceSingleton.queries.SetDeletionSubrequestStatus(ctx, db.SetDeletionSubrequestStatusParams{
		ID:           subrequestID,
		Status:       subrequestStatusFromError(callErr),
		ErrorMessage: errMsg,
	}); err != nil {
		log.WithError(err).Error("failed to update deletion subrequest status")
	}
}

func UpdateDeletionRequestStatus(ctx context.Context, state *Deletion) db.PrivacyDeletionRequestStatus {
	statusCtx := context.WithoutCancel(ctx)

	anyFailed := state.CoreDeletion.Result == db.PrivacyDeletionSubrequestStatusFailed
	for _, sub := range state.ExternalDeletions {
		if sub.Result == db.PrivacyDeletionSubrequestStatusFailed {
			anyFailed = true
			break
		}
	}

	finalStatus := db.PrivacyDeletionRequestStatusSucceeded
	if anyFailed {
		finalStatus = db.PrivacyDeletionRequestStatusFailed
	}

	if _, err := PrivacyServiceSingleton.queries.SetDeletionRequestStatus(statusCtx, db.SetDeletionRequestStatusParams{
		ID:     state.Record.ID,
		Status: finalStatus,
	}); err != nil {
		log.WithError(err).Error("failed to set deletion request final status")
	}
	return finalStatus
}
