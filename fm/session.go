package fm

/*
#include "FoundationModels.h"
#include <stdlib.h>

typedef void (*ToolCallable)(FMGeneratedContentRef, unsigned int);
extern ToolCallable fmGetToolShim(int index);
extern void fmStreamIterate(FMLanguageModelSessionResponseStreamRef stream, void* userInfo);
extern FMTaskRef fmSessionRespondStructured(FMLanguageModelSessionRef session, const char* prompt, FMGenerationSchemaRef schema, void* userInfo);
*/
import "C"

import (
	"context"
	"fmt"
	"runtime"
	"unsafe"
)

// SessionOptions configures a language model session.
type SessionOptions struct {
	// Model to use. Defaults to DefaultModel() if nil.
	Model *Model
	// Instructions (system prompt) for the session. Optional.
	Instructions string
	// Tools the model may call. Supports up to 8 tools.
	Tools []*Tool
}

// Session wraps an FMLanguageModelSessionRef.
// A session maintains conversation history across multiple Respond calls.
type Session struct {
	ref       C.FMLanguageModelSessionRef
	toolRefs  []C.FMBridgedToolRef // C-side tool refs owned by this session
	toolSlots []int                // shim slot indices in use
}

// NewSession creates a new language model session.
func NewSession(opts SessionOptions) (*Session, error) {
	if len(opts.Tools) > 8 {
		return nil, ErrTooManyTools
	}

	model := opts.Model
	if model == nil {
		model = DefaultModel()
	}

	if ok, err := model.IsAvailable(); !ok {
		return nil, err
	}

	var cInstructions *C.char
	if opts.Instructions != "" {
		cInstructions = C.CString(opts.Instructions)
		defer C.free(unsafe.Pointer(cInstructions))
	}

	s := &Session{}

	// Build C tool refs using the shim pool.
	for i, tool := range opts.Tools {
		shim := C.fmGetToolShim(C.int(i))
		if shim == nil {
			s.releaseToolRefs()
			return nil, fmt.Errorf("no tool shim for slot %d", i)
		}

		cName := C.CString(tool.Name)
		cDesc := C.CString(tool.Description)
		var errCode C.int
		var errDesc *C.char

		ref := C.FMBridgedToolCreate(cName, cDesc, tool.Params.ref, shim, &errCode, &errDesc)
		C.free(unsafe.Pointer(cName))
		C.free(unsafe.Pointer(cDesc))

		if errDesc != nil {
			msg := C.GoString(errDesc)
			C.FMFreeString(errDesc)
			s.releaseToolRefs()
			return nil, fmt.Errorf("creating tool %q: %s", tool.Name, msg)
		}
		if ref == nil {
			s.releaseToolRefs()
			return nil, fmt.Errorf("creating tool %q: unknown error (code %d)", tool.Name, int(errCode))
		}

		s.toolRefs = append(s.toolRefs, ref)
		s.toolSlots = append(s.toolSlots, i)

		toolMu.Lock()
		toolSlots[i].ref = ref
		toolSlots[i].handler = tool.Handler
		toolMu.Unlock()
	}

	// Build the session, passing tools as a contiguous C array.
	var toolsPtr *C.FMBridgedToolRef
	if len(s.toolRefs) > 0 {
		toolsPtr = &s.toolRefs[0]
	}

	s.ref = C.FMLanguageModelSessionCreateFromSystemLanguageModel(
		model.ref,
		cInstructions,
		toolsPtr,
		C.int(len(s.toolRefs)),
	)

	runtime.SetFinalizer(s, func(s *Session) { s.close() })
	return s, nil
}

func (s *Session) releaseToolRefs() {
	for _, r := range s.toolRefs {
		C.FMRelease(unsafe.Pointer(r))
	}
	s.toolRefs = nil
}

