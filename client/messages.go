package client

import (
	"context"
	"encoding/json"
	"fmt"

	sxtypes "simplex-chat-bot/types"
)

// SendMessages sends one or more messages to a chat.
//
// Command: /_send <str(sendRef)> json <json(composedMessages)>
func (c *Client) SendMessages(ctx context.Context, chatRef sxtypes.ChatRef, live bool, ttl *int, messages []sxtypes.ComposedMessage) (json.RawMessage, error) {
	payload, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("marshal messages: %w", err)
	}
	cmd := fmt.Sprintf("/_send %s", chatRef.String())
	if live {
		cmd += " live=on"
	}
	if ttl != nil {
		cmd += fmt.Sprintf(" ttl=%d", *ttl)
	}
	cmd += " json " + string(payload)
	return c.Send(ctx, cmd)
}

// SendText is a convenience method to send a single text message.
func (c *Client) SendText(ctx context.Context, chatRef sxtypes.ChatRef, text string) (json.RawMessage, error) {
	msg := sxtypes.ComposedMessage{
		MsgContent: sxtypes.MsgContent{
			Type: "text",
			Text: text,
		},
	}
	return c.SendMessages(ctx, chatRef, false, nil, []sxtypes.ComposedMessage{msg})
}

// UpdateChatItem updates an existing message.
//
// Command: /_update item <str(chatRef)> <chatItemId>[ live=on] json <json(updatedMessage)>
func (c *Client) UpdateChatItem(ctx context.Context, chatRef sxtypes.ChatRef, chatItemID int64, live bool, msg sxtypes.UpdatedMessage) (json.RawMessage, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal updatedMessage: %w", err)
	}
	cmd := fmt.Sprintf("/_update item %s %d", chatRef.String(), chatItemID)
	if live {
		cmd += " live=on"
	}
	cmd += " json " + string(payload)
	return c.Send(ctx, cmd)
}

// DeleteChatItem deletes messages.
//
// Command: /_delete item <str(chatRef)> <chatItemIds> broadcast|internal|internalMark
func (c *Client) DeleteChatItem(ctx context.Context, chatRef sxtypes.ChatRef, chatItemIDs []int64, deleteMode string) (json.RawMessage, error) {
	cmd := fmt.Sprintf("/_delete item %s %s %s", chatRef.String(), joinInt64(chatItemIDs), deleteMode)
	return c.Send(ctx, cmd)
}

// DeleteMemberChatItem moderates messages in a group. Requires Moderator role.
//
// Command: /_delete member item #<groupId> <chatItemIds>
func (c *Client) DeleteMemberChatItem(ctx context.Context, groupID int64, chatItemIDs []int64) (json.RawMessage, error) {
	cmd := fmt.Sprintf("/_delete member item #%d %s", groupID, joinInt64(chatItemIDs))
	return c.Send(ctx, cmd)
}

// ChatItemReaction adds or removes a reaction on a message.
//
// Command: /_reaction <str(chatRef)> <chatItemId> on|off <json(reaction)>
func (c *Client) ChatItemReaction(ctx context.Context, chatRef sxtypes.ChatRef, chatItemID int64, add bool, reaction sxtypes.MsgReaction) (json.RawMessage, error) {
	payload, err := json.Marshal(reaction)
	if err != nil {
		return nil, fmt.Errorf("marshal reaction: %w", err)
	}
	state := "off"
	if add {
		state = "on"
	}
	cmd := fmt.Sprintf("/_reaction %s %d %s %s", chatRef.String(), chatItemID, state, string(payload))
	return c.Send(ctx, cmd)
}
