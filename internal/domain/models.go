package domain

import "time"

type Hospital struct {
	ID         int64  `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	HISBaseURL string `json:"-"`
}

type Staff struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	HospitalID   int64  `json:"-"`
}

type Patient struct {
	ID           int64      `json:"-"`
	HospitalID   int64      `json:"-"`
	FirstNameTH  string     `json:"first_name_th"`
	MiddleNameTH string     `json:"middle_name_th"`
	LastNameTH   string     `json:"last_name_th"`
	FirstNameEN  string     `json:"first_name_en"`
	MiddleNameEN string     `json:"middle_name_en"`
	LastNameEN   string     `json:"last_name_en"`
	DateOfBirth  *time.Time `json:"date_of_birth,omitempty"`
	PatientHN    string     `json:"patient_hn"`
	NationalID   string     `json:"national_id,omitempty"`
	PassportID   string     `json:"passport_id,omitempty"`
	PhoneNumber  string     `json:"phone_number,omitempty"`
	Email        string     `json:"email,omitempty"`
	Gender       string     `json:"gender,omitempty"`
}

type PatientFilter struct {
	NationalID  string
	PassportID  string
	FirstName   string
	MiddleName  string
	LastName    string
	DateOfBirth *time.Time
	PhoneNumber string
	Email       string
}

type StaffCreateInput struct {
	Username     string
	Password     string
	HospitalCode string
}

type LoginInput = StaffCreateInput

type AuthResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}
