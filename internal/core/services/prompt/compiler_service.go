package prompt

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	promptDomain "brokle/internal/core/domain/prompt"
)

// Variable pattern matches Mustache-style variables: {{variable_name}}
// Variables must start with a letter and contain only alphanumeric characters and underscores
var variablePattern = regexp.MustCompile(`\{\{([a-zA-Z][a-zA-Z0-9_]*)\}\}`)

// compilerService implements promptDomain.CompilerService
type compilerService struct{}

// NewCompilerService creates a new compiler service instance
func NewCompilerService() promptDomain.CompilerService {
	return &compilerService{}
}

// ExtractVariables extracts variable names from a template
func (s *compilerService) ExtractVariables(template interface{}, promptType promptDomain.PromptType) ([]string, error) {
	switch promptType {
	case promptDomain.PromptTypeText:
		return s.extractTextVariables(template)
	case promptDomain.PromptTypeChat:
		return s.extractChatVariables(template)
	default:
		return nil, promptDomain.ErrInvalidPromptType
	}
}

// extractTextVariables extracts variables from a text template
func (s *compilerService) extractTextVariables(template interface{}) ([]string, error) {
	if raw, ok := template.(json.RawMessage); ok {
		var textTemplate promptDomain.TextTemplate
		if err := json.Unmarshal(raw, &textTemplate); err != nil {
			return nil, fmt.Errorf("%w: %v", promptDomain.ErrInvalidTemplateFormat, err)
		}
		return s.extractFromString(textTemplate.Content), nil
	}

	if m, ok := template.(map[string]interface{}); ok {
		if content, ok := m["content"].(string); ok {
			return s.extractFromString(content), nil
		}
		return nil, promptDomain.NewInvalidTemplateError("text template must have 'content' field")
	}

	if str, ok := template.(string); ok {
		return s.extractFromString(str), nil
	}

	return nil, promptDomain.NewInvalidTemplateError("unsupported template format for text type")
}

// extractChatVariables extracts variables from a chat template
func (s *compilerService) extractChatVariables(template interface{}) ([]string, error) {
	if raw, ok := template.(json.RawMessage); ok {
		var chatTemplate promptDomain.ChatTemplate
		if err := json.Unmarshal(raw, &chatTemplate); err != nil {
			return nil, fmt.Errorf("%w: %v", promptDomain.ErrInvalidTemplateFormat, err)
		}
		return s.extractFromMessages(chatTemplate.Messages), nil
	}

	if m, ok := template.(map[string]interface{}); ok {
		messages, err := s.parseMessagesFromMap(m)
		if err != nil {
			return nil, err
		}
		return s.extractFromMessages(messages), nil
	}

	return nil, promptDomain.NewInvalidTemplateError("unsupported template format for chat type")
}

