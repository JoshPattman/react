package react

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/JoshPattman/jpf"
)

// Create the pipeline to get a react step given the messages
func getAgentReActPipeline(model jpf.Model) jpf.Pipeline[[]Message, reasonResponse] {
	enc := getAgentEncoder()
	dec := getAgentReActDecoder()
	return jpf.NewOneShotPipeline(enc, dec, nil, model)
}

// Create the pipeline to create a final response given the messages
func getAgentFinalAnswerPipeline(model jpf.Model) jpf.Pipeline[[]Message, string] {
	enc := getAgentEncoder()
	dec := getAgentAnswerDecoder()
	return jpf.NewOneShotPipeline(enc, dec, nil, model)
}

func getAgentEncoder() jpf.Encoder[[]Message] {
	return &messagesEncoder{}
}

func getAgentReActDecoder() jpf.Parser[reasonResponse] {
	return jpf.NewJsonParser[reasonResponse]()
}

func getAgentAnswerDecoder() jpf.Parser[string] {
	return jpf.NewStringParser()
}

type toolArg struct {
	ArgName  string `json:"arg_name"`
	ArgValue any    `json:"arg_value"`
}

type toolCall struct {
	ToolName string    `json:"tool_name"`
	ToolArgs []toolArg `json:"tool_args"`
}

type reasonResponse struct {
	Reasoning string     `json:"reasoning"`
	ToolCalls []toolCall `json:"tool_calls"`
}

type messagesEncoder struct{}

func (m *messagesEncoder) BuildInputMessages(msgs []Message) ([]jpf.Message, error) {
	result := make([]jpf.Message, 0)
	for _, msg := range msgs {
		var resultMsg jpf.Message
		var dontMessage bool
		switch msg := msg.(type) {
		case SystemMessage:
			personality := getLastPersonality(msgs)
			skills := getLastInsertedSkills(msgs)
			prompt := executeSystemPrompt(msg.Template, systemPromptData{
				personality,
				skills,
			})
			resultMsg = jpf.Message{
				Role:    jpf.SystemRole,
				Content: prompt,
			}
		case UserMessage:
			resultMsg = jpf.Message{
				Role:    jpf.UserRole,
				Content: msg.Content,
			}
		case AgentMessage:
			resultMsg = jpf.Message{
				Role:    jpf.AssistantRole,
				Content: msg.Content,
			}
		case ToolCallsMessage:
			resp := responseFromToolCallsMessage(msg)
			content, _ := json.MarshalIndent(resp, "", "    ")
			resultMsg = jpf.Message{
				Role:    jpf.AssistantRole,
				Content: string(content),
			}
		case ToolResponseMessage:
			results := make([]string, len(msg.Responses))
			for i, r := range msg.Responses {
				results[i] = r.Response
			}
			toolSep := "\n==========\n"
			resultMsg = jpf.Message{
				Role:    jpf.SystemRole,
				Content: "Tool Responses:" + toolSep + strings.Join(results, toolSep),
			}
		case ModeSwitchMessage:
			switch msg.Mode {
			case ModeReasonAct:
				resultMsg = jpf.Message{
					Role:    jpf.SystemRole,
					Content: "**Mode Change**\nYou are now in reason-action mode. Use the reason-action json format when answering questions here. The user will not see the followin responses.",
				}
			case ModeAnswerUser:
				resultMsg = jpf.Message{
					Role:    jpf.SystemRole,
					Content: "**Mode Change**\nYou are now in final answer mode. Your full response will be show to the user. You can respond in any format.",
				}
			default:
				// We don't notify the agent about other mode switches
				dontMessage = true
			}
		case NotificationMessage:
			resultMsg = jpf.Message{
				Role:    jpf.SystemRole,
				Content: fmt.Sprintf("**Notification of type '%s'**\n%s", msg.Kind, msg.Content),
			}
		case SkillMessage:
			dontMessage = true // Skills are handled by the system prompt
		case PersonalityMessage:
			dontMessage = true
		case ToolsMessage:
			if len(msg.Tools) == 0 {
				resultMsg = jpf.Message{
					Role:    jpf.SystemRole,
					Content: "The available tools have changed, there are now no tools available.",
				}
			} else {
				tools := make([]string, len(msg.Tools))
				for i, t := range msg.Tools {
					s := fmt.Sprintf("- Tool `%s`", t.Name)
					for _, d := range t.Description {
						s += fmt.Sprintf("\n  - %s", d)
					}
					tools[i] = s
				}
				resultMsg = jpf.Message{
					Role:    jpf.SystemRole,
					Content: "The available tools have changed, here are the current available tools:\n" + strings.Join(tools, "\n"),
				}
			}

		default:
			return nil, errors.New("unknown message type")
		}
		if !dontMessage {
			result = append(result, resultMsg)
		}
	}
	return result, nil
}

func toolCallsMessageFromResponse(response reasonResponse) ToolCallsMessage {
	finalMessage := ToolCallsMessage{
		Reasoning: response.Reasoning,
	}
	for _, tc := range response.ToolCalls {
		args := make([]ToolCallArg, 0)
		for _, a := range tc.ToolArgs {
			args = append(args, ToolCallArg(a))
		}
		finalMessage.ToolCalls = append(finalMessage.ToolCalls, ToolCall{
			ToolName: tc.ToolName,
			ToolArgs: args,
		})
	}
	return finalMessage
}

func responseFromToolCallsMessage(msg ToolCallsMessage) reasonResponse {
	finalMessage := reasonResponse{
		Reasoning: msg.Reasoning,
	}
	for _, tc := range msg.ToolCalls {
		args := make([]toolArg, 0)
		for _, a := range tc.ToolArgs {
			args = append(args, toolArg(a))
		}
		finalMessage.ToolCalls = append(finalMessage.ToolCalls, toolCall{
			ToolName: tc.ToolName,
			ToolArgs: args,
		})
	}
	return finalMessage
}

type systemPromptData struct {
	Personality string
	Skills      []InsertedSkill
}

func executeSystemPrompt(temp string, data systemPromptData) string {
	tmp := template.Must(template.New("system").Parse(temp))
	result := bytes.NewBuffer(nil)
	err := tmp.Execute(result, data)
	if err != nil {
		panic(err)
	}
	return result.String()
}
