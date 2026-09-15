//go:build !example

package main

import (
	"database/sql"

	"sechelper-auth-template/api/internal/application"
)

func newBusinessRuntime(_ *sql.DB) (*businessRuntime, error) {
	return &businessRuntime{modules: application.NewRegistry()}, nil
}
