package client

import (
	"context"
	"encoding/json"
	"fmt"

	sxtypes "simplex-chat-bot/types"
)

// ShowActiveUser returns the currently active user profile.
//
// Command: /user
func (c *Client) ShowActiveUser(ctx context.Context) (*sxtypes.User, error) {
	raw, err := c.Send(ctx, "/user")
	if err != nil {
		return nil, err
	}
	var resp sxtypes.User
	if err := unmarshalResponse(raw, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateActiveUser creates a new user profile and sets it as active.
//
// Command: /_create user <json(newUser)>
func (c *Client) CreateActiveUser(ctx context.Context, newUser sxtypes.NewUser) (*sxtypes.User, error) {
	payload, err := json.Marshal(newUser)
	if err != nil {
		return nil, fmt.Errorf("marshal newUser: %w", err)
	}
	cmd := "/_create user " + string(payload)
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var resp sxtypes.User
	if err := unmarshalResponse(raw, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListUsers returns all user profiles.
//
// Command: /users
func (c *Client) ListUsers(ctx context.Context) ([]sxtypes.UserUI, error) {
	raw, err := c.Send(ctx, "/users")
	if err != nil {
		return nil, err
	}
	var resp []sxtypes.UserUI
	if err := unmarshalResponse(raw, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// SetActiveUser switches to the user with the given ID.
//
// Command: /_user <userId>
func (c *Client) SetActiveUser(ctx context.Context, userID int64, viewPwd *string) error {
	cmd := fmt.Sprintf("/_user %d", userID)
	if viewPwd != nil {
		pwdJSON, _ := json.Marshal(*viewPwd)
		cmd += " " + string(pwdJSON)
	}
	_, err := c.Send(ctx, cmd)
	return err
}

// DeleteUser deletes the user profile with the given ID.
//
// Command: /_delete user <userId> del_smp=on|off
func (c *Client) DeleteUser(ctx context.Context, userID int64, delSMPQueues bool, viewPwd *string) error {
	delStr := "off"
	if delSMPQueues {
		delStr = "on"
	}
	cmd := fmt.Sprintf("/_delete user %d del_smp=%s", userID, delStr)
	if viewPwd != nil {
		pwdJSON, _ := json.Marshal(*viewPwd)
		cmd += " " + string(pwdJSON)
	}
	_, err := c.Send(ctx, cmd)
	return err
}

// UpdateProfile updates the active user's profile.
//
// Command: /_profile <userId> <json(profile)>
func (c *Client) UpdateProfile(ctx context.Context, userID int64, profile sxtypes.Profile) (*sxtypes.UserProfileUpdated, error) {
	payload, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshal profile: %w", err)
	}
	cmd := fmt.Sprintf("/_profile %d %s", userID, string(payload))
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var resp sxtypes.UserProfileUpdated
	if err := unmarshalResponse(raw, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetContactPrefs configures chat preference overrides for a contact.
//
// Command: /_set prefs @<contactId> <json(preferences)>
func (c *Client) SetContactPrefs(ctx context.Context, contactID int64, prefs sxtypes.Preferences) error {
	payload, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("marshal prefs: %w", err)
	}
	cmd := fmt.Sprintf("/_set prefs @%d %s", contactID, string(payload))
	_, err = c.Send(ctx, cmd)
	return err
}

// CreateBotProfile creates a new bot profile.
//
// Command: /create bot [files=on] <display_name>
func (c *Client) CreateBotProfile(ctx context.Context, allowFiles bool, displayName string) error {
	cmd := "/create bot"
	if allowFiles {
		cmd += " files=on"
	}
	cmd += " " + displayName
	_, err := c.Send(ctx, cmd)
	return err
}

// SetBotCommands configures the bot command menu.
//
// Command: /set bot commands <commands>
func (c *Client) SetBotCommands(ctx context.Context, commands string) error {
	cmd := "/set bot commands " + commands
	_, err := c.Send(ctx, cmd)
	return err
}
