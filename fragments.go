package react

// FragmentSelector defines an object that can choose relevant fragments to a conversation.
type FragmentSelector interface {
	// Select any relevant fragments that should be added tp the conversation.
	SelectFragments([]PromptFragment, []Message) ([]PromptFragment, error)
}

// Create a frament selector that never selects any fragments.
func NewNoFragmentSelector() FragmentSelector {
	return &noFragmentSelector{}
}

type noFragmentSelector struct{}

func (*noFragmentSelector) SelectFragments([]PromptFragment, []Message) ([]PromptFragment, error) {
	return nil, nil
}
