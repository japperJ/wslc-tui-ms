package discovery

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"wslc-tui-ms/internal/commands"
)

type recordingRunner struct {
	args   [][]string
	output []byte
	err    error
}

func (r *recordingRunner) Run(_ context.Context, args []string) ([]byte, error) {
	r.args = append(r.args, append([]string(nil), args...))
	return r.output, r.err
}

func TestDiscoverParsesJSONAndBuildsExplicitArgv(t *testing.T) {
	runner := &recordingRunner{output: []byte(`[{"Name":"web"},{"Name":"worker"}]`)}
	client := NewClient(runner)

	got, err := client.Refresh(context.Background(), commands.ResourceTypeContainer)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if want := []string{"web", "worker"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens = %#v, want %#v", got, want)
	}
	if want := []string{"wslc", "container", "ls", "--all", "--format", "json"}; !reflect.DeepEqual(runner.args[0], want) {
		t.Fatalf("argv = %#v, want %#v", runner.args[0], want)
	}
}

func TestImageDiscoveryNormalizesRepositoryAndTag(t *testing.T) {
	runner := &recordingRunner{output: []byte(`[{"Repository":"ubuntu","Tag":"latest"},{"Repository":"busybox","Tag":""}]`)}
	client := NewClient(runner)

	got, err := client.Refresh(context.Background(), commands.ResourceTypeImage)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if want := []string{"ubuntu:latest", "busybox"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens = %#v, want %#v", got, want)
	}
}

func TestRefreshFailurePreservesValuesAndRecovers(t *testing.T) {
	runner := &recordingRunner{output: []byte(`[{"Name":"web"}]`)}
	client := NewClient(runner)
	if _, err := client.Refresh(context.Background(), commands.ResourceTypeContainer); err != nil {
		t.Fatal(err)
	}

	runner.err = errors.New("wslc unavailable")
	if _, err := client.Refresh(context.Background(), commands.ResourceTypeContainer); err == nil {
		t.Fatal("expected refresh error")
	}
	state := client.State(commands.ResourceTypeContainer)
	if !state.Disabled || state.Err == nil || !reflect.DeepEqual(state.Values, []string{"web"}) {
		t.Fatalf("failed refresh state = %#v", state)
	}

	runner.err = nil
	runner.output = []byte(`[{"Name":"api"}]`)
	if _, err := client.Refresh(context.Background(), commands.ResourceTypeContainer); err != nil {
		t.Fatal(err)
	}
	state = client.State(commands.ResourceTypeContainer)
	if state.Disabled || state.Err != nil || !reflect.DeepEqual(state.Values, []string{"api"}) {
		t.Fatalf("recovered refresh state = %#v", state)
	}
}

func TestCacheCopiesValuesAndIsolatesResourceTypes(t *testing.T) {
	runner := &recordingRunner{output: []byte(`[{"Name":"web"}]`)}
	client := NewClient(runner)
	values, err := client.Refresh(context.Background(), commands.ResourceTypeContainer)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = "changed"
	if got := client.State(commands.ResourceTypeContainer).Values[0]; got != "web" {
		t.Fatalf("cache was mutated through returned values: %q", got)
	}
	if got := client.State(commands.ResourceTypeImage).Values; got != nil {
		t.Fatalf("image cache leaked container values: %#v", got)
	}
}

func TestMalformedJSONAndEmptyResults(t *testing.T) {
	runner := &recordingRunner{output: []byte(`not-json`)}
	client := NewClient(runner)
	if _, err := client.Refresh(context.Background(), commands.ResourceTypeNetwork); err == nil {
		t.Fatal("expected malformed JSON error")
	}
	runner.output = []byte(`[]`)
	values, err := client.Refresh(context.Background(), commands.ResourceTypeNetwork)
	if err != nil || len(values) != 0 {
		t.Fatalf("empty result = %#v, %v", values, err)
	}
}

func TestUnknownResourceTypeFailsWithoutRunningCommand(t *testing.T) {
	runner := &recordingRunner{}
	client := NewClient(runner)
	if _, err := client.Refresh(context.Background(), commands.ResourceType("unknown")); err == nil {
		t.Fatal("expected unknown resource type error")
	}
	if len(runner.args) != 0 {
		t.Fatalf("unknown resource ran command: %#v", runner.args)
	}
}

func TestRefreshPropagatesCancellationToRunner(t *testing.T) {
	runner := RunnerFunc(func(ctx context.Context, _ []string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})
	client := NewClient(runner)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Discover(ctx, commands.ResourceTypeVolume); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context cancellation", err)
	}
}
