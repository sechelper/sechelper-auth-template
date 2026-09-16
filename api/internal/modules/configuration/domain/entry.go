package domain

import "time"

type Entry struct {
	Key         string    `json:"key"`
	Description string    `json:"description,omitempty"`
	IsSecret    bool      `json:"isSecret"`
	Version     int64     `json:"version"`
	UpdatedBy   string    `json:"updatedBy,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Value struct {
	Entry
	Value string `json:"value"`
}
