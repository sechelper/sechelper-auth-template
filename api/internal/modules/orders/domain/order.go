//go:build example

package domain

import "time"

type Order struct {
	ID         string
	Status     string
	TotalMinor int64
	Currency   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
