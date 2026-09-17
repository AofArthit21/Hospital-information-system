package service

import (
	"context"
	"testing"

	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type patientHospitalRepo struct{ hospital domain.Hospital }

func (r patientHospitalRepo) FindByCode(context.Context, string) (domain.Hospital, error) {
	return r.hospital, nil
}
func (r patientHospitalRepo) FindByID(_ context.Context, id int64) (domain.Hospital, error) {
	if id != r.hospital.ID {
		return domain.Hospital{}, domain.ErrNotFound
	}
	return r.hospital, nil
}

type patientRepo struct {
	upserted domain.Patient
	searched int64
}

func (r *patientRepo) Upsert(_ context.Context, patient domain.Patient) error {
	r.upserted = patient
	return nil
}
func (r *patientRepo) Search(_ context.Context, hospitalID int64, _ domain.PatientFilter) ([]domain.Patient, error) {
	r.searched = hospitalID
	if r.upserted.PatientHN == "" {
		return []domain.Patient{}, nil
	}
	return []domain.Patient{r.upserted}, nil
}

type patientHIS struct{ requestedID string }

func (h *patientHIS) SearchByID(_ context.Context, _ domain.Hospital, patientID string) (domain.Patient, error) {
	h.requestedID = patientID
	return domain.Patient{PatientHN: "HN001", NationalID: patientID}, nil
}

func TestPatientServiceScopesAndCachesHISResult(t *testing.T) {
	repository := &patientRepo{}
	his := &patientHIS{}
	service := NewPatientService(
		patientHospitalRepo{hospital: domain.Hospital{ID: 17, Code: "hospital-a"}}, repository, his,
	)
	patients, err := service.Search(context.Background(), 17, domain.PatientFilter{NationalID: "123"})
	if err != nil {
		t.Fatal(err)
	}
	if his.requestedID != "123" || repository.upserted.HospitalID != 17 || repository.searched != 17 {
		t.Fatalf("hospital scope was lost: requested=%s upsert=%d search=%d", his.requestedID, repository.upserted.HospitalID, repository.searched)
	}
	if len(patients) != 1 || patients[0].PatientHN != "HN001" {
		t.Fatalf("unexpected patients: %#v", patients)
	}
}
