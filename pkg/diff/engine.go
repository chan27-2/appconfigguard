package diff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/saichandankadarla/appconfigguard/pkg/azure"
)

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeTypeAdd    ChangeType = "add"
	ChangeTypeUpdate ChangeType = "update"
	ChangeTypeDelete ChangeType = "delete"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"

	// Bold colors
	colorBoldRed    = "\033[1;31m"
	colorBoldGreen  = "\033[1;32m"
	colorBoldYellow = "\033[1;33m"
	colorBoldBlue   = "\033[1;34m"
	colorBoldPurple = "\033[1;35m"
	colorBoldCyan   = "\033[1;36m"
	colorBoldWhite  = "\033[1;37m"

	// Background colors
	colorBgRed    = "\033[41m"
	colorBgGreen  = "\033[42m"
	colorBgYellow = "\033[43m"
)

// Change represents a single configuration change
type Change struct {
	Type     ChangeType
	Key      string
	OldValue string
	NewValue string
	Label    string
	Tags     map[string]string
}

// Summary provides a summary of changes
type Summary struct {
	Added   int
	Updated int
	Deleted int
	Total   int
}

// Helper functions for colored output
func colorize(text, color string) string {
	return color + text + colorReset
}

func bold(text string) string {
	return "\033[1m" + text + colorReset
}

func formatChangeSymbol(changeType ChangeType) string {
	switch changeType {
	case ChangeTypeAdd:
		return colorize("➕", colorBoldGreen)
	case ChangeTypeUpdate:
		return colorize("🔄", colorBoldYellow)
	case ChangeTypeDelete:
		return colorize("❌", colorBoldRed)
	default:
		return colorize("❓", colorGray)
	}
}

func formatChangeType(changeType ChangeType) string {
	switch changeType {
	case ChangeTypeAdd:
		return colorize("ADD", colorGreen)
	case ChangeTypeUpdate:
		return colorize("UPDATE", colorYellow)
	case ChangeTypeDelete:
		return colorize("DELETE", colorRed)
	default:
		return colorize("UNKNOWN", colorGray)
	}
}

// Engine handles diff operations between local and remote configurations
type Engine struct{}

// NewEngine creates a new diff engine
func NewEngine() *Engine {
	return &Engine{}
}

// Compare compares local configuration with remote configuration
func (e *Engine) Compare(local map[string]string, remote []azure.ConfigItem, strict bool) ([]Change, error) {
	changes := []Change{}

	// Create map of remote items for efficient lookup
	remoteMap := make(map[string]azure.ConfigItem)
	for _, item := range remote {
		remoteMap[item.Key] = item
	}

	// Check for additions and updates
	for key, localValue := range local {
		if remoteItem, exists := remoteMap[key]; exists {
			// Key exists, check if value changed
			if remoteItem.Value != localValue {
				changes = append(changes, Change{
					Type:     ChangeTypeUpdate,
					Key:      key,
					OldValue: remoteItem.Value,
					NewValue: localValue,
					Label:    remoteItem.Label,
					Tags:     remoteItem.Tags,
				})
			}
			// Remove from remoteMap to track what's left
			delete(remoteMap, key)
		} else {
			// Key doesn't exist in remote, it's an addition
			changes = append(changes, Change{
				Type:     ChangeTypeAdd,
				Key:      key,
				NewValue: localValue,
			})
		}
	}

	// Any remaining items in remoteMap are deletions (only if strict mode is enabled)
	if strict {
		for _, remoteItem := range remoteMap {
			changes = append(changes, Change{
				Type:     ChangeTypeDelete,
				Key:      remoteItem.Key,
				OldValue: remoteItem.Value,
				Label:    remoteItem.Label,
				Tags:     remoteItem.Tags,
			})
		}
	}

	// Sort changes for consistent output
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Key < changes[j].Key
	})

	return changes, nil
}

// GetSummary returns a summary of changes
func (e *Engine) GetSummary(changes []Change) Summary {
	summary := Summary{}

	for _, change := range changes {
		switch change.Type {
		case ChangeTypeAdd:
			summary.Added++
		case ChangeTypeUpdate:
			summary.Updated++
		case ChangeTypeDelete:
			summary.Deleted++
		}
		summary.Total++
	}

	return summary
}

