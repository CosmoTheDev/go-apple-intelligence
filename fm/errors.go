package fm

import (
	"errors"
	"fmt"
)

// Availability errors returned by Model.IsAvailable and Session methods.
var (
	ErrAppleIntelligenceNotEnabled = errors.New("Apple Intelligence is not enabled — go to Settings > Apple Intelligence & Siri")
	ErrDeviceNotEligible           = errors.New("device is not eligible for Apple Intelligence (requires Apple Silicon Mac, iPhone 15 Pro/Pro Max or later, or iPad with M1 or later)")
	ErrModelNotReady               = errors.New("Apple Intelligence model is not yet ready — it may still be downloading")
	ErrUnknownAvailability         = errors.New("Apple Intelligence availability is unknown")
	ErrModelUnavailable            = errors.New("Apple Intelligence model is unavailable")
	ErrTooManyTools                = errors.New("a session supports at most 8 tools")
)

// ModelError wraps a status code returned by the C layer.
type ModelError struct {
	Status int
}

func (e *ModelError) Error() string {
	switch e.Status {
	case 1:
		return "Apple Intelligence not available"
	case 2:
		return "device not eligible for Apple Intelligence"
	case 3:
		return "Apple Intelligence not enabled in Settings"
	case 4:
		return "Apple Intelligence model not ready (still downloading)"
	case 7:
		return "rate limited — Apple Intelligence CLI rate limit reached"
	case 10:
		return "invalid schema — check that field type names are lowercase (e.g. \"string\", \"integer\", \"boolean\")"
	default:
		return fmt.Sprintf("Foundation Models error (status %d)", e.Status)
	}
}

