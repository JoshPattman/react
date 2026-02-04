package react

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/JoshPattman/jpf"
)

// SkillSelector defines an object that can choose relevant [Skill]s to a conversation.
type SkillSelector interface {
	// Select any relevant [Skill]s that should be added to the conversation.
	SelectSkills([]Skill, []Message) ([]Skill, error)
}

func NewSkillSelector(modelBuilder FragmentSelectorModelBuilder, dontRepeatN int) SkillSelector {
	if modelBuilder == nil {
		return &noSkillSelector{}
	}
	var selec SkillSelector = &conversationLLMSkillSelector{modelBuilder}
	if dontRepeatN > 0 {
		selec = &dontRepeatInNFragmentSelector{dontRepeatN, selec}
	}
	return selec
}

type noSkillSelector struct{}

func (*noSkillSelector) SelectSkills([]Skill, []Message) ([]Skill, error) {
	return nil, nil
}

type conversationLLMSkillSelector struct {
	modelBuilder FragmentSelectorModelBuilder
}

type conversationLLMSkillSelectorInput struct {
	Frags    []Skill
	Messages []Message
}

type conversationLLMSkillSelectorOutput struct {
	RelevantFragmentIDs []string `json:"relevant_fragment_ids"`
}

func (selector *conversationLLMSkillSelector) SelectSkills(frags []Skill, messages []Message) ([]Skill, error) {
	model := selector.modelBuilder.BuildFragmentSelectorModel(conversationLLMSkillSelectorOutput{})
	encoder := selector
	decoder := jpf.NewJsonParser[conversationLLMSkillSelectorOutput]()
	mf := jpf.NewOneShotPipeline(encoder, decoder, nil, model)
	result, _, err := mf.Call(context.Background(), conversationLLMSkillSelectorInput{frags, messages})
	if err != nil {
		return nil, err
	}
	fragLookup := make(map[string]Skill)
	for _, f := range frags {
		fragLookup[f.Key] = f
	}
	relevantFrags := make([]Skill, 0)
	for _, fragID := range result.RelevantFragmentIDs {
		frag, ok := fragLookup[fragID]
		if !ok {
			continue // TODO: Maybe should check in a decoder so we can retry rather than ignoring
		}
		relevantFrags = append(relevantFrags, frag)
	}
	return relevantFrags, nil
}

func (selector *conversationLLMSkillSelector) BuildInputMessages(input conversationLLMSkillSelectorInput) ([]jpf.Message, error) {
	conv := make([]string, 0)
	for _, msg := range input.Messages {
		switch msg := msg.(type) {
		case UserMessage:
			conv = append(conv, fmt.Sprintf("<user-message>%s</user-message>", msg.Content))
		case AgentMessage:
			conv = append(conv, fmt.Sprintf("<agent-message>%s</agent-message>", msg.Content))
		}
	}
	if len(conv) > 10 {
		conv = conv[len(conv)-10:]
	}
	frags := make([]string, 0)
	for _, f := range input.Frags {
		frags = append(frags, fmt.Sprintf(`<fragment id="%s">%s</fragment>`, f.Key, f.When))
	}
	systemPrompt := `You are a fast AI who decides if any "fragments" are relevant to an agent's conversation.
	- You will list all fragment IDs in your response that you think might be relevant to the current turn in the conversation (the last message).
	- This means any fragments that may help an agent continue the conversation should be included.
	- Prefer recall over precision.
	- It may be the case that none are relevant, in that case respond with an empty list.
	- You will respond with a json object with a key "relevant_fragment_ids", which is a list of string IDs that exactly match the IDs of the provided fragments.`

	userPrompt := fmt.Sprintf(
		"Here is the conversation and messages:\n\n%s\n\n%s",
		strings.Join(conv, "\n"),
		strings.Join(frags, "\n"),
	)

	return []jpf.Message{
		{
			Role:    jpf.SystemRole,
			Content: systemPrompt,
		},
		{
			Role:    jpf.UserRole,
			Content: userPrompt,
		},
	}, nil
}

type dontRepeatInNFragmentSelector struct {
	n        int
	selector SkillSelector
}

func (selector *dontRepeatInNFragmentSelector) SelectSkills(frags []Skill, messages []Message) ([]Skill, error) {
	allowedSelectFrags := slices.Clone(frags)
	for i := len(messages) - 1; i >= 0 && len(messages)-1-i < selector.n; i-- {
		msg, ok := messages[i].(SkillMessage)
		if !ok {
			continue
		}
		for _, f := range msg.Skills {
			allowedSelectFrags = slices.DeleteFunc(allowedSelectFrags, func(x Skill) bool { return x == f })
		}
	}
	return selector.selector.SelectSkills(allowedSelectFrags, messages)
}
