package ui

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// InputDialog shows a native macOS input dialog and returns the entered text.
// Returns ("", err) if the user cancels.
func InputDialog(prompt string) (string, error) {
	script := fmt.Sprintf(
		`set r to display dialog %s default answer ""
text returned of r`,
		asStr(prompt),
	)
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ChooseFromList shows a native macOS multi-select list dialog.
// Items are numbered internally to avoid comma-parsing ambiguity.
// Returns nil if the user cancels.
func ChooseFromList(title string, items []string) ([]int, error) {
	// prefix each item with its 1-based index so we can parse selection unambiguously
	numbered := make([]string, len(items))
	for i, item := range items {
		numbered[i] = fmt.Sprintf("%d. %s", i+1, item)
	}

	script := fmt.Sprintf(`set r to choose from list %s with title %s with multiple selections allowed
if r is false then return ""
set AppleScript's text item delimiters to "\n"
r as text`,
		buildList(numbered),
		asStr(title),
	)

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return nil, err
	}

	s := strings.TrimSpace(string(out))
	if s == "" {
		return nil, nil
	}

	var indices []int
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if dot := strings.Index(line, "."); dot > 0 {
			if n, err := strconv.Atoi(line[:dot]); err == nil && n >= 1 && n <= len(items) {
				indices = append(indices, n-1)
			}
		}
	}
	return indices, nil
}

// Notify sends a macOS notification. Errors are silently ignored.
func Notify(title, message string) {
	script := fmt.Sprintf(`display notification %s with title %s`, asStr(message), asStr(title))
	exec.Command("osascript", "-e", script).Run()
}

// asStr converts a Go string to an AppleScript string literal.
func asStr(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// buildList converts a Go string slice to an AppleScript list literal.
func buildList(items []string) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = asStr(item)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
