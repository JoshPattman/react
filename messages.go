package react

// A message converter is an object that reads through a conversation,
// adding messages in order, without type switching.
type MessageConverter interface {
	AddSystem(template string)
	AddUser(content string)
	AddAgent(content string)
	AddToolCalls(reasoning string, toolCalls []ToolCall)
	AddToolResponse(responses []ToolResponse)
	AddModeSwitch(mode AgentMode)
	AddNotification(kind string, content string)
	AddPersonality(personality string)
	AddSkills(skills []InsertedSkill)
	AddToolDefs(defs []AvailableToolDefinition)
}

func ConvertMessages(converter MessageConverter, messages []Message) {
	for _, m := range messages {
		m.convert(converter)
	}
}

// Message is a sum type defining the structured data that can live in agent history.
type Message interface {
	convert(MessageConverter)
}

type systemMessage struct {
	Template string
}

func (m systemMessage) convert(c MessageConverter) {
	c.AddSystem(m.Template)
}

type userMessage struct {
	Content string
}

func (m userMessage) convert(c MessageConverter) {
	c.AddUser(m.Content)
}

type agentMessage struct {
	Content string
}

func (m agentMessage) convert(c MessageConverter) {
	c.AddAgent(m.Content)
}

type toolCallsMessage struct {
	Reasoning string
	ToolCalls []ToolCall
}

func (m toolCallsMessage) convert(c MessageConverter) {
	c.AddToolCalls(m.Reasoning, m.ToolCalls)
}

type toolResponseMessage struct {
	Responses []ToolResponse
}

func (m toolResponseMessage) convert(c MessageConverter) {
	c.AddToolResponse(m.Responses)
}

type modeSwitchMessage struct {
	Mode AgentMode
}

func (m modeSwitchMessage) convert(c MessageConverter) {
	c.AddModeSwitch(m.Mode)
}

type Notification struct {
	Kind    string
	Content string
}

type notificationMessage struct {
	Notification
}

func (m notificationMessage) convert(c MessageConverter) {
	c.AddNotification(m.Kind, m.Content)
}

type personalityMessage struct {
	Personality string
}

func (m personalityMessage) convert(c MessageConverter) {
	c.AddPersonality(m.Personality)
}

type skillMessage struct {
	Skills []InsertedSkill
}

func (m skillMessage) convert(c MessageConverter) {
	c.AddSkills(m.Skills)
}

type toolsMessage struct {
	Tools []AvailableToolDefinition
}

func (m toolsMessage) convert(c MessageConverter) {
	c.AddToolDefs(m.Tools)
}

type AvailableToolDefinition struct {
	Name        string
	Description []string
}

type Skill struct {
	// A sensible snake_case key
	Key string
	// When should this be applied? If empty will always be applied.
	When string
	// The content that is not seen by the selector but is seen by the agent when chosen.
	Content string
	// How many turns after the turn it is inserted will the skill remain in context
	RemainFor int
}

type InsertedSkill struct {
	Skill
	NowRemainFor int
}

func (f Skill) IsConditional() bool { return f.When != "" }

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
