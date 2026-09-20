// Package entity contains agent concepts and rules, independent of providers.
package entity

import (
	"errors"
	"strings"
)

var ErrEmptyPrompt = errors.New("prompt must contain text")

// Prompt is a value object: its text is immutable after construction.
type Prompt struct {
	text string
}

func NewPrompt(text string) (Prompt, error) {
	if strings.TrimSpace(text) == "" {
		return Prompt{}, ErrEmptyPrompt
	}
	// Preserve formatting that may matter to the model, such as code indentation.
	return Prompt{text: text}, nil
}

func (p Prompt) Text() string { return p.text }
