package quickaction

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dicestats"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/repository"
)

type ActionResult struct {
	Cmd    tea.Cmd
	ErrMsg string
	Result string
}

type Action interface {
	Name() string
	ArgHint() string
	Execute(agg *repository.CharacterAggregate, args string) ActionResult
}

type LongRestAction struct{}

func (a LongRestAction) Name() string    { return "longrest" }
func (a LongRestAction) ArgHint() string { return "" }

func (a LongRestAction) Execute(agg *repository.CharacterAggregate, args string) ActionResult {
	c := agg.Character
	c.CurrHitPoints = c.MaxHitPoints
	c.DeathSaveSuccesses = 0
	c.DeathSaveFailures = 0
	for i := range c.SpellSlotsUsed {
		c.SpellSlotsUsed[i] = 0
	}
	return ActionResult{Cmd: command.WriteBackRequest}
}

type CastAction struct{}

func (a CastAction) Name() string    { return "cast" }
func (a CastAction) ArgHint() string { return "<level>" }

func (a CastAction) Execute(agg *repository.CharacterAggregate, args string) ActionResult {
	level, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil || level < 1 || level > 9 {
		return ActionResult{ErrMsg: "usage: cast <1-9>"}
	}
	c := agg.Character
	if level >= len(c.SpellSlots) || c.SpellSlots[level] <= 0 {
		return ActionResult{ErrMsg: fmt.Sprintf("no spell slots at level %d", level)}
	}
	if c.SpellSlotsUsed[level] >= c.SpellSlots[level] {
		return ActionResult{ErrMsg: fmt.Sprintf("no available slots at level %d", level)}
	}
	c.SpellSlotsUsed[level]++
	return ActionResult{Cmd: command.WriteBackRequest}
}

type HealAction struct{}

func (a HealAction) Name() string    { return "heal" }
func (a HealAction) ArgHint() string { return "<amount>" }

func (a HealAction) Execute(agg *repository.CharacterAggregate, args string) ActionResult {
	amount, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil || amount < 0 {
		return ActionResult{ErrMsg: "usage: heal <amount>"}
	}
	c := agg.Character
	c.CurrHitPoints = min(c.CurrHitPoints+amount, c.MaxHitPoints)
	return ActionResult{Cmd: command.WriteBackRequest}
}

type DmgAction struct{}

func (a DmgAction) Name() string    { return "dmg" }
func (a DmgAction) ArgHint() string { return "<amount>" }

func (a DmgAction) Execute(agg *repository.CharacterAggregate, args string) ActionResult {
	amount, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil || amount < 0 {
		return ActionResult{ErrMsg: "usage: dmg <amount>"}
	}
	c := agg.Character
	c.CurrHitPoints = max(c.CurrHitPoints-amount, 0)
	return ActionResult{Cmd: command.WriteBackRequest}
}

type ProbAction struct{}

func (a ProbAction) Name() string    { return "prob" }
func (a ProbAction) ArgHint() string { return "<expr cmp value>" }

func (a ProbAction) Execute(_ *repository.CharacterAggregate, args string) ActionResult {
	args = strings.TrimSpace(args)
	if args == "" {
		return ActionResult{ErrMsg: "usage: prob <expr cmp value>"}
	}
	qr, err := dicestats.Query("P[" + args + "]")
	if err != nil {
		return ActionResult{ErrMsg: err.Error()}
	}
	prefix := ""
	if qr.Approximate {
		prefix = "~"
	}
	return ActionResult{Result: fmt.Sprintf("P = %s%.4f", prefix, qr.Value)}
}

type EvAction struct{}

func (a EvAction) Name() string    { return "ev" }
func (a EvAction) ArgHint() string { return "<expression>" }

func (a EvAction) Execute(_ *repository.CharacterAggregate, args string) ActionResult {
	args = strings.TrimSpace(args)
	if args == "" {
		return ActionResult{ErrMsg: "usage: ev <expression>"}
	}
	qr, err := dicestats.Query("E[" + args + "]")
	if err != nil {
		return ActionResult{ErrMsg: err.Error()}
	}
	prefix := ""
	if qr.Approximate {
		prefix = "~"
	}
	return ActionResult{Result: fmt.Sprintf("E = %s%.4f", prefix, qr.Value)}
}

type DistAction struct{}

func (a DistAction) Name() string    { return "dist" }
func (a DistAction) ArgHint() string { return "<expression>" }

func (a DistAction) Execute(_ *repository.CharacterAggregate, args string) ActionResult {
	args = strings.TrimSpace(args)
	if args == "" {
		return ActionResult{ErrMsg: "usage: dist <expression>"}
	}
	qr, err := dicestats.Query("D[" + args + "]")
	if err != nil {
		return ActionResult{ErrMsg: err.Error()}
	}
	d := qr.Distribution
	prefix := ""
	if qr.Approximate {
		prefix = "~"
	}
	result := fmt.Sprintf(
		"%smean: %.2f  std: %.2f\nmin:  %d      max: %d\nmode: %d      med: %d",
		prefix,
		d.Expected(), d.StdDev(),
		d.Min(), d.Max(),
		d.Mode(), d.Median(),
	)
	return ActionResult{Result: result}
}
