package models

type Item struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Weight      float64 `json:"weight"`
}
