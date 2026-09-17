package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/AofArthit21/Hospital-information-system/internal/auth"
	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type authHospitalRepo struct{ hospital domain.Hospital }

func (r authHospitalRepo) FindByCode(context.Context, string) (domain.Hospital, error) {
	return r.hospital, nil
}
func (r authHospitalRepo) FindByID(context.Context, int64) (domain.Hospital, error) {
	return r.hospital, nil
}

type authStaffRepo struct{ staff domain.Staff }

func (r *authStaffRepo) Create(_ context.Context, staff domain.Staff) (domain.Staff, error) {
	staff.ID = 9
	r.staff = staff
	return staff, nil
}
func (r *authStaffRepo) FindByUsernameAndHospital(context.Context, string, int64) (domain.Staff, error) {
	if r.staff.ID == 0 {
		return domain.Staff{}, domain.ErrNotFound
	}
	return r.staff, nil
}

func TestAuthServiceCreateAndLogin(t *testing.T) {
	repository := &authStaffRepo{}
	tokens := auth.NewManager(strings.Repeat("x", 32), time.Hour)
	service := NewAuthService(authHospitalRepo{hospital: domain.Hospital{ID: 3, Code: "hospital-a"}}, repository, tokens)

	created, err := service.CreateStaff(context.Background(), domain.StaffCreateInput{
		Username: "alice", Password: "password123", HospitalCode: "HOSPITAL-A",
	})
	if err != nil || created.ID != 9 || created.HospitalID != 3 {
		t.Fatalf("unexpected create result: %#v %v", created, err)
	}
	if repository.staff.PasswordHash == "password123" || bcrypt.CompareHashAndPassword([]byte(repository.staff.PasswordHash), []byte("password123")) != nil {
		t.Fatal("password was not bcrypt-hashed")
	}

	result, err := service.Login(context.Background(), domain.LoginInput{
		Username: "alice", Password: "password123", HospitalCode: "hospital-a",
	})
	if err != nil || result.AccessToken == "" {
		t.Fatalf("unexpected login result: %#v %v", result, err)
	}
	staffID, hospitalID, err := tokens.Parse(result.AccessToken)
	if err != nil || staffID != 9 || hospitalID != 3 {
		t.Fatalf("wrong token claims: staff=%d hospital=%d err=%v", staffID, hospitalID, err)
	}

	_, err = service.Login(context.Background(), domain.LoginInput{
		Username: "alice", Password: "wrong-password", HospitalCode: "hospital-a",
	})
	if !errors.Is(err, domain.ErrInvalidCredential) {
		t.Fatalf("expected invalid credential, got %v", err)
	}
}
