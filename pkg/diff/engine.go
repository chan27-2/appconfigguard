package diff

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/saichandankadarla/appconfigguard/pkg/azure"
)

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeTypeAdd    ChangeType = "add"
	ChangeTypeUpdate ChangeType = "update"
	ChangeTypeDelete ChangeType = "delete"
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

// Engine handles diff operations between local and remote configurations
type Engine struct{}

// NewEngine creates a new diff engine
func NewEngine() *Engine {
	return &Engine{}
}

// Compare compares local configuration with remote configuration
func (e *Engine) Compare(local map[string]string, remote []azure.ConfigItem) ([]Change, error) {
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

	// Any remaining items in remoteMap are deletions
	for _, remoteItem := range remoteMap {
		changes = append(changes, Change{
			Type:     ChangeTypeDelete,
			Key:      remoteItem.Key,
			OldValue: remoteItem.Value,
			Label:    remoteItem.Label,
			Tags:     remoteItem.Tags,
		})
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
		return "No changes detected."
	}

	output := "Configuration changes:\n\n"

	for _, change := range changes {
		switch change.Type {
		case ChangeTypeAdd:
			output += fmt.Sprintf("+ %s = %s\n", change.Key, e.truncateValue(change.NewValue))
		case ChangeTypeUpdate:
			output += fmt.Sprintf("~ %s = %s (was: %s)\n", change.Key, e.truncateValue(change.NewValue), e.truncateValue(change.OldValue))
		case ChangeTypeDelete:
			output += fmt.Sprintf("- %s = %s\n", change.Key, e.truncateValue(change.OldValue))
		}
	}

	summary := e.GetSummary(changes)
	output += fmt.Sprintf("\nSummary: %d added, %d updated, %d deleted (%d total)\n",
		summary.Added, summary.Updated, summary.Deleted, summary.Total)

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

// truncateValue truncates long values for display
func (e *Engine) truncateValue(value string) string {
	maxLen := 50
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}

// HasChanges returns true if there are any changes
func (e *Engine) HasChanges(changes []Change) bool {
	return len(changes) > 0
}
