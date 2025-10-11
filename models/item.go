package models

import "github.com/google/uuid"

type Item struct {
	Id              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Equipped        bool      `json:"equipped"`
	AttunementSlots int       `json:"attunement_slots"`
	Description     string    `json:"description"`
	Quantity        int       `json:"quantity"`
	Weight          float64   `json:"weight"`
}

type Wallet struct {
	Copper   int `json:"copper"`
	Silver   int `json:"silver"`
	Electrum int `json:"electrum"`
	Gold     int `json:"gold"`
	Platinum int `json:"platinum"`
}
