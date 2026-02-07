package react

import "slices"

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
	var lastToolMessage *ToolsMessage
	for _, h := range history {
		h, ok := h.(ToolsMessage)
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

func getLastInsertedSkills(msgs []Message) []InsertedSkill {
	for i := len(msgs) - 1; i >= 0; i-- {
		msg, ok := msgs[i].(SkillMessage)
		if !ok {
			continue
		}
		return slices.Clone(msg.Skills)
	}
	return nil
}

func getLastPersonality(msgs []Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		msg, ok := msgs[i].(PersonalityMessage)
		if !ok {
			continue
		}
		return msg.Personality
	}
	return ""
}
