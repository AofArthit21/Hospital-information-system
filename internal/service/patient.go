package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type PatientRepository interface {
	Search(ctx context.Context, hospitalID int64, filter domain.PatientFilter) ([]domain.Patient, error)
	Upsert(ctx context.Context, patient domain.Patient) error
}

type HISClient interface {
	SearchByID(ctx context.Context, hospital domain.Hospital, patientID string) (domain.Patient, error)
}

type PatientService struct {
	hospitals HospitalRepository
	patients  PatientRepository
	his       HISClient
}

func NewPatientService(hospitals HospitalRepository, patients PatientRepository, his HISClient) *PatientService {
	return &PatientService{hospitals: hospitals, patients: patients, his: his}
}

func (s *PatientService) Search(ctx context.Context, hospitalID int64, filter domain.PatientFilter) ([]domain.Patient, error) {
	patientID := filter.NationalID
	if patientID == "" {
		patientID = filter.PassportID
	}
	if patientID != "" {
		hospital, err := s.hospitals.FindByID(ctx, hospitalID)
		if err != nil {
			return nil, err
		}
		patient, err := s.his.SearchByID(ctx, hospital, patientID)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		if err == nil {
			patient.HospitalID = hospitalID
			if err := s.patients.Upsert(ctx, patient); err != nil {
				return nil, fmt.Errorf("cache patient: %w", err)
			}
		}
	}
	return s.patients.Search(ctx, hospitalID, filter)
}