// FormatConsole formats changes for console output with colors
func (e *Engine) FormatConsole(changes []Change) string {
	if len(changes) == 0 {
		return colorize("✨ No changes detected. Your configuration is up to date!", colorBoldGreen)
	}

	output := colorize("🔍 Configuration Changes", colorBoldCyan) + "\n"
	output += colorize(strings.Repeat("═", 60), colorGray) + "\n\n"

	for i, change := range changes {
		// Add separator between changes
		if i > 0 {
			output += "\n"
		}

		symbol := formatChangeSymbol(change.Type)
		changeType := formatChangeType(change.Type)
		key := colorize(bold(change.Key), colorBoldBlue)

		switch change.Type {
		case ChangeTypeAdd:
			output += fmt.Sprintf("%s %s %s\n", symbol, changeType, key)
			output += fmt.Sprintf("   %s %s\n", colorize("New value:", colorCyan), e.truncateValue(change.NewValue))

		case ChangeTypeUpdate:
			output += fmt.Sprintf("%s %s %s\n", symbol, changeType, key)
			output += fmt.Sprintf("   %s %s\n", colorize("New value:", colorCyan), e.truncateValue(change.NewValue))
			output += fmt.Sprintf("   %s %s\n", colorize("Old value:", colorGray), e.truncateValue(change.OldValue))

		case ChangeTypeDelete:
			output += fmt.Sprintf("%s %s %s\n", symbol, changeType, key)
			output += fmt.Sprintf("   %s %s\n", colorize("Old value:", colorGray), e.truncateValue(change.OldValue))
		}
	}

	// Add summary section
	summary := e.GetSummary(changes)
	output += "\n" + colorize(strings.Repeat("═", 60), colorGray) + "\n"
	output += colorize("📊 Summary", colorBoldCyan) + "\n"

	if summary.Added > 0 {
		output += fmt.Sprintf("   %s %d %s\n", colorize("➕", colorGreen), summary.Added, colorize("added", colorGreen))
	}
	if summary.Updated > 0 {
		output += fmt.Sprintf("   %s %d %s\n", colorize("🔄", colorYellow), summary.Updated, colorize("updated", colorYellow))
	}
	if summary.Deleted > 0 {
		output += fmt.Sprintf("   %s %d %s\n", colorize("❌", colorRed), summary.Deleted, colorize("deleted", colorRed))
	}

	output += fmt.Sprintf("\n   %s %d %s\n",
		colorize("📈", colorBoldPurple),
		summary.Total,
		colorize("total changes", colorBoldPurple))

	return output
}

// FormatJSON formats changes as JSON for machine-readable output
func (e *Engine) FormatJSON(changes []Change) ([]byte, error) {
	type jsonChange struct {
		Type     string            `json:"type"`
		Key      string            `json:"key"`
		OldValue string            `json:"old_value,omitempty"`
		NewValue string            `json:"new_value,omitempty"`
		Label    string            `json:"label,omitempty"`
		Tags     map[string]string `json:"tags,omitempty"`
	}

	type jsonOutput struct {
		Changes []jsonChange `json:"changes"`
		Summary Summary      `json:"summary"`
	}

	jsonChanges := make([]jsonChange, len(changes))
	for i, change := range changes {
		jsonChanges[i] = jsonChange{
			Type:     string(change.Type),
			Key:      change.Key,
			OldValue: change.OldValue,
			NewValue: change.NewValue,
			Label:    change.Label,
			Tags:     change.Tags,
		}
	}

	output := jsonOutput{
		Changes: jsonChanges,
		Summary: e.GetSummary(changes),
	}

	return json.MarshalIndent(output, "", "  ")
}

// truncateValue truncates long values for display with better formatting
func (e *Engine) truncateValue(value string) string {
	maxLen := 80 // Increased from 50 for better readability
	if len(value) <= maxLen {
		return colorize(`"`+value+`"`, colorWhite)
	}

	// Show first part and indicate truncation
	truncated := value[:maxLen-3] + "..."
	return colorize(`"`+truncated+`"`, colorWhite) +
		   colorize(fmt.Sprintf(" (%d chars total)", len(value)), colorGray)
}

// HasChanges returns true if there are any changes
func (e *Engine) HasChanges(changes []Change) bool {
	return len(changes) > 0
}
