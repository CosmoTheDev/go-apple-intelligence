// Package fm provides a native Go interface to Apple's on-device Foundation
// Models (Apple Intelligence) via CGo bindings to Apple's official C API.
//
// # Requirements
//
//   - macOS 26.0 or later (Apple Silicon)
//   - Apple Intelligence enabled in System Settings > Apple Intelligence & Siri
//   - The libFoundationModels.dylib in lib/ (run `make build-native` to compile it)
//   - CGO_ENABLED=1 (the default on macOS)
//
// # Quick start
//
//	model := fm.DefaultModel()
//	if ok, err := model.IsAvailable(); !ok {
//	    log.Fatal(err)
//	}
//
//	session, err := fm.NewSession(fm.SessionOptions{
//	    Instructions: "You are a concise assistant.",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	response, err := session.Respond(context.Background(), "What is Go?")
//
// # Rate limiting
//
// Apple applies rate limits to Foundation Model calls from CLI processes.
// Foreground GUI applications are not rate-limited. For interactive or
// high-throughput workloads, consider embedding this library in a GUI app.
// For development and personal tooling the default limits are generally
// sufficient.
package fm
