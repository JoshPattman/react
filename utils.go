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
	var lastToolMessage *AvailableToolDefinitionsMessage
	for _, h := range history {
		h, ok := h.(AvailableToolDefinitionsMessage)
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
