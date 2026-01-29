package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"google.golang.org/genai"
)

type Credential struct {
	Service  string `json:"service"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Note     string `json:"note"`
}

func ParseInput(ctx context.Context, mimeType string, data []byte) (*Credential, error) {
	// Initialize the client.
	cfg := &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	}

	client, err := genai.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	promptText := "Extract service name, login, and password from this. If it's an image or audio, analyze carefully. Return ONLY JSON: {\"service\": \"\", \"login\": \"\", \"password\": \"\", \"note\": \"\"} without markdown formatting."

	// Create the content part for the prompt.
	promptPart := &genai.Part{Text: promptText}

	// Create the content part for the data (image/audio) if present.
	var parts []*genai.Part
	parts = append(parts, promptPart)

	if len(data) > 0 {
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: mimeType,
				Data:     data,
			},
		})
	}

	// Construct the request content.
	contents := []*genai.Content{
		{
			Parts: parts,
		},
	}

	// Call GenerateContent.
	// Note: We use "gemini-2.0-flash" as requested.
	resp, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash", contents, nil)
	if err != nil {
		return nil, fmt.Errorf("generate content error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from AI")
	}

	// Extract text result.
	// The new SDK structure for response might vary slightly, but generally it's Candidates[0].Content.Parts[0].Text
	// Ensure we handle potential nil pointers safely if needed, though simple checks above help.
	rawText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			rawText += part.Text
		}
	}

	// Clean up markdown code blocks if present
	rawText = strings.TrimSpace(rawText)
	rawText = strings.TrimPrefix(rawText, "```json")
	rawText = strings.TrimPrefix(rawText, "```")
	rawText = strings.TrimSuffix(rawText, "```")
	rawText = strings.TrimSpace(rawText)

	var cred Credential
	err = json.Unmarshal([]byte(rawText), &cred)
	if err != nil {
		// Fallback or detailed error
		return nil, fmt.Errorf("failed to parse JSON: %w | raw: %s", err, rawText)
	}

	return &cred, nil
}
