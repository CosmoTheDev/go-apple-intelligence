// chat demonstrates a simple one-shot and multi-turn conversation using the
// Apple on-device Foundation Model.
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
		log.Fatalf("Apple Intelligence unavailable: %v", err)
	}
	fmt.Println("Apple Intelligence is available.")

	session, err := fm.NewSession(fm.SessionOptions{
		Instructions: "You are a concise assistant. Keep answers under three sentences.",
	})
	if err != nil {
		log.Fatalf("NewSession: %v", err)
	}

	ctx := context.Background()

	// First turn.
	resp, err := session.Respond(ctx, "What is the Go programming language?")
	if err != nil {
		log.Fatalf("Respond: %v", err)
	}
	fmt.Println("Turn 1:", resp)

	// Second turn — session remembers context.
	resp, err = session.Respond(ctx, "Who created it?")
	if err != nil {
		log.Fatalf("Respond: %v", err)
	}
	fmt.Println("Turn 2:", resp)
}
