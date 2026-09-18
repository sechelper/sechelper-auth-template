package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"sechelper-auth-template/api/internal/modules/manifest/domain"
)

type Registry struct {
	applicationCode string
	permissions     map[string]domain.Permission
}

var allowedRiskLevels = map[string]struct{}{
	"normal":     {},
	"privileged": {},
	"critical":   {},
}

func NewRegistry(applicationCode string) *Registry {
	return &Registry{applicationCode: applicationCode, permissions: map[string]domain.Permission{}}
}
func (r *Registry) Register(p domain.Permission) error {
	p.Code = strings.TrimSpace(p.Code)
	if p.Code == "" {
		return fmt.Errorf("permission code is required")
	}
	p.Name = strings.TrimSpace(p.Name)
	p.Description = strings.TrimSpace(p.Description)
	p.RiskLevel = strings.ToLower(strings.TrimSpace(p.RiskLevel))
	if p.RiskLevel == "" {
		p.RiskLevel = "normal"
	}
	if _, ok := allowedRiskLevels[p.RiskLevel]; !ok {
		return fmt.Errorf("unsupported risk level %q: must be one of normal, privileged, critical", p.RiskLevel)
	}
	for i := range p.APIs {
		p.APIs[i].Method = strings.ToUpper(strings.TrimSpace(p.APIs[i].Method))
		p.APIs[i].Path = strings.TrimSpace(p.APIs[i].Path)
	}
	if old, ok := r.permissions[p.Code]; ok && old.Name != p.Name {
		return domain.ErrDuplicatePermission
	}
	r.permissions[p.Code] = p
	return nil
}
func (r *Registry) Permissions() []domain.Permission {
	values := make([]domain.Permission, 0, len(r.permissions))
	for _, value := range r.permissions {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Code < values[j].Code })
	return values
}
func (r *Registry) Snapshot(version int64) (domain.Snapshot, string, error) {
	if version < 1 {
		version = 1
	}
	values := make([]domain.Permission, 0, len(r.permissions))
	for _, p := range r.permissions {
		p.Name = strings.TrimSpace(p.Name)
		p.Description = strings.TrimSpace(p.Description)
		p.RiskLevel = strings.ToLower(strings.TrimSpace(p.RiskLevel))
		if p.RiskLevel == "" {
			p.RiskLevel = "normal"
		}
		if _, ok := allowedRiskLevels[p.RiskLevel]; !ok {
			return domain.Snapshot{}, "", fmt.Errorf("unsupported risk level %q: must be one of normal, privileged, critical", p.RiskLevel)
		}
		for i := range p.APIs {
			p.APIs[i].Method = strings.ToUpper(strings.TrimSpace(p.APIs[i].Method))
			p.APIs[i].Path = strings.TrimSpace(p.APIs[i].Path)
		}
		sort.Slice(p.APIs, func(i, j int) bool {
			if p.APIs[i].Method == p.APIs[j].Method {
				return p.APIs[i].Path < p.APIs[j].Path
			}
			return p.APIs[i].Method < p.APIs[j].Method
		})
		values = append(values, p)
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Code < values[j].Code })
	snapshot := domain.Snapshot{ApplicationCode: r.applicationCode, ManifestVersion: version, Permissions: values}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return domain.Snapshot{}, "", err
	}
	sum := sha256.Sum256(raw)
	return snapshot, "sha256:" + hex.EncodeToString(sum[:]), nil
}
