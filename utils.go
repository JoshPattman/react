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
type baseMessageConverter struct{}

func (*baseMessageConverter) AddSystem(template string)                           {}
func (*baseMessageConverter) AddUser(content string)                              {}
func (*baseMessageConverter) AddAgent(content string)                             {}
func (*baseMessageConverter) AddToolCalls(reasoning string, toolCalls []ToolCall) {}
func (*baseMessageConverter) AddToolResponse(responses []ToolResponse)            {}
func (*baseMessageConverter) AddModeSwitch(mode AgentMode)                        {}
func (*baseMessageConverter) AddNotification(kind string, content string)         {}
func (*baseMessageConverter) AddPersonality(personality string)                   {}
func (*baseMessageConverter) AddSkills(skills []InsertedSkill)                    {}
func (*baseMessageConverter) AddToolDefs(defs []AvailableToolDefinition)          {}

// An encoder that tracks current state of the agent without actually noting down messages
type currentStateMessageConverter struct {
	baseMessageConverter
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
	convertMessages(enc, msgs)
	return enc
}

func getLastInsertedSkills(msgs []Message) []InsertedSkill {
	return getCurrentState(msgs).skills
}