// parseMessagesFromMap parses chat messages from a map structure
func (s *compilerService) parseMessagesFromMap(m map[string]interface{}) ([]promptDomain.ChatMessage, error) {
	messagesRaw, ok := m["messages"]
	if !ok {
		return nil, promptDomain.NewInvalidTemplateError("chat template must have 'messages' field")
	}

	messagesSlice, ok := messagesRaw.([]interface{})
	if !ok {
		return nil, promptDomain.NewInvalidTemplateError("messages must be an array")
	}

	var messages []promptDomain.ChatMessage
	for _, msgRaw := range messagesSlice {
		msgMap, ok := msgRaw.(map[string]interface{})
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

// extractFromString extracts unique variable names from a string
func (s *compilerService) extractFromString(content string) []string {
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

// extractFromMessages extracts variables from chat messages
func (s *compilerService) extractFromMessages(messages []promptDomain.ChatMessage) []string {
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

// Compile compiles a template with variable substitution
func (s *compilerService) Compile(template interface{}, promptType promptDomain.PromptType, variables map[string]string) (interface{}, error) {
	switch promptType {
	case promptDomain.PromptTypeText:
		return s.compileText(template, variables)
	case promptDomain.PromptTypeChat:
		return s.compileChat(template, variables)
	default:
		return nil, promptDomain.ErrInvalidPromptType
	}
}

// compileText compiles a text template
func (s *compilerService) compileText(template interface{}, variables map[string]string) (string, error) {
	var content string

	if raw, ok := template.(json.RawMessage); ok {
		var textTemplate promptDomain.TextTemplate
		if err := json.Unmarshal(raw, &textTemplate); err != nil {
			return "", fmt.Errorf("%w: %v", promptDomain.ErrInvalidTemplateFormat, err)
		}
		content = textTemplate.Content
	} else if m, ok := template.(map[string]interface{}); ok {
		if c, ok := m["content"].(string); ok {
			content = c
		} else {
			return "", promptDomain.NewInvalidTemplateError("text template must have 'content' field")
		}
	} else if str, ok := template.(string); ok {
		content = str
	} else {
		return "", promptDomain.NewInvalidTemplateError("unsupported template format")
	}

	return s.CompileText(content, variables)
}

// CompileText compiles a text string with variable substitution
func (s *compilerService) CompileText(template string, variables map[string]string) (string, error) {
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

// compileChat compiles a chat template
func (s *compilerService) compileChat(template interface{}, variables map[string]string) ([]promptDomain.ChatMessage, error) {
	var messages []promptDomain.ChatMessage

	if raw, ok := template.(json.RawMessage); ok {
		var chatTemplate promptDomain.ChatTemplate
		if err := json.Unmarshal(raw, &chatTemplate); err != nil {
			return nil, fmt.Errorf("%w: %v", promptDomain.ErrInvalidTemplateFormat, err)
		}
		messages = chatTemplate.Messages
	} else if m, ok := template.(map[string]interface{}); ok {
		var err error
		messages, err = s.parseMessagesFromMap(m)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, promptDomain.NewInvalidTemplateError("unsupported template format for chat")
	}

	return s.CompileChat(messages, variables)
}

// CompileChat compiles chat messages with variable substitution
func (s *compilerService) CompileChat(messages []promptDomain.ChatMessage, variables map[string]string) ([]promptDomain.ChatMessage, error) {
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

// ValidateTemplate validates a template structure
func (s *compilerService) ValidateTemplate(template interface{}, promptType promptDomain.PromptType) error {
	switch promptType {
	case promptDomain.PromptTypeText:
		return s.validateTextTemplate(template)
	case promptDomain.PromptTypeChat:
		return s.validateChatTemplate(template)
	default:
		return promptDomain.ErrInvalidPromptType
	}
}

// validateTextTemplate validates a text template
func (s *compilerService) validateTextTemplate(template interface{}) error {
	if raw, ok := template.(json.RawMessage); ok {
		var textTemplate promptDomain.TextTemplate
		if err := json.Unmarshal(raw, &textTemplate); err != nil {
			return fmt.Errorf("%w: %v", promptDomain.ErrInvalidTemplateFormat, err)
		}
		if textTemplate.Content == "" {
			return promptDomain.NewInvalidTemplateError("content cannot be empty")
		}
		return nil
	}

	if m, ok := template.(map[string]interface{}); ok {
		content, ok := m["content"].(string)
		if !ok {
			return promptDomain.NewInvalidTemplateError("text template must have 'content' field")
		}
		if content == "" {
			return promptDomain.NewInvalidTemplateError("content cannot be empty")
		}
		return nil
	}

	if str, ok := template.(string); ok {
		if str == "" {
			return promptDomain.NewInvalidTemplateError("content cannot be empty")
		}
		return nil
	}

	return promptDomain.NewInvalidTemplateError("unsupported template format")
}

// validateChatTemplate validates a chat template
func (s *compilerService) validateChatTemplate(template interface{}) error {
	if raw, ok := template.(json.RawMessage); ok {
		var chatTemplate promptDomain.ChatTemplate
		if err := json.Unmarshal(raw, &chatTemplate); err != nil {
			return fmt.Errorf("%w: %v", promptDomain.ErrInvalidTemplateFormat, err)
		}
		if len(chatTemplate.Messages) == 0 {
			return promptDomain.NewInvalidTemplateError("messages cannot be empty")
		}
		return s.validateMessages(chatTemplate.Messages)
	}

	if m, ok := template.(map[string]interface{}); ok {
		messages, err := s.parseMessagesFromMap(m)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return promptDomain.NewInvalidTemplateError("messages cannot be empty")
		}
		return s.validateMessages(messages)
	}

	return promptDomain.NewInvalidTemplateError("unsupported template format for chat")
}

// validateMessages validates chat messages
func (s *compilerService) validateMessages(messages []promptDomain.ChatMessage) error {
	validRoles := map[string]bool{"system": true, "user": true, "assistant": true}
	validTypes := map[string]bool{"message": true, "placeholder": true, "": true}

	for i, msg := range messages {
		if !validTypes[msg.Type] {
			return promptDomain.NewInvalidTemplateError(fmt.Sprintf("invalid message type at index %d: %s", i, msg.Type))
		}

		if msg.Type == "message" || msg.Type == "" {
			if !validRoles[msg.Role] {
				return promptDomain.NewInvalidTemplateError(fmt.Sprintf("invalid role at index %d: %s", i, msg.Role))
			}
			if msg.Content == "" {
				return promptDomain.NewInvalidTemplateError(fmt.Sprintf("empty content at index %d", i))
			}
		}

		if msg.Type == "placeholder" && msg.Name == "" {
			return promptDomain.NewInvalidTemplateError(fmt.Sprintf("placeholder at index %d must have a name", i))
		}
	}

	return nil
}

// ValidateVariables checks that all required variables are provided
func (s *compilerService) ValidateVariables(required []string, provided map[string]string) error {
	var missing []string
	for _, v := range required {
		if _, ok := provided[v]; !ok {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		return promptDomain.NewVariableMissingError(strings.Join(missing, ", "))
	}

	return nil
}
