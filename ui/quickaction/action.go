package quickaction

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/repository"
)

type Action interface {
	Name() string
	ArgHint() string
	Execute(agg *repository.CharacterAggregate, args string) (tea.Cmd, string)
}

type LongRestAction struct{}

func (a LongRestAction) Name() string    { return "longrest" }
func (a LongRestAction) ArgHint() string { return "" }

func (a LongRestAction) Execute(agg *repository.CharacterAggregate, args string) (tea.Cmd, string) {
	c := agg.Character
	c.CurrHitPoints = c.MaxHitPoints
	c.DeathSaveSuccesses = 0
	c.DeathSaveFailures = 0
	for i := range c.SpellSlotsUsed {
		c.SpellSlotsUsed[i] = 0
	}
	return command.WriteBackRequest, ""
}

type CastAction struct{}

func (a CastAction) Name() string    { return "cast" }
func (a CastAction) ArgHint() string { return "<level>" }

func (a CastAction) Execute(agg *repository.CharacterAggregate, args string) (tea.Cmd, string) {
	level, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil || level < 1 || level > 9 {
		return nil, "usage: cast <1-9>"
	}
	c := agg.Character
	if level >= len(c.SpellSlots) || c.SpellSlots[level] <= 0 {
		return nil, fmt.Sprintf("no spell slots at level %d", level)
	}
	if c.SpellSlotsUsed[level] >= c.SpellSlots[level] {
		return nil, fmt.Sprintf("no available slots at level %d", level)
	}
	c.SpellSlotsUsed[level]++
	return command.WriteBackRequest, ""
}

type HealAction struct{}

func (a HealAction) Name() string    { return "heal" }
func (a HealAction) ArgHint() string { return "<amount>" }

func (a HealAction) Execute(agg *repository.CharacterAggregate, args string) (tea.Cmd, string) {
	amount, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil || amount < 0 {
		return nil, "usage: heal <amount>"
	}
	c := agg.Character
	c.CurrHitPoints = min(c.CurrHitPoints+amount, c.MaxHitPoints)
	return command.WriteBackRequest, ""
}

type DmgAction struct{}

func (a DmgAction) Name() string    { return "dmg" }
func (a DmgAction) ArgHint() string { return "<amount>" }

func (a DmgAction) Execute(agg *repository.CharacterAggregate, args string) (tea.Cmd, string) {
	amount, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil || amount < 0 {
		return nil, "usage: dmg <amount>"
	}
	c := agg.Character
	c.CurrHitPoints = max(c.CurrHitPoints-amount, 0)
	return command.WriteBackRequest, ""
}
