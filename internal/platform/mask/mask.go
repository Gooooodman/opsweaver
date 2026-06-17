package mask

import "strings"

const Redacted = "[REDACTED]"

var sensitiveKeyParts = []string{
	"token",
	"password",
	"authorization",
	"secret",
	"api_key",
	"apikey",
	"master_key",
}

func Mask(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			if isSensitiveKey(key) {
				out[key] = Redacted
				continue
			}
			out[key] = Mask(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = Mask(item)
		}
		return out
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, part := range sensitiveKeyParts {
		if strings.Contains(lower, part) {
			return true
		}
	}
	return false
}
