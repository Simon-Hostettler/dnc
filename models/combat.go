package models

type DeathSaves struct {
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
}

type Attack struct {
	Name       string `json:"name"`
	Bonus      int    `json:"bonus"`
	Damage     string `json:"damage"` // e.g., "1d8+3"
	DamageType string `json:"damage_type"`
}
