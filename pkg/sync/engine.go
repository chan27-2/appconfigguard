package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/chan27-2/appconfigguard/pkg/azure"
	"github.com/chan27-2/appconfigguard/pkg/diff"
)

// Engine handles synchronization operations
type Engine struct {
	azureClient *azure.Client
	maxRetries  int
	baseDelay   time.Duration
}

// NewEngine creates a new sync engine
func NewEngine(azureClient *azure.Client) *Engine {
	return &Engine{
		azureClient: azureClient,
		maxRetries:  3,
		baseDelay:   time.Second,
	}
}

// ApplyChanges applies the given changes to Azure App Configuration
func (e *Engine) ApplyChanges(ctx context.Context, changes []diff.Change, strict bool) error {
	if len(changes) == 0 {
		return nil // Nothing to do
	}

	// Convert diff.Changes to azure.ChangeOperations
	operations := e.convertToOperations(changes)

	// Apply changes with retry logic
	return e.applyWithRetry(ctx, operations)
}

// PreviewChanges shows what would be changed without applying
func (e *Engine) PreviewChanges(changes []diff.Change) {
	if len(changes) == 0 {
		fmt.Println("No changes to preview.")
		return
	}

	fmt.Println("Preview of changes to apply:")
	fmt.Println("============================")

	for _, change := range changes {
		switch change.Type {
		case diff.ChangeTypeAdd:
			fmt.Printf("ADD: %s = %s\n", change.Key, e.truncateValue(change.NewValue))
		case diff.ChangeTypeUpdate:
			fmt.Printf("UPDATE: %s = %s (was: %s)\n", change.Key, e.truncateValue(change.NewValue), e.truncateValue(change.OldValue))
		case diff.ChangeTypeDelete:
			fmt.Printf("DELETE: %s (was: %s)\n", change.Key, e.truncateValue(change.OldValue))
		}
	}

	summary := e.getSummary(changes)
	fmt.Printf("\nSummary: %d added, %d updated, %d deleted\n",
		summary.Added, summary.Updated, summary.Deleted)
}

// convertToOperations converts diff.Changes to azure.ChangeOperations
func (e *Engine) convertToOperations(changes []diff.Change) []azure.ChangeOperation {
	operations := make([]azure.ChangeOperation, len(changes))

	for i, change := range changes {
		op := azure.ChangeOperation{
			Key:   change.Key,
			Label: change.Label,
			Tags:  change.Tags,
		}

		switch change.Type {
		case diff.ChangeTypeAdd:
			op.Operation = "add"
			op.Value = change.NewValue
		case diff.ChangeTypeUpdate:
			op.Operation = "update"
			op.Value = change.NewValue
		case diff.ChangeTypeDelete:
			op.Operation = "delete"
			op.Value = change.OldValue
		}

		operations[i] = op
	}

	return operations
}

// applyWithRetry applies operations with exponential backoff retry logic
func (e *Engine) applyWithRetry(ctx context.Context, operations []azure.ChangeOperation) error {
	var lastErr error

	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		err := e.azureClient.ApplyChanges(ctx, operations)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt == e.maxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(attempt+1) * e.baseDelay

		fmt.Printf("Attempt %d failed, retrying in %v: %v\n", attempt+1, delay, err)

		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", e.maxRetries+1, lastErr)
}

// getSummary creates a summary of changes
func (e *Engine) getSummary(changes []diff.Change) diff.Summary {
	return diff.Summary{
		Added:   e.countChanges(changes, diff.ChangeTypeAdd),
		Updated: e.countChanges(changes, diff.ChangeTypeUpdate),
		Deleted: e.countChanges(changes, diff.ChangeTypeDelete),
		Total:   len(changes),
	}
}

// countChanges counts changes of a specific type
func (e *Engine) countChanges(changes []diff.Change, changeType diff.ChangeType) int {
	count := 0
	for _, change := range changes {
		if change.Type == changeType {
			count++
		}
	}
	return count
}

// truncateValue truncates long values for display
func (e *Engine) truncateValue(value string) string {
	maxLen := 50
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}

// ValidateChanges performs basic validation on changes before applying
func (e *Engine) ValidateChanges(changes []diff.Change) error {
	for _, change := range changes {
		if change.Key == "" {
			return fmt.Errorf("empty key found in changes")
		}

		// Additional validations can be added here
		// e.g., key format validation, value size limits, etc.
	}

	return nil
}
