package react

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"
	"text/template"

	"github.com/JoshPattman/jpf"
)

//go:embed system.tpl
var craigSystemPrompt string

// Create a new CRAIG agent with the given tools.
func NewCraig(mb ModelBuilder, opts ...CraigOpt) Agent {
	kwargs := getCraigKwargs(opts)
	messages := []Message{
		SystemMessage{
			Content: createCraigSystemMessage(kwargs.personality),
		},
	}
	return newCraigHelper(mb, messages, kwargs)
}

// Create a new CRAIG agent from the given state (from a previous CRAIG agent).
// The tools provided should match the tools originally provided to NewCraig (in schema).
func NewCraigFromSaved(mb ModelBuilder, messages []Message, opts ...CraigOpt) Agent {
	return newCraigHelper(mb, messages, getCraigKwargs(opts))
}

func getCraigKwargs(opts []CraigOpt) craigKwargs {
	kwargs := craigKwargs{
		personality: "Your name is CRAIG, a helpful assistant.",
	}
	for _, o := range opts {
		o(&kwargs)
	}
	return kwargs
}

func newCraigHelper(mb ModelBuilder, messages []Message, kwargs craigKwargs) Agent {

	ag := &craig{
		messages:         messages,
		modelBuilder:     mb,
		tools:            kwargs.tools,
		allFragments:     kwargs.fragments,
		fragmentSelector: NewFragmentSelector(mb, 10),
	}
	return ag
}

type CraigOpt func(*craigKwargs)

func WithCraigFragments(frags ...PromptFragment) func(kw *craigKwargs) {
	return func(kw *craigKwargs) { kw.fragments = append(kw.fragments, frags...) }
}

func WithCraigTools(tools ...Tool) func(kw *craigKwargs) {
	return func(kw *craigKwargs) { kw.tools = append(kw.tools, tools...) }
}

func WithCraigPersonality(personality string) func(kw *craigKwargs) {
	return func(kw *craigKwargs) { kw.personality = personality }
}

type craigKwargs struct {
	fragments   []PromptFragment
	tools       []Tool
	personality string
}

func createCraigSystemMessage(personality string) string {
	tmp := template.Must(template.New("system").Parse(craigSystemPrompt))
	result := bytes.NewBuffer(nil)
	err := tmp.Execute(result, struct {
		Personality string
	}{personality})
	if err != nil {
		panic(err)
	}
	return result.String()
}

type craig struct {
	messages         []Message
	modelBuilder     AgentModelBuilder
	tools            []Tool
	fragmentSelector FragmentSelector
	allFragments     []PromptFragment
}

func (ag *craig) Send(msg string, opts ...SendMessageOpt) (string, error) {
	streamers := getStreamers(opts)

	// Update tool message if tools have changed
	if toolsHaveChanged(ag.messages, ag.tools) {
		ag.addMessages(streamers, AvailableToolDefinitionsMessage{
			Tools: getToolDefs(ag.tools),
		})
	}

	// Update notifications
	for _, msg := range streamers.notifications {
		ag.addMessages(streamers, msg)
	}

	// Add the initial user message
	ag.addMessages(streamers, UserMessage{msg})

	// Signal we are collecting context and add any relevant fragments
	if len(ag.allFragments) > 0 {
		ag.addMessages(streamers, ModeSwitchMessage{ModeCollectContext})
		fragments, err := ag.fragmentSelector.SelectFragments(ag.allFragments, ag.messages)
		if err != nil {
			return "", err
		}
		if len(fragments) > 0 {
			ag.addMessages(streamers, PromptFragmentMessage{fragments})
		}
	}

	ag.addMessages(streamers, ModeSwitchMessage{ModeReasonAct})

	// React loop
	for {
		// Ask agent for any new tool calls and break if there are no calls
		toolCalls, err := ag.getToolCalls()
		if err != nil {
			return "", err
		}
		ag.addMessages(streamers, toolCalls)
		if len(toolCalls.ToolCalls) == 0 {
			break
		}
		// Execute tool calls
		toolResults := ag.executeToolCalls(toolCalls.ToolCalls)
		ag.addMessages(streamers, ToolResponseMessage{toolResults})
	}

	// Set the agent to final answer mode and get the response
	ag.addMessages(streamers, ModeSwitchMessage{ModeAnswerUser})
	finalResp, err := ag.getFinalResponse(streamers)
	if err != nil {
		return "", err
	}
	ag.addMessages(streamers, AgentMessage{finalResp})
	return finalResp, nil
}

