// Package types contains all the JSON types for the SimpleX Chat Bot API.
package types

import "encoding/json"

// Command represents a command sent to the SimpleX Chat CLI via WebSocket.
type Command struct {
	CorrID string `json:"corrId"`
	Cmd    string `json:"cmd"`
}

// Response represents a response from the SimpleX Chat CLI via WebSocket.
type Response struct {
	CorrID string          `json:"corrId,omitempty"`
	Resp   json.RawMessage `json:"resp"`
}

// ResponseType is used to extract the "type" field from a response.
type ResponseType struct {
	Type string `json:"type"`
}

// ChatRef represents a reference to a chat (direct, group, or local).
type ChatRef struct {
	ChatType  string `json:"chatType"`
	ChatID    int64  `json:"chatId"`
	ChatScope string `json:"chatScope,omitempty"`
}

// String returns the string representation used in CLI commands.
func (c ChatRef) String() string {
	prefix := ""
	switch c.ChatType {
	case "direct":
		prefix = "@"
	case "group":
		prefix = "#"
	case "local":
		prefix = "*"
	}
	return prefix + int64ToString(c.ChatID)
}

func int64ToString(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		d := byte('0' + byte(n%10))
		s = string(d) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

// --- Core Types ---

// User represents a SimpleX Chat user.
type User struct {
	UserID           int64        `json:"userId"`
	AgentUserID      string       `json:"agentUserId"`
	UserContactID    int64        `json:"userContactId"`
	LocalDisplayName string       `json:"localDisplayName"`
	Profile          LocalProfile `json:"profile"`
	ActiveUser       bool         `json:"activeUser"`
}

// Profile represents a user profile.
type Profile struct {
	DisplayName string       `json:"displayName"`
	FullName    string       `json:"fullName"`
	ShortDescr  *string      `json:"shortDescr,omitempty"`
	Image       *string      `json:"image,omitempty"`
	ContactLink *string      `json:"contactLink,omitempty"`
	Preferences *Preferences `json:"preferences,omitempty"`
	PeerType    *string      `json:"peerType,omitempty"`
}

// LocalProfile extends Profile with local-only fields.
type LocalProfile struct {
	ProfileID   int64        `json:"profileId"`
	DisplayName string       `json:"displayName"`
	FullName    string       `json:"fullName"`
	ShortDescr  *string      `json:"shortDescr,omitempty"`
	Image       *string      `json:"image,omitempty"`
	ContactLink *string      `json:"contactLink,omitempty"`
	Preferences *Preferences `json:"preferences,omitempty"`
	PeerType    *string      `json:"peerType,omitempty"`
	LocalAlias  string       `json:"localAlias"`
}

// NewUser is used when creating a new user profile.
type NewUser struct {
	Profile           Profile `json:"profile"`
	ShortName         string  `json:"shortName"`
	FullName          string  `json:"fullName"`
	OwnLinkModeration *bool   `json:"ownLinkModeration,omitempty"`
}

// --- Contact Types ---

// Contact represents a contact.
type Contact struct {
	ContactID         int64                  `json:"contactId"`
	LocalDisplayName  string                 `json:"localDisplayName"`
	Profile           LocalProfile           `json:"profile"`
	ActiveConn        *Connection            `json:"activeConn,omitempty"`
	ViaGroup          *int64                 `json:"viaGroup,omitempty"`
	ContactUsed       bool                   `json:"contactUsed"`
	ContactStatus     string                 `json:"contactStatus"`
	ChatSettings      ChatSettings           `json:"chatSettings"`
	UserPreferences   Preferences            `json:"userPreferences"`
	MergedPreferences ContactUserPreferences `json:"mergedPreferences"`
	CreatedAt         string                 `json:"createdAt"`
	UpdatedAt         string                 `json:"updatedAt"`
	ChatTs            *string                `json:"chatTs,omitempty"`
}

// Connection represents a connection to another user.
type Connection struct {
	ConnID          int64  `json:"connId"`
	AgentConnID     string `json:"agentConnId"`
	ConnChatVersion int    `json:"connChatVersion"`
	ConnLevel       int    `json:"connLevel"`
	ConnType        string `json:"connType"`
	ConnStatus      string `json:"connStatus"`
	CreatedAt       string `json:"createdAt"`
}

// CreatedConnLink represents a created connection link.
type CreatedConnLink struct {
	ConnFullLink  string  `json:"connFullLink"`
	ConnShortLink *string `json:"connShortLink,omitempty"`
}

// ChatSettings represents per-chat settings.
type ChatSettings struct {
	EnableNtfs string `json:"enableNtfs"`
	SendRcpts  *bool  `json:"sendRcpts,omitempty"`
	Favorite   bool   `json:"favorite"`
}

// ContactUserPreferences holds merged preferences for a contact.
type ContactUserPreferences struct {
	TimedMessages *ContactUserPreference `json:"timedMessages,omitempty"`
	FullDelete    *ContactUserPreference `json:"fullDelete,omitempty"`
	Reactions     *ContactUserPreference `json:"reactions,omitempty"`
	Voice         *ContactUserPreference `json:"voice,omitempty"`
	Files         *ContactUserPreference `json:"files,omitempty"`
	Calls         *ContactUserPreference `json:"calls,omitempty"`
}

// ContactUserPreference is a preference with user/contact enable states.
type ContactUserPreference struct {
	Enabled        PrefEnabled `json:"enabled"`
	UserPreference string      `json:"userPreference"`
}

// --- Preference Types ---

// Preferences represents user or contact preferences.
type Preferences struct {
	TimedMessages *TimedMessagesPreference `json:"timedMessages,omitempty"`
	FullDelete    *SimplePreference        `json:"fullDelete,omitempty"`
	Reactions     *SimplePreference        `json:"reactions,omitempty"`
	Voice         *SimplePreference        `json:"voice,omitempty"`
	Files         *SimplePreference        `json:"files,omitempty"`
	Calls         *SimplePreference        `json:"calls,omitempty"`
	Sessions      *SimplePreference        `json:"sessions,omitempty"`
}

// SimplePreference represents a simple on/off preference.
type SimplePreference struct {
	Enable PrefEnabled `json:"enable"`
}

// PrefEnabled represents preference enable states for user and contact.
type PrefEnabled struct {
	ForUser    bool `json:"forUser"`
	ForContact bool `json:"forContact"`
}

// TimedMessagesPreference represents timed message settings.
type TimedMessagesPreference struct {
	Allow string `json:"allow"`
	TTL   *int   `json:"ttl,omitempty"`
}

// --- Group Types ---

// GroupInfo represents group information.
type GroupInfo struct {
	GroupID              int64                `json:"groupId"`
	LocalDisplayName     string               `json:"localDisplayName"`
	GroupProfile         GroupProfile         `json:"groupProfile"`
	LocalAlias           string               `json:"localAlias"`
	FullGroupPreferences FullGroupPreferences `json:"fullGroupPreferences"`
	Membership           GroupMember          `json:"membership"`
	ChatSettings         ChatSettings         `json:"chatSettings"`
	CreatedAt            string               `json:"createdAt"`
	UpdatedAt            string               `json:"updatedAt"`
}

// GroupProfile represents a group's profile.
type GroupProfile struct {
	DisplayName      string            `json:"displayName"`
	FullName         string            `json:"fullName"`
	ShortDescr       *string           `json:"shortDescr,omitempty"`
	Description      *string           `json:"description,omitempty"`
	Image            *string           `json:"image,omitempty"`
	GroupPreferences *GroupPreferences `json:"groupPreferences,omitempty"`
	MemberAdmission  *json.RawMessage  `json:"memberAdmission,omitempty"`
}

// GroupMember represents a group member.
type GroupMember struct {
	GroupMemberID    int64               `json:"groupMemberId"`
	GroupID          int64               `json:"groupId"`
	MemberID         string              `json:"memberId"`
	MemberRole       string              `json:"memberRole"`
	MemberCategory   string              `json:"memberCategory"`
	MemberStatus     string              `json:"memberStatus"`
	MemberSettings   GroupMemberSettings `json:"memberSettings"`
	LocalDisplayName string              `json:"localDisplayName"`
	MemberProfile    LocalProfile        `json:"memberProfile"`
	ActiveConn       *Connection         `json:"activeConn,omitempty"`
	CreatedAt        string              `json:"createdAt"`
}

// GroupMemberSettings represents settings for a group member.
type GroupMemberSettings struct {
	ShowMessages bool `json:"showMessages"`
}

// FullGroupPreferences represents all group preferences with values.
type FullGroupPreferences struct {
	TimedMessages *TimedGroupPreference `json:"timedMessages,omitempty"`
	FullDelete    *GroupPreference      `json:"fullDelete,omitempty"`
	Reactions     *GroupPreference      `json:"reactions,omitempty"`
	Voice         *GroupPreference      `json:"voice,omitempty"`
	Files         *GroupPreference      `json:"files,omitempty"`
	Calls         *GroupPreference      `json:"calls,omitempty"`
}

// GroupPreference represents a group-level preference.
type GroupPreference struct {
	Enable string `json:"enable"`
}

// TimedGroupPreference represents timed message group preference.
type TimedGroupPreference struct {
	Enable string `json:"enable"`
	TTL    *int   `json:"ttl,omitempty"`
}

// GroupPreferences represents group profile preferences.
type GroupPreferences struct {
	TimedMessages *TimedGroupPreference `json:"timedMessages,omitempty"`
	FullDelete    *GroupPreference      `json:"fullDelete,omitempty"`
	Reactions     *GroupPreference      `json:"reactions,omitempty"`
	Voice         *GroupPreference      `json:"voice,omitempty"`
	Files         *GroupPreference      `json:"files,omitempty"`
	Calls         *GroupPreference      `json:"calls,omitempty"`
}

// --- Chat Item Types ---

// AChatItem wraps a ChatItem with its ChatInfo context.
type AChatItem struct {
	ChatInfo ChatInfo `json:"chatInfo"`
	ChatItem ChatItem `json:"chatItem"`
}

// ChatInfo represents information about a chat.
type ChatInfo struct {
	Type              string                    `json:"type"`
	Contact           *Contact                  `json:"contact,omitempty"`
	GroupInfo         *GroupInfo                `json:"groupInfo,omitempty"`
	NoteFolder        *NoteFolder               `json:"noteFolder,omitempty"`
	ContactRequest    *UserContactRequest       `json:"contactRequest,omitempty"`
	ContactConnection *PendingContactConnection `json:"contactConnection,omitempty"`
}

// ChatItem represents a single chat message or item.
type ChatItem struct {
	ChatDir       CIDirection       `json:"chatDir"`
	Meta          CIMeta            `json:"meta"`
	Content       CIContent         `json:"content"`
	FormattedText []FormattedText   `json:"formattedText,omitempty"`
	QuotedItem    *CIQuote          `json:"quotedItem,omitempty"`
	Reactions     []CIReactionCount `json:"reactions"`
	File          *CIFile           `json:"file,omitempty"`
}

// CIDirection represents the direction of a chat item.
type CIDirection struct {
	Type       string `json:"type"`
	MemberID   string `json:"memberId,omitempty"`
	MemberRole string `json:"memberRole,omitempty"`
}

// CIMeta holds metadata for a chat item.
type CIMeta struct {
	ItemID     int64    `json:"itemId"`
	ItemTs     string   `json:"itemTs"`
	ItemText   string   `json:"itemText"`
	ItemStatus CIStatus `json:"itemStatus"`
	ItemEdited bool     `json:"itemEdited"`
	CreatedAt  string   `json:"createdAt"`
}

// CIStatus represents the delivery/read status of a chat item.
type CIStatus struct {
	Type     string `json:"type"`
	Progress int    `json:"progress,omitempty"`
}

// CIContent represents the content of a chat item.
type CIContent struct {
	Type       string      `json:"type"`
	MsgContent *MsgContent `json:"msgContent,omitempty"`
	DeleteMode string      `json:"deleteMode,omitempty"`
}

// MsgContent represents the content of a message.
type MsgContent struct {
	Type     string       `json:"type"`
	Text     string       `json:"text"`
	Image    string       `json:"image,omitempty"`
	Duration int          `json:"duration,omitempty"`
	Preview  *LinkPreview `json:"preview,omitempty"`
}

// LinkPreview represents a link preview in a message.
type LinkPreview struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
}

