package categories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/schemaModification"
	"github.com/prompt-edu/prompt/servers/assessment/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	orderCoursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	orderSchemaID      = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	versionControlID = uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e")
	userInterfaceID  = uuid.MustParse("815b159b-cab3-49b4-8060-c4722d59241d")
	fundamentalsID   = uuid.MustParse("9107c0aa-15b7-4967-bf62-6fa131f08bee")
)

func TestValidateSchemaOrder(t *testing.T) {
	categoryA, categoryB := uuid.New(), uuid.New()
	competencyA1, competencyA2, competencyB1 := uuid.New(), uuid.New(), uuid.New()
	current := []categoryDTO.CategoryWithCompetencies{
		{ID: categoryA, Competencies: []competencyDTO.Competency{{ID: competencyA1, Name: "Shared"}, {ID: competencyA2, Name: "Only A"}}},
		{ID: categoryB, Competencies: []competencyDTO.Competency{{ID: competencyB1, Name: "Shared"}}},
	}

	tests := []struct {
		name  string
		order []categoryDTO.CategoryOrder
		want  error
	}{
		{
			name: "reordered and moved",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1, competencyA2}},
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1}},
			},
		},
		{
			name:  "missing category",
			order: []categoryDTO.CategoryOrder{{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2, competencyB1}}},
			want:  ErrIncompleteSchemaOrder,
		},
		{
			name: "unknown category",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2}},
				{ID: uuid.New(), CompetencyIDs: []uuid.UUID{competencyB1}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "duplicate category",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2}},
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyB1}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "missing competency",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1}},
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "competency listed twice",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2}},
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1, competencyA2}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "unknown competency",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2, uuid.New()}},
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "duplicate name after move",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA2}},
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1, competencyA1}},
			},
			want: ErrDuplicateCompetencyName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchemaOrder(current, tt.order)
			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.want)
			}
		})
	}
}

type SchemaOrderTestSuite struct {
	suite.Suite
	suiteCtx                 context.Context
	cleanup                  func()
	mockCoreCleanup          func()
	categoryService          *CategoryService
	schemaService            *assessmentSchemas.AssessmentSchemaService
	coursePhaseConfigService *coursePhaseConfig.CoursePhaseConfigService
}

func (suite *SchemaOrderTestSuite) SetupTest() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.suiteCtx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/categories.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	_, mockCleanup := testutils.SetupMockCoreService()
	suite.mockCoreCleanup = mockCleanup

	suite.schemaService = assessmentSchemas.NewAssessmentSchemaService(*testDB.Queries, testDB.Conn)
	suite.coursePhaseConfigService = coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn, suite.schemaService)
	suite.categoryService = NewCategoryService(*testDB.Queries, testDB.Conn, suite.schemaService, schemaModification.NewSchemaModificationService(suite.schemaService, suite.coursePhaseConfigService, *testDB.Queries), suite.coursePhaseConfigService)
}

func (suite *SchemaOrderTestSuite) TearDownTest() {
	if suite.mockCoreCleanup != nil {
		suite.mockCoreCleanup()
	}
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *SchemaOrderTestSuite) categoryByID(schemaID uuid.UUID) map[uuid.UUID]categoryDTO.CategoryWithCompetencies {
	categories, err := suite.categoryService.GetCategoriesWithCompetencies(suite.suiteCtx, schemaID)
	require.NoError(suite.T(), err)
	byID := make(map[uuid.UUID]categoryDTO.CategoryWithCompetencies, len(categories))
	for _, category := range categories {
		byID[category.ID] = category
	}
	return byID
}

func layoutOf(categories []categoryDTO.CategoryWithCompetencies) map[string][]string {
	layout := make(map[string][]string, len(categories))
	for _, category := range categories {
		names := make([]string, 0, len(category.Competencies))
		for _, competency := range category.Competencies {
			names = append(names, competency.Name)
		}
		layout[category.Name] = names
	}
	return layout
}

func categoryNames(categories []categoryDTO.CategoryWithCompetencies) []string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		names = append(names, category.Name)
	}
	return names
}

func competencyIDs(competencies []competencyDTO.Competency) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(competencies))
	for _, competency := range competencies {
		ids = append(ids, competency.ID)
	}
	return ids
}

func (suite *SchemaOrderTestSuite) TestExistingSchemaKeepsAlphabeticalOrder() {
	categories, err := suite.categoryService.GetCategoriesWithCompetencies(suite.suiteCtx, orderSchemaID)
	suite.Require().NoError(err)

	suite.Equal([]string{"Fundamentals in Software Engineering", "User Interface", "Version Control"}, categoryNames(categories))
}

