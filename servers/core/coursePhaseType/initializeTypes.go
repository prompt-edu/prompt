package coursePhaseType

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
)

// InitializePhaseTypes creates every built-in course phase type and repairs the
// descriptors of the ones that already exist. A failure is fatal: core must not
// serve a course catalogue with missing phase types.
func (s *CoursePhaseTypeService) InitializePhaseTypes() {
	err := s.initInterview()
	if err != nil {
		log.Fatal("failed to init interview phase type: ", err)
	}

	err = s.initMatching()
	if err != nil {
		log.Fatal("failed to init matching phase type: ", err)
	}

	err = s.initIntroCourseDeveloper()
	if err != nil {
		log.Fatal("failed to init intro course developer phase type: ", err)
	}

	err = s.initDevOpsChallenge()
	if err != nil {
		log.Fatal("failed to init dev ops challenge phase type: ", err)
	}

	err = s.initAssessment()
	if err != nil {
		log.Fatal("failed to init assessment phase type: ", err)
	}

	err = s.initTeamAllocation()
	if err != nil {
		log.Fatal("failed to init team allocation phase type: ", err)
	}

	err = s.initSelfTeamAllocation()
	if err != nil {
		log.Fatal("failed to init self team allocation phase type: ", err)
	}

	err = s.initCertificate()
	if err != nil {
		log.Fatal("failed to init certificate phase type: ", err)
	}

	err = s.initPresentation()
	if err != nil {
		log.Fatal("failed to init presentation phase type: ", err)
	}

	err = s.initInfrastructureSetup()
	if err != nil {
		log.Fatal("failed to init infrastructure setup phase type: ", err)
	}
}

func getScoreLevelSpecificationBytes() ([]byte, error) {
	scoreLevelSpecificationJson := meta.MetaData{}
	scoreLevelSpecificationJson["type"] = "string"
	scoreLevelSpecificationJson["enum"] = []string{"veryBad", "bad", "ok", "good", "veryGood"}
	return scoreLevelSpecificationJson.GetDBModel()
}

func (s *CoursePhaseTypeService) initInterview() error {
	ctx := context.Background()
	exists, err := s.queries.TestInterviewPhaseTypeExists(ctx)
	if err != nil {
		log.Error("failed to check if interview phase type exists: ", err)
		return err
	}
	if !exists {
		tx, err := s.conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer sdkUtils.DeferRollback(tx, ctx)
		qtx := s.queries.WithTx(tx)

		// 1.) Create the phase
		baseURL := "{CORE_HOST}/interview/api"
		if s.isDevEnvironment {
			baseURL = "http://localhost:8087/interview/api"
		}

		newInterviewPhase := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "Interview",
			InitialPhase: false,
			BaseUrl:      baseURL,
			Description:  pgtype.Text{String: "Interview phase for student assessments and scheduling.", Valid: true},
		}
		err = qtx.CreateCoursePhaseType(ctx, newInterviewPhase)
		if err != nil {
			log.Error("failed to create matching module: ", err)
			return err
		}

		// 2.) Create the required input meta data
		scoreSpecificationJson := meta.MetaData{}
		scoreSpecificationJson["type"] = "integer"
		scoreSpecificationBytes, err := scoreSpecificationJson.GetDBModel()
		if err != nil {
			log.Error("failed to parse score specification")
			return err
		}

		scoreLevelSpecificationBytes, err := getScoreLevelSpecificationBytes()
		if err != nil {
			log.Error("failed to parse score level specification")
			return err
		}

		newRequiredScoreInput := db.CreateCoursePhaseTypeRequiredInputParams{
			ID:                uuid.New(),
			CoursePhaseTypeID: newInterviewPhase.ID,
			DtoName:           "score",
			Specification:     scoreSpecificationBytes,
		}
		err = qtx.CreateCoursePhaseTypeRequiredInput(ctx, newRequiredScoreInput)
		if err != nil {
			log.Error("failed to create required score input: ", err)
			return err
		}

		err = qtx.CreateRequiredApplicationAnswers(ctx, newInterviewPhase.ID)
		if err != nil {
			log.Error("failed to create required application answers: ", err)
			return err
		}

		// 3.) Specify the provided output meta data
		newProvidedOutput := db.CreateCoursePhaseTypeProvidedOutputParams{
			ID:                uuid.New(),
			CoursePhaseTypeID: newInterviewPhase.ID,
			DtoName:           "score",
			Specification:     scoreSpecificationBytes,
			VersionNumber:     1,
			EndpointPath:      "/interview-review/score",
		}
		err = qtx.CreateCoursePhaseTypeProvidedOutput(ctx, newProvidedOutput)
		if err != nil {
			log.Error("failed to create required score input: ", err)
			return err
		}

		newProvidedScoreLevelOutput := db.CreateCoursePhaseTypeProvidedOutputParams{
			ID:                uuid.New(),
			CoursePhaseTypeID: newInterviewPhase.ID,
			DtoName:           "scoreLevel",
			Specification:     scoreLevelSpecificationBytes,
			VersionNumber:     1,
			EndpointPath:      "/interview-review/scoreLevel",
		}
		err = qtx.CreateCoursePhaseTypeProvidedOutput(ctx, newProvidedScoreLevelOutput)
		if err != nil {
			log.Error("failed to create score level output: ", err)
			return err
		}

		// 4.) Commit the transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	} else {
		log.Debug("interview module already exists")
	}
	return nil
}

