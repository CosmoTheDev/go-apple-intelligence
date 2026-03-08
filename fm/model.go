package fm

/*
#include "FoundationModels.h"
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// UseCase configures the intended use of the model.
type UseCase int

const (
	UseCaseGeneral        UseCase = 0
	UseCaseContentTagging UseCase = 1
)

// Guardrails configures the safety guardrails applied to model output.
type Guardrails int

const (
	GuardrailsDefault                         Guardrails = 0
	GuardrailsPermissiveContentTransformations Guardrails = 1
)

// AvailabilityStatus describes why the model is unavailable.
type AvailabilityStatus int

const (
	AvailabilityStatusAppleIntelligenceNotEnabled AvailabilityStatus = 0
	AvailabilityStatusDeviceNotEligible           AvailabilityStatus = 1
	AvailabilityStatusModelNotReady               AvailabilityStatus = 2
	AvailabilityStatusUnknown                     AvailabilityStatus = 0xFF
)

// Model is a reference to an Apple on-device SystemLanguageModel.
type Model struct {
	ref C.FMSystemLanguageModelRef
}

// DefaultModel returns the system's default Foundation Model (Apple Intelligence).
func DefaultModel() *Model {
	ref := C.FMSystemLanguageModelGetDefault()
	m := &Model{ref: ref}
	runtime.SetFinalizer(m, func(m *Model) {
		if m.ref != nil {
			C.FMRelease(unsafe.Pointer(m.ref))
		}
	})
	return m
}

// NewModel creates a model with specific use-case and guardrail settings.
func NewModel(useCase UseCase, guardrails Guardrails) *Model {
	ref := C.FMSystemLanguageModelCreate(
		C.FMSystemLanguageModelUseCase(useCase),
		C.FMSystemLanguageModelGuardrails(guardrails),
	)
	m := &Model{ref: ref}
	runtime.SetFinalizer(m, func(m *Model) {
		if m.ref != nil {
			C.FMRelease(unsafe.Pointer(m.ref))
		}
	})
	return m
}

// IsAvailable reports whether Apple Intelligence is ready to use.
// If unavailable, it returns a descriptive error explaining why.
func (m *Model) IsAvailable() (bool, error) {
	var reason C.FMSystemLanguageModelUnavailableReason
	available := bool(C.FMSystemLanguageModelIsAvailable(m.ref, &reason))
	if available {
		return true, nil
	}
	switch AvailabilityStatus(reason) {
	case AvailabilityStatusAppleIntelligenceNotEnabled:
		return false, ErrAppleIntelligenceNotEnabled
	case AvailabilityStatusDeviceNotEligible:
		return false, ErrDeviceNotEligible
	case AvailabilityStatusModelNotReady:
		return false, ErrModelNotReady
	default:
		return false, ErrUnknownAvailability
	}
}
