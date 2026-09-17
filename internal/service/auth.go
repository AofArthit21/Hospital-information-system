package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/AofArthit21/Hospital-information-system/internal/auth"
	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type HospitalRepository interface {
	FindByCode(ctx context.Context, code string) (domain.Hospital, error)
	FindByID(ctx context.Context, id int64) (domain.Hospital, error)
}

type StaffRepository interface {
	Create(ctx context.Context, staff domain.Staff) (domain.Staff, error)
	FindByUsernameAndHospital(ctx context.Context, username string, hospitalID int64) (domain.Staff, error)
}

type AuthService struct {
	hospitals HospitalRepository
	staff     StaffRepository
	tokens    *auth.Manager
}

func NewAuthService(hospitals HospitalRepository, staff StaffRepository, tokens *auth.Manager) *AuthService {
	return &AuthService{hospitals: hospitals, staff: staff, tokens: tokens}
}

func (s *AuthService) CreateStaff(ctx context.Context, input domain.StaffCreateInput) (domain.Staff, error) {
	hospital, err := s.hospitals.FindByCode(ctx, normalizeCode(input.HospitalCode))
	if err != nil {
		return domain.Staff{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.Staff{}, fmt.Errorf("hash password: %w", err)
	}
	created, err := s.staff.Create(ctx, domain.Staff{
		Username:     strings.TrimSpace(input.Username),
		PasswordHash: string(hash),
		HospitalID:   hospital.ID,
	})
	if err != nil {
		return domain.Staff{}, err
	}
	return created, nil
}

func (s *AuthService) Login(ctx context.Context, input domain.LoginInput) (domain.AuthResult, error) {
	hospital, err := s.hospitals.FindByCode(ctx, normalizeCode(input.HospitalCode))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.AuthResult{}, domain.ErrInvalidCredential
		}
		return domain.AuthResult{}, err
	}
	staff, err := s.staff.FindByUsernameAndHospital(ctx, strings.TrimSpace(input.Username), hospital.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.AuthResult{}, domain.ErrInvalidCredential
		}
		return domain.AuthResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(input.Password)) != nil {
		return domain.AuthResult{}, domain.ErrInvalidCredential
	}
	token, expiresIn, err := s.tokens.Generate(staff.ID, staff.HospitalID)
	if err != nil {
		return domain.AuthResult{}, fmt.Errorf("generate token: %w", err)
	}
	return domain.AuthResult{AccessToken: token, TokenType: "Bearer", ExpiresIn: expiresIn}, nil
}

func normalizeCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}
