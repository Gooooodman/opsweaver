package mask_test

import (
	"reflect"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/mask"
)

func TestMask_RedactsSensitiveKeysRecursively(t *testing.T) {
	in := map[string]any{
		"service":       "github",
		"token":         "tok_123",
		"Password":      "p@ss",
		"authorization": "Bearer secret-token",
		"client_secret": "secret-value",
		"api_key":       "api-key-value",
		"apikey":        "api-key-value-2",
		"master_key":    "master-key-value",
		"nested": map[string]any{
			"accessToken": "nested-token",
			"enabled":     true,
		},
		"items": []any{
			map[string]any{
				"refresh_token": "refresh-token",
				"name":          "worker",
			},
			"plain-value",
		},
	}

	got := mask.Mask(in)

	want := map[string]any{
		"service":       "github",
		"token":         mask.Redacted,
		"Password":      mask.Redacted,
		"authorization": mask.Redacted,
		"client_secret": mask.Redacted,
		"api_key":       mask.Redacted,
		"apikey":        mask.Redacted,
		"master_key":    mask.Redacted,
		"nested": map[string]any{
			"accessToken": mask.Redacted,
			"enabled":     true,
		},
		"items": []any{
			map[string]any{
				"refresh_token": mask.Redacted,
				"name":          "worker",
			},
			"plain-value",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Mask() = %#v, want %#v", got, want)
	}
}

func TestMask_DoesNotMutateInput(t *testing.T) {
	in := map[string]any{
		"token": "tok_123",
		"nested": map[string]any{
			"password": "p@ss",
			"name":     "database",
		},
		"items": []any{
			map[string]any{
				"secret": "value",
			},
		},
	}

	original := map[string]any{
		"token": "tok_123",
		"nested": map[string]any{
			"password": "p@ss",
			"name":     "database",
		},
		"items": []any{
			map[string]any{
				"secret": "value",
			},
		},
	}

	_ = mask.Mask(in)

	if !reflect.DeepEqual(in, original) {
		t.Errorf("input mutated to %#v, want %#v", in, original)
	}
}