// FormattedText represents a formatted text segment.
type FormattedText struct {
	Text   string      `json:"text"`
	Format *TextFormat `json:"format,omitempty"`
}

// TextFormat represents a text formatting type.
type TextFormat struct {
	Type  string `json:"type"`
	Color string `json:"color,omitempty"`
}

// CIQuote represents a quoted chat item.
type CIQuote struct {
	ChatDir       string          `json:"chatDir,omitempty"`
	ItemID        *int64          `json:"itemId,omitempty"`
	SentAt        string          `json:"sentAt,omitempty"`
	Content       *CIContent      `json:"content,omitempty"`
	FormattedText []FormattedText `json:"formattedText,omitempty"`
}

// CIReactionCount represents a reaction with its count.
type CIReactionCount struct {
	Reaction    MsgReaction `json:"reaction"`
	UserReacted bool        `json:"userReacted"`
	TotalCount  int         `json:"totalCount"`
}

// MsgReaction represents a message reaction.
type MsgReaction struct {
	Emoji string `json:"emoji,omitempty"`
}

// CIFile represents a file attached to a chat item.
type CIFile struct {
	FileID       int64       `json:"fileId"`
	FileName     string      `json:"fileName"`
	FileSize     int64       `json:"fileSize"`
	FileSource   *CryptoFile `json:"fileSource,omitempty"`
	FileStatus   string      `json:"fileStatus"`
	FileProtocol string      `json:"fileProtocol"`
}