func (s *CoursePhaseTypeService) initMatching() error {
	ctx := context.Background()
	exists, err := s.queries.TestMatchingPhaseTypeExists(ctx)

	if err != nil {
		log.Error("failed to check if matching phase type exists: ", err)
		return err
	}
	if !exists {
		tx, err := s.conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer sdkUtils.DeferRollback(tx, ctx)
		qtx := s.queries.WithTx(tx)

		// 1.) Create the phase
		newMatchingPhase := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "Matching",
			InitialPhase: false,
			BaseUrl:      "core",
			Description:  pgtype.Text{String: "A placeholder description for this course phase type. Detailed description will follow.", Valid: true},
		}
		err = qtx.CreateCoursePhaseType(ctx, newMatchingPhase)
		if err != nil {
			log.Error("failed to create matching module: ", err)
			return err
		}

		// 2.) Create the required input meta data
		scoreSpecificationJson := meta.MetaData{}
		scoreSpecificationJson["type"] = "integer"
		scoreSpecificationBytes, err := scoreSpecificationJson.GetDBModel()
		if err != nil {
			log.Error("failed to parse score specification")
			return err
		}

		newRequiredScoreInput := db.CreateCoursePhaseTypeRequiredInputParams{
			ID:                uuid.New(),
			CoursePhaseTypeID: newMatchingPhase.ID,
			DtoName:           "score",
			Specification:     scoreSpecificationBytes,
		}
		err = qtx.CreateCoursePhaseTypeRequiredInput(ctx, newRequiredScoreInput)
		if err != nil {
			log.Error("failed to create required score input: ", err)
			return err
		}

		// 3.) Commit the transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

	} else {
		log.Debug("matching module already exists")
	}

	return nil
}

func (s *CoursePhaseTypeService) initIntroCourseDeveloper() error {
	ctx := context.Background()
	exists, err := s.queries.TestIntroCourseDeveloperPhaseTypeExists(ctx)

	if err != nil {
		log.Error("failed to check if intro course developer phase type exists: ", err)
		return err
	}
	if !exists {
		tx, err := s.conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer sdkUtils.DeferRollback(tx, ctx)
		qtx := s.queries.WithTx(tx)

		// 1.) Create the phase
		baseURL := "{CORE_HOST}/intro-course/api"
		if s.isDevEnvironment {
			baseURL = "http://localhost:8082/intro-course/api"
		}
		newIntroCourseDeveloper := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "Intro Course Developer",
			InitialPhase: false,
			BaseUrl:      baseURL,
			Description:  pgtype.Text{String: "A placeholder description for this course phase type. Detailed description will follow.", Valid: true},
		}
		err = qtx.CreateCoursePhaseType(ctx, newIntroCourseDeveloper)
		if err != nil {
			log.Error("failed to create intro course developer module: ", err)
			return err
		}

		// 2.) Provided Output
		err = qtx.InsertProvidedOutputDevices(ctx, newIntroCourseDeveloper.ID)
		if err != nil {
			log.Error("failed to create required application answers output: ", err)
			return err
		}

		// 3.) Commit the transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

	} else {
		log.Debug("matching module already exists")
	}

	return nil
}

