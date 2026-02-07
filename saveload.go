package react

type SerialisedMessageKind string

const (
	KindSystem         SerialisedMessageKind = "system"
	KindUser           SerialisedMessageKind = "user"
	KindAgent          SerialisedMessageKind = "agent"
	KindToolCalls      SerialisedMessageKind = "tool_calls"
	KindToolResponse   SerialisedMessageKind = "tool_response"
	KindNotification   SerialisedMessageKind = "notification"
	KindSkills         SerialisedMessageKind = "skills"
	KindAvailableTools SerialisedMessageKind = "available_tools"
	KindModeSwitch     SerialisedMessageKind = "mode_switch"
	KindPersonality    SerialisedMessageKind = "personality"
)

type SerialisedMessage struct {
	Kind             SerialisedMessageKind     `json:"kind"`
	Content          string                    `json:"content,omitempty"`
	Reasoning        string                    `json:"reasoning,omitempty"`
	NotificationKind string                    `json:"notification_kind,omitempty"`
	ToolCalls        []ToolCall                `json:"tool_calls,omitempty"`
	Responses        []ToolResponse            `json:"responses,omitempty"`
	Skills           []InsertedSkill           `json:"skills,omitempty"`
	AvailableTools   []AvailableToolDefinition `json:"available_tools,omitempty"`
	Mode             AgentMode                 `json:"mode,omitempty"`
	Personality      string                    `json:"personality,omitempty"`
}

func SerialiseMessages(msgs []Message) []SerialisedMessage {
	converter := &serialisingConverter{}
	convertMessages(converter, msgs)
	return converter.out
}

func DeserialiseMessages(smsgs []SerialisedMessage) []Message {
	msgs := make([]Message, len(smsgs))
	for i, sm := range smsgs {
		msgs[i] = dtoToMessage(sm)
	}
	return msgs
}

func dtoToMessage(d SerialisedMessage) Message {
	switch d.Kind {
	case KindSystem:
		return systemMessage{Template: d.Content}
	case KindUser:
		return userMessage{Content: d.Content}
	case KindAgent:
		return agentMessage{Content: d.Content}
	case KindToolCalls:
		return toolCallsMessage{
			Reasoning: d.Reasoning,
			ToolCalls: d.ToolCalls,
		}
	case KindToolResponse:
		return toolResponseMessage{
			Responses: d.Responses,
		}
	case KindNotification:
		return notificationMessage{Notification{Content: d.Content, Kind: d.NotificationKind}}
	case KindSkills:
		return skillMessage{Skills: d.Skills}
	case KindAvailableTools:
		return toolsMessage{Tools: d.AvailableTools}
	case KindModeSwitch:
		return modeSwitchMessage{Mode: d.Mode}
	case KindPersonality:
		return personalityMessage{d.Personality}
	default:
		panic("unknown message kind")
	}
}

type serialisingConverter struct {
	out []SerialisedMessage
}

func (c *serialisingConverter) AddSystem(template string) {
	c.out = append(c.out, SerialisedMessage{
		Kind:    KindSystem,
		Content: template,
	})
}

func (c *serialisingConverter) AddUser(content string) {
	c.out = append(c.out, SerialisedMessage{
		Kind:    KindUser,
		Content: content,
	})
}

func (c *serialisingConverter) AddAgent(content string) {
	c.out = append(c.out, SerialisedMessage{
		Kind:    KindAgent,
		Content: content,
	})
}

func (c *serialisingConverter) AddToolCalls(reasoning string, toolCalls []ToolCall) {
	c.out = append(c.out, SerialisedMessage{
		Kind:      KindToolCalls,
		Reasoning: reasoning,
		ToolCalls: toolCalls,
	})
}

func (c *serialisingConverter) AddToolResponse(responses []ToolResponse) {
	c.out = append(c.out, SerialisedMessage{
		Kind:      KindToolResponse,
		Responses: responses,
	})
}

func (c *serialisingConverter) AddModeSwitch(mode AgentMode) {
	c.out = append(c.out, SerialisedMessage{
		Kind: KindModeSwitch,
		Mode: mode,
	})
}

func (c *serialisingConverter) AddNotification(kind string, content string) {
	c.out = append(c.out, SerialisedMessage{
		Kind:             KindNotification,
		Content:          content,
		NotificationKind: kind,
	})
}

func (c *serialisingConverter) AddPersonality(personality string) {
	c.out = append(c.out, SerialisedMessage{
		Kind:        KindPersonality,
		Personality: personality,
	})
}

func (c *serialisingConverter) AddSkills(skills []InsertedSkill) {
	c.out = append(c.out, SerialisedMessage{
		Kind:   KindSkills,
		Skills: skills,
	})
}

func (c *serialisingConverter) AddToolDefs(defs []AvailableToolDefinition) {
	c.out = append(c.out, SerialisedMessage{
		Kind:           KindAvailableTools,
		AvailableTools: defs,
	})
}
