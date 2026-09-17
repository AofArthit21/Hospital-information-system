package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/AofArthit21/Hospital-information-system/internal/domain"
)

func TestBuildPatientSearchQueryReusesNamePlaceholder(t *testing.T) {
	date := time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)
	query, args := buildPatientSearchQuery(17, domain.PatientFilter{
		FirstName: "Ali", DateOfBirth: &date,
	})
	if strings.Contains(query, "%!") {
		t.Fatalf("query contains a formatting error: %s", query)
	}
	if !strings.Contains(query, "first_name_th ILIKE $2 OR first_name_en ILIKE $2") {
		t.Fatalf("name fields do not reuse the same placeholder: %s", query)
	}
	if !strings.Contains(query, "date_of_birth = $3") || len(args) != 3 {
		t.Fatalf("later placeholders shifted: query=%s args=%d", query, len(args))
	}
}