func (s *CoursePhaseTypeService) initDevOpsChallenge() error {
	ctx := context.Background()
	exists, err := s.queries.TestDevOpsChallengeTypeExists(ctx)

	if err != nil {
		log.Error("failed to check if dev ops challenge phase type exists: ", err)
		return err
	}
	if !exists {
		// 1.) Create the phase
		newDevOps := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "DevOps Challenge",
			InitialPhase: false,
			BaseUrl:      "core", // We use core here, as the server does not provide any exported DTOs
			Description:  pgtype.Text{String: "A placeholder description for this course phase type. Detailed description will follow.", Valid: true},
		}
		err = s.queries.CreateCoursePhaseType(ctx, newDevOps)
		if err != nil {
			log.Error("failed to create intro course developer module: ", err)
			return err
		}

		// No requires inputs and no provided outputs

	} else {
		log.Debug("dev ops challenge module already exists")
	}
	return nil
}

func (s *CoursePhaseTypeService) initAssessment() error {
	ctx := context.Background()
	exists, err := s.queries.TestAssessmentTypeExists(ctx)

	if err != nil {
		log.Error("failed to check if assessment phase type exists: ", err)
		return err
	}
	if !exists {
		tx, err := s.conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer sdkUtils.DeferRollback(tx, ctx)
		qtx := s.queries.WithTx(tx)

		// 1.) Create the phase
		baseURL := "{CORE_HOST}/assessment/api"
		if s.isDevEnvironment {
			baseURL = "http://localhost:8085/assessment/api"
		}

		// create the phase
		newAssessment := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "Assessment",
			InitialPhase: false,
			BaseUrl:      baseURL,
			Description:  pgtype.Text{String: "A placeholder description for this course phase type. Detailed description will follow.", Valid: true},
		}
		err = qtx.CreateCoursePhaseType(ctx, newAssessment)
		if err != nil {
			log.Error("failed to create assessment module: ", err)
			return err
		}

		// create the required phase input
		err = qtx.InsertTeamRequiredInput(ctx, db.InsertTeamRequiredInputParams{
			CoursePhaseTypeID: newAssessment.ID,
			Optional:          false,
		})
		if err != nil {
			log.Error("failed to create required team input: ", err)
			return err
		}

		// create the required participation input
		err = qtx.InsertTeamAllocationRequiredInput(ctx, db.InsertTeamAllocationRequiredInputParams{
			CoursePhaseTypeID: newAssessment.ID,
			Optional:          false,
		})
		if err != nil {
			log.Error("failed to create required team allocation input: ", err)
			return err
		}

		// create the participation output
		err = qtx.InsertAssessmentScoreOutput(ctx, newAssessment.ID)
		if err != nil {
			log.Error("failed to create required assessment output: ", err)
			return err
		}

		// create the actionItems output
		err = qtx.InsertActionItemsOutput(ctx, newAssessment.ID)
		if err != nil {
			log.Error("failed to create required action items output: ", err)
			return err
		}

		// create grade output
		err = qtx.InsertGradeOutput(ctx, newAssessment.ID)
		if err != nil {
			log.Error("failed to create required grade output: ", err)
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

	} else {
		log.Debug("assessment module already exists")
	}

	return nil
}

