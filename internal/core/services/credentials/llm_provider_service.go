// Package credentials provides service implementations for credential management.
package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	credentialsDomain "brokle/internal/core/domain/credentials"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/encryption"
	"brokle/pkg/ulid"
)

// llmProviderCredentialService implements credentialsDomain.LLMProviderCredentialService
type llmProviderCredentialService struct {
	repo       credentialsDomain.LLMProviderCredentialRepository
	encryptor  *encryption.Service
	logger     *slog.Logger
	httpClient *http.Client
}

// NewLLMProviderCredentialService creates a new service instance.
func NewLLMProviderCredentialService(
	repo credentialsDomain.LLMProviderCredentialRepository,
	encryptor *encryption.Service,
	logger *slog.Logger,
) credentialsDomain.LLMProviderCredentialService {
	return &llmProviderCredentialService{
		repo:      repo,
		encryptor: encryptor,
		logger:    logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateOrUpdate creates or updates a credential for a project/provider.
func (s *llmProviderCredentialService) CreateOrUpdate(ctx context.Context, req *credentialsDomain.CreateCredentialRequest) (*credentialsDomain.LLMProviderCredentialResponse, error) {
	if !req.Provider.IsValid() {
		return nil, appErrors.NewValidationError("Invalid provider", fmt.Sprintf("provider must be one of: %v", credentialsDomain.ValidProviders()))
	}

	if len(req.APIKey) < 10 {
		return nil, appErrors.NewValidationError("Invalid API key", "API key is too short")
	}

	if err := s.ValidateKey(ctx, req.Provider, req.APIKey, req.BaseURL); err != nil {
		return nil, err
	}

	encryptedKey, err := s.encryptor.Encrypt(req.APIKey)
	if err != nil {
		s.logger.Error("failed to encrypt API key",
			"error", err,
			"project_id", req.ProjectID,
			"provider", req.Provider,
		)
		return nil, appErrors.NewInternalError("Failed to secure API key", err)
	}

	keyPreview := credentialsDomain.MaskAPIKey(req.APIKey)

	existing, err := s.repo.GetByProjectAndProvider(ctx, req.ProjectID, req.Provider)
	if err != nil && !errors.Is(err, credentialsDomain.ErrCredentialNotFound) {
		s.logger.Error("failed to check existing credential",
			"error", err,
			"project_id", req.ProjectID,
			"provider", req.Provider,
		)
		return nil, appErrors.NewInternalError("Failed to check existing credential", err)
	}

	var credential *credentialsDomain.LLMProviderCredential

	if existing != nil {
		existing.EncryptedKey = encryptedKey
		existing.KeyPreview = keyPreview
		existing.BaseURL = req.BaseURL
		existing.UpdatedAt = time.Now()

		if err := s.repo.Update(ctx, existing); err != nil {
			s.logger.Error("failed to update credential",
				"error", err,
				"project_id", req.ProjectID,
				"provider", req.Provider,
			)
			return nil, appErrors.NewInternalError("Failed to update credential", err)
		}
		credential = existing

		s.logger.Info("LLM provider credential updated",
			"project_id", req.ProjectID,
			"provider", req.Provider,
		)
	} else {
		credential = &credentialsDomain.LLMProviderCredential{
			ID:           ulid.New(),
			ProjectID:    req.ProjectID,
			Provider:     req.Provider,
			EncryptedKey: encryptedKey,
			KeyPreview:   keyPreview,
			BaseURL:      req.BaseURL,
			CreatedBy:    req.CreatedBy,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.repo.Create(ctx, credential); err != nil {
			s.logger.Error("failed to create credential",
				"error", err,
				"project_id", req.ProjectID,
				"provider", req.Provider,
			)
			return nil, appErrors.NewInternalError("Failed to create credential", err)
		}

		s.logger.Info("LLM provider credential created",
			"project_id", req.ProjectID,
			"provider", req.Provider,
			"credential_id", credential.ID,
		)
	}

	return credential.ToResponse(), nil
}

// Get retrieves a credential by project and provider.
func (s *llmProviderCredentialService) Get(ctx context.Context, projectID ulid.ULID, provider credentialsDomain.LLMProvider) (*credentialsDomain.LLMProviderCredentialResponse, error) {
	credential, err := s.repo.GetByProjectAndProvider(ctx, projectID, provider)
	if err != nil {
		if errors.Is(err, credentialsDomain.ErrCredentialNotFound) {
			return nil, appErrors.NewNotFoundError("Credential not found")
		}
		return nil, appErrors.NewInternalError("Failed to retrieve credential", err)
	}
	return credential.ToResponse(), nil
}

// List retrieves all credentials for a project.
func (s *llmProviderCredentialService) List(ctx context.Context, projectID ulid.ULID) ([]*credentialsDomain.LLMProviderCredentialResponse, error) {
	credentials, err := s.repo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to list credentials", err)
	}

	responses := make([]*credentialsDomain.LLMProviderCredentialResponse, len(credentials))
	for i, cred := range credentials {
		responses[i] = cred.ToResponse()
	}
	return responses, nil
}

// Delete removes a credential.
func (s *llmProviderCredentialService) Delete(ctx context.Context, projectID ulid.ULID, provider credentialsDomain.LLMProvider) error {
	if err := s.repo.DeleteByProjectAndProvider(ctx, projectID, provider); err != nil {
		if errors.Is(err, credentialsDomain.ErrCredentialNotFound) {
			return appErrors.NewNotFoundError("Credential not found")
		}
		return appErrors.NewInternalError("Failed to delete credential", err)
	}

	s.logger.Info("LLM provider credential deleted",
		"project_id", projectID,
		"provider", provider,
	)
	return nil
}

// GetDecrypted retrieves the decrypted key configuration (internal use only).
func (s *llmProviderCredentialService) GetDecrypted(ctx context.Context, projectID ulid.ULID, provider credentialsDomain.LLMProvider) (*credentialsDomain.DecryptedKeyConfig, error) {
	credential, err := s.repo.GetByProjectAndProvider(ctx, projectID, provider)
	if err != nil {
		if errors.Is(err, credentialsDomain.ErrCredentialNotFound) {
			return nil, credentialsDomain.ErrCredentialNotFound
		}
		return nil, err
	}

	decryptedKey, err := s.encryptor.Decrypt(credential.EncryptedKey)
	if err != nil {
		s.logger.Error("failed to decrypt API key",
			"error", err,
			"credential_id", credential.ID,
		)
		return nil, credentialsDomain.ErrDecryptionFailed
	}

	config := &credentialsDomain.DecryptedKeyConfig{
		Provider: provider,
		APIKey:   decryptedKey,
	}

	if credential.BaseURL != nil {
		config.BaseURL = *credential.BaseURL
	}

	return config, nil
}

// GetExecutionConfig returns the decrypted key config for a project/provider.
// No environment fallback - project credentials are required.
func (s *llmProviderCredentialService) GetExecutionConfig(ctx context.Context, projectID ulid.ULID, provider credentialsDomain.LLMProvider) (*credentialsDomain.DecryptedKeyConfig, error) {
	if !provider.IsValid() {
		return nil, credentialsDomain.NewInvalidProviderError(string(provider))
	}

	config, err := s.GetDecrypted(ctx, projectID, provider)
	if err != nil {
		if errors.Is(err, credentialsDomain.ErrCredentialNotFound) {
			return nil, credentialsDomain.NewNoKeyConfiguredError(string(provider))
		}
		s.logger.Error("failed to get project credential",
			"error", err,
			"project_id", projectID,
			"provider", provider,
		)
		return nil, err
	}

	return config, nil
}

// ValidateKey validates an API key with the provider.
func (s *llmProviderCredentialService) ValidateKey(ctx context.Context, provider credentialsDomain.LLMProvider, apiKey string, baseURL *string) error {
	switch provider {
	case credentialsDomain.ProviderOpenAI:
		return s.validateOpenAIKey(ctx, apiKey, baseURL)
	case credentialsDomain.ProviderAnthropic:
		return s.validateAnthropicKey(ctx, apiKey, baseURL)
	default:
		return credentialsDomain.NewInvalidProviderError(string(provider))
	}
}

// validateOpenAIKey validates an OpenAI API key by calling the models endpoint.
func (s *llmProviderCredentialService) validateOpenAIKey(ctx context.Context, apiKey string, baseURL *string) error {
	endpoint := "https://api.openai.com/v1/models"
	if baseURL != nil && *baseURL != "" {
		endpoint = strings.TrimSuffix(*baseURL, "/") + "/models"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return appErrors.NewInternalError("Failed to create validation request", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return appErrors.NewValidationError("API key validation failed", "Could not connect to OpenAI: "+err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return appErrors.NewValidationError("Invalid API key", "OpenAI rejected the API key")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return appErrors.NewValidationError("API key validation failed", fmt.Sprintf("OpenAI returned status %d: %s", resp.StatusCode, string(body)))
	}

	return nil
}

// validateAnthropicKey validates an Anthropic API key by making a minimal API call.
func (s *llmProviderCredentialService) validateAnthropicKey(ctx context.Context, apiKey string, baseURL *string) error {
	endpoint := "https://api.anthropic.com/v1/messages"
	if baseURL != nil && *baseURL != "" {
		endpoint = strings.TrimSuffix(*baseURL, "/") + "/v1/messages"
	}

	// Create a minimal request to validate the key
	// Using max_tokens=1 to minimize cost
	reqBody := `{"model":"claude-3-haiku-20240307","max_tokens":1,"messages":[{"role":"user","content":"hi"}]}`

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(reqBody))
	if err != nil {
		return appErrors.NewInternalError("Failed to create validation request", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return appErrors.NewValidationError("API key validation failed", "Could not connect to Anthropic: "+err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return appErrors.NewValidationError("Invalid API key", "Anthropic rejected the API key")
	}

	// 400 or 200 both indicate the key is valid (400 might be model access)
	// We mainly care about 401/403 which indicate invalid key
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest {
		// For 400, check if it's an auth error or just a request error
		if resp.StatusCode == http.StatusBadRequest {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			var errResp struct {
				Error struct {
					Type string `json:"type"`
				} `json:"error"`
			}
			if json.Unmarshal(body, &errResp) == nil {
				if errResp.Error.Type == "authentication_error" {
					return appErrors.NewValidationError("Invalid API key", "Anthropic authentication failed")
				}
			}
		}
		return nil // Key is valid
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return appErrors.NewValidationError("API key validation failed", fmt.Sprintf("Anthropic returned status %d: %s", resp.StatusCode, string(body)))
}
