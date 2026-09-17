package httpapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AofArthit21/Hospital-information-system/internal/auth"
	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type fakeStaffService struct {
	created domain.Staff
	result  domain.AuthResult
	create  error
	login   error
}

func (f *fakeStaffService) CreateStaff(context.Context, domain.StaffCreateInput) (domain.Staff, error) {
	return f.created, f.create
}

func (f *fakeStaffService) Login(context.Context, domain.LoginInput) (domain.AuthResult, error) {
	return f.result, f.login
}

type fakePatientService struct {
	patients   []domain.Patient
	err        error
	hospitalID int64
	filter     domain.PatientFilter
}

func (f *fakePatientService) Search(_ context.Context, hospitalID int64, filter domain.PatientFilter) ([]domain.Patient, error) {
	f.hospitalID = hospitalID
	f.filter = filter
	return f.patients, f.err
}

func testRouter(staff *fakeStaffService, patients *fakePatientService) (*gin.Engine, *auth.Manager) {
	gin.SetMode(gin.TestMode)
	tokens := auth.NewManager(strings.Repeat("s", 32), time.Hour)
	return NewRouter(NewHandler(staff, patients), tokens), tokens
}

func TestCreateStaff(t *testing.T) {
	t.Run("created", func(t *testing.T) {
		router, _ := testRouter(&fakeStaffService{created: domain.Staff{ID: 7, Username: "alice"}}, &fakePatientService{})
		response := performRequest(router, http.MethodPost, "/staff/create", `{"username":"alice","password":"password123","hospital":"hospital-a"}`, "")
		if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"id":7`) {
			t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		router, _ := testRouter(&fakeStaffService{create: domain.ErrConflict}, &fakePatientService{})
		response := performRequest(router, http.MethodPost, "/staff/create", `{"username":"alice","password":"password123","hospital":"hospital-a"}`, "")
		if response.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", response.Code)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		router, _ := testRouter(&fakeStaffService{}, &fakePatientService{})
		response := performRequest(router, http.MethodPost, "/staff/create", `{"username":"a"}`, "")
		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", response.Code)
		}
	})
}

func TestLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		staff := &fakeStaffService{result: domain.AuthResult{AccessToken: "jwt", TokenType: "Bearer", ExpiresIn: 3600}}
		router, _ := testRouter(staff, &fakePatientService{})
		response := performRequest(router, http.MethodPost, "/staff/login", `{"username":"alice","password":"password123","hospital":"hospital-a"}`, "")
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"access_token":"jwt"`) {
			t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
		}
	})

	t.Run("wrong credentials", func(t *testing.T) {
		router, _ := testRouter(&fakeStaffService{login: domain.ErrInvalidCredential}, &fakePatientService{})
		response := performRequest(router, http.MethodPost, "/staff/login", `{"username":"alice","password":"incorrect","hospital":"hospital-a"}`, "")
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", response.Code)
		}
	})
}

func TestSearchPatients(t *testing.T) {
	t.Run("authenticated and scoped by token", func(t *testing.T) {
		date := time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)
		patients := &fakePatientService{patients: []domain.Patient{{PatientHN: "HN001", FirstNameEN: "Alice", DateOfBirth: &date}}}
		router, tokens := testRouter(&fakeStaffService{}, patients)
		token, _, err := tokens.Generate(4, 23)
		if err != nil {
			t.Fatal(err)
		}
		response := performRequest(router, http.MethodGet, "/patient/search?first_name=Ali&date_of_birth=1990-01-02", "", token)
		if response.Code != http.StatusOK || patients.hospitalID != 23 || patients.filter.FirstName != "Ali" {
			t.Fatalf("unexpected response/scope: %d hospital=%d body=%s", response.Code, patients.hospitalID, response.Body.String())
		}
		if !strings.Contains(response.Body.String(), `"date_of_birth":"1990-01-02"`) {
			t.Fatalf("date response is not API format: %s", response.Body.String())
		}
	})

	t.Run("missing token", func(t *testing.T) {
		router, _ := testRouter(&fakeStaffService{}, &fakePatientService{})
		response := performRequest(router, http.MethodGet, "/patient/search", "", "")
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", response.Code)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		router, tokens := testRouter(&fakeStaffService{}, &fakePatientService{})
		token, _, _ := tokens.Generate(4, 23)
		response := performRequest(router, http.MethodGet, "/patient/search?date_of_birth=02-01-1990", "", token)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", response.Code)
		}
	})

	t.Run("HIS unavailable", func(t *testing.T) {
		router, tokens := testRouter(&fakeStaffService{}, &fakePatientService{err: errors.Join(domain.ErrUpstream, errors.New("timeout"))})
		token, _, _ := tokens.Generate(4, 23)
		response := performRequest(router, http.MethodGet, "/patient/search?national_id=123", "", token)
		if response.Code != http.StatusBadGateway {
			t.Fatalf("expected 502, got %d", response.Code)
		}
	})
}

func TestAccessLogOmitsQueryString(t *testing.T) {
	var logs bytes.Buffer
	previousWriter := gin.DefaultWriter
	gin.DefaultWriter = &logs
	defer func() { gin.DefaultWriter = previousWriter }()

	router, _ := testRouter(&fakeStaffService{}, &fakePatientService{})
	response := performRequest(router, http.MethodGet, "/healthz?national_id=secret-patient-id", "", "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if strings.Contains(logs.String(), "secret-patient-id") || strings.Contains(logs.String(), "national_id") {
		t.Fatalf("access log leaked query string: %s", logs.String())
	}
}

func performRequest(router http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
