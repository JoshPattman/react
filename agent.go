package react

import (
	"context"

	"fmt"
	"iter"
	"slices"
)

type Agent struct {
	messages         []Message
	modelBuilder     AgentModelBuilder
	tools            []Tool
	skillSelector    SkillSelector
	dynamicFragments []Skill
}

func (ag *Agent) Send(msg string, opts ...SendMessageOpt) (string, error) {
	kwargs := getKwargs(opts)
	streamers := kwargs.Streamers()

	// Update notifications
	for _, msg := range kwargs.notifications {
		ag.addMessages(streamers, notificationMessage{msg})
	}

	// Add the initial user message
	ag.addMessages(streamers, userMessage{msg})

	// Signal we are collecting context and add any relevant fragments
	if len(ag.dynamicFragments) > 0 {
		ag.addMessages(streamers, modeSwitchMessage{ModeCollectContext})
		nextSkills, err := ag.getNextSelectedSkills()
		if err != nil {
			return "", err
		}
		ag.addMessages(streamers, skillMessage{nextSkills})
	}

	ag.addMessages(streamers, modeSwitchMessage{ModeReasonAct})

	// React loop
	for {
		// Ask agent for any new tool calls and break if there are no calls
		toolCalls, err := ag.answerReAct()
		if err != nil {
			return "", err
		}
		ag.addMessages(streamers, toolCalls)
		if len(toolCalls.ToolCalls) == 0 {
			break
		}
		// Execute tool calls
		toolResults := ag.executeToolCalls(toolCalls.ToolCalls)
		ag.addMessages(streamers, toolResponseMessage{toolResults})
	}

	// Set the agent to final answer mode and get the response
	ag.addMessages(streamers, modeSwitchMessage{ModeAnswerUser})
	finalResp, err := ag.answerFinalResponse(streamers)
	if err != nil {
		return "", err
	}
	ag.addMessages(streamers, agentMessage{finalResp})
	return finalResp, nil
}

func (ag *Agent) Messages() iter.Seq[Message] {
	return slices.Values(ag.messages)
}

func (ag *Agent) getNextSelectedSkills() ([]InsertedSkill, error) {
	// Find any carry forward skills
	prevSkills := getLastInsertedSkills(ag.messages)
	skillsToPersist := make([]InsertedSkill, 0)
	for _, f := range prevSkills {
		if f.NowRemainFor > 0 {
			skillsToPersist = append(skillsToPersist, InsertedSkill{f.Skill, f.NowRemainFor - 1})
		}
	}
	// Select new skills
	newSkills, err := ag.skillSelector.SelectSkills(ag.dynamicFragments, ag.messages)
	if err != nil {
		return nil, err
	}
	skillsToInsert := make([]InsertedSkill, len(newSkills))
	for i, s := range newSkills {
		skillsToInsert[i] = InsertedSkill{s, s.RemainFor}
	}
	return append(skillsToPersist, skillsToInsert...), nil
}

func (ag *Agent) answerReAct() (toolCallsMessage, error) {
	model := ag.modelBuilder.BuildAgentModel(reasonResponse{}, nil, nil)
	pipeline := getAgentReActPipeline(model)
	result, _, err := pipeline.Call(context.Background(), ag.messages)
	if err != nil {
		return toolCallsMessage{}, err
	}
	return toolCallsMessageFromResponse(result), nil
}

func (ag *Agent) answerFinalResponse(streamer TextStreamer) (string, error) {
	model := ag.modelBuilder.BuildAgentModel(nil, nil, streamer.TrySendTextChunk)
	pipeline := getAgentFinalAnswerPipeline(model)
	result, _, err := pipeline.Call(context.Background(), ag.messages)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (ag *Agent) addMessages(msgStreamer MessageStreamer, msgs ...Message) {
	ag.messages = append(ag.messages, msgs...)
	if msgStreamer != nil {
		for _, msg := range msgs {
			msgStreamer.TrySendMessage(msg)
		}
	}
}

func (ag *Agent) executeToolCalls(calls []ToolCall) []ToolResponse {
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

func (ag *Agent) findToolByName(toolName string) Tool {
	for _, t := range ag.tools {
		if t.Name() == toolName {
			return t
		}
	}
	return nil
}
