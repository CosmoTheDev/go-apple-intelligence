// tools demonstrates function/tool calling: the model can invoke a Go function
// during generation and incorporate its result into the response.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	fm "github.com/CosmoTheDev/go-apple-intelligence/fm"
)

func main() {
	model := fm.DefaultModel()
	if ok, err := model.IsAvailable(); !ok {
		log.Fatalf("Apple Intelligence unavailable: %v", err)
	}

	// Define the tool's parameter schema.
	params := fm.NewSchema("WeatherParams", "Parameters for the weather tool").
		AddField("city", "The city name to look up", "string", false)

	// Create the tool with a Go handler.
	weatherTool := fm.NewTool(
		"get_weather",
		"Returns the current weather for a given city",
		params,
		func(argsJSON string) string {
			var args map[string]any
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return `{"error": "invalid args"}`
			}
			city, _ := args["city"].(string)
			// Simulated weather data.
			weather := map[string]any{
				"city":        strings.Title(city),
				"temperature": 72,
				"condition":   "sunny",
			}
			out, _ := json.Marshal(weather)
			return string(out)
		},
	)

	session, err := fm.NewSession(fm.SessionOptions{
		Instructions: "You are a helpful assistant with access to real-time weather data.",
		Tools:        []*fm.Tool{weatherTool},
	})
	if err != nil {
		log.Fatalf("NewSession: %v", err)
	}

	resp, err := session.Respond(
		context.Background(),
		"What's the weather like in San Francisco right now?",
	)
	if err != nil {
		log.Fatalf("Respond: %v", err)
	}
	fmt.Println(resp)
}
