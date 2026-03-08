/*
 * helpers.c — CGo bridge helpers for go-apple-intelligence.
 * Compiled by CGo using the CFLAGS from cgo.go (package-wide).
 */

#include "FoundationModels.h"
#include <stdlib.h>

/* Forward-declare the Go-exported callback functions produced by //export. */
extern void goFMTextCallback(int status, const char *content, size_t length, void *userInfo);
extern void goFMStructuredCallback(int status, FMGeneratedContentRef content, void *userInfo);
extern void goFMToolCallable0(FMGeneratedContentRef content, unsigned int callID);
extern void goFMToolCallable1(FMGeneratedContentRef content, unsigned int callID);
extern void goFMToolCallable2(FMGeneratedContentRef content, unsigned int callID);
extern void goFMToolCallable3(FMGeneratedContentRef content, unsigned int callID);
extern void goFMToolCallable4(FMGeneratedContentRef content, unsigned int callID);
extern void goFMToolCallable5(FMGeneratedContentRef content, unsigned int callID);
extern void goFMToolCallable6(FMGeneratedContentRef content, unsigned int callID);
extern void goFMToolCallable7(FMGeneratedContentRef content, unsigned int callID);

/* Tool shim pool — maps slot index → per-slot C function pointer. */
typedef void (*ToolCallable)(FMGeneratedContentRef, unsigned int);

static ToolCallable toolShims[8] = {
    goFMToolCallable0,
    goFMToolCallable1,
    goFMToolCallable2,
    goFMToolCallable3,
    goFMToolCallable4,
    goFMToolCallable5,
    goFMToolCallable6,
    goFMToolCallable7,
};

ToolCallable fmGetToolShim(int index) {
    if (index < 0 || index >= 8) return NULL;
    return toolShims[index];
}

/* Shim: non-streaming respond via callback. */
FMTaskRef fmSessionRespond(
    FMLanguageModelSessionRef session,
    const char *prompt,
    void *userInfo)
{
    return FMLanguageModelSessionRespond(session, prompt, userInfo, goFMTextCallback);
}

/* Shim: iterate a response stream via callback. */
void fmStreamIterate(
    FMLanguageModelSessionResponseStreamRef stream,
    void *userInfo)
{
    FMLanguageModelSessionResponseStreamIterate(stream, userInfo, goFMTextCallback);
}

/* Shim: structured generation via callback. */
FMTaskRef fmSessionRespondStructured(
    FMLanguageModelSessionRef session,
    const char *prompt,
    FMGenerationSchemaRef schema,
    void *userInfo)
{
    return FMLanguageModelSessionRespondWithSchema(session, prompt, schema, userInfo, goFMStructuredCallback);
}