// CryptoFile represents a file with optional encryption.
type CryptoFile struct {
	FilePath   string          `json:"filePath"`
	CryptoArgs *CryptoFileArgs `json:"cryptoArgs,omitempty"`
}

// CryptoFileArgs holds encryption key and nonce for a file.
type CryptoFileArgs struct {
	FileKey   string `json:"fileKey"`
	FileNonce string `json:"fileNonce"`
}

// ComposedMessage is used when sending a new message.
type ComposedMessage struct {
	FileSource   *CryptoFile      `json:"fileSource,omitempty"`
	QuotedItemID *int64           `json:"quotedItemId,omitempty"`
	MsgContent   MsgContent       `json:"msgContent"`
	Mentions     map[string]int64 `json:"mentions,omitempty"`
}

// UpdatedMessage is used when editing an existing message.
type UpdatedMessage struct {
	MsgContent MsgContent `json:"msgContent"`
}

// --- Contact Request Types ---

// UserContactRequest represents a pending contact request.
type UserContactRequest struct {
	ContactRequestID int64        `json:"contactRequestId"`
	LocalDisplayName string       `json:"localDisplayName"`
	Profile          LocalProfile `json:"profile"`
	CreatedAt        string       `json:"createdAt"`
}

// PendingContactConnection represents a pending connection.
type PendingContactConnection struct {
	ConnID     int64  `json:"connId"`
	ConnStatus string `json:"connStatus"`
	CreatedAt  string `json:"createdAt"`
}

