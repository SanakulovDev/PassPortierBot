package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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
		if mimeType == "text/plain" {
			parts = append(parts, &genai.Part{Text: string(data)})
		} else {
			parts = append(parts, &genai.Part{
				InlineData: &genai.Blob{
					MIMEType: mimeType,
					Data:     data,
				},
			})
		}
	}

	// Construct the request content.
	contents := []*genai.Content{
		{
			Parts: parts,
		},
	}

	// Call GenerateContent with retry logic.
	// Note: We use "gemini-2.0-flash" as requested.
	var resp *genai.GenerateContentResponse

	maxRetries := 5
	baseDelay := 2 * time.Second
	retryRegex := regexp.MustCompile(`Please retry in (\d+(\.\d+)?)s`)

	for i := 0; i <= maxRetries; i++ {
		resp, err = client.Models.GenerateContent(ctx, "gemini-2.0-flash", contents, nil)
		if err == nil {
			break
		}

		// Check for rate limit errors
		errMsg := err.Error()
		if strings.Contains(errMsg, "429") || strings.Contains(errMsg, "RESOURCE_EXHAUSTED") || strings.Contains(errMsg, "quota") {
			if i < maxRetries {
				waitDuration := baseDelay * time.Duration(1<<i)

				// Try to parse specific wait time from error
				matches := retryRegex.FindStringSubmatch(errMsg)
				if len(matches) > 1 {
					if val, parseErr := strconv.ParseFloat(matches[1], 64); parseErr == nil {
						// Add 1 second buffer
						waitDuration = time.Duration(val*float64(time.Second)) + 1*time.Second
					}
				}

				log.Printf("[WARNING] Gemini rate limit hit. Retrying in %v... (Attempt %d/%d)", waitDuration, i+1, maxRetries)
				time.Sleep(waitDuration)
				continue
			}
		}

		// If other error or retries exhausted
		return nil, fmt.Errorf("generate content error (after %d retries): %w", i, err)
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
