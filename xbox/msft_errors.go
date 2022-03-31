package xbox

import (
	"github.com/brawaru/marct/utils/slices"
)

type TokenError struct {
	Type          string `json:"error"`
	Description   string `json:"error_description"`
	Codes         []int  `json:"error_codes"`
	Timestamp     string `json:"timestamp"`
	TraceID       string `json:"trace_id"`
	CorrelationID string `json:"correlation_id"`
	ErrorURI      string `json:"error_uri"`
}

func (e *TokenError) Error() string {
	return e.Description
}

// Is checks whether provided error equals to this error, ignoring auto-generated properties.
func (e *TokenError) Is(target error) bool {
	t, ok := target.(*TokenError)
	return ok && (t.Type == "" || t.Type == e.Type) &&
		(t.Codes == nil || slices.Equal(e.Codes, t.Codes)) &&
		(t.CorrelationID == "" || t.CorrelationID == e.CorrelationID) &&
		(t.TraceID == "" || t.TraceID == e.TraceID) &&
		(t.Timestamp == "" || t.Timestamp == e.Timestamp)
}
