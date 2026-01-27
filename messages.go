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
func (ModeSwitchMessage) message()               {}

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

type ModeSwitchMessage struct {
	Mode AgentMode
}

type NotificationMessage struct {
	Kind    string
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
	// A sensible snake_case key
	Key string
	// When should this be applied? If empty will always be applied.
	When string
	// The content that is not seen by the selector but is seen by the agent when chosen.
	Content string
}

func (f PromptFragment) IsConditional() bool { return f.When != "" }

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

type AgentMode uint8

const (
	ModeCollectContext = iota
	ModeReasonAct
	ModeAnswerUser
)
