package react

import _ "embed"

func New(mb ModelBuilder, opts ...NewOpt) *Agent {
	kwargs := getNewKwargs(opts)
	messages := []Message{
		personalityMessage{
			kwargs.personality,
		},
		systemMessage{
			Template: createCraigSystemTemplate(),
		},
	}
	return newHelper(mb, messages, kwargs)
}

func NewFromSaved(mb ModelBuilder, messages []Message, opts ...NewOpt) *Agent {
	return newHelper(mb, messages, getNewKwargs(opts))
}

func newHelper(mb ModelBuilder, messages []Message, kwargs newKwargs) *Agent {
	// Add persistent skills by default forever
	dyn, pers := getDynamicAndPersistent(kwargs.skills)
	insertPersistentSkills := make([]InsertedSkill, len(pers))
	for i, s := range pers {
		insertPersistentSkills[i] = InsertedSkill{s, 999999999999999999}
	}
	messages = append(messages, skillMessage{insertPersistentSkills})

	// Add tool definitions if the tools were changed since the last agent
	if toolsHaveChanged(messages, kwargs.tools) {
		messages = append(messages, toolsMessage{
			getToolDefs(kwargs.tools),
		})
	}

	// Build
	ag := &Agent{
		messages:         messages,
		modelBuilder:     mb,
		dynamicFragments: dyn,
		skillSelector:    NewSkillSelector(mb),
	}
	return ag
}

func getNewKwargs(opts []NewOpt) newKwargs {
	kwargs := newKwargs{
		personality: "Your name is CRAIG, a helpful assistant.",
	}
	for _, o := range opts {
		o(&kwargs)
	}
	return kwargs
}

type NewOpt func(*newKwargs)

func WithSkills(skills ...Skill) func(kw *newKwargs) {
	return func(kw *newKwargs) { kw.skills = append(kw.skills, skills...) }
}

func WithTools(tools ...Tool) func(kw *newKwargs) {
	return func(kw *newKwargs) { kw.tools = append(kw.tools, tools...) }
}

func WithPersonality(personality string) func(kw *newKwargs) {
	return func(kw *newKwargs) { kw.personality = personality }
}

type newKwargs struct {
	skills      []Skill
	tools       []Tool
	personality string
}

//go:embed system.tpl
var craigSystemPrompt string

func createCraigSystemTemplate() string {
	return craigSystemPrompt
}