// The phase consumes a global schema, so the first reorder copies it; the second one then edits the copy directly.
func (suite *SchemaOrderTestSuite) TestUpdateSchemaOrderCopiesSharedSchemaThenEditsCopy() {
	original := suite.categoryByID(orderSchemaID)
	versionControl := competencyIDs(original[versionControlID].Competencies)
	userInterface := competencyIDs(original[userInterfaceID].Competencies)
	fundamentals := competencyIDs(original[fundamentalsID].Competencies)

	movedCompetency := versionControl[0]
	err := suite.categoryService.UpdateSchemaOrder(suite.suiteCtx, orderCoursePhaseID, categoryDTO.UpdateSchemaOrderRequest{
		Categories: []categoryDTO.CategoryOrder{
			{ID: userInterfaceID, CompetencyIDs: []uuid.UUID{userInterface[1], movedCompetency, userInterface[0]}},
			{ID: versionControlID, CompetencyIDs: versionControl[1:]},
			{ID: fundamentalsID, CompetencyIDs: []uuid.UUID{fundamentals[2], fundamentals[0], fundamentals[1]}},
		},
	})
	suite.Require().NoError(err)

	config, err := suite.coursePhaseConfigService.GetCoursePhaseConfig(suite.suiteCtx, orderCoursePhaseID)
	suite.Require().NoError(err)
	copiedSchemaID := config.AssessmentSchemaID
	suite.NotEqual(orderSchemaID, copiedSchemaID, "reordering a shared schema should give the phase its own copy")

	reordered, err := suite.categoryService.GetCategoriesWithCompetencies(suite.suiteCtx, copiedSchemaID)
	suite.Require().NoError(err)
	suite.Equal([]string{"User Interface", "Version Control", "Fundamentals in Software Engineering"}, categoryNames(reordered))

	competencyName := func(category uuid.UUID, index int) string {
		return original[category].Competencies[index].Name
	}
	suite.Equal(map[string][]string{
		"User Interface":                       {competencyName(userInterfaceID, 1), competencyName(versionControlID, 0), competencyName(userInterfaceID, 0)},
		"Version Control":                      {competencyName(versionControlID, 1)},
		"Fundamentals in Software Engineering": {competencyName(fundamentalsID, 2), competencyName(fundamentalsID, 0), competencyName(fundamentalsID, 1)},
	}, layoutOf(reordered))

	untouched, err := suite.categoryService.GetCategoriesWithCompetencies(suite.suiteCtx, orderSchemaID)
	suite.Require().NoError(err)
	suite.Equal([]string{"Fundamentals in Software Engineering", "User Interface", "Version Control"}, categoryNames(untouched),
		"the shared schema should keep its order")

	schemasBefore, err := suite.schemaService.ListAssessmentSchemas(suite.suiteCtx)
	suite.Require().NoError(err)

	reversed := make([]categoryDTO.CategoryOrder, 0, len(reordered))
	for i := len(reordered) - 1; i >= 0; i-- {
		reversed = append(reversed, categoryDTO.CategoryOrder{ID: reordered[i].ID, CompetencyIDs: competencyIDs(reordered[i].Competencies)})
	}
	err = suite.categoryService.UpdateSchemaOrder(suite.suiteCtx, orderCoursePhaseID, categoryDTO.UpdateSchemaOrderRequest{
		Categories: reversed,
	})
	suite.Require().NoError(err)

	schemasAfter, err := suite.schemaService.ListAssessmentSchemas(suite.suiteCtx)
	suite.Require().NoError(err)
	suite.Len(schemasAfter, len(schemasBefore), "the phase owns its copy now, so no further copy is made")

	final, err := suite.categoryService.GetCategoriesWithCompetencies(suite.suiteCtx, copiedSchemaID)
	suite.Require().NoError(err)
	suite.Equal([]string{"Fundamentals in Software Engineering", "Version Control", "User Interface"}, categoryNames(final))
}

func (suite *SchemaOrderTestSuite) TestUpdateSchemaOrderRejectsIncompleteOrderWithoutCopying() {
	schemasBefore, err := suite.schemaService.ListAssessmentSchemas(suite.suiteCtx)
	suite.Require().NoError(err)

	err = suite.categoryService.UpdateSchemaOrder(suite.suiteCtx, orderCoursePhaseID, categoryDTO.UpdateSchemaOrderRequest{
		Categories: []categoryDTO.CategoryOrder{{ID: userInterfaceID}},
	})
	suite.ErrorIs(err, ErrIncompleteSchemaOrder)

	schemasAfter, err := suite.schemaService.ListAssessmentSchemas(suite.suiteCtx)
	suite.Require().NoError(err)
	suite.Len(schemasAfter, len(schemasBefore))
}

func (suite *SchemaOrderTestSuite) TestCreatedEntriesAreAppended() {
	created, err := suite.categoryService.CreateCategory(suite.suiteCtx, orderCoursePhaseID, categoryDTO.CreateCategoryRequest{
		Name:               "Accessibility",
		ShortName:          "A11y",
		Weight:             1,
		AssessmentSchemaID: orderSchemaID,
	})
	suite.Require().NoError(err)

	categories, err := suite.categoryService.GetCategoriesWithCompetencies(suite.suiteCtx, created.AssessmentSchemaID)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(categories)
	suite.Equal("Accessibility", categories[len(categories)-1].Name, "a new category should come last despite its name")
}

func TestSchemaOrderTestSuite(t *testing.T) {
	suite.Run(t, new(SchemaOrderTestSuite))
}
