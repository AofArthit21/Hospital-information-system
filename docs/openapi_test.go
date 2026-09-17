package docs

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOpenAPISpecIsValidYAML(t *testing.T) {
	contents, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		OpenAPI string                    `yaml:"openapi"`
		Paths   map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(contents, &document); err != nil {
		t.Fatalf("invalid OpenAPI YAML: %v", err)
	}
	if document.OpenAPI == "" {
		t.Fatal("OpenAPI version is missing")
	}
	for _, path := range []string{"/staff/create", "/staff/login", "/patient/search"} {
		if _, ok := document.Paths[path]; !ok {
			t.Fatalf("required path %s is missing", path)
		}
	}
}
