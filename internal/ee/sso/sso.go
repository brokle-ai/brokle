package sso

import (
	"context"
	"errors"
)

// SSOProvider interface for enterprise SSO authentication
type SSOProvider interface {
	Authenticate(ctx context.Context, token string) (*User, error)
	GetLoginURL(ctx context.Context) (string, error)
	ValidateAssertion(ctx context.Context, assertion string) (*User, error)
	GetSupportedProviders(ctx context.Context) ([]string, error)
	ConfigureProvider(ctx context.Context, provider, config string) error
}

// User represents an authenticated SSO user
type User struct {
	ID          string            `json:"id"`
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	Roles       []string          `json:"roles"`
	Attributes  map[string]string `json:"attributes"`
	Provider    string            `json:"provider"`
}

// StubSSO provides stub implementation for OSS version
type StubSSO struct{}

// New returns the SSO provider implementation (stub or real based on build tags)
func New() SSOProvider {
	return &StubSSO{}
}

func (s *StubSSO) Authenticate(ctx context.Context, token string) (*User, error) {
	return nil, errors.New("SSO authentication requires Enterprise license")
}

func (s *StubSSO) GetLoginURL(ctx context.Context) (string, error) {
	return "", errors.New("SSO login requires Enterprise license")
}

func (s *StubSSO) ValidateAssertion(ctx context.Context, assertion string) (*User, error) {
	return nil, errors.New("SSO assertion validation requires Enterprise license")
}

func (s *StubSSO) GetSupportedProviders(ctx context.Context) ([]string, error) {
	return []string{}, errors.New("SSO providers require Enterprise license")
}

func (s *StubSSO) ConfigureProvider(ctx context.Context, provider, config string) error {
	return errors.New("SSO configuration requires Enterprise license")
}