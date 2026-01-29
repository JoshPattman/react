# `react` - Simple yet powerful agent framework for Go
- A ReAct agent is a class of AI agent that is capable of advanced tasks through tool calling.
- This package implements a version of a ReAct agent in Go, with a few extra features.

## Features
- **Backed by the power of [jpf](https://github.com/JoshPattman/jpf)**: The LLM handling is extremely robust and model-agnostic.
- **Easy to create custom tools**: To create a tool, you just need to implement a simple interface.
- **Built-in dynamic skill retrieval**: Named `PromptFragment` in this package, you can add skills to the agent and they will be intelligently injected into its context when relevant.
- **Supports Streaming**: You can inject callbacks for both streaming back the text of the final response, and streaming back messages as they are created.
- **Clean message abstraction**: The agent history is stored as a typed list of messages, differentiated by use. These only get converted into the API messages right before sending to the LLM.

## Usage
- Create an agent and ask it a question:
```go
agent := NewCraig(
    modelBuilder,
    WithCraigTools(tool1, tool2),
    WithCraigPersonality("You are an ai assistant"),
    WithCraigFragments(PromptFragment{
        Key: "time_tool",
        When: "The user asks the agent for the time or date",
        Content: "When asked for the time, the agent will respond in HH:MM 24hr format. When asked the date, DD/MM/YYYY.",
    }),
)

response, err := agent.Send("Whats the time")
```

- Stream the response back to the terminal

```go
type printStreamer struct{}
func (*printStreamer) TrySendTextChunk(chunk string) {fmt.Print(chunk)}

response, err := agent.Send("Write 3 haikus", WithResponseStreamer(&printStreamer{}))
// The response will be printed to the terminal as the API streams it back
```

- Agents use model builders, which are the method of providing the agent with the llm to use

```go
type ModelBuilder interface {
	AgentModelBuilder
	FragmentSelectorModelBuilder
}

type AgentModelBuilder interface {
	// Build a model for the agent.
	// A struct may be passed as the response type for a json schema, or it may be nil.
	// The stream callbacks may also be nil.
	BuildAgentModel(responseType any, onInitFinalStream func(), onDataFinalStream func(string)) jpf.Model
}

type FragmentSelectorModelBuilder interface {
	// Build a model for a fragment selector.
	// A struct may be passed as the response type for a json schema, or it may be nil.
	BuildFragmentSelectorModel(responseType any) jpf.Model
}
```