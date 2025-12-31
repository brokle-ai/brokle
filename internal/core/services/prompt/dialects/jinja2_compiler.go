package dialects

import (
	"bytes"
	"regexp"
	"sort"
	"strings"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"

	promptDomain "brokle/internal/core/domain/prompt"
)

var (
	jinja2VarPattern    = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_.]*?)(?:\s*\|[^}]*)?\s*\}\}`) // {{ variable }} or {{ variable | filter }}
	jinja2ForPattern    = regexp.MustCompile(`\{%\s*for\s+\w+\s+in\s+([a-zA-Z_][a-zA-Z0-9_.]*?)(?:\s|\|)`)
	jinja2IfPattern     = regexp.MustCompile(`\{%\s*if\s+([a-zA-Z_][a-zA-Z0-9_.]*?)(?:\s|\|)`)
	jinja2BlockPattern  = regexp.MustCompile(`\{%[^%]*%\}`)
	jinja2EndForPattern = regexp.MustCompile(`\{%\s*endfor\s*%\}`)
	jinja2EndIfPattern  = regexp.MustCompile(`\{%\s*endif\s*%\}`)
	jinja2ElsePattern   = regexp.MustCompile(`\{%\s*(?:else|elif)[^%]*%\}`)
)

type jinja2Compiler struct{}

func NewJinja2Compiler() promptDomain.DialectCompiler {
	return &jinja2Compiler{}
}

func (c *jinja2Compiler) Dialect() promptDomain.TemplateDialect {
	return promptDomain.DialectJinja2
}

func (c *jinja2Compiler) ExtractVariables(content string) ([]string, error) {
	seen := make(map[string]bool)
	var vars []string

	for _, match := range jinja2VarPattern.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			varName := match[1]
			rootVar := strings.Split(varName, ".")[0] // For nested paths like user.name, extract root
			if !seen[rootVar] && !isBuiltinVariable(rootVar) {
				seen[rootVar] = true
				vars = append(vars, rootVar)
			}
		}
	}

	for _, match := range jinja2ForPattern.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			varName := match[1]
			rootVar := strings.Split(varName, ".")[0]
			if !seen[rootVar] && !isBuiltinVariable(rootVar) {
				seen[rootVar] = true
				vars = append(vars, rootVar)
			}
		}
	}

	for _, match := range jinja2IfPattern.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			varName := match[1]
			rootVar := strings.Split(varName, ".")[0]
			if !seen[rootVar] && !isBuiltinVariable(rootVar) {
				seen[rootVar] = true
				vars = append(vars, rootVar)
			}
		}
	}

	sort.Strings(vars)
	return vars, nil
}

func isBuiltinVariable(name string) bool {
	builtins := map[string]bool{
		"loop": true, "self": true, "super": true,
		"true": true, "false": true, "none": true,
		"True": true, "False": true, "None": true,
	}
	return builtins[name]
}

// Compile renders the template with the provided variables.
//
// Security note: The gonja library is a Go-native Jinja2 implementation that provides
// a sandboxed template execution environment. Unlike Python's Jinja2:
//   - No file system access (no include/extends with paths)
//   - No Python builtins (os, sys, subprocess not exposed)
//   - No arbitrary code execution (Go's type system prevents it)
//   - Variables are passed explicitly via context
//
// This makes it safe for user-provided templates as long as the variables
// passed to the context don't contain sensitive data or dangerous callbacks.
func (c *jinja2Compiler) Compile(content string, variables map[string]any) (string, error) {
	if len(content) > promptDomain.MaxTemplateSize {
		return "", promptDomain.NewTemplateTooLargeError(len(content), promptDomain.MaxTemplateSize)
	}

	tmpl, err := gonja.FromString(content)
	if err != nil {
		return "", promptDomain.NewDialectCompilationError("jinja2", err.Error())
	}

	ctx := exec.NewContext(variables)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", promptDomain.NewDialectCompilationError("jinja2", err.Error())
	}

	return buf.String(), nil
}

func (c *jinja2Compiler) Validate(content string) (*promptDomain.ValidationResult, error) {
	result := promptDomain.NewValidationResult(true, promptDomain.DialectJinja2)

	if len(content) > promptDomain.MaxTemplateSize {
		result.AddError(0, 0, "template exceeds maximum size limit", promptDomain.ErrCodeTemplateTooLarge)
		return result, nil
	}

	_, err := gonja.FromString(content)
	if err != nil {
		result.AddError(0, 0, "jinja2 syntax error: "+err.Error(), promptDomain.ErrCodeInvalidSyntax)
		return result, nil
	}

	c.validateBlocks(result, content)

	vars, _ := c.ExtractVariables(content)
	if len(vars) > promptDomain.MaxVariables {
		result.AddWarning(0, 0, "template has many variables, consider simplifying", promptDomain.WarnCodeUnusedVariable)
	}

	if depth := c.calculateNestingDepth(content); depth > promptDomain.MaxNestingDepth {
		result.AddError(0, 0, "template nesting exceeds maximum depth", promptDomain.ErrCodeNestedTooDeep)
	}

	return result, nil
}

func (c *jinja2Compiler) validateBlocks(result *promptDomain.ValidationResult, content string) {
	lines := strings.Split(content, "\n")
	forStack := make([]int, 0)
	ifStack := make([]int, 0)
	forStartPattern := regexp.MustCompile(`\{%\s*for\s+`)
	ifStartPattern := regexp.MustCompile(`\{%\s*if\s+`)

	for lineNum, line := range lines {
		forStarts := len(forStartPattern.FindAllString(line, -1))
		forEnds := len(jinja2EndForPattern.FindAllString(line, -1))

		for i := 0; i < forStarts; i++ {
			forStack = append(forStack, lineNum+1)
		}
		for i := 0; i < forEnds; i++ {
			if len(forStack) == 0 {
				result.AddError(lineNum+1, 0, "{% endfor %} without matching {% for %}", promptDomain.ErrCodeUnmatchedClosing)
			} else {
				forStack = forStack[:len(forStack)-1]
			}
		}

		ifStarts := len(ifStartPattern.FindAllString(line, -1))
		ifEnds := len(jinja2EndIfPattern.FindAllString(line, -1))

		for i := 0; i < ifStarts; i++ {
			ifStack = append(ifStack, lineNum+1)
		}
		for i := 0; i < ifEnds; i++ {
			if len(ifStack) == 0 {
				result.AddError(lineNum+1, 0, "{% endif %} without matching {% if %}", promptDomain.ErrCodeUnmatchedClosing)
			} else {
				ifStack = ifStack[:len(ifStack)-1]
			}
		}
	}

	for _, line := range forStack {
		result.AddError(line, 0, "unclosed {% for %} block", promptDomain.ErrCodeUnmatchedOpening)
	}
	for _, line := range ifStack {
		result.AddError(line, 0, "unclosed {% if %} block", promptDomain.ErrCodeUnmatchedOpening)
	}
}

func (c *jinja2Compiler) calculateNestingDepth(content string) int {
	maxDepth := 0
	currentDepth := 0
	forStartPattern := regexp.MustCompile(`\{%\s*for\s+`)
	ifStartPattern := regexp.MustCompile(`\{%\s*if\s+`)

	for _, line := range strings.Split(content, "\n") {
		starts := len(forStartPattern.FindAllString(line, -1)) + len(ifStartPattern.FindAllString(line, -1))
		ends := len(jinja2EndForPattern.FindAllString(line, -1)) + len(jinja2EndIfPattern.FindAllString(line, -1))

		currentDepth += starts
		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}
		currentDepth -= ends
	}

	return maxDepth
}
