//go:build example

package main

import (
	"database/sql"

	"sechelper-auth-template/api/internal/application"
	"sechelper-auth-template/api/internal/business/orders"
)

func newBusinessRuntime(db *sql.DB) (*businessRuntime, error) {
	modules := application.NewRegistry()
	if err := modules.Add(orders.New(db)); err != nil {
		return nil, err
	}
	return &businessRuntime{modules: modules}, nil
}
