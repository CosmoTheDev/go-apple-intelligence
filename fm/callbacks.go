package fm

/*
#include "FoundationModels.h"
#include <stdlib.h>
*/
import "C"

import (
	"sync"
	"unsafe"
)

// textCallback is the payload delivered to a text-response waiter.
type textCallback struct {
	content string // cumulative content so far (empty when done)
	done    bool
	err     error
}

// structuredCallback is the payload delivered to a structured-response waiter.
type structuredCallback struct {
	ref C.FMGeneratedContentRef
	err error
}

var (
	textMu   sync.Mutex
	textMap  = map[uintptr]chan textCallback{}
	textNext uintptr = 1

	structMu   sync.Mutex
	structMap  = map[uintptr]chan structuredCallback{}
	structNext uintptr = 1

	// toolSlots holds, for each shim index, the active Go Tool handler and its
	// associated FMBridgedToolRef so we can call FMBridgedToolFinishCall.
	toolMu    sync.Mutex
	toolSlots [8]struct {
		ref     C.FMBridgedToolRef
		handler func(args string) string
	}
)

// registerTextHandler allocates a channel keyed by a uintptr handle.
func registerTextHandler() (uintptr, chan textCallback) {
	ch := make(chan textCallback, 64)
	textMu.Lock()
	id := textNext
	textNext++
	textMap[id] = ch
	textMu.Unlock()
	return id, ch
}

func unregisterTextHandler(id uintptr) {
	textMu.Lock()
	delete(textMap, id)
	textMu.Unlock()
}

// registerStructuredHandler allocates a channel keyed by a uintptr handle.
func registerStructuredHandler() (uintptr, chan structuredCallback) {
	ch := make(chan structuredCallback, 1)
	structMu.Lock()
	id := structNext
	structNext++
	structMap[id] = ch
	structMu.Unlock()
	return id, ch
}

func unregisterStructuredHandler(id uintptr) {
	structMu.Lock()
	delete(structMap, id)
	structMu.Unlock()
}

// goFMTextCallback is called by the C layer for both streaming and non-streaming
// text responses. content==nil signals completion.
//
//export goFMTextCallback
func goFMTextCallback(status C.int, content *C.char, length C.size_t, userInfo unsafe.Pointer) {
	id := uintptr(userInfo)
	textMu.Lock()
	ch, ok := textMap[id]
	textMu.Unlock()
	if !ok {
		return
	}

	if int(status) != 0 {
		ch <- textCallback{err: &ModelError{Status: int(status)}}
		return
	}
	if content == nil {
		ch <- textCallback{done: true}
		return
	}
	ch <- textCallback{content: C.GoStringN(content, C.int(length))}
}

// goFMStructuredCallback is called by the C layer when structured generation completes.
//
//export goFMStructuredCallback
func goFMStructuredCallback(status C.int, content C.FMGeneratedContentRef, userInfo unsafe.Pointer) {
	id := uintptr(userInfo)
	structMu.Lock()
	ch, ok := structMap[id]
	structMu.Unlock()
	if !ok {
		return
	}

	if int(status) != 0 {
		ch <- structuredCallback{err: &ModelError{Status: int(status)}}
		return
	}
	if content != nil {
		C.FMRetain(unsafe.Pointer(content))
	}
	ch <- structuredCallback{ref: content}
}

// Tool callables — one per slot in the shim pool.

//export goFMToolCallable0
func goFMToolCallable0(content C.FMGeneratedContentRef, callID C.uint) {
	dispatchToolCall(0, content, callID)
}

//export goFMToolCallable1
func goFMToolCallable1(content C.FMGeneratedContentRef, callID C.uint) {
	dispatchToolCall(1, content, callID)
}

//export goFMToolCallable2
func goFMToolCallable2(content C.FMGeneratedContentRef, callID C.uint) {
	dispatchToolCall(2, content, callID)
}

//export goFMToolCallable3
func goFMToolCallable3(content C.FMGeneratedContentRef, callID C.uint) {
	dispatchToolCall(3, content, callID)
}

//export goFMToolCallable4
func goFMToolCallable4(content C.FMGeneratedContentRef, callID C.uint) {
	dispatchToolCall(4, content, callID)
}

//export goFMToolCallable5
func goFMToolCallable5(content C.FMGeneratedContentRef, callID C.uint) {
	dispatchToolCall(5, content, callID)
}

//export goFMToolCallable6
func goFMToolCallable6(content C.FMGeneratedContentRef, callID C.uint) {
	dispatchToolCall(6, content, callID)
}

//export goFMToolCallable7
func goFMToolCallable7(content C.FMGeneratedContentRef, callID C.uint) {
	dispatchToolCall(7, content, callID)
}

func dispatchToolCall(slot int, content C.FMGeneratedContentRef, callID C.uint) {
	toolMu.Lock()
	entry := toolSlots[slot]
	toolMu.Unlock()

	if entry.handler == nil || entry.ref == nil {
		return
	}

	// Extract args as JSON string.
	var argsJSON string
	if content != nil {
		raw := C.FMGeneratedContentGetJSONString(content)
		if raw != nil {
			argsJSON = C.GoString(raw)
			C.FMFreeString(raw)
		}
	}

	// Invoke the Go handler in a goroutine so we don't block the C callback thread.
	go func() {
		result := entry.handler(argsJSON)
		cResult := C.CString(result)
		defer C.free(unsafe.Pointer(cResult))
		C.FMBridgedToolFinishCall(entry.ref, callID, cResult)
	}()
}
