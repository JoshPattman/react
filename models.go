package react

import "github.com/JoshPattman/jpf"

type ModelBuilder interface {
	AgentModelBuilder
	FragmentSelectorModelBuilder
}

type AgentModelBuilder interface {
	// Build a model for the agent.
	// A struct may be passed as the response type for a json schema, or it may be nil.
	// The stream callbacks may also be nil.
	BuildAgentModel(responseType any, onInitFinalStream func(), onDataFinalStream func(string)) jpf.Model
}

type FragmentSelectorModelBuilder interface {
	// Build a model for a fragment selector.
	// A struct may be passed as the response type for a json schema, or it may be nil.
	BuildFragmentSelectorModel(responseType any) jpf.Model
}
