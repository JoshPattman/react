package react

import (
	"encoding/json"
	"io"
)

type messageKind string

const (
	kindSystem         messageKind = "system"
	kindUser           messageKind = "user"
	kindAgent          messageKind = "agent"
	kindToolCalls      messageKind = "tool_calls"
	kindToolResponse   messageKind = "tool_response"
	kindNotification   messageKind = "notification"
	kindSkills         messageKind = "skills"
	kindAvailableTools messageKind = "available_tools"
	kindModeSwitch     messageKind = "mode_switch"
	kindPersonality    messageKind = "personality"
)

// EncodeMessages into a json format on the writer.
func EncodeMessages(w io.Writer, messages []Message) error {
	dtos := make([]messageDTO, len(messages))
	for i, m := range messages {
		dtos[i] = messageToDTO(m)
	}
	return json.NewEncoder(w).Encode(dtos)
}

// DecodeMessages from the json format (from EncodeMessages) on the writer.
func DecodeMessages(r io.Reader) ([]Message, error) {
	var dtos []messageDTO
	if err := json.NewDecoder(r).Decode(&dtos); err != nil {
		return nil, err
	}

	msgs := make([]Message, len(dtos))
	for i, d := range dtos {
		msgs[i] = dtoToMessage(d)
	}
	return msgs, nil
}

type messageDTO struct {
	Kind messageKind `json:"kind"`

	Content          string `json:"content,omitempty"`
	Reasoning        string `json:"reasoning,omitempty"`
	NotificationKind string `json:"notification_kind,omitempty"`

	ToolCalls []ToolCall     `json:"tool_calls,omitempty"`
	Responses []ToolResponse `json:"responses,omitempty"`

	Skills         []InsertedSkill           `json:"skills,omitempty"`
	AvailableTools []AvailableToolDefinition `json:"available_tools,omitempty"`
	Mode           AgentMode                 `json:"mode,omitempty"`
	Personality    string                    `json:"personality,omitempty"`
}

func messageToDTO(m Message) messageDTO {
	switch v := m.(type) {
	case SystemMessage:
		return messageDTO{
			Kind:    kindSystem,
			Content: v.Template,
		}

	case UserMessage:
		return messageDTO{
			Kind:    kindUser,
			Content: v.Content,
		}

	case AgentMessage:
		return messageDTO{
			Kind:    kindAgent,
			Content: v.Content,
		}

	case ToolCallsMessage:
		return messageDTO{
			Kind:      kindToolCalls,
			Reasoning: v.Reasoning,
			ToolCalls: v.ToolCalls,
		}

	case ToolResponseMessage:
		return messageDTO{
			Kind:      kindToolResponse,
			Responses: v.Responses,
		}

	case NotificationMessage:
		return messageDTO{
			Kind:             kindNotification,
			Content:          v.Content,
			NotificationKind: v.Kind,
		}
	case SkillMessage:
		return messageDTO{
			Kind:   kindSkills,
			Skills: v.Skills,
		}
	case ToolsMessage:
		return messageDTO{
			Kind:           kindSkills,
			AvailableTools: v.Tools,
		}
	case ModeSwitchMessage:
		return messageDTO{
			Kind: kindModeSwitch,
			Mode: v.Mode,
		}
	case PersonalityMessage:
		return messageDTO{
			Kind:        kindPersonality,
			Personality: v.Personality,
		}

	default:
		panic("unknown Message type")
	}
}

func dtoToMessage(d messageDTO) Message {
	switch d.Kind {
	case kindSystem:
		return SystemMessage{Template: d.Content}
	case kindUser:
		return UserMessage{Content: d.Content}
	case kindAgent:
		return AgentMessage{Content: d.Content}
	case kindToolCalls:
		return ToolCallsMessage{
			Reasoning: d.Reasoning,
			ToolCalls: d.ToolCalls,
		}
	case kindToolResponse:
		return ToolResponseMessage{
			Responses: d.Responses,
		}
	case kindNotification:
		return NotificationMessage{Content: d.Content, Kind: d.NotificationKind}
	case kindSkills:
		return SkillMessage{Skills: d.Skills}
	case kindAvailableTools:
		return ToolsMessage{Tools: d.AvailableTools}
	case kindModeSwitch:
		return ModeSwitchMessage{Mode: d.Mode}
	case kindPersonality:
		return PersonalityMessage{d.Personality}
	default:
		panic("unknown message kind")
	}
}
