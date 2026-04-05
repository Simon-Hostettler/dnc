package quickaction

import "strings"

type Registry struct {
	actions []Action
}

func NewRegistry() *Registry {
	r := &Registry{}
	r.Register(LongRestAction{})
	r.Register(CastAction{})
	r.Register(HealAction{})
	r.Register(DmgAction{})
	return r
}

func (r *Registry) Register(a Action) {
	r.actions = append(r.actions, a)
}

func (r *Registry) All() []Action {
	return r.actions
}

func (r *Registry) Match(prefix string) []Action {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "" {
		return r.actions
	}
	var matches []Action
	for _, a := range r.actions {
		if strings.HasPrefix(a.Name(), prefix) {
			matches = append(matches, a)
		}
	}
	return matches
}

func (r *Registry) Parse(input string) (Action, string, bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, "", false
	}
	parts := strings.SplitN(input, " ", 2)
	name := strings.ToLower(parts[0])
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}
	for _, a := range r.actions {
		if a.Name() == name {
			return a, args, true
		}
	}
	return nil, "", false
}
