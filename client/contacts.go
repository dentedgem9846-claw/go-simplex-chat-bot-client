package client

import (
	"context"
	"encoding/json"
	"fmt"

	sxtypes "simplex-chat-bot/types"
)

// --- Address Commands ---

// CreateMyAddress creates a bot contact address.
//
// Command: /_address <userId>
func (c *Client) CreateMyAddress(ctx context.Context, userID int64) (*sxtypes.UserContactLink, error) {
	cmd := fmt.Sprintf("/_address %d", userID)
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var resp sxtypes.UserContactLinkCreated
	if err := unmarshalResponse(raw, &resp); err != nil {
		return nil, err
	}
	return &resp.UserContactLink, nil
}

// DeleteMyAddress deletes the bot contact address.
//
// Command: /_delete_address <userId>
func (c *Client) DeleteMyAddress(ctx context.Context, userID int64) error {
	cmd := fmt.Sprintf("/_delete_address %d", userID)
	_, err := c.Send(ctx, cmd)
	return err
}

// ShowMyAddress returns the bot contact address and settings.
//
// Command: /_show_address <userId>
func (c *Client) ShowMyAddress(ctx context.Context, userID int64) (*sxtypes.UserContactLink, error) {
	cmd := fmt.Sprintf("/_show_address %d", userID)
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var link sxtypes.UserContactLink
	if err := unmarshalResponse(raw, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// SetProfileAddress adds or removes the address from the bot profile.
//
// Command: /_profile_address <userId> on|off
func (c *Client) SetProfileAddress(ctx context.Context, userID int64, enable bool) error {
	state := "off"
	if enable {
		state = "on"
	}
	cmd := fmt.Sprintf("/_profile_address %d %s", userID, state)
	_, err := c.Send(ctx, cmd)
	return err
}

// SetAddressSettings updates the bot address settings.
//
// Command: /_address_settings <userId> <json(settings)>
func (c *Client) SetAddressSettings(ctx context.Context, userID int64, settings sxtypes.AddressSettings) error {
	payload, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	cmd := fmt.Sprintf("/_address_settings %d %s", userID, string(payload))
	_, err = c.Send(ctx, cmd)
	return err
}

// --- Connection Commands ---

// AddContact creates a 1-time invitation link.
//
// Command: /_connect <userId>
func (c *Client) AddContact(ctx context.Context, userID int64, incognito bool) (*sxtypes.CreatedConnLink, error) {
	cmd := fmt.Sprintf("/_connect %d", userID)
	if incognito {
		cmd += " incognito=on"
	}
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var link sxtypes.CreatedConnLink
	if err := unmarshalResponse(raw, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// ConnectPlan determines the link type and if already connected.
//
// Command: /_connect plan <userId> <connectionLink>
func (c *Client) ConnectPlan(ctx context.Context, userID int64, connectionLink string) (json.RawMessage, error) {
	cmd := fmt.Sprintf("/_connect plan %d %s", userID, connectionLink)
	return c.Send(ctx, cmd)
}

// APIConnect connects via a prepared SimpleX link.
//
// Command: /_connect <userId> <str(preparedLink_)>
func (c *Client) APIConnect(ctx context.Context, userID int64, incognito bool, preparedLink *sxtypes.CreatedConnLink) error {
	cmd := fmt.Sprintf("/_connect %d", userID)
	if incognito {
		cmd += " incognito=on"
	}
	if preparedLink != nil {
		cmd += " " + preparedLink.ConnFullLink
	}
	_, err := c.Send(ctx, cmd)
	return err
}

// ConnectViaLink connects via a SimpleX link as a string.
//
// Command: /connect <connLink_>
func (c *Client) ConnectViaLink(ctx context.Context, incognito bool, connLink *string) error {
	cmd := "/connect"
	if incognito {
		cmd += " incognito=on"
	}
	if connLink != nil {
		cmd += " " + *connLink
	}
	_, err := c.Send(ctx, cmd)
	return err
}

// AcceptContact accepts a contact request.
//
// Command: /_accept <contactReqId>
func (c *Client) AcceptContact(ctx context.Context, contactReqID int64) error {
	cmd := fmt.Sprintf("/_accept %d", contactReqID)
	_, err := c.Send(ctx, cmd)
	return err
}

// RejectContact rejects a contact request.
//
// Command: /_reject <contactReqId>
func (c *Client) RejectContact(ctx context.Context, contactReqID int64) error {
	cmd := fmt.Sprintf("/_reject %d", contactReqID)
	_, err := c.Send(ctx, cmd)
	return err
}

// --- Chat Commands ---

// ListContacts returns all contacts for the given user.
//
// Command: /_contacts <userId>
func (c *Client) ListContacts(ctx context.Context, userID int64) ([]sxtypes.Contact, error) {
	cmd := fmt.Sprintf("/_contacts %d", userID)
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var contacts []sxtypes.Contact
	if err := unmarshalResponse(raw, &contacts); err != nil {
		return nil, err
	}
	return contacts, nil
}

// ListGroups returns all groups for the given user.
//
// Command: /_groups <userId>
func (c *Client) ListGroups(ctx context.Context, userID int64, contactID *int64, search *string) ([]sxtypes.GroupInfo, error) {
	cmd := fmt.Sprintf("/_groups %d", userID)
	if contactID != nil {
		cmd += fmt.Sprintf(" @%d", *contactID)
	}
	if search != nil {
		cmd += " " + *search
	}
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var groups []sxtypes.GroupInfo
	if err := unmarshalResponse(raw, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// DeleteChat deletes a chat.
//
// Command: /_delete <str(chatRef)> <str(chatDeleteMode)>
func (c *Client) DeleteChat(ctx context.Context, chatRef sxtypes.ChatRef, deleteMode string, notify bool) error {
	notifyStr := ""
	if deleteMode == "full" || deleteMode == "entity" {
		if notify {
			notifyStr = " notify=on"
		} else {
			notifyStr = " notify=off"
		}
	}
	cmd := fmt.Sprintf("/_delete %s %s%s", chatRef.String(), deleteMode, notifyStr)
	_, err := c.Send(ctx, cmd)
	return err
}
