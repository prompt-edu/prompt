package skills

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/skills/skillDTO"
	log "github.com/sirupsen/logrus"
)

type SkillsService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewSkillsService(queries db.Queries, conn *pgxpool.Pool) *SkillsService {
	return &SkillsService{
		queries: queries,
		conn:    conn,
	}
}

func (s *SkillsService) GetAllSkills(ctx context.Context, coursePhaseID uuid.UUID) ([]skillDTO.Skill, error) {
	dbSkills, err := s.queries.GetSkillsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get the skills from the database: ", err)
		return nil, errors.New("could not get the skills from the database")
	}
	return skillDTO.GetSkillDTOsFromDBModels(dbSkills), nil
}

func (s *SkillsService) CreateNewSkills(ctx context.Context, skillNames []string, coursePhaseID uuid.UUID) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	for _, skillName := range skillNames {
		err := qtx.CreateSkill(ctx, db.CreateSkillParams{
			ID:            uuid.New(),
			Name:          skillName,
			CoursePhaseID: coursePhaseID,
		})
		if err != nil {
			log.Error("error creating the skills: ", err)
			return errors.New("error creating the skills")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *SkillsService) UpdateSkill(ctx context.Context, coursePhaseID, skillID uuid.UUID, newSkillName string) error {
	err := s.queries.UpdateSkill(ctx, db.UpdateSkillParams{
		ID:            skillID,
		CoursePhaseID: coursePhaseID,
		Name:          newSkillName,
	})
	if err != nil {
		log.Error("could not update the skill: ", err)
		return errors.New("could not update the skill")
	}
	return nil
}

func (s *SkillsService) DeleteSkill(ctx context.Context, coursePhaseID, skillID uuid.UUID) error {
	err := s.queries.DeleteSkill(ctx, db.DeleteSkillParams{
		ID:            skillID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		log.Error("could not delete the skill: ", err)
		return errors.New("could not delete the skill")
	}
	return nil
}
