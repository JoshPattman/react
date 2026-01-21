package react

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/JoshPattman/jpf"
)

// FragmentSelector defines an object that can choose relevant fragments to a conversation.
type FragmentSelector interface {
	// Select any relevant fragments that should be added tp the conversation.
	SelectFragments([]PromptFragment, []Message) ([]PromptFragment, error)
}

func NewFragmentSelector(modelBuilder FragmentSelectorModelBuilder, dontRepeatN int) FragmentSelector {
	if modelBuilder == nil {
		return &noFragmentSelector{}
	}
	var selec FragmentSelector = &llmFragmentSelector{modelBuilder}
	if dontRepeatN > 0 {
		selec = &dontRepeatInNFragmentSelector{dontRepeatN, selec}
	}
	return selec
}

type noFragmentSelector struct{}

func (*noFragmentSelector) SelectFragments([]PromptFragment, []Message) ([]PromptFragment, error) {
	return nil, nil
}

type llmFragmentSelector struct {
	modelBuilder FragmentSelectorModelBuilder
}

type llmFragmentSelectorInput struct {
	Frags    []PromptFragment
	Messages []Message
}

type llmFragmentSelectorOutput struct {
	RelevantFragmentIDs []string `json:"relevant_fragment_ids"`
}

func (selector *llmFragmentSelector) SelectFragments(frags []PromptFragment, messages []Message) ([]PromptFragment, error) {
	model := selector.modelBuilder.BuildFragmentSelectorModel(llmFragmentSelectorOutput{})
	encoder := selector
	decoder := jpf.NewJsonResponseDecoder[llmFragmentSelectorInput, llmFragmentSelectorOutput]()
	mf := jpf.NewOneShotMapFunc(encoder, decoder, model)
	result, _, err := mf.Call(context.Background(), llmFragmentSelectorInput{frags, messages})
	if err != nil {
		return nil, err
	}
	fragLookup := make(map[string]PromptFragment)
	for _, f := range frags {
		fragLookup[f.Key] = f
	}
	relevantFrags := make([]PromptFragment, 0)
	for _, fragID := range result.RelevantFragmentIDs {
		frag, ok := fragLookup[fragID]
		if !ok {
			continue // TODO: Maybe should check in a decoder so we can retry rather than ignoring
		}
		relevantFrags = append(relevantFrags, frag)
	}
	return relevantFrags, nil
}

func (selector *llmFragmentSelector) BuildInputMessages(input llmFragmentSelectorInput) ([]jpf.Message, error) {
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
	selector FragmentSelector
}

func (selector *dontRepeatInNFragmentSelector) SelectFragments(frags []PromptFragment, messages []Message) ([]PromptFragment, error) {
	allowedSelectFrags := slices.Clone(frags)
	for i := len(messages) - 1; i >= 0 && len(messages)-1-i < selector.n; i-- {
		msg, ok := messages[i].(PromptFragmentMessage)
		if !ok {
			continue
		}
		for _, f := range msg.Fragments {
			allowedSelectFrags = slices.DeleteFunc(allowedSelectFrags, func(x PromptFragment) bool { return x == f })
		}
	}
	return selector.selector.SelectFragments(allowedSelectFrags, messages)
}
