package studyProgram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/studyProgram/studyProgramDTO"
)

var (
	ErrDuplicateStudyProgram = errors.New("a study program with this name already exists")
	ErrStudyProgramNotFound  = errors.New("study program not found")
)

type StudyProgramService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewStudyProgramService(queries db.Queries, conn *pgxpool.Pool) *StudyProgramService {
	return &StudyProgramService{queries: queries, conn: conn}
}

func (s *StudyProgramService) ListStudyPrograms(ctx context.Context) ([]studyProgramDTO.StudyProgram, error) {
	studyPrograms, err := s.queries.ListStudyPrograms(ctx)
	if err != nil {
		return nil, err
	}
	return studyProgramDTO.StudyProgramsFromDBModels(studyPrograms), nil
}

func (s *StudyProgramService) GetStudentCounts(ctx context.Context) ([]studyProgramDTO.StudyProgramStudentCount, error) {
	rows, err := s.queries.CountStudentsPerStudyProgram(ctx)
	if err != nil {
		return nil, err
	}
	return studyProgramDTO.StudentCountsFromDBModels(rows), nil
}

func (s *StudyProgramService) CreateStudyProgram(ctx context.Context, input studyProgramDTO.CreateStudyProgram) (studyProgramDTO.StudyProgram, error) {
	created, err := s.queries.CreateStudyProgram(ctx, db.CreateStudyProgramParams{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(input.Name),
		ShortName: optionalText(input.ShortName),
	})
	if err != nil {
		return studyProgramDTO.StudyProgram{}, mapUniqueViolation(err)
	}
	return studyProgramDTO.StudyProgramFromDBModel(created), nil
}

func (s *StudyProgramService) UpdateStudyProgram(ctx context.Context, id uuid.UUID, input studyProgramDTO.UpdateStudyProgram) (studyProgramDTO.StudyProgram, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return studyProgramDTO.StudyProgram{}, err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	previous, err := qtx.GetStudyProgramByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return studyProgramDTO.StudyProgram{}, ErrStudyProgramNotFound
	}
	if err != nil {
		return studyProgramDTO.StudyProgram{}, err
	}

	updated, err := qtx.UpdateStudyProgram(ctx, db.UpdateStudyProgramParams{
		ID:        id,
		Name:      strings.TrimSpace(input.Name),
		ShortName: optionalText(input.ShortName),
	})
	if err != nil {
		return studyProgramDTO.StudyProgram{}, mapUniqueViolation(err)
	}

	if updated.Name != previous.Name {
		_, err := qtx.RenameStudentStudyProgram(ctx, db.RenameStudentStudyProgramParams{
			NewName: updated.Name,
			OldName: previous.Name,
		})
		if err != nil {
			return studyProgramDTO.StudyProgram{}, fmt.Errorf("failed to rename the study program of students: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return studyProgramDTO.StudyProgram{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return studyProgramDTO.StudyProgramFromDBModel(updated), nil
}

func (s *StudyProgramService) DeleteStudyProgram(ctx context.Context, id uuid.UUID) error {
	deleted, err := s.queries.DeleteStudyProgram(ctx, id)
	if err != nil {
		return err
	}
	if deleted == 0 {
		return ErrStudyProgramNotFound
	}
	return nil
}

func optionalText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	return pgtype.Text{String: trimmed, Valid: trimmed != ""}
}

func mapUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicateStudyProgram
	}
	return err
}
