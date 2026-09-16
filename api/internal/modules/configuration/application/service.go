package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sechelper-auth-template/api/internal/modules/configuration/domain"
	"strings"
)

var keyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,127}$`)
var ErrNotFound = errors.New("configuration entry not found")

const InstallationMarker = "FRAMEWORK_INSTALLATION_COMPLETE"

type InstallEntry struct {
	Key, Description, Value string
	IsSecret                bool
}

type Repository interface {
	List(context.Context) ([]domain.Entry, error)
	Values(context.Context) (map[string]string, error)
	Get(context.Context, string) (domain.Value, error)
	Upsert(context.Context, string, string, bool, string, string) (domain.Entry, error)
	Delete(context.Context, string) error
}
type Protector interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}
type Service struct {
	repository Repository
	protector  Protector
}

func NewService(repository Repository, protector Protector) *Service {
	return &Service{repository: repository, protector: protector}
}
func (s *Service) GetValue(ctx context.Context, key string) (string, bool) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return "", false
	}
	return value.Value, true
}
func (s *Service) List(ctx context.Context) ([]domain.Entry, error) { return s.repository.List(ctx) }
func (s *Service) Resolve(ctx context.Context) (map[string]string, error) {
	values, err := s.repository.Values(ctx)
	if err != nil {
		return nil, err
	}
	for key, encrypted := range values {
		value, err := s.protector.Decrypt(encrypted)
		if err != nil {
			return nil, fmt.Errorf("decrypt configuration %s: %w", key, err)
		}
		values[key] = value
	}
	return values, nil
}
func (s *Service) Get(ctx context.Context, key string) (domain.Value, error) {
	value, err := s.repository.Get(ctx, key)
	if err != nil {
		return domain.Value{}, err
	}
	value.Value, err = s.protector.Decrypt(value.Value)
	return value, err
}
func (s *Service) Upsert(ctx context.Context, key, description string, secret bool, value, actor string) (domain.Entry, error) {
	key = strings.TrimSpace(key)
	if !keyPattern.MatchString(key) {
		return domain.Entry{}, errors.New("key must be an uppercase environment variable name")
	}
	if strings.TrimSpace(value) == "" {
		return domain.Entry{}, errors.New("value is required")
	}
	encrypted, err := s.protector.Encrypt(value)
	if err != nil {
		return domain.Entry{}, err
	}
	return s.repository.Upsert(ctx, key, strings.TrimSpace(description), secret, encrypted, actor)
}
func (s *Service) Delete(ctx context.Context, key string) error {
	if !keyPattern.MatchString(strings.TrimSpace(key)) {
		return errors.New("invalid key")
	}
	return s.repository.Delete(ctx, strings.TrimSpace(key))
}

func (s *Service) IsInstalled(ctx context.Context) (bool, error) {
	value, err := s.Get(ctx, InstallationMarker)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return value.Value == "true", nil
}

func (s *Service) Install(ctx context.Context, entries []InstallEntry, actor string) error {
	installed, err := s.IsInstalled(ctx)
	if err != nil {
		return err
	}
	if installed {
		return errors.New("configuration center is already installed")
	}
	for _, entry := range entries {
		if _, err := s.Upsert(ctx, entry.Key, entry.Description, entry.IsSecret, entry.Value, actor); err != nil {
			return err
		}
	}
	_, err = s.Upsert(ctx, InstallationMarker, "framework installation marker", false, "true", actor)
	return err
}
