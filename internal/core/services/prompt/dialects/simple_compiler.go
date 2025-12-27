package dialects

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	promptDomain "brokle/internal/core/domain/prompt"
)

// Variable pattern matches simple Mustache-style variables: {{variable_name}}
// Variables must start with a letter and contain only alphanumeric characters and underscores
var simpleVarPattern = regexp.MustCompile(`\{\{([a-zA-Z][a-zA-Z0-9_]*)\}\}`)

type simpleCompiler struct{}

func NewSimpleCompiler() promptDomain.DialectCompiler {
	return &simpleCompiler{}
}

func (c *simpleCompiler) Dialect() promptDomain.TemplateDialect {
	return promptDomain.DialectSimple
}

func (c *simpleCompiler) ExtractVariables(content string) ([]string, error) {
	matches := simpleVarPattern.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	var vars []string

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			seen[match[1]] = true
			vars = append(vars, match[1])
		}
	}

	sort.Strings(vars)
	return vars, nil
}

// Compile renders with string substitution; non-string values are JSON-serialized.
func (c *simpleCompiler) Compile(content string, variables map[string]any) (string, error) {
	if len(content) > promptDomain.MaxTemplateSize {
		return "", promptDomain.NewTemplateTooLargeError(len(content), promptDomain.MaxTemplateSize)
	}

	result := simpleVarPattern.ReplaceAllStringFunc(content, func(match string) string {
		varName := match[2 : len(match)-2] // Extract variable name from {{var}}
		if val, ok := variables[varName]; ok {
			return anyToString(val)
		}
		// Preserve missing variables (don't replace with empty)
		return match
	})

	return result, nil
}

func (c *simpleCompiler) Validate(content string) (*promptDomain.ValidationResult, error) {
	result := promptDomain.NewValidationResult(true, promptDomain.DialectSimple)

	if len(content) > promptDomain.MaxTemplateSize {
		result.AddError(0, 0, "template exceeds maximum size limit", promptDomain.ErrCodeTemplateTooLarge)
		return result, nil
	}

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		c.validateLine(result, line, lineNum+1)
	}

	vars, _ := c.ExtractVariables(content)
	if len(vars) > promptDomain.MaxVariables {
		result.AddWarning(0, 0, "template has many variables, consider simplifying", promptDomain.WarnCodeUnusedVariable)
	}

	return result, nil
}

func (c *simpleCompiler) validateLine(result *promptDomain.ValidationResult, line string, lineNum int) {
	openCount := strings.Count(line, "{{")
	closeCount := strings.Count(line, "}}")

	if openCount > closeCount {
		col := strings.Index(line, "{{") + 1
		result.AddError(lineNum, col, "unmatched opening braces '{{'", promptDomain.ErrCodeUnmatchedOpening)
	} else if closeCount > openCount {
		col := strings.LastIndex(line, "}}") + 1
		result.AddError(lineNum, col, "unmatched closing braces '}}'", promptDomain.ErrCodeUnmatchedClosing)
	}

	varContentPattern := regexp.MustCompile(`\{\{([^}]*)\}\}`)
	matches := varContentPattern.FindAllStringSubmatchIndex(line, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			varContent := line[match[2]:match[3]]
			if varContent != "" && !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(varContent) {
				// Detect advanced dialect syntax and suggest appropriate dialect
				if strings.HasPrefix(varContent, "#") || strings.HasPrefix(varContent, "^") ||
					strings.HasPrefix(varContent, "/") || strings.HasPrefix(varContent, ">") {
					result.AddWarning(lineNum, match[0]+1, "Mustache syntax detected; consider using 'mustache' dialect", promptDomain.WarnCodeDeprecatedSyntax)
				} else if strings.Contains(varContent, "|") {
					result.AddWarning(lineNum, match[0]+1, "Jinja2 filter syntax detected; consider using 'jinja2' dialect", promptDomain.WarnCodeDeprecatedSyntax)
				} else if varContent != "" {
					result.AddError(lineNum, match[0]+1, "invalid variable name: must start with a letter and contain only alphanumeric characters and underscores", promptDomain.ErrCodeInvalidVariableName)
				}
			}
		}
	}
}

// anyToString converts any value to a string representation.
// Strings are returned as-is; other types are converted appropriately.
func anyToString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case nil:
		return ""
	case fmt.Stringer:
		return v.String()
	default:
		// For complex types, use JSON serialization
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(data)
	}
}
