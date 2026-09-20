package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/insomnius/almira/cmd"
)

func TestHelpNeedsNoCredentials(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_MODEL", "")
	root := cmd.NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "--prompt") || !strings.Contains(out.String(), "--model") {
		t.Fatalf("flags missing from help: %s", out.String())
	}
}

func TestInvalidArguments(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_MODEL", "test-model")
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"missing prompt", nil, `required flag(s) "prompt" not set`},
		{"blank prompt", []string{"--prompt", " \t"}, "prompt must contain text"},
		{"positional input", []string{"--prompt", "hello", "extra"}, "unknown command"},
		{"unknown flag", []string{"--unknown"}, "unknown flag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := cmd.NewRootCommand()
			root.SetArgs(tc.args)
			err := root.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
}

func TestCommandCallsProvider(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		model string
	}{
		{"environment model", []string{"--prompt", "Hello"}, "env-model"},
		{"flag override", []string{"--prompt", "Hello", "--model", "flag-model"}, "flag-model"},
		{"short flags", []string{"-p", "Hello", "-m", "short-model"}, "short-model"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Model string `json:"model"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if body.Model != tc.model {
					t.Errorf("model = %q, want %q", body.Model, tc.model)
				}
				fmt.Fprint(w, `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"Hello from Almira."}}]}`)
			}))
			defer server.Close()
			t.Setenv("OPENAI_API_KEY", "test-key")
			t.Setenv("OPENAI_MODEL", "env-model")
			t.Setenv("OPENAI_BASE_URL", server.URL+"/v1")
			root := cmd.NewRootCommand()
			var out, stderr bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&stderr)
			root.SetArgs(tc.args)
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			if out.String() != "Hello from Almira.\n" || stderr.Len() != 0 {
				t.Fatalf("stdout=%q stderr=%q", out.String(), stderr.String())
			}
		})
	}
}

func TestCommandPropagatesCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("canceled command reached provider")
	}))
	defer server.Close()
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_MODEL", "test-model")
	t.Setenv("OPENAI_BASE_URL", server.URL)
	root := cmd.NewRootCommand()
	root.SetArgs([]string{"--prompt", "Hello"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := root.ExecuteContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}
