package react

// Tool is a runnable object that can be both described to and called by an agent.
type Tool interface {
	// The name of the tool to be used by the agent. Should probably be snake_case.
	Name() string
	// Describe the tool in short bullet points to the agent.
	// Includes which parameters to provide.
	// Parameters can be described as any json type.
	Description() []string
	// Call the tool, providing a formatted response or an error if the tool call failed.
	// The values of the args will be the direct result of json decoding the tool call args.
	Call(map[string]any) (string, error)
}
