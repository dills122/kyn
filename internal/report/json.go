package report

import (
	"encoding/json"

	"kyn/internal/rules"
)

func RenderJSON(summary rules.Summary) ([]byte, error) {
	return json.MarshalIndent(summary, "", "  ")
}