// NoteFolder represents a local note folder.
type NoteFolder struct {
	NoteFolderID int64  `json:"noteFolderId"`
	FolderName   string `json:"folderName"`
}

// --- Address Types ---

// UserContactLink represents the user's contact address.
type UserContactLink struct {
	ConnReqContact   string       `json:"connReqContact"`
	ShortLink        *string      `json:"shortLink,omitempty"`
	ShortLinkDataSet *interface{} `json:"shortLinkDataSet,omitempty"`
	AutoAccept       *AutoAccept  `json:"autoAccept,omitempty"`
}

// AutoAccept represents auto-accept settings for an address.
type AutoAccept struct {
	AcceptIncognito bool        `json:"acceptIncognito"`
	AutoReply       *MsgContent `json:"autoReply,omitempty"`
}

// AddressSettings represents settings for a contact address.
type AddressSettings struct {
	AutoAccept *AutoAccept `json:"autoAccept,omitempty"`
}

// --- File Transfer Types ---

// RcvFileTransfer represents a receiving file transfer.
type RcvFileTransfer struct {
	FileID            int64  `json:"fileId"`
	SenderDisplayName string `json:"senderDisplayName"`
}

// FileTransferMeta represents metadata for a file transfer.
type FileTransferMeta struct {
	FileID   int64  `json:"fileId"`
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
}

// RcvFileDescr represents a file description for receiving.
type RcvFileDescr struct {
	FileDescrID int64  `json:"fileDescrId"`
	FileDescr   string `json:"fileDescr"`
}

// --- Error Types ---

// ChatError represents an error from the chat API.
type ChatError struct {
	Type       string          `json:"type"`
	ErrorType  *ChatErrorType  `json:"errorType,omitempty"`
	AgentError *AgentErrorType `json:"agentError,omitempty"`
	StoreError *StoreError     `json:"storeError,omitempty"`
}

// ChatErrorType represents a chat error type.
type ChatErrorType struct {
	Type string `json:"type"`
}

// AgentErrorType represents an agent error type.
type AgentErrorType struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

// StoreError represents a store error.
type StoreError struct {
	Type string `json:"type"`
}

// --- Response-Specific Types ---

// UserContactLinkCreated is the response for APICreateMyAddress.
type UserContactLinkCreated struct {
	Type            string          `json:"type"`
	UserContactLink UserContactLink `json:"userContactLink"`
}

// UserContactLinkDeleted is the response for APIDeleteMyAddress.
type UserContactLinkDeleted struct {
	Type string `json:"type"`
}

