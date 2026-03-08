# go-apple-intelligence

An unofficial, open-source Go SDK for Apple's on-device Foundation Models
(Apple Intelligence) — **no SwiftUI app required**.

Bridges directly to Apple's official C API ([`apple/python-apple-fm-sdk`](https://github.com/apple/python-apple-fm-sdk))
via CGo. The same approach Apple uses for their own Python SDK.

## Features

- One-shot and multi-turn **text generation**
- Real-time **token streaming**
- **Structured output** via schema-guided generation
- **Tool calling** — the model can invoke Go functions during generation
- Idiomatic Go API with `context.Context` support
- Full memory management via finalizers

## Requirements

| Requirement | Version |
|---|---|
| macOS | 26.0+ (Apple Silicon) |
| Apple Intelligence | Enabled in Settings > Apple Intelligence & Siri |
| Go | 1.24+ |
| Xcode | 26.0+ (only to build the dylib) |

> **Rate limiting**: Apple applies rate limits to Foundation Model calls from
> CLI processes. For development and personal tooling the limits are generally
> fine. GUI foreground apps have no rate limit.

---

## Getting Started

### 1. Clone the repo

```bash
git clone https://github.com/CosmoTheDev/go-apple-intelligence.git
cd go-apple-intelligence
```

### 2. Build the native dylib

This compiles Apple's Foundation Models C bindings (requires Xcode 26+):

```bash
make build-native
```

This produces `lib/libFoundationModels.dylib`. You only need to do this once.

> **No Xcode?** If you're on macOS 26 but don't have Xcode installed, you can
> download a pre-built binary instead (once a release is published):
> ```bash
> make download-native
> ```

### 3. Run an example

```bash
make run-chat       # multi-turn conversation
make run-stream     # real-time streaming output
make run-structured # schema-guided structured output
make run-tools      # tool/function calling
make run-chatbot    # interactive chatbot REPL
```

That's it. The Makefile handles the CGo flags and rpath automatically.

---

## Using as a library

### 1. Add the dependency

```bash
go get github.com/CosmoTheDev/go-apple-intelligence/fm
```

### 2. Get the native dylib

The Go package is a CGo bridge to a native dylib. You need
`libFoundationModels.dylib` on your system — **do this once**.

**Option A — download the pre-built binary** (no Xcode required):

```bash
curl -fL \
  https://github.com/CosmoTheDev/go-apple-intelligence/releases/latest/download/libFoundationModels.dylib \
  -o /tmp/libFoundationModels.dylib
sudo install -m 755 /tmp/libFoundationModels.dylib /usr/local/lib/
```

**Option B — build from source** (requires Xcode 26+):

```bash
git clone https://github.com/CosmoTheDev/go-apple-intelligence.git
cd go-apple-intelligence
make build-native && make install-dylib
```

### 3. Build your app

After the dylib is installed to `/usr/local/lib/`, plain `go build` works:

```bash
CGO_ENABLED=1 go build ./...
```

If you want to keep the dylib local (not system-wide), embed the rpath instead:

```bash
CGO_ENABLED=1 CGO_LDFLAGS="-Wl,-rpath,/path/to/lib" go build ./...
```

---

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    fm "github.com/CosmoTheDev/go-apple-intelligence/fm"
)

func main() {
    model := fm.DefaultModel()
    if ok, err := model.IsAvailable(); !ok {
        log.Fatal(err)
    }

    session, err := fm.NewSession(fm.SessionOptions{
        Instructions: "You are a concise assistant.",
    })
    if err != nil {
        log.Fatal(err)
    }

    resp, err := session.Respond(context.Background(), "What is Go?")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(resp)
}
```

---

## Examples

| Example | Makefile target | Description |
|---|---|---|
| `example/chat/` | `make run-chat` | One-shot and multi-turn conversation |
| `example/stream/` | `make run-stream` | Real-time streaming output |
| `example/structured/` | `make run-structured` | Schema-guided structured generation |
| `example/tools/` | `make run-tools` | Tool/function calling |
| `example/chatbot/` | `make run-chatbot` | Interactive multi-turn chatbot REPL |
| `example/chatbot-memory/` | `make run-chatbot-with-memory` | Chatbot with cross-session memory |

All examples are run from the project root via `make`. Running them directly
with `go run` requires the dylib to be on the search path — either install it
system-wide (`make install-dylib`) or set `DYLD_LIBRARY_PATH`:

```bash
# From the project root, without make:
DYLD_LIBRARY_PATH=$(pwd)/lib CGO_ENABLED=1 go run ./example/chat/
```

---

## API overview

```go
// Check model availability
model := fm.DefaultModel()
available, err := model.IsAvailable()

// Create a session (Instructions is an optional system prompt)
session, err := fm.NewSession(fm.SessionOptions{
    Instructions: "You are a helpful assistant.",
})

// One-shot or multi-turn text generation
// (session keeps conversation history automatically across calls)
response, err := session.Respond(ctx, "What is Go?")
response, err  = session.Respond(ctx, "Who made it?")  // remembers context

// Reset conversation history
session.Reset()

// Real-time streaming
err = session.StreamResponse(ctx, "Tell me a story", func(chunk string, done bool) {
    if !done {
        fmt.Print(chunk)
    }
})

// Structured output
schema := fm.NewSchema("Person", "").
    AddField("name", "full name", "string", false).
    AddField("age",  "age in years", "integer", false)
result, err := session.RespondStructured(ctx, "Describe a fictional person", schema)
name, _ := result.Get("name")
fmt.Println(result.JSON())  // {"name":"Alice","age":30}

// Tool calling — the model invokes Go functions during generation
params := fm.NewSchema("Params", "").AddField("city", "city name", "string", false)
tool := fm.NewTool(
    "get_weather",
    "Returns current weather for a city",
    params,
    func(argsJSON string) string {
        return `{"temp": 72, "condition": "sunny"}`
    },
)
session, err = fm.NewSession(fm.SessionOptions{Tools: []*fm.Tool{tool}})
response, err = session.Respond(ctx, "What's the weather in SF?")
```

---

## Makefile targets

```
make build-native    # compile the Swift dylib (requires Xcode 26+)
make download-native # download a pre-built dylib from GitHub Releases
make install-dylib   # install dylib to /usr/local/lib/ (system-wide)
make build           # go build ./...
make run-chat        # run example/chat/
make run-stream      # run example/stream/
make run-structured  # run example/structured/
make run-tools       # run example/tools/
make run-chatbot              # run example/chatbot/ (interactive REPL)
make run-chatbot-with-memory  # run example/chatbot-memory/ (persists memory across sessions)
make test            # go test ./...
make clean           # remove build artifacts
```

---

## Architecture

```
Your Go app
    └── fm/  (this package, CGo)
         └── lib/libFoundationModels.dylib  (built from foundation-models-c/)
              └── Apple FoundationModels.framework
                   └── On-device Apple Intelligence model
```

The `foundation-models-c/` directory is vendored from Apple's official
[`python-apple-fm-sdk`](https://github.com/apple/python-apple-fm-sdk) (Apache 2.0).
It provides the C header and Swift `@_cdecl` exports that make the FoundationModels
framework callable from non-Swift languages.

## License

This SDK (Go code) is MIT licensed.

The vendored C bindings in `foundation-models-c/` are Copyright © 2026 Apple Inc.,
licensed under the Apache License 2.0 — see [`foundation-models-c/LICENSE.md`](foundation-models-c/LICENSE.md).
