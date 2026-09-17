package his

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{httpClient: &http.Client{Timeout: timeout}}
}

func (c *Client) SearchByID(ctx context.Context, hospital domain.Hospital, patientID string) (domain.Patient, error) {
	if hospital.HISBaseURL == "" {
		return domain.Patient{}, fmt.Errorf("%w: hospital HIS URL is not configured", domain.ErrUpstream)
	}
	endpoint := strings.TrimRight(hospital.HISBaseURL, "/") + "/patient/search/" + url.PathEscape(patientID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("build HIS request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("%w: %v", domain.ErrUpstream, err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return domain.Patient{}, domain.ErrNotFound
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<20))
		return domain.Patient{}, fmt.Errorf("%w: HIS returned status %d", domain.ErrUpstream, res.StatusCode)
	}

	var payload struct {
		FirstNameTH  string `json:"first_name_th"`
		MiddleNameTH string `json:"middle_name_th"`
		LastNameTH   string `json:"last_name_th"`
		FirstNameEN  string `json:"first_name_en"`
		MiddleNameEN string `json:"middle_name_en"`
		LastNameEN   string `json:"last_name_en"`
		DateOfBirth  string `json:"date_of_birth"`
		PatientHN    string `json:"patient_hn"`
		NationalID   string `json:"national_id"`
		PassportID   string `json:"passport_id"`
		PhoneNumber  string `json:"phone_number"`
		Email        string `json:"email"`
		Gender       string `json:"gender"`
	}
	decoder := json.NewDecoder(io.LimitReader(res.Body, 1<<20))
	if err := decoder.Decode(&payload); err != nil {
		return domain.Patient{}, fmt.Errorf("%w: invalid JSON: %v", domain.ErrUpstream, err)
	}
	if payload.PatientHN == "" {
		return domain.Patient{}, fmt.Errorf("%w: patient_hn is missing", domain.ErrUpstream)
	}
	if payload.NationalID == "" && payload.PassportID == "" {
		return domain.Patient{}, fmt.Errorf("%w: a patient identifier is missing", domain.ErrUpstream)
	}
	var dob *time.Time
	if payload.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", payload.DateOfBirth)
		if err != nil {
			return domain.Patient{}, fmt.Errorf("%w: invalid date_of_birth", domain.ErrUpstream)
		}
		dob = &parsed
	}
	gender := strings.ToUpper(payload.Gender)
	if gender != "" && gender != "M" && gender != "F" {
		return domain.Patient{}, fmt.Errorf("%w: invalid gender", domain.ErrUpstream)
	}
	return domain.Patient{
		FirstNameTH: payload.FirstNameTH, MiddleNameTH: payload.MiddleNameTH, LastNameTH: payload.LastNameTH,
		FirstNameEN: payload.FirstNameEN, MiddleNameEN: payload.MiddleNameEN, LastNameEN: payload.LastNameEN,
		DateOfBirth: dob, PatientHN: payload.PatientHN, NationalID: payload.NationalID,
		PassportID: payload.PassportID, PhoneNumber: payload.PhoneNumber, Email: payload.Email, Gender: gender,
	}, nil
}
