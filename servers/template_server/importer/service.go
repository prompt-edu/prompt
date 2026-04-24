package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/template_server/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type ImporterService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var ImporterServiceSingleton *ImporterService

func (s *ImporterService) ImportPromptReadyTemplate(
	ctx context.Context,
	authHeader string,
	payload PromptReadyTemplate,
) (*ImportPromptReadyTemplateResponse, error) {
	if err := validatePromptReadyTemplate(payload); err != nil {
		return nil, err
	}

	coreBaseURL := strings.TrimRight(sdkUtils.GetCoreUrl(), "/")

	phaseTypes, err := fetchCoursePhaseTypes(coreBaseURL, authHeader)
	if err != nil {
		return nil, err
	}

	applicationPhaseType, err := findPhaseTypeByName(phaseTypes, "Application")
	if err != nil {
		return nil, err
	}

	teamAllocationPhaseType, err := findPhaseTypeByName(phaseTypes, "Team Allocation")
	if err != nil {
		return nil, err
	}

	if payload.TemplateMeta.ReplaceExistingByID {
		if err := replaceExistingTemplateCourses(
			ctx,
			coreBaseURL,
			authHeader,
			payload.TemplateMeta.TemplateID,
			payload.CourseCreationDefaults.Name,
			payload.CourseCreationDefaults.SemesterTag,
		); err != nil {
			return nil, err
		}
	}

	courseResp, err := createCourse(ctx, coreBaseURL, authHeader, payload)
	if err != nil {
		return nil, err
	}

	applicationPhaseResp, err := createCoursePhase(
		ctx,
		coreBaseURL,
		authHeader,
		courseResp.ID,
		applicationPhaseType.ID,
		"Application",
		true,
		map[string]any{},
		map[string]any{},
	)
	if err != nil {
		return nil, err
	}

	teamAllocationRestrictedData := map[string]any{
		"teamAllocationMode":      payload.TemplateMeta.TeamAllocationMode,
		"defaultMatchingStrategy": payload.TemplateMeta.DefaultStrategyInTease,
	}

	teamAllocationPhaseResp, err := createCoursePhase(
		ctx,
		coreBaseURL,
		authHeader,
		courseResp.ID,
		teamAllocationPhaseType.ID,
		"Team Allocation",
		false,
		teamAllocationRestrictedData,
		map[string]any{},
	)
	if err != nil {
		return nil, err
	}

	if err := updateCoursePhaseGraph(
		ctx,
		coreBaseURL,
		authHeader,
		courseResp.ID,
		applicationPhaseResp.ID,
		[]coursePhaseGraphItem{
			{
				FromCoursePhaseID: applicationPhaseResp.ID,
				ToCoursePhaseID:   teamAllocationPhaseResp.ID,
			},
		},
	); err != nil {
		return nil, err
	}

	applicationAnswersOutputID, err := findParticipationOutputDTOID(applicationPhaseType, "applicationAnswers")
	if err != nil {
		return nil, err
	}
	applicationAnswersInputID, err := findParticipationInputDTOID(teamAllocationPhaseType, "applicationAnswers")
	if err != nil {
		return nil, err
	}

	if err := updateParticipationDataGraph(
		ctx,
		coreBaseURL,
		authHeader,
		courseResp.ID,
		[]metaDataGraphItem{
			{
				FromCoursePhaseID:    applicationPhaseResp.ID,
				ToCoursePhaseID:      teamAllocationPhaseResp.ID,
				FromCoursePhaseDtoID: applicationAnswersOutputID,
				ToCoursePhaseDtoID:   applicationAnswersInputID,
			},
		},
	); err != nil {
		return nil, err
	}

	if err := updateApplicationForm(
		ctx,
		coreBaseURL,
		authHeader,
		applicationPhaseResp.ID,
		payload.ApplicationFormUpdatePayload.withCoursePhaseID(applicationPhaseResp.ID),
	); err != nil {
		return nil, err
	}

	teamAllocationBaseURL := resolveBaseURL(teamAllocationPhaseType.BaseURL)

	if len(payload.TeamAllocationPhaseSetup.Skills) > 0 {
		if err := createTeamAllocationSkills(
			ctx,
			teamAllocationBaseURL,
			authHeader,
			teamAllocationPhaseResp.ID,
			payload.TeamAllocationPhaseSetup.Skills,
		); err != nil {
			return nil, err
		}
	}

	if len(payload.TeamAllocationPhaseSetup.Teams) > 0 {
		if err := createTeamAllocationTeams(
			ctx,
			teamAllocationBaseURL,
			authHeader,
			teamAllocationPhaseResp.ID,
			payload.TeamAllocationPhaseSetup.Teams,
		); err != nil {
			return nil, err
		}
	}

	warnings := buildImportWarnings(teamAllocationPhaseType)

	return &ImportPromptReadyTemplateResponse{
		CourseID:                courseResp.ID,
		ApplicationPhaseID:      applicationPhaseResp.ID,
		TeamAllocationPhaseID:   teamAllocationPhaseResp.ID,
		TeamAllocationMode:      payload.TemplateMeta.TeamAllocationMode,
		DefaultMatchingStrategy: payload.TemplateMeta.DefaultStrategyInTease,
		Warnings:                warnings,
	}, nil
}

