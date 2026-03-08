// stream demonstrates real-time token streaming from the Apple on-device model.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	fm "github.com/CosmoTheDev/go-apple-intelligence/fm"
)

func main() {
	model := fm.DefaultModel()
	if ok, err := model.IsAvailable(); !ok {
		log.Fatalf("Apple Intelligence unavailable: %v", err)
	}

	session, err := fm.NewSession(fm.SessionOptions{})
	if err != nil {
		log.Fatalf("NewSession: %v", err)
	}

	fmt.Print("Response: ")
	err = session.StreamResponse(
		context.Background(),
		"Tell me a short two-sentence story about a gopher exploring a cave.",
		func(partial string, done bool) {
			if done {
				fmt.Println()
				return
			}
			fmt.Fprint(os.Stdout, partial)
		},
	)
	if err != nil {
		log.Fatalf("StreamResponse: %v", err)
	}
}
