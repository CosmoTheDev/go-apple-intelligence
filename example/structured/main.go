// structured demonstrates schema-guided generation that returns typed JSON output.
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

	// Build a schema for a recipe.
	schema := fm.NewSchema("Recipe", "A simple cooking recipe").
		AddField("title", "The recipe title", "string", false).
		AddField("description", "A one-sentence description", "string", false).
		AddField("prepTimeMinutes", "Preparation time in minutes", "integer", false)

	session, err := fm.NewSession(fm.SessionOptions{})
	if err != nil {
		log.Fatalf("NewSession: %v", err)
	}

	result, err := session.RespondStructured(
		context.Background(),
		"Give me a quick pasta recipe.",
		schema,
	)
	if err != nil {
		log.Fatalf("RespondStructured: %v", err)
	}

	title, _ := result.Get("title")
	desc, _ := result.Get("description")
	prep, _ := result.Get("prepTimeMinutes")

	fmt.Printf("Title:    %s\n", title)
	fmt.Printf("Desc:     %s\n", desc)
	fmt.Printf("Prep:     %s min\n", prep)
	fmt.Printf("Full JSON: %s\n", result.JSON())
}
