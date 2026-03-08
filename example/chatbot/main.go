// chatbot is an interactive multi-turn chatbot powered by Apple Intelligence.
// Type a message and press Enter to chat. Type "quit" or press Ctrl-C to exit.
// Type "reset" to clear the conversation history.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	fm "github.com/CosmoTheDev/go-apple-intelligence/fm"
)

func main() {
	model := fm.DefaultModel()
	if ok, err := model.IsAvailable(); !ok {
		log.Fatalf("Apple Intelligence unavailable: %v", err)
	}

	session, err := fm.NewSession(fm.SessionOptions{
		Instructions: "You are a helpful, concise assistant.",
	})
	if err != nil {
		log.Fatalf("NewSession: %v", err)
	}

	fmt.Println("Apple Intelligence Chatbot")
	fmt.Println("Commands: \"reset\" to clear history, \"quit\" or Ctrl-C to exit")
	fmt.Println(strings.Repeat("─", 40))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			fmt.Println("Goodbye!")
			break
		}
		if input == "reset" {
			session.Reset()
			fmt.Println("(conversation history cleared)")
			continue
		}

		fmt.Print("AI:  ")
		err := session.StreamResponse(context.Background(), input, func(chunk string, done bool) {
			if !done {
				fmt.Print(chunk)
			} else {
				fmt.Println()
			}
		})
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		if transcript, err := session.TranscriptJSON(); err == nil {
			chars := len(transcript)
			fmt.Printf("     \033[2m[context: ~%d chars / ~%d tokens]\033[0m\n", chars, chars/4)
		}
	}
}
