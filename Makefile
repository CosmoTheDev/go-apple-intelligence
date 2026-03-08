SWIFT_PKG  := foundation-models-c
LIB_DIR    := lib
DYLIB_NAME := libFoundationModels.dylib
DYLIB_PATH := $(LIB_DIR)/$(DYLIB_NAME)
ABS_LIB    := $(abspath $(LIB_DIR))

RELEASE_URL ?= https://github.com/CosmoTheDev/go-apple-intelligence/releases/latest/download/$(DYLIB_NAME)

# Embed the rpath so binaries find the dylib without DYLD_LIBRARY_PATH.
# CGO_LDFLAGS is an env var and is not subject to Go's #cgo flag security filter.
export CGO_LDFLAGS := -Wl,-rpath,$(ABS_LIB)

.PHONY: build-native download-native install-dylib build test clean check-dylib help

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/^## //' | column -t -s ':'

## build-native: compile the Swift Foundation Models C bindings (requires Xcode 26+)
build-native:
	@echo "==> Building Foundation Models C bindings (requires macOS 26 + Xcode 26)..."
	swift build --package-path $(SWIFT_PKG) -c release
	mkdir -p $(LIB_DIR)
	cp $(SWIFT_PKG)/.build/release/$(DYLIB_NAME) $(DYLIB_PATH)
	# Set the dylib's own install name to @rpath so LC_RPATH entries work correctly.
	install_name_tool -id @rpath/$(DYLIB_NAME) $(DYLIB_PATH)
	@echo "==> dylib written to $(DYLIB_PATH)"

## download-native: download a pre-built dylib from the latest GitHub release
download-native:
	@echo "==> Downloading pre-built $(DYLIB_NAME) from $(RELEASE_URL) ..."
	mkdir -p $(LIB_DIR)
	curl -fL --progress-bar -o $(DYLIB_PATH) "$(RELEASE_URL)"
	install_name_tool -id @rpath/$(DYLIB_NAME) $(DYLIB_PATH)
	@echo "==> dylib saved to $(DYLIB_PATH)"

## install-dylib: copy the dylib to /usr/local/lib for system-wide access
##                (allows plain `go run` without any env vars)
install-dylib: check-dylib
	sudo cp $(DYLIB_PATH) /usr/local/lib/$(DYLIB_NAME)
	@echo "==> Installed to /usr/local/lib/$(DYLIB_NAME)"

## check-dylib: ensure the dylib exists
check-dylib:
	@if [ ! -f "$(DYLIB_PATH)" ]; then \
		echo "$(DYLIB_PATH) not found. Run 'make build-native' or 'make download-native'."; \
		exit 1; \
	fi

## build: build all packages
build: check-dylib
	CGO_ENABLED=1 go build ./...

## run-chat: run the chat example
run-chat: check-dylib
	CGO_ENABLED=1 go run ./example/chat/

## run-stream: run the streaming example
run-stream: check-dylib
	CGO_ENABLED=1 go run ./example/stream/

## run-structured: run the structured output example
run-structured: check-dylib
	CGO_ENABLED=1 go run ./example/structured/

## run-tools: run the tool calling example
run-tools: check-dylib
	CGO_ENABLED=1 go run ./example/tools/

## run-chatbot: run the interactive multi-turn chatbot
run-chatbot: check-dylib
	CGO_ENABLED=1 go run ./example/chatbot/

## run-chatbot-with-memory: run the chatbot with cross-session memory
run-chatbot-with-memory: check-dylib
	CGO_ENABLED=1 go run ./example/chatbot-memory/

## test: run all Go tests
test: check-dylib
	CGO_ENABLED=1 go test ./...

## clean: remove build artifacts
clean:
	swift package --package-path $(SWIFT_PKG) clean 2>/dev/null || true
	rm -rf $(LIB_DIR)
	go clean ./...
