// Package discovery lists picker-eligible WSLC resources for the current process.
package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"wslc-tui-ms/internal/commands"
)

// Runner executes one machine-readable WSLC command.
type Runner interface {
	Run(context.Context, []string) ([]byte, error)
}

// Discovery lists resources by their explicit resource type.
type Discovery interface {
	Discover(context.Context, commands.ResourceType) ([]string, error)
}

// RunnerFunc adapts a function to Runner.
type RunnerFunc func(context.Context, []string) ([]byte, error)

func (f RunnerFunc) Run(ctx context.Context, args []string) ([]byte, error) {
	return f(ctx, args)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("discovery command is empty")
	}
	return exec.CommandContext(ctx, args[0], args[1:]...).Output()
}

// ResourceDefinition describes the fixed command and parser for one resource type.
type ResourceDefinition struct {
	Args   []string
	Parser func([]byte) ([]string, error)
}

var definitions = map[commands.ResourceType]ResourceDefinition{
	commands.ResourceTypeContainer: {
		Args:   []string{"wslc", "container", "ls", "--all", "--format", "json"},
		Parser: parseNames("Name", "ID", "Id"),
	},
	commands.ResourceTypeImage: {
		Args:   []string{"wslc", "image", "ls", "--format", "json"},
		Parser: parseImages,
	},
	commands.ResourceTypeNetwork: {
		Args:   []string{"wslc", "network", "ls", "--format", "json"},
		Parser: parseNames("Name", "ID", "Id"),
	},
	commands.ResourceTypeVolume: {
		Args:   []string{"wslc", "volume", "ls", "--format", "json"},
		Parser: parseNames("Name", "ID", "Id"),
	},
	commands.ResourceTypeSession: {
		Args:   []string{"wslc", "system", "session", "list", "--verbose"},
		Parser: parseNames("Name", "DisplayName", "ID", "Id"),
	},
}

// Definitions returns a copy of the supported discovery definitions.
func Definitions() map[commands.ResourceType]ResourceDefinition {
	result := make(map[commands.ResourceType]ResourceDefinition, len(definitions))
	for resourceType, definition := range definitions {
		definition.Args = append([]string(nil), definition.Args...)
		result[resourceType] = definition
	}
	return result
}

// CacheState is the last refresh state for a resource type.
type CacheState struct {
	Values   []string
	Err      error
	Disabled bool
}

// Client discovers resources and retains their results for the current session.
type Client struct {
	runner Runner
	mu     sync.RWMutex
	cache  map[commands.ResourceType]CacheState
}

func NewClient(runner Runner) *Client {
	if runner == nil {
		runner = execRunner{}
	}
	return &Client{runner: runner, cache: make(map[commands.ResourceType]CacheState)}
}

// Refresh runs the resource's fixed list command. Failed refreshes retain stale values
// but mark selection disabled until a later refresh succeeds.
func (c *Client) Refresh(ctx context.Context, resourceType commands.ResourceType) ([]string, error) {
	definition, ok := definitions[resourceType]
	if !ok {
		return nil, fmt.Errorf("unsupported resource type %q", resourceType)
	}
	output, err := c.runner.Run(ctx, definition.Args)
	if err != nil {
		c.mu.Lock()
		state := c.cache[resourceType]
		state.Err = err
		state.Disabled = true
		state.Values = append([]string(nil), state.Values...)
		c.cache[resourceType] = state
		c.mu.Unlock()
		return append([]string(nil), state.Values...), err
	}
	values, err := definition.Parser(output)
	if err != nil {
		c.mu.Lock()
		state := c.cache[resourceType]
		state.Err = err
		state.Disabled = true
		state.Values = append([]string(nil), state.Values...)
		c.cache[resourceType] = state
		c.mu.Unlock()
		return append([]string(nil), state.Values...), err
	}

	c.mu.Lock()
	c.cache[resourceType] = CacheState{Values: append([]string(nil), values...)}
	c.mu.Unlock()
	return append([]string(nil), values...), nil
}

// Discover implements Discovery and refreshes the session cache.
func (c *Client) Discover(ctx context.Context, resourceType commands.ResourceType) ([]string, error) {
	return c.Refresh(ctx, resourceType)
}

// State returns a copy of the last refresh state.
func (c *Client) State(resourceType commands.ResourceType) CacheState {
	c.mu.RLock()
	state := c.cache[resourceType]
	c.mu.RUnlock()
	state.Values = append([]string(nil), state.Values...)
	return state
}

func parseImages(data []byte) ([]string, error) {
	var records []map[string]json.RawMessage
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse image list: %w", err)
	}
	values := make([]string, 0, len(records))
	for index, record := range records {
		repository, err := jsonString(record, "Repository")
		if err != nil {
			return nil, fmt.Errorf("parse image list item %d: %w", index, err)
		}
		tag, err := optionalJSONText(record, "Tag")
		if err != nil {
			return nil, fmt.Errorf("parse image list item %d: %w", index, err)
		}
		value := repository
		if tag != "" {
			value += ":" + tag
		}
		values = append(values, value)
	}
	return values, nil
}

func parseNames(keys ...string) func([]byte) ([]string, error) {
	return func(data []byte) ([]string, error) {
		var records []map[string]json.RawMessage
		if err := json.Unmarshal(data, &records); err != nil {
			return nil, fmt.Errorf("parse resource list: %w", err)
		}
		values := make([]string, 0, len(records))
		for index, record := range records {
			value := ""
			for _, key := range keys {
				candidate, err := optionalJSONText(record, key)
				if err != nil {
					return nil, fmt.Errorf("parse resource list item %d: %w", index, err)
				}
				if candidate != "" {
					value = candidate
					break
				}
			}
			if value == "" {
				return nil, fmt.Errorf("parse resource list item %d: missing resource name", index)
			}
			values = append(values, value)
		}
		return values, nil
	}
}

func jsonString(record map[string]json.RawMessage, key string) (string, error) {
	value, err := optionalJSONText(record, key)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("missing %s", key)
	}
	return value, nil
}

func optionalJSONText(record map[string]json.RawMessage, key string) (string, error) {
	raw, ok := record[key]
	if !ok || string(raw) == "null" {
		return "", nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("field %s is not a string: %w", key, err)
	}
	return strings.TrimSpace(value), nil
}
