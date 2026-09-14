package application

import (
	"context"
	"errors"
	"fmt"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
	"sort"
	"strings"
	"sync"
)

var (
	ErrUnknownResource = errors.New("unknown resource type")
	ErrUnknownAction   = errors.New("unknown resource action")
)

type ResourceRegistry struct {
	mu          sync.RWMutex
	definitions map[string]domain.ResourceDefinition
	policies    map[string]domain.ResourcePolicy
}

func NewResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{definitions: map[string]domain.ResourceDefinition{}, policies: map[string]domain.ResourcePolicy{}}
}
func (r *ResourceRegistry) Register(definition domain.ResourceDefinition, policy domain.ResourcePolicy) error {
	definition.Type = strings.TrimSpace(definition.Type)
	if definition.Type == "" {
		return errors.New("resource type is required")
	}
	if len(definition.Actions) == 0 {
		return fmt.Errorf("resource %q must define at least one action", definition.Type)
	}
	seen := map[string]struct{}{}
	for _, action := range definition.Actions {
		if strings.TrimSpace(action.Name) == "" || strings.TrimSpace(action.Permission) == "" {
			return fmt.Errorf("resource %q has an invalid action", definition.Type)
		}
		if _, ok := seen[action.Name]; ok {
			return fmt.Errorf("resource %q has duplicate action %q", definition.Type, action.Name)
		}
		seen[action.Name] = struct{}{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, ok := r.definitions[definition.Type]; ok && old.Name != definition.Name {
		return fmt.Errorf("resource %q is already registered", definition.Type)
	}
	r.definitions[definition.Type] = definition
	if policy != nil {
		r.policies[definition.Type] = policy
	}
	return nil
}
func (r *ResourceRegistry) Definitions() []domain.ResourceDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.ResourceDefinition, 0, len(r.definitions))
	for _, definition := range r.definitions {
		result = append(result, definition)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Type < result[j].Type })
	return result
}

type ResourceService struct{ registry *ResourceRegistry }

func NewResourceService(registry *ResourceRegistry) *ResourceService {
	return &ResourceService{registry: registry}
}
func (s *ResourceService) Definitions() []domain.ResourceDefinition { return s.registry.Definitions() }
func (s *ResourceService) Check(ctx context.Context, request domain.AccessRequest) (domain.Decision, error) {
	decision := domain.Decision{ResourceType: request.Resource.Type, ResourceID: request.Resource.ID, Action: request.Action}
	if request.Actor.Subject == "" {
		decision.ReasonCode = "UNAUTHENTICATED"
		return decision, nil
	}
	s.registry.mu.RLock()
	definition, ok := s.registry.definitions[request.Resource.Type]
	policy := s.registry.policies[request.Resource.Type]
	s.registry.mu.RUnlock()
	if !ok {
		return decision, ErrUnknownResource
	}
	var action domain.ActionDefinition
	for _, candidate := range definition.Actions {
		if candidate.Name == request.Action {
			action = candidate
			break
		}
	}
	if action.Name == "" {
		return decision, ErrUnknownAction
	}
	decision.Permission = action.Permission
	if !request.Actor.HasPermission(action.Permission) {
		decision.ReasonCode = "MISSING_PERMISSION"
		return decision, nil
	}
	if policy == nil {
		decision.Allowed = true
		decision.ReasonCode = "ALLOWED_BY_PERMISSION"
		return decision, nil
	}
	result, err := policy.Evaluate(ctx, request)
	if err != nil {
		return decision, err
	}
	result.Permission = decision.Permission
	result.ResourceType = request.Resource.Type
	result.ResourceID = request.Resource.ID
	result.Action = request.Action
	if result.ReasonCode == "" {
		result.ReasonCode = "RESOURCE_POLICY"
	}
	return result, nil
}
