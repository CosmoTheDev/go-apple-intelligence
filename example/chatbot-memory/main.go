// chatbot-memory is an interactive chatbot that persists conversation summaries
// across sessions. Memory is saved to ~/.apple-intelligence-memory.json.
//
// Commands:
//
//	"reset"   — clear current conversation history (memory is preserved)
//	"forget"  — wipe all saved memory
//	"memory"  — print what has been remembered
//	"quit"    — exit
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	fm "github.com/CosmoTheDev/go-apple-intelligence/fm"
)

const maxMemoryEntries = 20

type memoryEntry struct {
	User string `json:"user"`
	AI   string `json:"ai"`
	At   string `json:"at"`
}

type memoryStore struct {
	Entries []memoryEntry `json:"entries"`
}

func memoryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".apple-intelligence-memory.json")
}

func loadMemory() memoryStore {
	data, err := os.ReadFile(memoryPath())
	if err != nil {
		return memoryStore{}
	}
	var store memoryStore
	_ = json.Unmarshal(data, &store)
	return store
}

func saveMemory(store memoryStore) {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(memoryPath(), data, 0600)
}

func buildInstructions(store memoryStore) string {
	base := "You are a helpful, concise assistant with memory of past conversations."
	if len(store.Entries) == 0 {
		return base
	}
	var sb strings.Builder
	sb.WriteString(base)
	sb.WriteString("\n\nPrevious conversation history (most recent last):\n")
	// Include up to the last 10 entries as context.
	start := 0
	if len(store.Entries) > 10 {
		start = len(store.Entries) - 10
	}
	for _, e := range store.Entries[start:] {
		sb.WriteString(fmt.Sprintf("User: %s\nAI: %s\n\n", e.User, e.AI))
	}
	return sb.String()
}

func main() {
	model := fm.DefaultModel()
	if ok, err := model.IsAvailable(); !ok {
		log.Fatalf("Apple Intelligence unavailable: %v", err)
	}

	store := loadMemory()

	session, err := fm.NewSession(fm.SessionOptions{
		Instructions: buildInstructions(store),
	})
	if err != nil {
		log.Fatalf("NewSession: %v", err)
	}

	fmt.Println("Apple Intelligence Chatbot  (with memory)")
	fmt.Printf("Memory file: %s\n", memoryPath())
	fmt.Printf("Remembered exchanges: %d\n", len(store.Entries))
	fmt.Println("Commands: \"reset\", \"forget\", \"memory\", \"quit\"")
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

		switch input {
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		case "reset":
			session.Reset()
			fmt.Println("(conversation history cleared — memory preserved)")
			continue
		case "forget":
			store = memoryStore{}
			saveMemory(store)
			// Rebuild session without old memory.
			session, err = fm.NewSession(fm.SessionOptions{
				Instructions: buildInstructions(store),
			})
			if err != nil {
				log.Fatalf("NewSession: %v", err)
			}
			fmt.Println("(all memory wiped)")
			continue
		case "memory":
			if len(store.Entries) == 0 {
				fmt.Println("(no memory saved yet)")
			} else {
				fmt.Printf("(%d exchanges remembered)\n", len(store.Entries))
				for i, e := range store.Entries {
					fmt.Printf("  [%d] %s  You: %s\n        AI: %s\n", i+1, e.At, e.User, e.AI)
				}
			}
			continue
		}

		var reply strings.Builder
		fmt.Print("AI:  ")
		err := session.StreamResponse(context.Background(), input, func(chunk string, done bool) {
			if !done {
				fmt.Print(chunk)
				reply.WriteString(chunk)
			} else {
				fmt.Println()
			}
		})
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		// Save exchange to memory.
		store.Entries = append(store.Entries, memoryEntry{
			User: input,
			AI:   strings.TrimSpace(reply.String()),
			At:   time.Now().Format("2006-01-02 15:04"),
		})
		if len(store.Entries) > maxMemoryEntries {
			store.Entries = store.Entries[len(store.Entries)-maxMemoryEntries:]
		}
		saveMemory(store)

		// Show context length.
		if transcript, err := session.TranscriptJSON(); err == nil {
			chars := len(transcript)
			fmt.Printf("     \033[2m[context: ~%d chars / ~%d tokens | memory: %d exchanges]\033[0m\n",
				chars, chars/4, len(store.Entries))
		}
	}
}
