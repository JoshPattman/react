package react

import (
	"slices"
)

func getToolDefs(tools []Tool) []AvailableToolDefinition {
	defs := make([]AvailableToolDefinition, len(tools))
	for i, t := range tools {
		defs[i] = AvailableToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
		}
	}
	return defs
}

func toolsHaveChanged(history []Message, tools []Tool) bool {
	var lastToolMessage *toolsMessage
	for _, h := range history {
		h, ok := h.(toolsMessage)
		if !ok {
			continue
		}
		lastToolMessage = &h
	}
	if lastToolMessage == nil {
		return true
	}
	newToolNames := make([]string, len(tools))
	for i, t := range tools {
		newToolNames[i] = t.Name()
	}
	if len(newToolNames) != len(lastToolMessage.Tools) {
		return true
	}
	for _, t := range lastToolMessage.Tools {
		if !slices.Contains(newToolNames, t.Name) {
			return true
		}
	}
	return false
}

// Compose on this to get no-op behaviour for all message types
type BaseMessageConverter struct{}

func (*BaseMessageConverter) AddSystem(template string)                           {}
func (*BaseMessageConverter) AddUser(content string)                              {}
func (*BaseMessageConverter) AddAgent(content string)                             {}
func (*BaseMessageConverter) AddToolCalls(reasoning string, toolCalls []ToolCall) {}
func (*BaseMessageConverter) AddToolResponse(responses []ToolResponse)            {}
func (*BaseMessageConverter) AddModeSwitch(mode AgentMode)                        {}
func (*BaseMessageConverter) AddNotification(kind string, content string)         {}
func (*BaseMessageConverter) AddPersonality(personality string)                   {}
func (*BaseMessageConverter) AddSkills(skills []InsertedSkill)                    {}
func (*BaseMessageConverter) AddToolDefs(defs []AvailableToolDefinition)          {}

// An encoder that tracks current state of the agent without actually noting down messages
type currentStateMessageConverter struct {
	BaseMessageConverter
	systemTemplate string
	personality    string
	skills         []InsertedSkill
	toolDefs       []AvailableToolDefinition
	mode           AgentMode
}

func (conv *currentStateMessageConverter) AddSystem(template string) {
	conv.systemTemplate = template
}
func (conv *currentStateMessageConverter) AddModeSwitch(mode AgentMode) {
	conv.mode = mode
}
func (conv *currentStateMessageConverter) AddPersonality(personality string) {
	conv.personality = personality
}
func (conv *currentStateMessageConverter) AddSkills(skills []InsertedSkill) {
	conv.skills = skills
}
func (conv *currentStateMessageConverter) AddToolDefs(defs []AvailableToolDefinition) {
	conv.toolDefs = defs
}

func getCurrentState(msgs []Message) *currentStateMessageConverter {
	enc := &currentStateMessageConverter{}
	ConvertMessages(enc, msgs)
	return enc
}

func getLastInsertedSkills(msgs []Message) []InsertedSkill {
	return getCurrentState(msgs).skills
}
