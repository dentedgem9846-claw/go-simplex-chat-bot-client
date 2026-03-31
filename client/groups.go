package client

import (
	"context"
	"encoding/json"
	"fmt"

	sxtypes "simplex-chat-bot/types"
)

// AddMember adds a contact to a group. Requires Admin role.
//
// Command: /_add #<groupId> <contactId> <memberRole>
func (c *Client) AddMember(ctx context.Context, groupID, contactID int64, memberRole string) error {
	cmd := fmt.Sprintf("/_add #%d %d %s", groupID, contactID, memberRole)
	_, err := c.Send(ctx, cmd)
	return err
}

// JoinGroup joins a group.
//
// Command: /_join #<groupId>
func (c *Client) JoinGroup(ctx context.Context, groupID int64) error {
	cmd := fmt.Sprintf("/_join #%d", groupID)
	_, err := c.Send(ctx, cmd)
	return err
}

// AcceptMember accepts a group member. Requires Admin role.
//
// Command: /_accept member #<groupId> <groupMemberId> <memberRole>
func (c *Client) AcceptMember(ctx context.Context, groupID, groupMemberID int64, memberRole string) error {
	cmd := fmt.Sprintf("/_accept member #%d %d %s", groupID, groupMemberID, memberRole)
	_, err := c.Send(ctx, cmd)
	return err
}

// MembersRole sets the role for members. Requires Admin role.
//
// Command: /_member role #<groupId> <memberIds> <memberRole>
func (c *Client) MembersRole(ctx context.Context, groupID int64, memberIDs []int64, memberRole string) error {
	cmd := fmt.Sprintf("/_member role #%d %s %s", groupID, joinInt64(memberIDs), memberRole)
	_, err := c.Send(ctx, cmd)
	return err
}

// BlockMembersForAll blocks members for all. Requires Moderator role.
//
// Command: /_block #<groupId> <memberIds> blocked=on|off
func (c *Client) BlockMembersForAll(ctx context.Context, groupID int64, memberIDs []int64, blocked bool) error {
	state := "off"
	if blocked {
		state = "on"
	}
	cmd := fmt.Sprintf("/_block #%d %s blocked=%s", groupID, joinInt64(memberIDs), state)
	_, err := c.Send(ctx, cmd)
	return err
}

// RemoveMembers removes members from a group. Requires Admin role.
//
// Command: /_remove #<groupId> <memberIds>[ messages=on]
func (c *Client) RemoveMembers(ctx context.Context, groupID int64, memberIDs []int64, withMessages bool) error {
	cmd := fmt.Sprintf("/_remove #%d %s", groupID, joinInt64(memberIDs))
	if withMessages {
		cmd += " messages=on"
	}
	_, err := c.Send(ctx, cmd)
	return err
}

// LeaveGroup leaves a group.
//
// Command: /_leave #<groupId>
func (c *Client) LeaveGroup(ctx context.Context, groupID int64) error {
	cmd := fmt.Sprintf("/_leave #%d", groupID)
	_, err := c.Send(ctx, cmd)
	return err
}

// ListMembers returns all members of a group.
//
// Command: /_members #<groupId>
func (c *Client) ListMembers(ctx context.Context, groupID int64) ([]sxtypes.GroupMember, error) {
	cmd := fmt.Sprintf("/_members #%d", groupID)
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var members []sxtypes.GroupMember
	if err := unmarshalResponse(raw, &members); err != nil {
		return nil, err
	}
	return members, nil
}

// NewGroup creates a new group.
//
// Command: /_group <userId>[ incognito=on] <json(groupProfile)>
func (c *Client) NewGroup(ctx context.Context, userID int64, incognito bool, profile sxtypes.GroupProfile) (*sxtypes.GroupInfo, error) {
	payload, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshal groupProfile: %w", err)
	}
	cmd := fmt.Sprintf("/_group %d", userID)
	if incognito {
		cmd += " incognito=on"
	}
	cmd += " " + string(payload)
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var info sxtypes.GroupInfo
	if err := unmarshalResponse(raw, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// UpdateGroupProfile updates a group's profile.
//
// Command: /_group_profile #<groupId> <json(groupProfile)>
func (c *Client) UpdateGroupProfile(ctx context.Context, groupID int64, profile sxtypes.GroupProfile) (*sxtypes.GroupInfo, error) {
	payload, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshal groupProfile: %w", err)
	}
	cmd := fmt.Sprintf("/_group_profile #%d %s", groupID, string(payload))
	raw, err := c.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var info sxtypes.GroupInfo
	if err := unmarshalResponse(raw, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// --- Group Link Commands ---

// CreateGroupLink creates an invitation link for a group.
//
// Command: /_create link #<groupId> <memberRole>
func (c *Client) CreateGroupLink(ctx context.Context, groupID int64, memberRole string) (*sxtypes.CreatedConnLink, error) {
	cmd := fmt.Sprintf("/_create link #%d %s", groupID, memberRole)
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

// GroupLinkMemberRole sets the member role for a group link.
//
// Command: /_set link role #<groupId> <memberRole>
func (c *Client) GroupLinkMemberRole(ctx context.Context, groupID int64, memberRole string) error {
	cmd := fmt.Sprintf("/_set link role #%d %s", groupID, memberRole)
	_, err := c.Send(ctx, cmd)
	return err
}

// DeleteGroupLink deletes the invitation link for a group.
//
// Command: /_delete link #<groupId>
func (c *Client) DeleteGroupLink(ctx context.Context, groupID int64) error {
	cmd := fmt.Sprintf("/_delete link #%d", groupID)
	_, err := c.Send(ctx, cmd)
	return err
}

// GetGroupLink returns the invitation link for a group.
//
// Command: /_get link #<groupId>
func (c *Client) GetGroupLink(ctx context.Context, groupID int64) (*sxtypes.CreatedConnLink, error) {
	cmd := fmt.Sprintf("/_get link #%d", groupID)
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

func joinInt64(ids []int64) string {
	s := ""
	for i, id := range ids {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%d", id)
	}
	return s
}