func validatePromptReadyTemplate(payload PromptReadyTemplate) error {
	if strings.TrimSpace(payload.CourseCreationDefaults.Name) == "" {
		return fmt.Errorf("courseCreationDefaults.name is required")
	}
	if strings.TrimSpace(payload.CourseCreationDefaults.SemesterTag) == "" {
		return fmt.Errorf("courseCreationDefaults.semesterTag is required")
	}
	if strings.TrimSpace(payload.TemplateMeta.TeamAllocationMode) == "" {
		return fmt.Errorf("templateMeta.teamAllocationMode is required")
	}
	if strings.TrimSpace(payload.TemplateMeta.DefaultStrategyInTease) == "" {
		return fmt.Errorf("templateMeta.defaultStrategyInTease is required")
	}
	if payload.TemplateMeta.ReplaceExistingByID && strings.TrimSpace(payload.TemplateMeta.TemplateID) == "" {
		return fmt.Errorf("templateMeta.templateId is required when replaceExistingById is true")
	}
	return nil
}

func fetchCoursePhaseTypes(coreBaseURL, authHeader string) ([]coursePhaseType, error) {
	body, err := sdkUtils.FetchJSON(fmt.Sprintf("%s/api/course_phase_types", coreBaseURL), authHeader)
	if err != nil {
		return nil, err
	}

	var phaseTypes []coursePhaseType
	if err := json.Unmarshal(body, &phaseTypes); err != nil {
		return nil, err
	}

	return phaseTypes, nil
}

func findPhaseTypeByName(phaseTypes []coursePhaseType, phaseTypeName string) (coursePhaseType, error) {
	for _, phaseType := range phaseTypes {
		if strings.EqualFold(phaseType.Name, phaseTypeName) {
			return phaseType, nil
		}
	}
	return coursePhaseType{}, fmt.Errorf("course phase type %q not found", phaseTypeName)
}

func findParticipationOutputDTOID(phaseType coursePhaseType, dtoName string) (uuid.UUID, error) {
	for _, dto := range phaseType.ProvidedParticipationOutputDTOs {
		if dto.DtoName == dtoName {
			return dto.ID, nil
		}
	}
	return uuid.UUID{}, fmt.Errorf("provided participation DTO %q not found in phase type %q", dtoName, phaseType.Name)
}

func findParticipationInputDTOID(phaseType coursePhaseType, dtoName string) (uuid.UUID, error) {
	for _, dto := range phaseType.RequiredParticipationInputDTOs {
		if dto.DtoName == dtoName {
			return dto.ID, nil
		}
	}
	return uuid.UUID{}, fmt.Errorf("required participation DTO %q not found in phase type %q", dtoName, phaseType.Name)
}

