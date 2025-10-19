package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type CharacterSummary struct {
	ID   uuid.UUID
	Name string
}

type Equippable int

const (
	NonEquippable Equippable = iota
	NotEquipped
	Equipped
)

type Proficiency int

const (
	NoProficiency Proficiency = iota
	Proficient
	Expertise
)

func (c CharacterSkillDetailTO) ToCharacterSkillTO() CharacterSkillTO {
	return CharacterSkillTO{
		ID:             c.ID,
		CharacterID:    c.CharacterID,
		SkillID:        c.SkillID,
		Proficiency:    c.Proficiency,
		CustomModifier: c.CustomModifier,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func (c CharacterSkillDetailTO) ToModifier(score int, profMod int) int {
	return (score-10)/2 + profMod*c.Proficiency
}

func (a AbilitiesTO) ToScoreByName(ability string) int {
	switch ability {
	case "strength":
		return a.Strength
	case "dexterity":
		return a.Dexterity
	case "constitution":
		return a.Constitution
	case "intelligence":
		return a.Intelligence
	case "wisdom":
		return a.Wisdom
	case "charisma":
		return a.Charisma
	}
	return 0
}

// IntList is a []int that implements sql.Scanner (and optionally driver.Valuer) to handle DuckDB LIST columns.
type IntList []int

// Scan implements sql.Scanner for DuckDB LIST to Go []int.
func (il *IntList) Scan(src any) error {
	if src == nil {
		*il = nil
		return nil
	}
	switch v := src.(type) {
	case []interface{}:
		out := make([]int, 0, len(v))
		for _, e := range v {
			switch t := e.(type) {
			case int64:
				out = append(out, int(t))
			case int32:
				out = append(out, int(t))
			case float64:
				out = append(out, int(t))
			case []byte:
				// fallback: parse numeric string
				n, err := strconv.Atoi(string(t))
				if err != nil {
					return err
				}
				out = append(out, n)
			case string:
				n, err := strconv.Atoi(t)
				if err != nil {
					return err
				}
				out = append(out, n)
			default:
				return fmt.Errorf("IntList.Scan: unsupported element type %T", e)
			}
		}
		*il = out
		return nil
	case string:
		parsed, err := parseBracketedIntList(v)
		if err != nil {
			return err
		}
		*il = parsed
		return nil
	case []byte:
		parsed, err := parseBracketedIntList(string(v))
		if err != nil {
			return err
		}
		*il = parsed
		return nil
	default:
		return fmt.Errorf("IntList.Scan: unsupported source type %T", src)
	}
}

// Value implements driver.Valuer; we return a string literal which we don't currently use for inserts (we build literals ourselves), but keep for completeness.
func (il IntList) Value() (driver.Value, error) {
	var b strings.Builder
	b.WriteByte('[')
	for i, n := range il {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(n))
	}
	b.WriteByte(']')
	return b.String(), nil
}

func parseBracketedIntList(s string) ([]int, error) {
	ss := strings.TrimSpace(s)
	if ss == "" || ss == "[]" {
		return []int{}, nil
	}
	if !strings.HasPrefix(ss, "[") || !strings.HasSuffix(ss, "]") {
		return nil, errors.New("invalid list literal")
	}
	inner := strings.TrimSpace(ss[1 : len(ss)-1])
	if inner == "" {
		return []int{}, nil
	}
	parts := strings.Split(inner, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}
