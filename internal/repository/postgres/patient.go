package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type PatientRepository struct{ db *pgxpool.Pool }

func NewPatientRepository(db *pgxpool.Pool) *PatientRepository {
	return &PatientRepository{db: db}
}

func (r *PatientRepository) Upsert(ctx context.Context, p domain.Patient) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO patients (
			hospital_id, first_name_th, middle_name_th, last_name_th,
			first_name_en, middle_name_en, last_name_en, date_of_birth,
			patient_hn, national_id, passport_id, phone_number, email, gender
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (hospital_id, patient_hn) DO UPDATE SET
			first_name_th=EXCLUDED.first_name_th, middle_name_th=EXCLUDED.middle_name_th,
			last_name_th=EXCLUDED.last_name_th, first_name_en=EXCLUDED.first_name_en,
			middle_name_en=EXCLUDED.middle_name_en, last_name_en=EXCLUDED.last_name_en,
			date_of_birth=EXCLUDED.date_of_birth, national_id=EXCLUDED.national_id,
			passport_id=EXCLUDED.passport_id, phone_number=EXCLUDED.phone_number,
			email=EXCLUDED.email, gender=EXCLUDED.gender, updated_at=NOW()`,
		p.HospitalID, p.FirstNameTH, p.MiddleNameTH, p.LastNameTH,
		p.FirstNameEN, p.MiddleNameEN, p.LastNameEN, p.DateOfBirth,
		p.PatientHN, nullIfEmpty(p.NationalID), nullIfEmpty(p.PassportID),
		nullIfEmpty(p.PhoneNumber), nullIfEmpty(p.Email), nullIfEmpty(p.Gender),
	)
	return err
}

func (r *PatientRepository) Search(ctx context.Context, hospitalID int64, filter domain.PatientFilter) ([]domain.Patient, error) {
	query, args := buildPatientSearchQuery(hospitalID, filter)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	patients := make([]domain.Patient, 0)
	for rows.Next() {
		var p domain.Patient
		if err := rows.Scan(
			&p.ID, &p.HospitalID, &p.FirstNameTH, &p.MiddleNameTH, &p.LastNameTH,
			&p.FirstNameEN, &p.MiddleNameEN, &p.LastNameEN, &p.DateOfBirth,
			&p.PatientHN, &p.NationalID, &p.PassportID, &p.PhoneNumber, &p.Email, &p.Gender,
		); err != nil {
			return nil, err
		}
		patients = append(patients, p)
	}
	return patients, rows.Err()
}

func buildPatientSearchQuery(hospitalID int64, filter domain.PatientFilter) (string, []any) {
	query := `SELECT id, hospital_id, first_name_th, middle_name_th, last_name_th,
		first_name_en, middle_name_en, last_name_en, date_of_birth,
		patient_hn, COALESCE(national_id,''), COALESCE(passport_id,''),
		COALESCE(phone_number,''), COALESCE(email,''), COALESCE(gender,'')
		FROM patients WHERE hospital_id = $1`
	args := []any{hospitalID}
	add := func(condition string, value any) {
		args = append(args, value)
		query += fmt.Sprintf(" AND "+condition, len(args))
	}
	if filter.NationalID != "" {
		add("national_id = $%d", filter.NationalID)
	}
	if filter.PassportID != "" {
		add("passport_id = $%d", filter.PassportID)
	}
	if filter.FirstName != "" {
		add("(first_name_th ILIKE $%[1]d OR first_name_en ILIKE $%[1]d)", "%"+escapeLike(filter.FirstName)+"%")
	}
	if filter.MiddleName != "" {
		add("(middle_name_th ILIKE $%[1]d OR middle_name_en ILIKE $%[1]d)", "%"+escapeLike(filter.MiddleName)+"%")
	}
	if filter.LastName != "" {
		add("(last_name_th ILIKE $%[1]d OR last_name_en ILIKE $%[1]d)", "%"+escapeLike(filter.LastName)+"%")
	}
	if filter.DateOfBirth != nil {
		add("date_of_birth = $%d", *filter.DateOfBirth)
	}
	if filter.PhoneNumber != "" {
		add("phone_number = $%d", filter.PhoneNumber)
	}
	if filter.Email != "" {
		add("LOWER(email) = LOWER($%d)", filter.Email)
	}
	query += " ORDER BY updated_at DESC LIMIT 100"
	return query, args
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}
