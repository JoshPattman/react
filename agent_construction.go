package react

import _ "embed"

func New(mb ModelBuilder, opts ...NewOpt) *Agent {
	kwargs := getNewKwargs(opts)
	messages := []Message{
		PersonalityMessage{
			kwargs.personality,
		},
		SystemMessage{
			Template: createCraigSystemTemplate(),
		},
	}
	return newHelper(mb, messages, kwargs)
}

func NewFromSaved(mb ModelBuilder, messages []Message, opts ...NewOpt) *Agent {
	return newHelper(mb, messages, getNewKwargs(opts))
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

func newHelper(mb ModelBuilder, messages []Message, kwargs newKwargs) *Agent {
	dyn, pers := getDynamicAndPersistent(kwargs.skills)
	insertPersistentSkills := make([]InsertedSkill, len(pers))
	for i, s := range pers {
		insertPersistentSkills[i] = InsertedSkill{s, 999999999999999999}
	}
	messages = append(messages, SkillMessage{insertPersistentSkills})
	ag := &Agent{
		messages:         messages,
		modelBuilder:     mb,
		tools:            kwargs.tools,
		dynamicFragments: dyn,
		skillSelector:    NewSkillSelector(mb),
	}
	return ag
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
