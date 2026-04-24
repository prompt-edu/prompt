package teams

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/team/teamDTO"
	log "github.com/sirupsen/logrus"
)

type TeamsService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var TeamsServiceSingleton *TeamsService

const (
	allocationProfileStandard        = "standard"
	allocationProfileProjectWeek1000 = "project_week_1000_plus"
)

func getTeamsWithMembersByType(ctx context.Context, coursePhaseID uuid.UUID, teamType string) ([]db.GetTeamsWithMembersRow, error) {
	rows, err := TeamsServiceSingleton.conn.Query(ctx, `
		SELECT t.id,
		       t.name,
		       COALESCE(members.team_members, '[]'::jsonb) AS team_members,
		       COALESCE(tutors.team_tutors, '[]'::jsonb)   AS team_tutors
		FROM team t
		         LEFT JOIN LATERAL (
		    SELECT jsonb_agg(
		                   jsonb_build_object(
		                           'id', a.course_participation_id,
		                           'firstName', a.student_first_name,
		                           'lastName', a.student_last_name
		                   )
		                   ORDER BY a.student_first_name
		           ) AS team_members
		    FROM allocations a
		    WHERE a.team_id = t.id
		    ) members ON TRUE
		         LEFT JOIN LATERAL (
		    SELECT jsonb_agg(
		                   jsonb_build_object(
		                           'id', tu.course_participation_id,
		                           'firstName', tu.first_name,
		                           'lastName', tu.last_name
		                   )
		                   ORDER BY tu.first_name
		           ) AS team_tutors
		    FROM tutor tu
		    WHERE tu.team_id = t.id
		    ) tutors ON TRUE
		WHERE t.course_phase_id = $1
		  AND t.team_type = $2
		ORDER BY t.name
	`, coursePhaseID, teamType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbTeams := []db.GetTeamsWithMembersRow{}
	for rows.Next() {
		var team db.GetTeamsWithMembersRow
		if scanErr := rows.Scan(&team.ID, &team.Name, &team.TeamMembers, &team.TeamTutors); scanErr != nil {
			return nil, scanErr
		}
		dbTeams = append(dbTeams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dbTeams, nil
}

func GetAllTeams(ctx context.Context, coursePhaseID uuid.UUID) ([]promptTypes.Team, error) {
	profile, err := TeamsServiceSingleton.queries.GetAllocationProfile(ctx, coursePhaseID)
	if err != nil {
		profile = allocationProfileStandard
	}

	var dbTeams []db.GetTeamsWithMembersRow
	switch profile {
	case allocationProfileProjectWeek1000:
		rows, queryErr := TeamsServiceSingleton.queries.GetFieldBucketTeamsByCoursePhase(ctx, coursePhaseID)
		if queryErr != nil {
			log.Error("could not get the teams from the database: ", queryErr)
			return nil, errors.New("could not get the teams from the database")
		}
		dbTeams = make([]db.GetTeamsWithMembersRow, 0, len(rows))
		for _, row := range rows {
			dbTeams = append(dbTeams, db.GetTeamsWithMembersRow{
				ID:          row.ID,
				Name:        row.Name,
				TeamMembers: []byte("[]"),
				TeamTutors:  []byte("[]"),
			})
		}
	default:
		dbTeams, err = getTeamsWithMembersByType(ctx, coursePhaseID, "standard")
		if err != nil {
			log.Error("could not get the teams from the database: ", err)
			return nil, errors.New("could not get the teams from the database")
		}
	}
	dtos, err := teamDTO.GetTeamsWithMembersDTOFromDBModel(dbTeams)
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return nil, errors.New("could not get the teams from the database")
	}
	return dtos, nil
}

func GetTeamByID(ctx context.Context, coursePhaseID uuid.UUID, teamID uuid.UUID) (promptTypes.Team, error) {
	dbTeam, err := TeamsServiceSingleton.queries.GetTeamByCoursePhaseAndTeamID(ctx, db.GetTeamByCoursePhaseAndTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return promptTypes.Team{}, errors.New("could not get the teams from the database")
	}
	return teamDTO.GetTeamDTOFromDBModel(dbTeam), nil
}

func CreateNewTeams(
	ctx context.Context,
	teamNames []string,
	coursePhaseID uuid.UUID,
	teamType string,
	replaceExisting bool,
	teamSizeConstraints map[string]teamDTO.TeamSizeConstraint,
) error {
	if teamType == "" {
		teamType = "standard"
	}
	tx, err := TeamsServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := TeamsServiceSingleton.queries.WithTx(tx)

	if replaceExisting {
		if teamType == "standard" {
			return errors.New("replacing standard teams is not supported")
		}
		if _, err := tx.Exec(ctx, "DELETE FROM team WHERE course_phase_id = $1 AND team_type = $2", coursePhaseID, teamType); err != nil {
			log.Error("error deleting existing teams by type ", err)
			return errors.New("error replacing existing teams")
		}
	}

	for _, teamName := range teamNames {
		teamSizeMin := pgtype.Int4{Valid: false}
		teamSizeMax := pgtype.Int4{Valid: false}
		if constraint, ok := teamSizeConstraints[teamName]; ok {
			teamSizeMin = pgtype.Int4{Int32: constraint.LowerBound, Valid: true}
			teamSizeMax = pgtype.Int4{Int32: constraint.UpperBound, Valid: true}
		}

		err := qtx.CreateTeam(ctx, db.CreateTeamParams{
			ID:            uuid.New(),
			Name:          teamName,
			CoursePhaseID: coursePhaseID,
			TeamType:      teamType,
			TeamSizeMin:   teamSizeMin,
			TeamSizeMax:   teamSizeMax,
		})
		if err != nil {
			log.Error("error creating the teams ", err)
			return errors.New("error creating the teams")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func UpdateTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID, newTeamName string) error {
	err := TeamsServiceSingleton.queries.UpdateTeam(ctx, db.UpdateTeamParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
		Name:          newTeamName,
	})
	if err != nil {
		log.Error("could not update the team: ", err)
		return errors.New("could not update the team")
	}
	return nil
}

func DeleteTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID) error {
	err := TeamsServiceSingleton.queries.DeleteTeam(ctx, db.DeleteTeamParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		log.Error("could not delete the team: ", err)
		return errors.New("could not delete the team")
	}
	return nil
}

func AddStudentNamesToAllocations(ctx context.Context, req teamDTO.StudentNameUpdateRequest) error {
	tx, err := TeamsServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := TeamsServiceSingleton.queries.WithTx(tx)

	for participationID, name := range req.StudentNamesPerID {
		err := qtx.UpdateStudentNameForAllocation(ctx, db.UpdateStudentNameForAllocationParams{
			StudentFirstName:      name.FirstName,
			StudentLastName:       name.LastName,
			CourseParticipationID: participationID,
			CoursePhaseID:         req.CoursePhaseID,
		})
		if err != nil {
			return fmt.Errorf("failed to update name for participation ID %s: %w", participationID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func ImportTutors(ctx context.Context, coursePhaseID uuid.UUID, tutors []teamDTO.Tutor) error {
	// add students to the keycloak group
	tx, err := TeamsServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := TeamsServiceSingleton.queries.WithTx(tx)

	for _, tutor := range tutors {
		// store tutor in database
		err := qtx.CreateTutor(ctx, db.CreateTutorParams{
			CoursePhaseID:         coursePhaseID,
			CourseParticipationID: tutor.CourseParticipationID,
			FirstName:             tutor.FirstName,
			LastName:              tutor.LastName,
			TeamID:                tutor.TeamID,
		})
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
