package his

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

func TestSearchByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/patient/search/1234567890123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"first_name_en":"Alice","last_name_en":"Doe","date_of_birth":"1990-01-02","patient_hn":"HN001","national_id":"1234567890123","gender":"F"}`))
	}))
	defer server.Close()

	patient, err := NewClient(time.Second).SearchByID(context.Background(), domain.Hospital{HISBaseURL: server.URL}, "1234567890123")
	if err != nil || patient.PatientHN != "HN001" || patient.DateOfBirth == nil {
		t.Fatalf("unexpected result: %#v, %v", patient, err)
	}
}

func TestSearchByIDFailures(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }))
		defer server.Close()
		_, err := NewClient(time.Second).SearchByID(context.Background(), domain.Hospital{HISBaseURL: server.URL}, "missing")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected not found, got %v", err)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"patient_hn":""}`)) }))
		defer server.Close()
		_, err := NewClient(time.Second).SearchByID(context.Background(), domain.Hospital{HISBaseURL: server.URL}, "bad")
		if !errors.Is(err, domain.ErrUpstream) {
			t.Fatalf("expected upstream error, got %v", err)
		}
	})
}
