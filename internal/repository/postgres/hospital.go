package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type HospitalRepository struct{ db *pgxpool.Pool }

func NewHospitalRepository(db *pgxpool.Pool) *HospitalRepository {
	return &HospitalRepository{db: db}
}

func (r *HospitalRepository) FindByCode(ctx context.Context, code string) (domain.Hospital, error) {
	return scanHospital(r.db.QueryRow(ctx,
		`SELECT id, code, name, his_base_url FROM hospitals WHERE code = $1`, code))
}

func (r *HospitalRepository) FindByID(ctx context.Context, id int64) (domain.Hospital, error) {
	return scanHospital(r.db.QueryRow(ctx,
		`SELECT id, code, name, his_base_url FROM hospitals WHERE id = $1`, id))
}

type rowScanner interface{ Scan(dest ...any) error }

func scanHospital(row rowScanner) (domain.Hospital, error) {
	var hospital domain.Hospital
	if err := row.Scan(&hospital.ID, &hospital.Code, &hospital.Name, &hospital.HISBaseURL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Hospital{}, domain.ErrNotFound
		}
		return domain.Hospital{}, err
	}
	return hospital, nil
}
