package fm

// Tool defines a callable function that the model can invoke during generation.
// Tools are lightweight Go-side definitions; the C bridge ref is created when
// the tool is attached to a Session.
type Tool struct {
	Name        string
	Description string
	// Params describes the input parameters the model must supply when calling
	// this tool. Build one with NewSchema(...).AddField(...).
	Params  *Schema
	Handler func(argsJSON string) string
}

// NewTool creates a tool definition.
//
// name        — machine-readable identifier (no spaces)
// description — natural-language description shown to the model
// params      — schema describing the tool's input parameters
// handler     — invoked when the model calls the tool; receives args as a JSON
//
//	string and must return a JSON string result
func NewTool(name, description string, params *Schema, handler func(argsJSON string) string) *Tool {
	return &Tool{
		Name:        name,
		Description: description,
		Params:      params,
		Handler:     handler,
	}
}
