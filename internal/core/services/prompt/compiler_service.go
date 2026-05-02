// Package prompt implements prompt management: templates, versions, labels, compilation, and execution routing.
package prompt

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	promptDomain "brokle/internal/core/domain/prompt"
	"brokle/internal/core/services/prompt/dialects"
	appErrors "brokle/pkg/errors"
)

// Variable pattern matches Mustache-style variables: {{variable_name}}
// Variables must start with a letter and contain only alphanumeric characters and underscores
var variablePattern = regexp.MustCompile(`\{\{([a-zA-Z][a-zA-Z0-9_]*)\}\}`)

type CompilerService struct {
	registry promptDomain.DialectRegistry
}

func NewCompilerService() *CompilerService {
	return &CompilerService{
		registry: dialects.NewRegistry(),
	}
}

func (s *CompilerService) ExtractVariables(template any, promptType promptDomain.PromptType) ([]string, error) {
	switch promptType {
	case promptDomain.PromptTypeText:
		return s.extractTextVariables(template)
	case promptDomain.PromptTypeChat:
		return s.extractChatVariables(template)
	default:
		return nil, appErrors.InvalidParam("type", "must be 'text' or 'chat'")
	}
}

func (s *CompilerService) extractTextVariables(template any) ([]string, error) {
	if raw, ok := template.(json.RawMessage); ok {
		var textTemplate promptDomain.TextTemplate
		if err := json.Unmarshal(raw, &textTemplate); err != nil {
			return nil, appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		return s.extractFromString(textTemplate.Content), nil
	}

	if m, ok := template.(map[string]any); ok {
		if content, ok := m["content"].(string); ok {
			return s.extractFromString(content), nil
		}
		return nil, appErrors.InvalidParam("template", "text template must have 'content' field")
	}

	if str, ok := template.(string); ok {
		return s.extractFromString(str), nil
	}

	return nil, appErrors.InvalidParam("template", "unsupported template format for text type")
}

func (s *CompilerService) extractChatVariables(template any) ([]string, error) {
	if raw, ok := template.(json.RawMessage); ok {
		var chatTemplate promptDomain.ChatTemplate
		if err := json.Unmarshal(raw, &chatTemplate); err != nil {
			return nil, appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		return s.extractFromMessages(chatTemplate.Messages), nil
	}

	if m, ok := template.(map[string]any); ok {
		messages, err := s.parseMessagesFromMap(m)
		if err != nil {
			return nil, err
		}
		return s.extractFromMessages(messages), nil
	}

	return nil, appErrors.InvalidParam("template", "unsupported template format for chat type")
}

func (s *CompilerService) parseMessagesFromMap(m map[string]any) ([]promptDomain.ChatMessage, error) {
	messagesRaw, ok := m["messages"]
	if !ok {
		return nil, appErrors.InvalidParam("template", "chat template must have 'messages' field")
	}

	messagesSlice, ok := messagesRaw.([]any)
	if !ok {
		return nil, appErrors.InvalidParam("template", "messages must be an array")
	}

	var messages []promptDomain.ChatMessage
	for _, msgRaw := range messagesSlice {
		msgMap, ok := msgRaw.(map[string]any)
		if !ok {
			continue
		}

		msg := promptDomain.ChatMessage{}
		if t, ok := msgMap["type"].(string); ok {
			msg.Type = t
		}
		if r, ok := msgMap["role"].(string); ok {
			msg.Role = r
		}
		if c, ok := msgMap["content"].(string); ok {
			msg.Content = c
		}
		if n, ok := msgMap["name"].(string); ok {
			msg.Name = n
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (s *CompilerService) extractFromString(content string) []string {
	matches := variablePattern.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	var vars []string

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			seen[match[1]] = true
			vars = append(vars, match[1])
		}
	}

	sort.Strings(vars)
	return vars
}

func (s *CompilerService) extractFromMessages(messages []promptDomain.ChatMessage) []string {
	seen := make(map[string]bool)
	var vars []string

	for _, msg := range messages {
		for _, v := range s.extractFromString(msg.Content) {
			if !seen[v] {
				seen[v] = true
				vars = append(vars, v)
			}
		}

		// Placeholders are also treated as variables
		if msg.Type == "placeholder" && msg.Name != "" && !seen[msg.Name] {
			seen[msg.Name] = true
			vars = append(vars, msg.Name)
		}
	}

	sort.Strings(vars)
	return vars
}

func (s *CompilerService) Compile(template any, promptType promptDomain.PromptType, variables map[string]string) (any, error) {
	switch promptType {
	case promptDomain.PromptTypeText:
		return s.compileText(template, variables)
	case promptDomain.PromptTypeChat:
		return s.compileChat(template, variables)
	default:
		return nil, appErrors.InvalidParam("type", "must be 'text' or 'chat'")
	}
}

func (s *CompilerService) compileText(template any, variables map[string]string) (string, error) {
	var content string

	if raw, ok := template.(json.RawMessage); ok {
		var textTemplate promptDomain.TextTemplate
		if err := json.Unmarshal(raw, &textTemplate); err != nil {
			return "", appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		content = textTemplate.Content
	} else if m, ok := template.(map[string]any); ok {
		if c, ok := m["content"].(string); ok {
			content = c
		} else {
			return "", appErrors.InvalidParam("template", "text template must have 'content' field")
		}
	} else if str, ok := template.(string); ok {
		content = str
	} else {
		return "", appErrors.InvalidParam("template", "unsupported template format")
	}

	return s.CompileText(content, variables)
}

func (s *CompilerService) CompileText(template string, variables map[string]string) (string, error) {
	required := s.extractFromString(template)

	if err := s.ValidateVariables(required, variables); err != nil {
		return "", err
	}

	result := variablePattern.ReplaceAllStringFunc(template, func(match string) string {
		varName := match[2 : len(match)-2]
		if val, ok := variables[varName]; ok {
			return val
		}
		return match
	})

	return result, nil
}

func (s *CompilerService) compileChat(template any, variables map[string]string) ([]promptDomain.ChatMessage, error) {
	var messages []promptDomain.ChatMessage

	if raw, ok := template.(json.RawMessage); ok {
		var chatTemplate promptDomain.ChatTemplate
		if err := json.Unmarshal(raw, &chatTemplate); err != nil {
			return nil, appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		messages = chatTemplate.Messages
	} else if m, ok := template.(map[string]any); ok {
		var err error
		messages, err = s.parseMessagesFromMap(m)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, appErrors.InvalidParam("template", "unsupported template format for chat")
	}

	return s.CompileChat(messages, variables)
}

func (s *CompilerService) CompileChat(messages []promptDomain.ChatMessage, variables map[string]string) ([]promptDomain.ChatMessage, error) {
	required := s.extractFromMessages(messages)

	if err := s.ValidateVariables(required, variables); err != nil {
		return nil, err
	}

	result := make([]promptDomain.ChatMessage, 0, len(messages))
	for _, msg := range messages {
		// Placeholders are replaced by user-provided content
		if msg.Type == "placeholder" {
			if val, ok := variables[msg.Name]; ok && val != "" {
				result = append(result, promptDomain.ChatMessage{
					Type:    "message",
					Role:    "user",
					Content: val,
				})
			}
			continue
		}

		compiledContent := variablePattern.ReplaceAllStringFunc(msg.Content, func(match string) string {
			varName := match[2 : len(match)-2]
			if val, ok := variables[varName]; ok {
				return val
			}
			return match
		})

		result = append(result, promptDomain.ChatMessage{
			Type:    msg.Type,
			Role:    msg.Role,
			Content: compiledContent,
			Name:    msg.Name,
		})
	}

	return result, nil
}

func (s *CompilerService) ValidateTemplate(template any, promptType promptDomain.PromptType) error {
	switch promptType {
	case promptDomain.PromptTypeText:
		return s.validateTextTemplate(template)
	case promptDomain.PromptTypeChat:
		return s.validateChatTemplate(template)
	default:
		return appErrors.InvalidParam("type", "must be 'text' or 'chat'")
	}
}

func (s *CompilerService) validateTextTemplate(template any) error {
	if raw, ok := template.(json.RawMessage); ok {
		var textTemplate promptDomain.TextTemplate
		if err := json.Unmarshal(raw, &textTemplate); err != nil {
			return appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		if textTemplate.Content == "" {
			return appErrors.InvalidParam("template", "content cannot be empty")
		}
		return nil
	}

	if m, ok := template.(map[string]any); ok {
		content, ok := m["content"].(string)
		if !ok {
			return appErrors.InvalidParam("template", "text template must have 'content' field")
		}
		if content == "" {
			return appErrors.InvalidParam("template", "content cannot be empty")
		}
		return nil
	}

	if str, ok := template.(string); ok {
		if str == "" {
			return appErrors.InvalidParam("template", "content cannot be empty")
		}
		return nil
	}

	return appErrors.InvalidParam("template", "unsupported template format")
}

func (s *CompilerService) validateChatTemplate(template any) error {
	if raw, ok := template.(json.RawMessage); ok {
		var chatTemplate promptDomain.ChatTemplate
		if err := json.Unmarshal(raw, &chatTemplate); err != nil {
			return appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		if len(chatTemplate.Messages) == 0 {
			return appErrors.InvalidParam("template", "messages cannot be empty")
		}
		return s.validateMessages(chatTemplate.Messages)
	}

	if m, ok := template.(map[string]any); ok {
		messages, err := s.parseMessagesFromMap(m)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return appErrors.InvalidParam("template", "messages cannot be empty")
		}
		return s.validateMessages(messages)
	}

	return appErrors.InvalidParam("template", "unsupported template format for chat")
}

func (s *CompilerService) validateMessages(messages []promptDomain.ChatMessage) error {
	validRoles := map[string]bool{"system": true, "user": true, "assistant": true}
	validTypes := map[string]bool{"message": true, "placeholder": true, "": true}

	for i, msg := range messages {
		if !validTypes[msg.Type] {
			return appErrors.InvalidParam("template", fmt.Sprintf("invalid message type at index %d: %s", i, msg.Type))
		}

		if msg.Type == "message" || msg.Type == "" {
			if !validRoles[msg.Role] {
				return appErrors.InvalidParam("template", fmt.Sprintf("invalid role at index %d: %s", i, msg.Role))
			}
			if msg.Content == "" {
				return appErrors.InvalidParam("template", fmt.Sprintf("empty content at index %d", i))
			}
		}

		if msg.Type == "placeholder" && msg.Name == "" {
			return appErrors.InvalidParam("template", fmt.Sprintf("placeholder at index %d must have a name", i))
		}
	}

	return nil
}

func (s *CompilerService) ValidateVariables(required []string, provided map[string]string) error {
	var missing []string
	for _, v := range required {
		if _, ok := provided[v]; !ok {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		return appErrors.InvalidParam("variables", fmt.Sprintf("required variable missing: %s", strings.Join(missing, ", ")))
	}

	return nil
}

func (s *CompilerService) GetDialectRegistry() promptDomain.DialectRegistry {
	return s.registry
}

func (s *CompilerService) DetectDialect(template any, promptType promptDomain.PromptType) (promptDomain.TemplateDialect, error) {
	content, err := s.extractContentForDetection(template, promptType)
	if err != nil {
		return promptDomain.DialectSimple, err
	}
	return s.registry.Detect(content), nil
}

func (s *CompilerService) extractContentForDetection(template any, promptType promptDomain.PromptType) (string, error) {
	switch promptType {
	case promptDomain.PromptTypeText:
		return s.extractTextContent(template)
	case promptDomain.PromptTypeChat:
		return s.extractChatContent(template)
	default:
		return "", appErrors.InvalidParam("type", "must be 'text' or 'chat'")
	}
}

func (s *CompilerService) extractTextContent(template any) (string, error) {
	if raw, ok := template.(json.RawMessage); ok {
		var textTemplate promptDomain.TextTemplate
		if err := json.Unmarshal(raw, &textTemplate); err != nil {
			return "", appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		return textTemplate.Content, nil
	}
	if m, ok := template.(map[string]any); ok {
		if content, ok := m["content"].(string); ok {
			return content, nil
		}
		return "", appErrors.InvalidParam("template", "text template must have 'content' field")
	}
	if str, ok := template.(string); ok {
		return str, nil
	}
	return "", appErrors.InvalidParam("template", "unsupported template format")
}

func (s *CompilerService) extractChatContent(template any) (string, error) {
	var messages []promptDomain.ChatMessage

	if raw, ok := template.(json.RawMessage); ok {
		var chatTemplate promptDomain.ChatTemplate
		if err := json.Unmarshal(raw, &chatTemplate); err != nil {
			return "", appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		messages = chatTemplate.Messages
	} else if m, ok := template.(map[string]any); ok {
		var err error
		messages, err = s.parseMessagesFromMap(m)
		if err != nil {
			return "", err
		}
	} else {
		return "", appErrors.InvalidParam("template", "unsupported template format for chat")
	}

	// Concatenate all message content for dialect detection
	var builder strings.Builder
	for _, msg := range messages {
		builder.WriteString(msg.Content)
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

func (s *CompilerService) ValidateSyntax(template any, promptType promptDomain.PromptType, dialect promptDomain.TemplateDialect) (*promptDomain.ValidationResult, error) {
	// Handle auto-detection
	if dialect == promptDomain.DialectAuto || dialect == "" {
		detected, err := s.DetectDialect(template, promptType)
		if err != nil {
			return nil, err
		}
		dialect = detected
	}

	// Get the compiler for this dialect
	compiler, err := s.registry.Get(dialect)
	if err != nil {
		return nil, err
	}

	// Extract content based on prompt type
	switch promptType {
	case promptDomain.PromptTypeText:
		content, err := s.extractTextContent(template)
		if err != nil {
			return nil, err
		}
		return compiler.Validate(content)

	case promptDomain.PromptTypeChat:
		content, err := s.extractChatContent(template)
		if err != nil {
			return nil, err
		}
		return compiler.Validate(content)

	default:
		return nil, appErrors.InvalidParam("type", "must be 'text' or 'chat'")
	}
}

func (s *CompilerService) ExtractVariablesWithDialect(template any, promptType promptDomain.PromptType, dialect promptDomain.TemplateDialect) ([]string, error) {
	// Handle auto-detection
	if dialect == promptDomain.DialectAuto || dialect == "" {
		detected, err := s.DetectDialect(template, promptType)
		if err != nil {
			return nil, err
		}
		dialect = detected
	}

	// Get the compiler for this dialect
	compiler, err := s.registry.Get(dialect)
	if err != nil {
		return nil, err
	}

	switch promptType {
	case promptDomain.PromptTypeText:
		content, err := s.extractTextContent(template)
		if err != nil {
			return nil, err
		}
		return compiler.ExtractVariables(content)

	case promptDomain.PromptTypeChat:
		return s.extractChatVariablesWithDialect(template, compiler)

	default:
		return nil, appErrors.InvalidParam("type", "must be 'text' or 'chat'")
	}
}

func (s *CompilerService) extractChatVariablesWithDialect(template any, compiler promptDomain.DialectCompiler) ([]string, error) {
	var messages []promptDomain.ChatMessage

	if raw, ok := template.(json.RawMessage); ok {
		var chatTemplate promptDomain.ChatTemplate
		if err := json.Unmarshal(raw, &chatTemplate); err != nil {
			return nil, appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		messages = chatTemplate.Messages
	} else if m, ok := template.(map[string]any); ok {
		var err error
		messages, err = s.parseMessagesFromMap(m)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, appErrors.InvalidParam("template", "unsupported template format for chat")
	}

	seen := make(map[string]bool)
	var vars []string

	for _, msg := range messages {
		// Extract variables from message content
		extracted, err := compiler.ExtractVariables(msg.Content)
		if err != nil {
			return nil, err
		}
		for _, v := range extracted {
			if !seen[v] {
				seen[v] = true
				vars = append(vars, v)
			}
		}

		// Placeholders are also treated as variables
		if msg.Type == "placeholder" && msg.Name != "" && !seen[msg.Name] {
			seen[msg.Name] = true
			vars = append(vars, msg.Name)
		}
	}

	sort.Strings(vars)
	return vars, nil
}

func (s *CompilerService) CompileWithDialect(template any, promptType promptDomain.PromptType, variables map[string]any, dialect promptDomain.TemplateDialect) (any, error) {
	// Handle auto-detection
	if dialect == promptDomain.DialectAuto || dialect == "" {
		detected, err := s.DetectDialect(template, promptType)
		if err != nil {
			return nil, err
		}
		dialect = detected
	}

	// Get the compiler for this dialect
	compiler, err := s.registry.Get(dialect)
	if err != nil {
		return nil, err
	}

	switch promptType {
	case promptDomain.PromptTypeText:
		return s.compileTextWithDialect(template, variables, compiler)
	case promptDomain.PromptTypeChat:
		return s.compileChatWithDialect(template, variables, compiler)
	default:
		return nil, appErrors.InvalidParam("type", "must be 'text' or 'chat'")
	}
}

func (s *CompilerService) compileTextWithDialect(template any, variables map[string]any, compiler promptDomain.DialectCompiler) (string, error) {
	content, err := s.extractTextContent(template)
	if err != nil {
		return "", err
	}
	return compiler.Compile(content, variables)
}

func (s *CompilerService) compileChatWithDialect(template any, variables map[string]any, compiler promptDomain.DialectCompiler) ([]promptDomain.ChatMessage, error) {
	var messages []promptDomain.ChatMessage

	if raw, ok := template.(json.RawMessage); ok {
		var chatTemplate promptDomain.ChatTemplate
		if err := json.Unmarshal(raw, &chatTemplate); err != nil {
			return nil, appErrors.InvalidParam("template", fmt.Sprintf("invalid template format: %v", err))
		}
		messages = chatTemplate.Messages
	} else if m, ok := template.(map[string]any); ok {
		var err error
		messages, err = s.parseMessagesFromMap(m)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, appErrors.InvalidParam("template", "unsupported template format for chat")
	}

	result := make([]promptDomain.ChatMessage, 0, len(messages))
	for _, msg := range messages {
		// Handle placeholders
		if msg.Type == "placeholder" {
			if val, ok := variables[msg.Name]; ok {
				// Handle different value types for placeholders
				injectedMessages := s.handlePlaceholderValue(val)
				result = append(result, injectedMessages...)
			}
			continue
		}

		// Compile message content using dialect compiler
		compiledContent, err := compiler.Compile(msg.Content, variables)
		if err != nil {
			return nil, err
		}

		result = append(result, promptDomain.ChatMessage{
			Type:    msg.Type,
			Role:    msg.Role,
			Content: compiledContent,
			Name:    msg.Name,
		})
	}

	return result, nil
}

// handlePlaceholderValue converts a placeholder value into one or more ChatMessages.
// Supports:
// - string: creates a single user message
// - []ChatMessage: injects messages directly (for history injection)
// - []map[string]any: converts to ChatMessages (from JSON)
// - []any: handles mixed arrays from JSON unmarshaling
// - other: JSON serializes and creates a single user message
func (s *CompilerService) handlePlaceholderValue(val any) []promptDomain.ChatMessage {
	switch v := val.(type) {
	case string:
		// Simple string value becomes a single user message
		if v == "" {
			return nil
		}
		return []promptDomain.ChatMessage{{
			Type:    "message",
			Role:    "user",
			Content: v,
		}}

	case []promptDomain.ChatMessage:
		// Direct ChatMessage array - inject as-is (for history)
		return v

	case []map[string]any:
		// Array of maps from JSON - convert to ChatMessages
		return s.convertMapsToMessages(v)

	case []any:
		// Mixed array from JSON unmarshaling - convert each element
		return s.convertInterfaceArrayToMessages(v)

	default:
		// Complex type - serialize to JSON
		data, err := json.Marshal(v)
		if err != nil || len(data) == 0 || string(data) == "null" {
			return nil
		}
		return []promptDomain.ChatMessage{{
			Type:    "message",
			Role:    "user",
			Content: string(data),
		}}
	}
}

func (s *CompilerService) convertMapsToMessages(maps []map[string]any) []promptDomain.ChatMessage {
	result := make([]promptDomain.ChatMessage, 0, len(maps))
	for _, m := range maps {
		msg := s.mapToChatMessage(m)
		if msg.Content != "" || msg.Type == "placeholder" {
			result = append(result, msg)
		}
	}
	return result
}

func (s *CompilerService) convertInterfaceArrayToMessages(arr []any) []promptDomain.ChatMessage {
	result := make([]promptDomain.ChatMessage, 0, len(arr))
	for _, item := range arr {
		switch v := item.(type) {
		case map[string]any:
			msg := s.mapToChatMessage(v)
			if msg.Content != "" || msg.Type == "placeholder" {
				result = append(result, msg)
			}
		case promptDomain.ChatMessage:
			result = append(result, v)
		case string:
			// String in array becomes user message
			if v != "" {
				result = append(result, promptDomain.ChatMessage{
					Type:    "message",
					Role:    "user",
					Content: v,
				})
			}
		}
	}
	return result
}

func (s *CompilerService) mapToChatMessage(m map[string]any) promptDomain.ChatMessage {
	msg := promptDomain.ChatMessage{
		Type: "message", // Default type
	}

	if t, ok := m["type"].(string); ok {
		msg.Type = t
	}
	if role, ok := m["role"].(string); ok {
		msg.Role = role
	}
	if content, ok := m["content"].(string); ok {
		msg.Content = content
	}
	if name, ok := m["name"].(string); ok {
		msg.Name = name
	}

	return msg
}
