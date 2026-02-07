package react

import (
	"bytes"
	"encoding/json"
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
	converter := &jpfMessageConverter{}
	ConvertMessages(converter, msgs)
	return converter.Messages(), nil
}

func toolCallsMessageFromResponse(response reasonResponse) toolCallsMessage {
	finalMessage := toolCallsMessage{
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

func responseFromToolCallsMessage(reasoning string, calls []ToolCall) reasonResponse {
	finalMessage := reasonResponse{
		Reasoning: reasoning,
	}
	for _, tc := range calls {
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

type jpfMessageConverter struct {
	systemTemplate string
	personality    string
	skills         []InsertedSkill
	activeMessages []jpf.Message
}

func (c *jpfMessageConverter) Messages() []jpf.Message {
	prompt := executeSystemPrompt(c.systemTemplate, systemPromptData{
		c.personality,
		c.skills,
	})
	return []jpf.Message{
		{
			Role:    jpf.SystemRole,
			Content: prompt,
		},
	}
}

func (conv *jpfMessageConverter) AddSystem(template string) {
	conv.systemTemplate = template
}
func (conv *jpfMessageConverter) AddUser(content string) {
	conv.activeMessages = append(conv.activeMessages, jpf.Message{
		Role:    jpf.UserRole,
		Content: content,
	})
}
func (conv *jpfMessageConverter) AddAgent(content string) {
	conv.activeMessages = append(conv.activeMessages, jpf.Message{
		Role:    jpf.AssistantRole,
		Content: content,
	})
}
func (conv *jpfMessageConverter) AddToolCalls(reasoning string, toolCalls []ToolCall) {
	resp := responseFromToolCallsMessage(reasoning, toolCalls)
	content, _ := json.MarshalIndent(resp, "", "    ")
	conv.activeMessages = append(conv.activeMessages, jpf.Message{
		Role:    jpf.AssistantRole,
		Content: string(content),
	})
}
func (conv *jpfMessageConverter) AddToolResponse(responses []ToolResponse) {
	results := make([]string, len(responses))
	for i, r := range responses {
		results[i] = r.Response
	}
	toolSep := "\n==========\n"
	conv.activeMessages = append(conv.activeMessages, jpf.Message{
		Role:    jpf.SystemRole,
		Content: "Tool Responses:" + toolSep + strings.Join(results, toolSep),
	})
}
func (conv *jpfMessageConverter) AddModeSwitch(mode AgentMode) {
	var resultMsg jpf.Message
	switch mode {
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
		return
	}
	conv.activeMessages = append(conv.activeMessages, resultMsg)
}
func (conv *jpfMessageConverter) AddNotification(kind string, content string) {
	conv.activeMessages = append(conv.activeMessages, jpf.Message{
		Role:    jpf.SystemRole,
		Content: fmt.Sprintf("**Notification of type '%s'**\n%s", kind, content),
	})
}
func (conv *jpfMessageConverter) AddPersonality(personality string) {
	conv.personality = personality
}
func (conv *jpfMessageConverter) AddSkills(skills []InsertedSkill) {
	conv.skills = skills
}
func (conv *jpfMessageConverter) AddToolDefs(defs []AvailableToolDefinition) {
	var resultMsg jpf.Message
	if len(defs) == 0 {
		resultMsg = jpf.Message{
			Role:    jpf.SystemRole,
			Content: "The available tools have changed, there are now no tools available.",
		}
	} else {
		tools := make([]string, len(defs))
		for i, t := range defs {
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
	conv.activeMessages = append(conv.activeMessages, resultMsg)
}
