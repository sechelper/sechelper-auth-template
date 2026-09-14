package domain

import "errors"

var ErrDuplicatePermission = errors.New("duplicate permission")

type API struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}
type Permission struct {
	Code        string `json:"permission_code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	RiskLevel   string `json:"risk_level"`
	APIs        []API  `json:"apis"`
}
type Snapshot struct {
	ApplicationCode string       `json:"application_code"`
	ManifestVersion int64        `json:"manifest_version"`
	Permissions     []Permission `json:"permissions"`
}
