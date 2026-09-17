package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type StaffRepository struct{ db *pgxpool.Pool }

func NewStaffRepository(db *pgxpool.Pool) *StaffRepository {
	return &StaffRepository{db: db}
}

func (r *StaffRepository) Create(ctx context.Context, staff domain.Staff) (domain.Staff, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO staff (username, password_hash, hospital_id)
		 VALUES ($1, $2, $3) RETURNING id`,
		staff.Username, staff.PasswordHash, staff.HospitalID,
	).Scan(&staff.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Staff{}, domain.ErrConflict
		}
		return domain.Staff{}, err
	}
	return staff, nil
}

func (r *StaffRepository) FindByUsernameAndHospital(ctx context.Context, username string, hospitalID int64) (domain.Staff, error) {
	var staff domain.Staff
	err := r.db.QueryRow(ctx,
		`SELECT id, username, password_hash, hospital_id
		 FROM staff WHERE username = $1 AND hospital_id = $2`, username, hospitalID,
	).Scan(&staff.ID, &staff.Username, &staff.PasswordHash, &staff.HospitalID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Staff{}, domain.ErrNotFound
	}
	return staff, err
}
