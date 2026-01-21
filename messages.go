package react

// Message is a sum type defining the structured data that can live in agent history.
type Message interface {
	message()
}

func (SystemMessage) message()                   {}
func (UserMessage) message()                     {}
func (AgentMessage) message()                    {}
func (ToolCallsMessage) message()                {}
func (ToolResponseMessage) message()             {}
func (NotificationMessage) message()             {}
func (PromptFragmentMessage) message()           {}
func (AvailableToolDefinitionsMessage) message() {}

type SystemMessage struct {
	Content string
}

type UserMessage struct {
	Content string
}

type AgentMessage struct {
	Content string
}

type ToolCallsMessage struct {
	Reasoning string
	ToolCalls []ToolCall
}

type ToolResponseMessage struct {
	Responses []ToolResponse
}

type NotificationMessage struct {
	Content string
}

type PromptFragmentMessage struct {
	Fragments []PromptFragment
}

type AvailableToolDefinitionsMessage struct {
	Tools []AvailableToolDefinition
}

type AvailableToolDefinition struct {
	Name        string
	Description []string
}

type PromptFragment struct {
	Key     string
	When    string
	Content string
}

type ToolCallArg struct {
	ArgName  string
	ArgValue any
}

type ToolCall struct {
	ToolName string
	ToolArgs []ToolCallArg
}

type ToolResponse struct {
	Response string
}