func (ag *craig) Messages() iter.Seq[Message] {
	return slices.Values(ag.messages)
}

func (ag *craig) addMessages(msgStreamer MessageStreamer, msgs ...Message) {
	ag.messages = append(ag.messages, msgs...)
	if msgStreamer != nil {
		for _, msg := range msgs {
			msgStreamer.TrySendMessage(msg)
		}
	}
}

func (ag *craig) executeToolCalls(calls []ToolCall) []ToolResponse {
	results := make([]ToolResponse, 0)
	for _, call := range calls {
		tool := ag.findToolByName(call.ToolName)
		if tool == nil {
			results = append(results, ToolResponse{fmt.Sprintf("Could not find tool. with name '%s'", call.ToolName)})
			continue
		}
		args := make(map[string]any)
		for _, arg := range call.ToolArgs {
			args[arg.ArgName] = arg.ArgValue
		}
		result, err := tool.Call(args)
		if err != nil {
			results = append(results, ToolResponse{fmt.Sprintf("There was an error calling the tool: %v", err)})
			continue
		}
		results = append(results, ToolResponse{result})
	}
	return results
}

func (ag *craig) findToolByName(toolName string) Tool {
	for _, t := range ag.tools {
		if t.Name() == toolName {
			return t
		}
	}
	return nil
}

func (ag *craig) getToolCalls() (ToolCallsMessage, error) {
	model := ag.modelBuilder.BuildAgentModel(reasonResponse{}, nil, nil)
	enc := ag.getEncoder()
	dec := jpf.NewJsonResponseDecoder[[]Message, reasonResponse]()
	mf := jpf.NewOneShotMapFunc(enc, dec, model)
	result, _, err := mf.Call(context.Background(), ag.messages)
	if err != nil {
		return ToolCallsMessage{}, err
	}
	return toolCallsMessageFromResponse(result), nil
}

func (ag *craig) getFinalResponse(streamer TextStreamer) (string, error) {
	model := ag.modelBuilder.BuildAgentModel(nil, nil, streamer.TrySendTextChunk)
	enc := ag.getEncoder()
	dec := jpf.NewRawStringResponseDecoder[[]Message]()
	mf := jpf.NewOneShotMapFunc(enc, dec, model)
	result, _, err := mf.Call(context.Background(), ag.messages)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (ag *craig) getEncoder() jpf.MessageEncoder[[]Message] {
	return &messagesEncoder{}
}

type messagesEncoder struct{}

func (m *messagesEncoder) BuildInputMessages(msgs []Message) ([]jpf.Message, error) {
	result := make([]jpf.Message, 0)
	for _, msg := range msgs {
		var resultMsg jpf.Message
		var dontMessage bool
		switch msg := msg.(type) {
		case SystemMessage:
			resultMsg = jpf.Message{
				Role:    jpf.SystemRole,
				Content: msg.Content,
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
		case PromptFragmentMessage:
			frags := make([]string, len(msg.Fragments))
			for i, r := range msg.Fragments {
				frags[i] = r.Content
			}
			sep := "\n\n"
			resultMsg = jpf.Message{
				Role:    jpf.SystemRole,
				Content: "Below is some potentially useful information (some of this may not be relevant):" + sep + strings.Join(frags, sep),
			}
		case AvailableToolDefinitionsMessage:
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
