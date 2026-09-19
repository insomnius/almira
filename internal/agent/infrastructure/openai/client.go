// Package openai adapts the OpenAI Chat Completions format to the agent port.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/insomnius/almira/internal/agent/domain"
)

const DefaultBaseURL = "https://api.openai.com/v1"
const maxResponseBytes = 1 << 20 // Bound memory use to 1 MiB per response.

type Client struct {
	apiKey   string
	model    string
	endpoint string
	http     *http.Client
}

func NewClient(apiKey, model, baseURL string) (*Client, error) {
	apiKey, model = strings.TrimSpace(apiKey), strings.TrimSpace(model)
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is required")
	}
	if model == "" {
		return nil, errors.New("set OPENAI_MODEL or pass -model")
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("OPENAI_BASE_URL must be an absolute URL without credentials, query, or fragment")
	}
	loopback := u.Hostname() == "localhost" || net.ParseIP(u.Hostname()).IsLoopback()
	if u.Scheme != "https" && !(u.Scheme == "http" && loopback) {
		return nil, errors.New("OPENAI_BASE_URL requires HTTPS (HTTP is allowed on loopback for local development)")
	}
	return &Client{
		apiKey:   apiKey,
		model:    model,
		endpoint: strings.TrimRight(baseURL, "/") + "/chat/completions",
		http: &http.Client{
			Timeout: 60 * time.Second,
			// An unexpected redirect must not forward credentials or prompt data.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}, nil
}

// The wire types stay private so OpenAI's schema cannot become our domain model.
type request struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Store    bool      `json:"store"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type response struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role    string  `json:"role"`
			Content *string `json:"content"`
			Refusal *string `json:"refusal"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) Generate(ctx context.Context, prompt domain.Prompt) (string, error) {
	// Go permits a zero-value Prompt even with an unexported field.
	if prompt.Text() == "" {
		return "", domain.ErrEmptyPrompt
	}
	payload, err := json.Marshal(request{
		Model:    c.model,
		Messages: []message{{Role: "user", Content: prompt.Text()}},
		Store:    false,
	})
	if err != nil {
		return "", fmt.Errorf("encode completion request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create completion request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("send completion request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		// Provider error bodies can echo credentials or private input.
		return "", fmt.Errorf("completion request failed: HTTP %d (%s)", res.StatusCode, http.StatusText(res.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("read completion response: %w", err)
	}
	if len(body) > maxResponseBytes {
		return "", errors.New("completion response exceeds 1 MiB")
	}
	var result response
	if err := json.Unmarshal(body, &result); err != nil {
		return "", errors.New("provider returned invalid completion JSON")
	}
	if len(result.Choices) != 1 {
		return "", errors.New("expected exactly one completion choice")
	}
	choice := result.Choices[0]
	if choice.Message.Refusal != nil && *choice.Message.Refusal != "" {
		return "", errors.New("model declined the request")
	}
	if choice.FinishReason != "stop" {
		return "", errors.New("model did not finish a text answer (it may have reached a limit or requested a tool)")
	}
	if choice.Message.Role != "assistant" || choice.Message.Content == nil || strings.TrimSpace(*choice.Message.Content) == "" {
		return "", errors.New("provider returned no assistant text")
	}
	return *choice.Message.Content, nil
}
