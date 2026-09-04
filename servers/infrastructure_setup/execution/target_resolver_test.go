package execution

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

func TestParseTeams(t *testing.T) {
	teamID := uuid.New()
	memberID := uuid.New()
	tutorID := uuid.New()

	teams, err := parseTeams([]interface{}{
		map[string]interface{}{
			"id":   teamID.String(),
			"name": "Team A",
			"members": []interface{}{
				map[string]interface{}{"id": memberID.String(), "firstName": "Max", "lastName": "Muster"},
			},
			"tutors": []interface{}{
				map[string]interface{}{"id": tutorID.String(), "firstName": "Tina", "lastName": "Tutor"},
			},
		},
	})
	if err != nil {
		t.Fatalf("parseTeams: %v", err)
	}
	if len(teams) != 1 {
		t.Fatalf("teams = %d, want 1", len(teams))
	}
	if teams[0].ID != teamID || teams[0].Name != "Team A" {
		t.Fatalf("team = %+v, want %s/Team A", teams[0], teamID)
	}
	if len(teams[0].Members) != 1 || teams[0].Members[0].ID != memberID {
		t.Fatalf("members = %+v, want one entry with id %s", teams[0].Members, memberID)
	}
	if len(teams[0].Tutors) != 1 || teams[0].Tutors[0].ID != tutorID {
		t.Fatalf("tutors = %+v, want one entry with id %s", teams[0].Tutors, tutorID)
	}
}

// Core is an upstream service, so a malformed entry must be skipped rather than abort
// the whole run. Only a payload that is not a list at all is an error.
func TestParseTeamsSkipsMalformedEntries(t *testing.T) {
	valid := uuid.New()

	teams, err := parseTeams([]interface{}{
		"not a map",
		map[string]interface{}{"name": "Missing ID"},
		map[string]interface{}{"id": "not-a-uuid", "name": "Bad ID"},
		map[string]interface{}{"id": uuid.New().String()},
		map[string]interface{}{"id": valid.String(), "name": "Team A"},
	})
	if err != nil {
		t.Fatalf("parseTeams: %v", err)
	}
	if len(teams) != 1 {
		t.Fatalf("teams = %d, want only the valid one", len(teams))
	}
	if teams[0].ID != valid {
		t.Fatalf("team id = %s, want %s", teams[0].ID, valid)
	}
}

func TestParseTeamsHandlesEmptyAndInvalidPayloads(t *testing.T) {
	teams, err := parseTeams(nil)
	if err != nil {
		t.Fatalf("parseTeams(nil): %v", err)
	}
	if len(teams) != 0 {
		t.Fatalf("teams = %d, want 0", len(teams))
	}

	if _, err := parseTeams(map[string]interface{}{"teams": "wrong shape"}); err == nil {
		t.Fatal("parseTeams returned no error for a payload that is not a list")
	}
}

func TestParsePersonsSkipsEntriesWithoutValidID(t *testing.T) {
	valid := uuid.New()

	persons := parsePersons([]interface{}{
		"not a map",
		map[string]interface{}{"firstName": "No ID"},
		map[string]interface{}{"id": "not-a-uuid"},
		map[string]interface{}{"id": valid.String(), "firstName": "Max", "lastName": "Muster"},
	})
	if len(persons) != 1 {
		t.Fatalf("persons = %d, want only the valid one", len(persons))
	}
	if persons[0].ID != valid || persons[0].FirstName != "Max" {
		t.Fatalf("person = %+v, want %s/Max", persons[0], valid)
	}

	if got := parsePersons(nil); len(got) != 0 {
		t.Fatalf("parsePersons(nil) = %d entries, want 0", len(got))
	}
	if got := parsePersons("wrong shape"); len(got) != 0 {
		t.Fatalf("parsePersons(non-list) = %d entries, want 0", len(got))
	}
}

// targetIndex is what lets the worker resolve every scope once per run instead of once
// per instance.
func TestTargetIndexLooksUpByTeamAndStudent(t *testing.T) {
	teamID := uuid.New()
	participationID := uuid.New()

	index := newTargetIndex([]ProvisioningTarget{
		{TeamID: &teamID, TeamName: "Team A"},
		{CourseParticipationID: &participationID},
	})

	if target, ok := index.find(instanceForTeam(teamID)); !ok || target.TeamName != "Team A" {
		t.Fatalf("team lookup = %+v (found: %v), want Team A", target, ok)
	}
	if target, ok := index.find(instanceForStudent(participationID)); !ok || target.CourseParticipationID == nil {
		t.Fatalf("student lookup = %+v (found: %v), want the student target", target, ok)
	}
	if _, ok := index.find(instanceForTeam(uuid.New())); ok {
		t.Fatal("lookup of an unknown team reported a match")
	}
}

func instanceForTeam(teamID uuid.UUID) db.ResourceInstance {
	return db.ResourceInstance{TeamID: &teamID}
}

func instanceForStudent(participationID uuid.UUID) db.ResourceInstance {
	return db.ResourceInstance{CourseParticipationID: &participationID}
}

// A phase whose setup page was never saved has no course_phase_config row. The only
// thing that row carries is the optional semester tag, so resolution must go ahead
// with an empty one instead of refusing to provision anything.
func TestResolveTargetsWithoutAConfigRow(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/participations") {
			t.Errorf("unexpected core request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"participations":[],"resolutions":[]}`))
	}))
	defer core.Close()

	resolver := &CoreTargetResolver{queries: testDB.Queries, coreURL: core.URL}
	targets, err := resolver.ResolveTargets(context.Background(), "Bearer test", coursePhaseID, db.ResourceScopePerStudent)
	if err != nil {
		t.Fatalf("ResolveTargets returned error: %v", err)
	}
	if len(targets) != 0 {
		t.Fatalf("targets = %d, want 0 for a phase without participants", len(targets))
	}
}
