package fm

/*
#include "FoundationModels.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// Schema wraps an FMGenerationSchemaRef, describing the shape of structured
// output or tool parameters.
type Schema struct {
	ref C.FMGenerationSchemaRef
}

// NewSchema creates a new generation schema.
//
// name        — identifier for the schema (e.g. "Recipe")
// description — optional description shown to the model
func NewSchema(name, description string) *Schema {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	var cDesc *C.char
	if description != "" {
		cDesc = C.CString(description)
		defer C.free(unsafe.Pointer(cDesc))
	}
	ref := C.FMGenerationSchemaCreate(cName, cDesc)
	s := &Schema{ref: ref}
	runtime.SetFinalizer(s, func(s *Schema) {
		if s.ref != nil {
			C.FMRelease(unsafe.Pointer(s.ref))
		}
	})
	return s
}

// AddField adds a typed property to the schema and returns the schema for
// chaining.
func (s *Schema) AddField(name, description, typeName string, optional bool) *Schema {
	prop := newProperty(name, description, typeName, optional)
	C.FMGenerationSchemaAddProperty(s.ref, prop)
	C.FMRelease(unsafe.Pointer(prop)) // schema now owns it
	return s
}

// AddReference nests another schema as a reference type inside this one.
func (s *Schema) AddReference(other *Schema) *Schema {
	C.FMGenerationSchemaAddReferenceSchema(s.ref, other.ref)
	return s
}

// JSON serialises the schema to its JSON representation (useful for debugging).
func (s *Schema) JSON() (string, error) {
	var errCode C.int
	var errDesc *C.char
	raw := C.FMGenerationSchemaGetJSONString(s.ref, &errCode, &errDesc)
	if errDesc != nil {
		msg := C.GoString(errDesc)
		C.FMFreeString(errDesc)
		return "", fmt.Errorf("schema JSON: %s", msg)
	}
	if raw == nil {
		return "", fmt.Errorf("schema JSON: nil result (code %d)", int(errCode))
	}
	defer C.FMFreeString(raw)
	return C.GoString(raw), nil
}

// newProperty creates a bare FMGenerationSchemaPropertyRef.
// The caller is responsible for releasing it (or transferring ownership via
// FMGenerationSchemaAddProperty).
func newProperty(name, description, typeName string, optional bool) C.FMGenerationSchemaPropertyRef {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	var cDesc *C.char
	if description != "" {
		cDesc = C.CString(description)
		defer C.free(unsafe.Pointer(cDesc))
	}
	cType := C.CString(typeName)
	defer C.free(unsafe.Pointer(cType))
	return C.FMGenerationSchemaPropertyCreate(cName, cDesc, cType, C.bool(optional))
}

// ---- Convenience field constructors ----------------------------------------

// StringField returns field metadata for a string property.
func StringField(name, description string) (string, string, string, bool) {
	return name, description, "string", false
}

// IntField returns field metadata for an integer property.
func IntField(name, description string) (string, string, string, bool) {
	return name, description, "integer", false
}

// DoubleField returns field metadata for a floating-point property.
func DoubleField(name, description string) (string, string, string, bool) {
	return name, description, "number", false
}

// BoolField returns field metadata for a boolean property.
func BoolField(name, description string) (string, string, string, bool) {
	return name, description, "boolean", false
}

// StringArrayField returns field metadata for an array-of-strings property.
func StringArrayField(name, description string) (string, string, string, bool) {
	return name, description, "array<string>", false
}

// IntArrayField returns field metadata for an array-of-integers property.
func IntArrayField(name, description string) (string, string, string, bool) {
	return name, description, "array<integer>", false
}

// GeneratedContent is the result of a structured generation call.
type GeneratedContent struct {
	ref C.FMGeneratedContentRef
}

func newGeneratedContent(ref C.FMGeneratedContentRef) *GeneratedContent {
	gc := &GeneratedContent{ref: ref}
	runtime.SetFinalizer(gc, func(gc *GeneratedContent) {
		if gc.ref != nil {
			C.FMRelease(unsafe.Pointer(gc.ref))
		}
	})
	return gc
}

// Get returns the JSON-encoded value of the named property, or an error.
func (gc *GeneratedContent) Get(property string) (string, error) {
	cProp := C.CString(property)
	defer C.free(unsafe.Pointer(cProp))
	var errCode C.int
	var errDesc *C.char
	raw := C.FMGeneratedContentGetPropertyValue(gc.ref, cProp, &errCode, &errDesc)
	if errDesc != nil {
		msg := C.GoString(errDesc)
		C.FMFreeString(errDesc)
		return "", fmt.Errorf("get property %q: %s", property, msg)
	}
	if raw == nil {
		return "", fmt.Errorf("get property %q: nil result (code %d)", property, int(errCode))
	}
	defer C.FMFreeString(raw)
	return C.GoString(raw), nil
}

// JSON returns the full generated content as a JSON string.
func (gc *GeneratedContent) JSON() string {
	raw := C.FMGeneratedContentGetJSONString(gc.ref)
	if raw == nil {
		return ""
	}
	defer C.FMFreeString(raw)
	return C.GoString(raw)
}

// IsComplete reports whether the generated content is fully formed.
func (gc *GeneratedContent) IsComplete() bool {
	return bool(C.FMGeneratedContentIsComplete(gc.ref))
}