func (s *CoursePhaseTypeService) initTeamAllocation() error {
	ctx := context.Background()
	exists, err := s.queries.TestTeamAllocationTypeExists(ctx)

	if err != nil {
		log.Error("failed to check if team allocation phase type exists: ", err)
		return err
	}
	if !exists {
		tx, err := s.conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer sdkUtils.DeferRollback(tx, ctx)
		qtx := s.queries.WithTx(tx)

		// 1.) Create the phase
		baseURL := "{CORE_HOST}/team-allocation/api"
		if s.isDevEnvironment {
			baseURL = "http://localhost:8083/team-allocation/api"
		}

		newTeamAllocation := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "Team Allocation",
			InitialPhase: false,
			BaseUrl:      baseURL,
			Description:  pgtype.Text{String: "A placeholder description for this course phase type. Detailed description will follow.", Valid: true},
		}
		err = qtx.CreateCoursePhaseType(ctx, newTeamAllocation)
		if err != nil {
			log.Error("failed to create assessment module: ", err)
			return err
		}

		// 2.) Create the required input meta data

		// Languages from the application
		err = qtx.CreateRequiredApplicationAnswers(ctx, newTeamAllocation.ID)
		if err != nil {
			log.Error("failed to create required application answers: ", err)
			return err
		}

		// devices from the intro course
		err = qtx.CreateRequiredDevices(ctx, newTeamAllocation.ID)
		if err != nil {
			log.Error("failed to create required devices: ", err)
			return err
		}

		// score level from the intro course
		err = qtx.InsertAssessmentScoreRequiredInput(ctx, newTeamAllocation.ID)
		if err != nil {
			log.Error("failed to create required score level: ", err)
			return err
		}

		// 3.) Provided Output
		err = qtx.InsertTeamAllocationOutput(ctx, newTeamAllocation.ID)
		if err != nil {
			log.Error("failed to create required provided team allocation: ", err)
			return err
		}

		err = qtx.InsertTeamOutput(ctx, newTeamAllocation.ID)
		if err != nil {
			log.Error("failed to create required provided teams: ", err)
			return err
		}

		// 3.) Commit the transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

	} else {
		log.Debug("team allocation module already exists")
	}

	return nil
}

func (s *CoursePhaseTypeService) initSelfTeamAllocation() error {
	ctx := context.Background()
	exists, err := s.queries.TestSelfTeamAllocationTypeExists(ctx)

	if err != nil {
		log.Error("failed to check if self team allocation phase type exists: ", err)
		return err
	}
	if !exists {
		tx, err := s.conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer sdkUtils.DeferRollback(tx, ctx)
		qtx := s.queries.WithTx(tx)

		// 1.) Create the phase
		baseURL := "{CORE_HOST}/self-team-allocation/api"
		if s.isDevEnvironment {
			baseURL = "http://localhost:8084/self-team-allocation/api"
		}

		newTeamAllocation := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "Self Team Allocation",
			InitialPhase: false,
			BaseUrl:      baseURL,
			Description:  pgtype.Text{String: "A placeholder description for this course phase type. Detailed description will follow.", Valid: true},
		}
		err = qtx.CreateCoursePhaseType(ctx, newTeamAllocation)
		if err != nil {
			log.Error("failed to create assessment module: ", err)
			return err
		}

		// 2.) Provided Output
		err = qtx.InsertTeamAllocationOutput(ctx, newTeamAllocation.ID)
		if err != nil {
			log.Error("failed to create required provided team allocation: ", err)
			return err
		}

		err = qtx.InsertTeamOutput(ctx, newTeamAllocation.ID)
		if err != nil {
			log.Error("failed to create required provided teams: ", err)
			return err
		}

		// 3.) Commit the transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

	} else {
		log.Debug("team allocation module already exists")
	}

	return nil
}

func (s *CoursePhaseTypeService) initCertificate() error {
	ctx := context.Background()
	exists, err := s.queries.TestCertificateTypeExists(ctx)

	if err != nil {
		log.Error("failed to check if certificate phase type exists: ", err)
		return err
	}
	if !exists {
		// Begin transaction
		tx, err := s.conn.Begin(ctx)
		if err != nil {
			log.Error("failed to begin transaction: ", err)
			return err
		}
		defer sdkUtils.DeferRollback(tx, ctx)
		qtx := s.queries.WithTx(tx)

		// 1.) Create the phase
		baseURL := "{CORE_HOST}/certificate/api"
		if s.isDevEnvironment {
			baseURL = "http://localhost:8088/certificate/api"
		}

		newCertificate := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "Certificate",
			Description:  pgtype.Text{String: "Certificate of completion generation and distribution.", Valid: true},
			InitialPhase: false,
			BaseUrl:      baseURL,
		}
		err = qtx.CreateCoursePhaseType(ctx, newCertificate)
		if err != nil {
			log.Error("failed to create certificate phase type: ", err)
			return err
		}

		// 2.) Create required inputs - Certificate phase typically needs team and student data
		// Team allocation input (to know which teams exist)
		err = qtx.InsertTeamAllocationRequiredInput(ctx, db.InsertTeamAllocationRequiredInputParams{
			CoursePhaseTypeID: newCertificate.ID,
			Optional:          false,
		})
		if err != nil {
			log.Error("failed to create required team allocation input: ", err)
			return err
		}

		// Team input (to get team information)
		err = qtx.InsertTeamRequiredInput(ctx, db.InsertTeamRequiredInputParams{
			CoursePhaseTypeID: newCertificate.ID,
			Optional:          false,
		})
		if err != nil {
			log.Error("failed to create required team input: ", err)
			return err
		}

		// No provided outputs for certificate phase (it's typically an end phase)

		// 3.) Commit the transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

	} else {
		log.Debug("certificate phase type already exists")
	}

	return nil
}

