package fm

/*
#cgo CFLAGS: -I${SRCDIR}/../foundation-models-c/Sources/FoundationModelsCBindings/include
#cgo LDFLAGS: -L${SRCDIR}/../lib -L/usr/local/lib -lFoundationModels

#include "FoundationModels.h"
#include <stdlib.h>

// Declarations of C helpers defined in helpers.c.
typedef void (*ToolCallable)(FMGeneratedContentRef, unsigned int);
extern ToolCallable fmGetToolShim(int index);
extern FMTaskRef fmSessionRespond(FMLanguageModelSessionRef session, const char* prompt, void* userInfo);
extern void fmStreamIterate(FMLanguageModelSessionResponseStreamRef stream, void* userInfo);
extern FMTaskRef fmSessionRespondStructured(FMLanguageModelSessionRef session, const char* prompt, FMGenerationSchemaRef schema, void* userInfo);
*/
import "C"