func (s *Session) close() {
	if s.ref != nil {
		C.FMRelease(unsafe.Pointer(s.ref))
		s.ref = nil
	}
	s.releaseToolRefs()

	toolMu.Lock()
	for _, slot := range s.toolSlots {
		toolSlots[slot].ref = nil
		toolSlots[slot].handler = nil
	}
	toolMu.Unlock()
	s.toolSlots = nil
}

// Reset clears the session's conversation history.
func (s *Session) Reset() {
	C.FMLanguageModelSessionReset(s.ref)
}

// Respond sends a prompt and waits for the complete response.
// The session retains conversation history, so subsequent calls form a
// multi-turn conversation.
//
// Internally uses the stream API (same as StreamResponse) because
// FMLanguageModelSessionRespond relies on Swift's Task executor which may
// not be running in CLI processes — the stream path is what Apple's own C
// example uses and is reliable in non-GUI contexts.
func (s *Session) Respond(ctx context.Context, prompt string) (string, error) {
	var buf []string
	err := s.StreamResponse(ctx, prompt, func(partial string, done bool) {
		if !done {
			buf = append(buf, partial)
		}
	})
	if err != nil {
		return "", err
	}
	result := ""
	for _, chunk := range buf {
		result += chunk
	}
	return result, nil
}

// StreamHandler is called for each delta chunk during streaming.
// partial is the new text since the last call. done is true on the final call.
type StreamHandler func(partial string, done bool)

// StreamResponse streams the model's response, calling handler with each delta.
// fmStreamIterate returns immediately; the C layer drives callbacks on a
// background thread. We wait on a channel instead of busy-polling.
func (s *Session) StreamResponse(ctx context.Context, prompt string, handler StreamHandler) error {
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	id, ch := registerTextHandler()
	defer unregisterTextHandler(id)

	stream := C.FMLanguageModelSessionStreamResponse(s.ref, cPrompt)
	if stream == nil {
		return ErrModelUnavailable
	}
	defer C.FMRelease(unsafe.Pointer(stream))

	// Start iteration; callbacks fire asynchronously on the C layer's thread.
	C.fmStreamIterate(stream, unsafe.Pointer(uintptr(id)))

	var lastLen int
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case cb := <-ch:
			if cb.err != nil {
				return cb.err
			}
			if cb.done {
				handler("", true)
				return nil
			}
			// Compute delta from cumulative content.
			if len(cb.content) > lastLen {
				delta := cb.content[lastLen:]
				lastLen = len(cb.content)
				handler(delta, false)
			}
		}
	}
}

// RespondStructured sends a prompt and returns structured output conforming to
// the provided schema.
func (s *Session) RespondStructured(ctx context.Context, prompt string, schema *Schema) (*GeneratedContent, error) {
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	id, ch := registerStructuredHandler()
	defer unregisterStructuredHandler(id)

	taskRef := C.fmSessionRespondStructured(s.ref, cPrompt, schema.ref, unsafe.Pointer(uintptr(id)))
	if taskRef != nil {
		defer C.FMRelease(unsafe.Pointer(taskRef))
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case cb := <-ch:
		if cb.err != nil {
			return nil, cb.err
		}
		if cb.ref == nil {
			return nil, fmt.Errorf("structured generation returned nil content")
		}
		return newGeneratedContent(cb.ref), nil
	}
}

// TranscriptJSON returns the session's full conversation transcript as JSON.
func (s *Session) TranscriptJSON() (string, error) {
	var errCode C.int
	var errDesc *C.char
	raw := C.FMLanguageModelSessionGetTranscriptJSONString(s.ref, &errCode, &errDesc)
	if errDesc != nil {
		msg := C.GoString(errDesc)
		C.FMFreeString(errDesc)
		return "", fmt.Errorf("transcript JSON: %s", msg)
	}
	if raw == nil {
		return "", fmt.Errorf("transcript JSON: nil result (code %d)", int(errCode))
	}
	defer C.FMFreeString(raw)
	return C.GoString(raw), nil
}