func createCourse(
	ctx context.Context,
	coreBaseURL, authHeader string,
	payload PromptReadyTemplate,
) (*courseResponse, error) {
	now := time.Now()
	requestBody := createCourseRequest{
		Name:        payload.CourseCreationDefaults.Name,
		StartDate:   now.Format("2006-01-02"),
		EndDate:     now.Format("2006-01-02"),
		SemesterTag: payload.CourseCreationDefaults.SemesterTag,
		RestrictedData: map[string]any{
			"templateSource": map[string]any{
				"templateId":      payload.TemplateMeta.TemplateID,
				"templateVersion": payload.TemplateMeta.TemplateVersion,
				"templateName":    payload.TemplateMeta.Name,
			},
			"teamAllocationMode":      payload.TemplateMeta.TeamAllocationMode,
			"defaultMatchingStrategy": payload.TemplateMeta.DefaultStrategyInTease,
		},
		StudentReadableData: payload.CourseCreationDefaults.StudentReadableData,
		ShortDescription:    payload.CourseCreationDefaults.ShortDescription,
		LongDescription:     payload.CourseCreationDefaults.LongDescription,
		CourseType:          normalizeCourseType(payload.TemplateMeta.RecommendedCourseType),
		Ects:                payload.CourseCreationDefaults.Ects,
		Template:            true,
	}

	var response courseResponse
	if err := sendJSON(ctx, http.MethodPost, fmt.Sprintf("%s/api/courses/", coreBaseURL), authHeader, requestBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func replaceExistingTemplateCourses(
	ctx context.Context,
	coreBaseURL, authHeader, templateID, courseName, semesterTag string,
) error {
	existingTemplates, err := fetchTemplateCourses(coreBaseURL, authHeader)
	if err != nil {
		return err
	}

	for _, templateCourse := range existingTemplates {
		if !shouldReplaceTemplateCourse(templateCourse, templateID, courseName, semesterTag) {
			continue
		}
		if err := deleteCourse(ctx, coreBaseURL, authHeader, templateCourse.ID); err != nil {
			return err
		}
	}

	return nil
}

func fetchTemplateCourses(coreBaseURL, authHeader string) ([]templateCourseResponse, error) {
	body, err := sdkUtils.FetchJSON(fmt.Sprintf("%s/api/courses/template", coreBaseURL), authHeader)
	if err != nil {
		return nil, err
	}

	var courses []templateCourseResponse
	if err := json.Unmarshal(body, &courses); err != nil {
		return nil, err
	}

	return courses, nil
}

func matchesTemplateID(restrictedData map[string]any, templateID string) bool {
	if restrictedData == nil {
		return false
	}

	templateSourceRaw, ok := restrictedData["templateSource"]
	if !ok {
		return false
	}

	templateSource, ok := templateSourceRaw.(map[string]any)
	if !ok {
		return false
	}

	storedTemplateID, ok := templateSource["templateId"].(string)
	if !ok {
		return false
	}

	return storedTemplateID == templateID
}

func matchesCourseIdentity(course templateCourseResponse, courseName, semesterTag string) bool {
	return course.Template &&
		course.Name == courseName &&
		course.SemesterTag == semesterTag
}

func shouldReplaceTemplateCourse(
	course templateCourseResponse,
	templateID, courseName, semesterTag string,
) bool {
	return matchesTemplateID(course.RestrictedData, templateID) ||
		matchesCourseIdentity(course, courseName, semesterTag)
}

func createCoursePhase(
	ctx context.Context,
	coreBaseURL, authHeader string,
	courseID, coursePhaseTypeID uuid.UUID,
	name string,
	isInitialPhase bool,
	restrictedData map[string]any,
	studentReadableData map[string]any,
) (*coursePhaseResponse, error) {
	requestBody := createCoursePhaseRequest{
		CourseID:            courseID,
		Name:                name,
		IsInitialPhase:      isInitialPhase,
		RestrictedData:      restrictedData,
		StudentReadableData: studentReadableData,
		CoursePhaseTypeID:   coursePhaseTypeID,
	}

	var response coursePhaseResponse
	if err := sendJSON(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/course_phases/course/%s", coreBaseURL, courseID),
		authHeader,
		requestBody,
		&response,
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func updateCoursePhaseGraph(
	ctx context.Context,
	coreBaseURL, authHeader string,
	courseID, initialPhaseID uuid.UUID,
	phaseGraph []coursePhaseGraphItem,
) error {
	requestBody := updateCoursePhaseGraphRequest{
		InitialPhase:     initialPhaseID,
		CoursePhaseGraph: phaseGraph,
	}

	return sendJSON(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s/api/courses/%s/phase_graph", coreBaseURL, courseID),
		authHeader,
		requestBody,
		nil,
	)
}

func updateParticipationDataGraph(
	ctx context.Context,
	coreBaseURL, authHeader string,
	courseID uuid.UUID,
	graph []metaDataGraphItem,
) error {
	return sendJSON(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s/api/courses/%s/participation_data_graph", coreBaseURL, courseID),
		authHeader,
		graph,
		nil,
	)
}

func updateApplicationForm(
	ctx context.Context,
	coreBaseURL, authHeader string,
	coursePhaseID uuid.UUID,
	payload applicationFormUpdatePayload,
) error {
	return sendJSON(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s/api/applications/%s/form", coreBaseURL, coursePhaseID),
		authHeader,
		payload,
		nil,
	)
}

func deleteCourse(
	ctx context.Context,
	coreBaseURL, authHeader string,
	courseID uuid.UUID,
) error {
	return sendJSON(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s/api/courses/%s", coreBaseURL, courseID),
		authHeader,
		nil,
		nil,
	)
}

func createTeamAllocationSkills(
	ctx context.Context,
	teamAllocationBaseURL, authHeader string,
	coursePhaseID uuid.UUID,
	skills []string,
) error {
	requestBody := createSkillsRequest{SkillNames: skills}
	return sendJSON(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/course_phase/%s/skill", teamAllocationBaseURL, coursePhaseID),
		authHeader,
		requestBody,
		nil,
	)
}

func createTeamAllocationTeams(
	ctx context.Context,
	teamAllocationBaseURL, authHeader string,
	coursePhaseID uuid.UUID,
	teams []string,
) error {
	requestBody := createTeamsRequest{TeamNames: teams}
	return sendJSON(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/course_phase/%s/team", teamAllocationBaseURL, coursePhaseID),
		authHeader,
		requestBody,
		nil,
	)
}

func resolveBaseURL(baseURL string) string {
	resolved := strings.TrimSpace(baseURL)
	if strings.Contains(resolved, "{CORE_HOST}") {
		resolved = strings.ReplaceAll(
			resolved,
			"{CORE_HOST}",
			promptCoreHost(),
		)
	}
	return strings.TrimRight(resolved, "/")
}

func promptCoreHost() string {
	return sdkUtils.GetEnv("CORE_HOST", "http://localhost:3000")
}

func buildImportWarnings(teamAllocationPhaseType coursePhaseType) []string {
	requiredInputNames := make(map[string]bool, len(teamAllocationPhaseType.RequiredParticipationInputDTOs))
	for _, dto := range teamAllocationPhaseType.RequiredParticipationInputDTOs {
		requiredInputNames[dto.DtoName] = true
	}

	warnings := []string{}
	if requiredInputNames["devices"] {
		warnings = append(
			warnings,
			`Team Allocation currently expects the "devices" participation input, but this importer only wires applicationAnswers.`,
		)
	}
	if requiredInputNames["scoreLevel"] {
		warnings = append(
			warnings,
			`Team Allocation currently expects the "scoreLevel" participation input, but this importer does not add an upstream provider for it.`,
		)
	}

	return warnings
}

func sendJSON(
	ctx context.Context,
	method, url, authHeader string,
	requestBody any,
	responseTarget any,
) error {
	var bodyReader io.Reader
	if requestBody != nil {
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	resp, err := sdkUtils.SendCoreRequest(ctx, method, authHeader, bodyReader, url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request %s %s failed with status %d: %s", method, url, resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	if responseTarget == nil {
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(bodyBytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(bodyBytes, responseTarget); err != nil {
		log.WithField("url", url).Error("failed to decode JSON response")
		return err
	}

	return nil
}

func normalizeCourseType(raw string) string {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	switch normalized {
	case "lecture":
		return "lecture"
	case "seminar":
		return "seminar"
	case "practical course", "practicalcourse":
		return "practical course"
	default:
		return "seminar"
	}
}
