package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

type StaffService interface {
	CreateStaff(context.Context, domain.StaffCreateInput) (domain.Staff, error)
	Login(context.Context, domain.LoginInput) (domain.AuthResult, error)
}

type PatientService interface {
	Search(context.Context, int64, domain.PatientFilter) ([]domain.Patient, error)
}

type Handler struct {
	staff    StaffService
	patients PatientService
}

func NewHandler(staff StaffService, patients PatientService) *Handler {
	return &Handler{staff: staff, patients: patients}
}

type credentialsRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Hospital string `json:"hospital" binding:"required,max=50"`
}

func (h *Handler) createStaff(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "username, password and hospital are required; password must be 8-72 characters")
		return
	}
	input := domain.StaffCreateInput{Username: req.Username, Password: req.Password, HospitalCode: req.Hospital}
	staff, err := h.staff.CreateStaff(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConflict):
			writeError(c, http.StatusConflict, "STAFF_EXISTS", "the username already exists in this hospital")
		case errors.Is(err, domain.ErrNotFound):
			writeError(c, http.StatusBadRequest, "UNKNOWN_HOSPITAL", "the hospital does not exist")
		default:
			writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "unable to create staff")
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": staff.ID, "username": staff.Username, "hospital": strings.ToLower(strings.TrimSpace(req.Hospital))})
}

func (h *Handler) login(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "username, password and hospital are required")
		return
	}
	result, err := h.staff.Login(c.Request.Context(), domain.LoginInput{
		Username: req.Username, Password: req.Password, HospitalCode: req.Hospital,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredential) {
			writeError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "username, password or hospital is incorrect")
			return
		}
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "unable to login")
		return
	}
	c.JSON(http.StatusOK, result)
}

type patientQuery struct {
	NationalID  string `form:"national_id" binding:"max=30"`
	PassportID  string `form:"passport_id" binding:"max=30"`
	FirstName   string `form:"first_name" binding:"max=100"`
	MiddleName  string `form:"middle_name" binding:"max=100"`
	LastName    string `form:"last_name" binding:"max=100"`
	DateOfBirth string `form:"date_of_birth" binding:"max=10"`
	PhoneNumber string `form:"phone_number" binding:"max=30"`
	Email       string `form:"email" binding:"omitempty,email,max=255"`
}

func (h *Handler) searchPatients(c *gin.Context) {
	var query patientQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "one or more search parameters are invalid")
		return
	}
	filter := domain.PatientFilter{
		NationalID: strings.TrimSpace(query.NationalID), PassportID: strings.TrimSpace(query.PassportID),
		FirstName: strings.TrimSpace(query.FirstName), MiddleName: strings.TrimSpace(query.MiddleName),
		LastName: strings.TrimSpace(query.LastName), PhoneNumber: strings.TrimSpace(query.PhoneNumber),
		Email: strings.TrimSpace(query.Email),
	}
	if query.DateOfBirth != "" {
		date, err := time.Parse("2006-01-02", query.DateOfBirth)
		if err != nil {
			writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "date_of_birth must use YYYY-MM-DD")
			return
		}
		filter.DateOfBirth = &date
	}
	patients, err := h.patients.Search(c.Request.Context(), hospitalIDFromContext(c), filter)
	if err != nil {
		if errors.Is(err, domain.ErrUpstream) {
			writeError(c, http.StatusBadGateway, "HIS_UNAVAILABLE", "the hospital information system is unavailable")
			return
		}
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "unable to search patients")
		return
	}
	data := make([]patientResponse, 0, len(patients))
	for _, patient := range patients {
		data = append(data, newPatientResponse(patient))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "count": len(data)})
}

type patientResponse struct {
	FirstNameTH  string `json:"first_name_th"`
	MiddleNameTH string `json:"middle_name_th"`
	LastNameTH   string `json:"last_name_th"`
	FirstNameEN  string `json:"first_name_en"`
	MiddleNameEN string `json:"middle_name_en"`
	LastNameEN   string `json:"last_name_en"`
	DateOfBirth  string `json:"date_of_birth,omitempty"`
	PatientHN    string `json:"patient_hn"`
	NationalID   string `json:"national_id,omitempty"`
	PassportID   string `json:"passport_id,omitempty"`
	PhoneNumber  string `json:"phone_number,omitempty"`
	Email        string `json:"email,omitempty"`
	Gender       string `json:"gender,omitempty"`
}

func newPatientResponse(p domain.Patient) patientResponse {
	response := patientResponse{
		FirstNameTH: p.FirstNameTH, MiddleNameTH: p.MiddleNameTH, LastNameTH: p.LastNameTH,
		FirstNameEN: p.FirstNameEN, MiddleNameEN: p.MiddleNameEN, LastNameEN: p.LastNameEN,
		PatientHN: p.PatientHN, NationalID: p.NationalID, PassportID: p.PassportID,
		PhoneNumber: p.PhoneNumber, Email: p.Email, Gender: p.Gender,
	}
	if p.DateOfBirth != nil {
		response.DateOfBirth = p.DateOfBirth.Format("2006-01-02")
	}
	return response
}