func (s *CoursePhaseTypeService) initPresentation() error {
	ctx := context.Background()
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	baseURL := "{CORE_HOST}/presentation/api"
	if s.isDevEnvironment {
		baseURL = "http://localhost:8089/presentation/api"
	}

	presentationPhaseID := uuid.New()
	presentationPhase, err := qtx.GetCoursePhaseTypeByName(ctx, "Presentation")
	switch {
	case err == nil:
		presentationPhaseID = presentationPhase.ID
		log.Debug("presentation phase type already exists; ensuring its input descriptors")
	case errors.Is(err, pgx.ErrNoRows):
		newPresentationPhase := db.CreateCoursePhaseTypeParams{
			ID:           presentationPhaseID,
			Name:         "Presentation",
			InitialPhase: false,
			BaseUrl:      baseURL,
			Description:  pgtype.Text{String: "Presentation scheduling, material submission, and instructor feedback.", Valid: true},
		}
		if err := qtx.CreateCoursePhaseType(ctx, newPresentationPhase); err != nil {
			log.Error("failed to create presentation phase type: ", err)
			return err
		}
	default:
		log.Error("failed to get presentation phase type: ", err)
		return err
	}

	if err := qtx.InsertTeamRequiredInput(ctx, db.InsertTeamRequiredInputParams{
		CoursePhaseTypeID: presentationPhaseID,
		Optional:          true,
	}); err != nil {
		log.Error("failed to create optional team input: ", err)
		return err
	}

	if err := qtx.InsertTeamAllocationRequiredInput(ctx, db.InsertTeamAllocationRequiredInputParams{
		CoursePhaseTypeID: presentationPhaseID,
		Optional:          true,
	}); err != nil {
		log.Error("failed to create optional team allocation input: ", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *CoursePhaseTypeService) initInfrastructureSetup() error {
	ctx := context.Background()
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	baseURL := "{CORE_HOST}/infrastructure-setup/api"
	if s.isDevEnvironment {
		baseURL = "http://localhost:8091/infrastructure-setup/api"
	}

	infrastructureSetupPhaseID := uuid.New()
	infrastructureSetupPhase, err := qtx.GetCoursePhaseTypeByName(ctx, "Infrastructure Setup")
	switch {
	case err == nil:
		infrastructureSetupPhaseID = infrastructureSetupPhase.ID
		log.Debug("infrastructure setup phase type already exists; ensuring its input descriptors")
	case errors.Is(err, pgx.ErrNoRows):
		newInfrastructureSetupPhase := db.CreateCoursePhaseTypeParams{
			ID:           infrastructureSetupPhaseID,
			Name:         "Infrastructure Setup",
			InitialPhase: false,
			BaseUrl:      baseURL,
			Description:  pgtype.Text{String: "Automated provisioning of external resources (GitLab, Slack, Outline, etc.) per team or student.", Valid: true},
		}
		if err := qtx.CreateCoursePhaseType(ctx, newInfrastructureSetupPhase); err != nil {
			log.Error("failed to create infrastructure setup phase type: ", err)
			return err
		}
	default:
		log.Error("failed to get infrastructure setup phase type: ", err)
		return err
	}

	if err := qtx.InsertTeamRequiredInput(ctx, db.InsertTeamRequiredInputParams{
		CoursePhaseTypeID: infrastructureSetupPhaseID,
		Optional:          false,
	}); err != nil {
		log.Error("failed to create required team input: ", err)
		return err
	}

	if err := qtx.InsertTeamAllocationRequiredInput(ctx, db.InsertTeamAllocationRequiredInputParams{
		CoursePhaseTypeID: infrastructureSetupPhaseID,
		Optional:          false,
	}); err != nil {
		log.Error("failed to create required team allocation input: ", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
