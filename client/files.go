package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ReceiveFile accepts and receives a file.
//
// Command: /freceive <fileId>[ approved_relays=on][ encrypt=on|off][ inline=on|off][ <filePath>]
func (c *Client) ReceiveFile(ctx context.Context, fileID int64, approvedRelays bool, storeEncrypted *bool, fileInline *bool, filePath *string) (json.RawMessage, error) {
	cmd := fmt.Sprintf("/freceive %d", fileID)
	if approvedRelays {
		cmd += " approved_relays=on"
	}
	if storeEncrypted != nil {
		if *storeEncrypted {
			cmd += " encrypt=on"
		} else {
			cmd += " encrypt=off"
		}
	}
	if fileInline != nil {
		if *fileInline {
			cmd += " inline=on"
		} else {
			cmd += " inline=off"
		}
	}
	if filePath != nil {
		cmd += " " + *filePath
	}
	return c.Send(ctx, cmd)
}

// CancelFile cancels a file send or receive.
//
// Command: /fcancel <fileId>
func (c *Client) CancelFile(ctx context.Context, fileID int64) (json.RawMessage, error) {
	cmd := fmt.Sprintf("/fcancel %d", fileID)
	return c.Send(ctx, cmd)
}
