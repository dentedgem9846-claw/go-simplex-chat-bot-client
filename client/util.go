package client

import (
	"encoding/json"
	"fmt"
)

// unmarshalResponse extracts a typed value from a raw JSON response.
// It first checks for a ChatCmdError and returns it as a descriptive error.
func unmarshalResponse(raw json.RawMessage, target interface{}) error {
	var rt struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &rt); err != nil {
		return fmt.Errorf("unmarshal response type: %w", err)
	}

	if rt.Type == "chatCmdError" {
		var chatErr struct {
			Type      string `json:"type"`
			ChatError struct {
				Type      string `json:"type"`
				ErrorType *struct {
					Type string `json:"type"`
				} `json:"errorType,omitempty"`
			} `json:"chatError"`
		}
		if err := json.Unmarshal(raw, &chatErr); err == nil {
			msg := "chat command error"
			if chatErr.ChatError.ErrorType != nil {
				msg = fmt.Sprintf("chat error: %s (%s)", chatErr.ChatError.Type, chatErr.ChatError.ErrorType.Type)
			}
			return fmt.Errorf("%s", msg)
		}
	}

	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}
	return nil
}