// UserContactLinkUpdated is the response for APISetAddressSettings.
type UserContactLinkUpdated struct {
	Type            string          `json:"type"`
	User            User            `json:"user"`
	UserContactLink UserContactLink `json:"userContactLink"`
}

// UserProfileUpdated is the response for APIUpdateProfile and APISetProfileAddress.
type UserProfileUpdated struct {
	Type        string       `json:"type"`
	User        User         `json:"user"`
	Profile     Profile      `json:"profile"`
	Preferences *Preferences `json:"preferences,omitempty"`
}

// NewChatItems is the response/event for APISendMessages and received messages.
type NewChatItems struct {
	Type      string      `json:"type"`
	User      User        `json:"user"`
	ChatItems []AChatItem `json:"chatItems"`
}

// ChatItemUpdated is the response/event for APIUpdateChatItem.
type ChatItemUpdated struct {
	Type     string    `json:"type"`
	User     User      `json:"user"`
	ChatItem AChatItem `json:"chatItem"`
}

// ChatItemsDeleted is the response/event for APIDeleteChatItem.
type ChatItemsDeleted struct {
	Type              string        `json:"type"`
	User              User          `json:"user"`
	ChatItemDeletions []interface{} `json:"chatItemDeletions"`
	ByUser            bool          `json:"byUser"`
	Timed             bool          `json:"timed"`
}

// ChatItemReaction is the response/event for APIChatItemReaction.
type ChatItemReaction struct {
	Type     string      `json:"type"`
	User     User        `json:"user"`
	Added    bool        `json:"added"`
	Reaction interface{} `json:"reaction"`
}

// ChatItemNotChanged is the response when an update results in no change.
type ChatItemNotChanged struct {
	Type     string    `json:"type"`
	User     User      `json:"user"`
	ChatItem AChatItem `json:"chatItem"`
}

// RcvFileAccepted is the response for ReceiveFile.
type RcvFileAccepted struct {
	Type     string    `json:"type"`
	User     User      `json:"user"`
	ChatItem AChatItem `json:"chatItem"`
}

// RcvFileAcceptedSndCancelled is the response when file was cancelled by sender.
type RcvFileAcceptedSndCancelled struct {
	Type string `json:"type"`
	User User   `json:"user"`
}

// SndFileCancelled is the response for CancelFile (sending).
type SndFileCancelled struct {
	Type             string      `json:"type"`
	User             User        `json:"user"`
	FileTransferMeta interface{} `json:"fileTransferMeta"`
	SndCancelled     bool        `json:"sndCancelled"`
}

// RcvFileCancelled is the response for CancelFile (receiving).
type RcvFileCancelled struct {
	Type            string          `json:"type"`
	User            User            `json:"user"`
	RcvFileTransfer RcvFileTransfer `json:"rcvFileTransfer"`
}

// ContactConnected is the event when a contact connects.
type ContactConnected struct {
	Type              string   `json:"type"`
	User              User     `json:"user"`
	Contact           Contact  `json:"contact"`
	UserCustomProfile *Profile `json:"userCustomProfile,omitempty"`
}

// ContactConnecting is the event when a contact is connecting.
type ContactConnecting struct {
	Type    string  `json:"type"`
	User    User    `json:"user"`
	Contact Contact `json:"contact"`
}

// ContactSndReady is the event when a contact is ready to receive messages.
type ContactSndReady struct {
	Type    string  `json:"type"`
	User    User    `json:"user"`
	Contact Contact `json:"contact"`
}

// ReceivedContactRequest is the event when a contact request is received.
type ReceivedContactRequest struct {
	Type           string             `json:"type"`
	User           User               `json:"user"`
	ContactRequest UserContactRequest `json:"contactRequest"`
}

// ChatCmdError is the response when a command fails.
type ChatCmdError struct {
	Type      string    `json:"type"`
	ChatError ChatError `json:"chatError"`
}

// ChatError_ is the event for a chat error.
type ChatError_ struct {
	Type      string    `json:"type"`
	ChatError ChatError `json:"chatError"`
}

// ListUsersResponse is the response for ListUsers.
type ListUsersResponse []UserUI

// UserUI represents a user in the list users response.
type UserUI struct {
	User                User `json:"user"`
	ActiveUser          bool `json:"activeUser"`
	ShowNtfs            bool `json:"showNtfs"`
	AutoAcceptAutomated bool `json:"autoAcceptAutomated"`
}

// --- Utility Types ---

// VersionRange represents a range of supported versions.
type VersionRange struct {
	MinVersion int `json:"minVersion"`
	MaxVersion int `json:"maxVersion"`
}
