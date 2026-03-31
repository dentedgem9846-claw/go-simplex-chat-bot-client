// Package main provides an example of using the SimpleX Chat Bot client.
//
// This example demonstrates:
//   - Connecting to a SimpleX Chat CLI WebSocket server
//   - Setting up an event handler to receive messages
//   - Getting the active user profile
//   - Listing contacts
//   - Sending a direct message (if a contact exists)
//
// Usage:
//
//	go run ./example
//
// Or with a custom WebSocket URL:
//
//	SIMPLEX_WS_URL=ws://localhost:5225 go run ./example
//
// Note: This example handles connection errors gracefully. If the SimpleX CLI
// is not running, it will print an error message and exit without panicking.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"simplex-chat-bot/client"
	sxtypes "simplex-chat-bot/types"
)

func main() {
	// Get WebSocket URL from environment or use default
	wsURL := os.Getenv("SIMPLEX_WS_URL")
	if wsURL == "" {
		wsURL = "ws://localhost:5225"
	}

	// Create a context that can be cancelled on interrupt signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Create a logger
	logger := log.New(os.Stdout, "[SimpleX] ", log.LstdFlags)

	// Create the client
	opts := &client.Options{
		Logger:    logger,
		Timeout:   30 * time.Second,
		Reconnect: true, // Enable automatic reconnection
	}
	c := client.New(ctx, wsURL, opts)

	// Register an event handler to receive messages and other events
	c.OnEvent(func(event json.RawMessage) {
		// Try to extract the event type
		var typeCheck struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(event, &typeCheck); err == nil {
			logger.Printf("Received event: %s", typeCheck.Type)
		}

		// Pretty print the event for demonstration
		var prettyEvent map[string]interface{}
		if err := json.Unmarshal(event, &prettyEvent); err == nil {
			prettyJSON, _ := json.MarshalIndent(prettyEvent, "", "  ")
			logger.Printf("Event details:\n%s", prettyJSON)
		}
	})

	// Connect to the SimpleX CLI
	logger.Printf("Connecting to %s...", wsURL)
	if err := c.Connect(ctx); err != nil {
		logger.Printf("Failed to connect: %v", err)
		logger.Printf("Note: Make sure the SimpleX Chat CLI is running with: simplex-chat -p 5225")
		os.Exit(1)
	}
	logger.Printf("Connected successfully")

	// Get the active user profile
	logger.Printf("Getting active user...")
	user, err := c.ShowActiveUser(ctx)
	if err != nil {
		logger.Printf("Failed to get active user: %v", err)
	} else {
		logger.Printf("Active user: %s (ID: %d)", user.LocalDisplayName, user.UserID)
	}

	// List all users
	logger.Printf("Listing all users...")
	users, err := c.ListUsers(ctx)
	if err != nil {
		logger.Printf("Failed to list users: %v", err)
	} else {
		logger.Printf("Found %d user(s):", len(users))
		for _, u := range users {
			active := ""
			if u.ActiveUser {
				active = " (active)"
			}
			logger.Printf("  - %s%s", u.User.LocalDisplayName, active)
		}
	}

	// List contacts
	logger.Printf("Listing contacts...")
	var contacts []sxtypes.Contact
	if user != nil {
		contacts, err = c.ListContacts(ctx, user.UserID)
		if err != nil {
			logger.Printf("Failed to list contacts: %v", err)
		} else {
			logger.Printf("Found %d contact(s):", len(contacts))
			for _, contact := range contacts {
				logger.Printf("  - %s (ID: %d)", contact.LocalDisplayName, contact.ContactID)
			}
		}
	}

	// If we have contacts, send a test message to the first one
	if len(contacts) > 0 {
		contact := contacts[0]
		logger.Printf("Sending test message to %s...", contact.LocalDisplayName)

		chatRef := sxtypes.ChatRef{
			ChatType: "direct",
			ChatID:   contact.ContactID,
		}

		messages := []sxtypes.ComposedMessage{
			{
				MsgContent: sxtypes.MsgContent{
					Type: "text",
					Text: "Hello from the SimpleX Chat Bot Go client! 👋",
				},
			},
		}

		chatItems, err := c.SendMessages(ctx, chatRef, false, nil, messages)
		if err != nil {
			logger.Printf("Failed to send message: %v", err)
		} else {
			logger.Printf("Message sent successfully! Created %d chat item(s).", len(chatItems))
		}
	} else {
		logger.Printf("No contacts found to send test message to.")
		logger.Printf("Tip: Use /connect or create a bot address to add contacts.")
	}

	// List groups
	logger.Printf("Listing groups...")
	if user != nil {
		groups, err := c.ListGroups(ctx, user.UserID, nil, nil)
		if err != nil {
			logger.Printf("Failed to list groups: %v", err)
		} else {
			logger.Printf("Found %d group(s):", len(groups))
			for _, group := range groups {
				logger.Printf("  - %s (ID: %d)", group.LocalDisplayName, group.GroupID)
			}
		}
	}

	// Keep the connection alive and wait for events until interrupted
	logger.Printf("Listening for events. Press Ctrl+C to exit.")
	<-ctx.Done()

	// Close the connection cleanly
	logger.Printf("Closing connection...")
	if err := c.Close(); err != nil {
		logger.Printf("Error during close: %v", err)
	}
	logger.Printf("Goodbye!")
}
