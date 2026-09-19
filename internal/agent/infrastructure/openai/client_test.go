package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/insomnius/almira/internal/agent/application"
	"github.com/insomnius/almira/internal/agent/domain"
	"github.com/insomnius/almira/internal/agent/infrastructure/openai"
)

const goodResponse = `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"Hello from Almira."}}]}`

func TestAskThroughOpenAIAdapter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("incorrect authentication or content type")
		}
		var body struct {
			Model    string                           `json:"model"`
			Store    *bool                            `json:"store"`
			Messages []struct{ Role, Content string } `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.Model != "test-model" || body.Store == nil || *body.Store {
			t.Error("model or explicit store:false missing")
		}
		if len(body.Messages) != 1 || body.Messages[0].Role != "user" || body.Messages[0].Content != "Hello" {
			t.Errorf("unexpected messages: %+v", body.Messages)
		}
		fmt.Fprint(w, goodResponse)
	}))
	defer server.Close()
	client := newClient(t, server.URL+"/v1/")
	answer, err := application.NewAsk(client).Execute(context.Background(), "Hello")
	if err != nil || answer != "Hello from Almira." {
		t.Fatalf("got %q, %v", answer, err)
	}
}

func TestProviderFailures(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		body, want string
	}{
		{"unauthorized", 401, `{"error":{"message":"test-key private prompt"}}`, "HTTP 401"},
		{"rate limit", 429, `{"error":{"message":"test-key private prompt"}}`, "HTTP 429"},
		{"unavailable", 503, "private prompt", "HTTP 503"},
		{"bad JSON", 200, "private prompt", "invalid completion JSON"},
		{"no choices", 200, `{"choices":[]}`, "exactly one"},
		{"truncated", 200, `{"choices":[{"finish_reason":"length","message":{"role":"assistant","content":"partial"}}]}`, "did not finish"},
		{"tool call", 200, `{"choices":[{"finish_reason":"tool_calls","message":{"role":"assistant","content":null}}]}`, "did not finish"},
		{"refusal", 200, `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","refusal":"private prompt"}}]}`, "declined"},
		{"null text", 200, `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":null}}]}`, "no assistant text"},
		{"blank text", 200, `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":" "}}]}`, "no assistant text"},
		{"oversized", 200, strings.Repeat("x", (1<<20)+1), "exceeds 1 MiB"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			answer, err := newClient(t, server.URL).Generate(context.Background(), prompt(t))
			if answer != "" || err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %q, %v; want error containing %q", answer, err, tc.want)
			}
			if strings.Contains(err.Error(), "private prompt") || strings.Contains(err.Error(), "test-key") {
				t.Fatal("provider response leaked into error")
			}
		})
	}
}

func TestCancellation(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := newClient(t, server.URL)
	go func() { <-started; cancel() }()
	_, err := client.Generate(ctx, prompt(t))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v; want context.Canceled", err)
	}
}

func TestRedirectIsNotFollowed(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("redirect target received a request")
	}))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	_, err := newClient(t, source.URL).Generate(context.Background(), prompt(t))
	if err == nil || !strings.Contains(err.Error(), "HTTP 307") {
		t.Fatalf("got %v; want HTTP 307", err)
	}
}

func TestConfiguration(t *testing.T) {
	for _, base := range []string{"/v1", "http://example.com/v1", "ftp://example.com", "https://user:secret@example.com", "https://example.com?key=secret", "https://example.com#fragment"} {
		if _, err := openai.NewClient("key", "model", base); err == nil {
			t.Errorf("accepted invalid URL %q", base)
		}
	}
	if _, err := openai.NewClient("", "model", ""); err == nil {
		t.Error("accepted missing key")
	}
	if _, err := openai.NewClient("key", "", ""); err == nil {
		t.Error("accepted missing model")
	}
}

func newClient(t *testing.T, base string) *openai.Client {
	t.Helper()
	c, err := openai.NewClient("test-key", "test-model", base)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func prompt(t *testing.T) domain.Prompt {
	t.Helper()
	p, err := domain.NewPrompt("Hello")
	if err != nil {
		t.Fatal(err)
	}
	return p
}
