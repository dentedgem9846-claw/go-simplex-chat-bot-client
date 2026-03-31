# API Types

This file documents all data types used in the SimpleX Chat Bot API.

> **Note:** This documentation is based on the official SimpleX Chat Bot API types reference. For the complete specification, see the [official repository](https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/TYPES.md).

## Core Types

### User

**Record type:**
- `userId`: int64
- `agentUserId`: string
- `userContactId`: int64
- `localDisplayName`: string
- `profile`: LocalProfile
- `activeUser`: bool

### Profile

**Record type:**
- `displayName`: string
- `fullName`: string
- `shortDescr`: string?
- `image`: string?
- `contactLink`: string?
- `preferences`: Preferences?
- `peerType`: ChatPeerType?

### LocalProfile

**Record type:**
- `profileId`: int64
- `displayName`: string
- `fullName`: string
- `shortDescr`: string?
- `image`: string?
- `contactLink`: string?
- `preferences`: Preferences?
- `peerType`: ChatPeerType?
- `localAlias`: string

### Contact

**Record type:**
- `contactId`: int64
- `localDisplayName`: string
- `profile`: LocalProfile
- `activeConn`: Connection?
- `viaGroup`: int64?
- `contactUsed`: bool
- `contactStatus`: ContactStatus
- `chatSettings`: ChatSettings
- `userPreferences`: Preferences
- `mergedPreferences`: ContactUserPreferences
- `createdAt`: UTCTime
- `updatedAt`: UTCTime
- `chatTs`: UTCTime?

## Chat Types

### ChatInfo

**Discriminated union type:**

**Direct:**
- type: "direct"
- contact: Contact

**Group:**
- type: "group"
- groupInfo: GroupInfo
- groupChatScope: GroupChatScopeInfo?

**Local:**
- type: "local"
- noteFolder: NoteFolder

**ContactRequest:**
- type: "contactRequest"
- contactRequest: UserContactRequest

**ContactConnection:**
- type: "contactConnection"
- contactConnection: PendingContactConnection

### ChatType

**Enum type:**
- "direct"
- "group"
- "local"

### ChatRef

**Record type:**
- `chatType`: ChatType
- `chatId`: int64
- `chatScope`: GroupChatScope?

## Chat Items

### AChatItem

**Record type:**
- `chatInfo`: ChatInfo
- `chatItem`: ChatItem

### ChatItem

**Record type:**
- `chatDir`: CIDirection
- `meta`: CIMeta
- `content`: CIContent
- `mentions`: {string : CIMention}
- `formattedText`: [FormattedText]?
- `quotedItem`: CIQuote?
- `reactions`: [CIReactionCount]
- `file`: CIFile?

### CIDirection

**Discriminated union type:**
- `directSnd`: Sent in direct chat
- `directRcv`: Received in direct chat
- `groupSnd`: Sent in group
- `groupRcv`: Received in group (with groupMember)
- `localSnd`: Local sent
- `localRcv`: Local received

### CIContent

**Discriminated union type:**

**SndMsgContent:**
- type: "sndMsgContent"
- msgContent: MsgContent

**RcvMsgContent:**
- type: "rcvMsgContent"
- msgContent: MsgContent

**SndDeleted/RcvDeleted:**
- type: "sndDeleted" | "rcvDeleted"
- deleteMode: CIDeleteMode

### CIMeta

**Record type:**
- `itemId`: int64
- `itemTs`: UTCTime
- `itemText`: string
- `itemStatus`: CIStatus
- `itemEdited`: bool
- `createdAt`: UTCTime

### CIStatus

**Discriminated union type:**
- `sndNew`: Sending new
- `sndSent`: Sent (with progress)
- `sndRcvd`: Received by recipient
- `rcvNew`: Newly received
- `rcvRead`: Read by user

## Message Types

### MsgContent

**Discriminated union type:**

**Text:**
- type: "text"
- text: string

**Link:**
- type: "link"
- text: string
- preview: LinkPreview

**Image:**
- type: "image"
- text: string
- image: string

**Video:**
- type: "video"
- text: string
- image: string
- duration: int

**Voice:**
- type: "voice"
- text: string
- duration: int

**File:**
- type: "file"
- text: string

### ComposedMessage

**Record type:**
- `fileSource`: CryptoFile?
- `quotedItemId`: int64?
- `msgContent`: MsgContent
- `mentions`: {string : int64}

### UpdatedMessage

**Record type:**
- `msgContent`: MsgContent

## Group Types

### Group

**Record type:**
- `groupInfo`: GroupInfo
- `members`: [GroupMember]

### GroupInfo

**Record type:**
- `groupId`: int64
- `localDisplayName`: string
- `groupProfile`: GroupProfile
- `localAlias`: string
- `fullGroupPreferences`: FullGroupPreferences
- `membership`: GroupMember
- `chatSettings`: ChatSettings
- `createdAt`: UTCTime
- `updatedAt`: UTCTime

### GroupProfile

**Record type:**
- `displayName`: string
- `fullName`: string
- `shortDescr`: string?
- `description`: string?
- `image`: string?
- `groupPreferences`: GroupPreferences?
- `memberAdmission`: GroupMemberAdmission?

### GroupMember

**Record type:**
- `groupMemberId`: int64
- `groupId`: int64
- `memberId`: string
- `memberRole`: GroupMemberRole
- `memberCategory`: GroupMemberCategory
- `memberStatus`: GroupMemberStatus
- `memberSettings`: GroupMemberSettings
- `localDisplayName`: string
- `memberProfile`: LocalProfile
- `activeConn`: Connection?
- `createdAt`: UTCTime

### GroupMemberRole

**Enum type:**
- "observer"
- "author"
- "member"
- "moderator"
- "admin"
- "owner"

### GroupMemberStatus

**Enum type:**
- "rejected"
- "removed"
- "left"
- "deleted"
- "unknown"
- "invited"
- "pending_approval"
- "pending_review"
- "introduced"
- "accepted"
- "announced"
- "connected"
- "complete"
- "creator"

## File Types

### CIFile

**Record type:**
- `fileId`: int64
- `fileName`: string
- `fileSize`: int64
- `fileSource`: CryptoFile?
- `fileStatus`: CIFileStatus
- `fileProtocol`: FileProtocol

### CIFileStatus

**Discriminated union type:**

**Sending:**
- sndStored
- sndTransfer (with progress)
- sndCancelled
- sndComplete

**Receiving:**
- rcvInvitation
- rcvAccepted
- rcvTransfer (with progress)
- rcvAborted
- rcvComplete
- rcvCancelled

### FileProtocol

**Enum type:**
- "smp"
- "xftp"
- "local"

### CryptoFile

**Record type:**
- `filePath`: string
- `cryptoArgs`: CryptoFileArgs?

### CryptoFileArgs

**Record type:**
- `fileKey`: string
- `fileNonce`: string

## Connection Types

### Connection

**Record type:**
- `connId`: int64
- `agentConnId`: string
- `connChatVersion`: int
- `peerChatVRange`: VersionRange
- `connLevel`: int
- `connType`: ConnType
- `connStatus`: ConnStatus
- `createdAt`: UTCTime

### ConnType

**Enum type:**
- "contact"
- "member"
- "user_contact"

### ConnStatus

**Enum type:**
- "new"
- "prepared"
- "joined"
- "requested"
- "accepted"
- "snd-ready"
- "ready"
- "deleted"

### CreatedConnLink

**Record type:**
- `connFullLink`: string
- `connShortLink`: string?

## Preference Types

### Preferences

**Record type:**
- `timedMessages`: TimedMessagesPreference?
- `fullDelete`: SimplePreference?
- `reactions`: SimplePreference?
- `voice`: SimplePreference?
- `files`: SimplePreference?
- `calls`: SimplePreference?
- `sessions`: SimplePreference?
- `commands`: [ChatBotCommand]?

### SimplePreference

**Record type:**
- `enable`: PrefEnabled

### PrefEnabled

**Record type:**
- `forUser`: bool
- `forContact`: bool

### ChatBotCommand

**Discriminated union type:**

**Command:**
- type: "command"
- keyword: string
- label: string
- params: string?

**Menu:**
- type: "menu"
- label: string
- commands: [ChatBotCommand]

## Error Types

### ChatError

**Discriminated union type:**

**Error:**
- type: "error"
- errorType: ChatErrorType

**ErrorAgent:**
- type: "errorAgent"
- agentError: AgentErrorType
- connectionEntity_: ConnectionEntity?

**ErrorStore:**
- type: "errorStore"
- storeError: StoreError

### ChatErrorType

**Discriminated union type:**
- `noActiveUser`
- `userExists` (with contactName)
- `chatNotStarted`
- `contactNotReady`
- `groupUserRole`
- `fileNotFound`
- And many more...

### AgentErrorType

**Discriminated union type:**
- `CMD` (command error)
- `CONN` (connection error)
- `SMP` (SMP server error)
- `XFTP` (file transfer error)
- `BROKER` (broker error)
- `INTERNAL` (internal error)
- And more...

## Utility Types

### ChatSettings

**Record type:**
- `enableNtfs`: MsgFilter
- `sendRcpts`: bool?
- `favorite`: bool

### ChatStats

**Record type:**
- `unreadCount`: int
- `unreadMentions`: int
- `minUnreadItemId`: int64
- `unreadChat`: bool

### CIDeleteMode

**Enum type:**
- "broadcast" - Delete for everyone
- "internal" - Delete locally only
- "internalMark" - Mark as deleted locally

### ChatDeleteMode

**Discriminated union type:**
- `full` (with notify option)
- `entity` (with notify option)
- `messages`

### MsgReaction

**Discriminated union type:**
- `emoji`: string
- `unknown`: tag, json

### Format

**Discriminated union type:**
- `bold`, `italic`, `strikeThrough`
- `snippet`, `secret`
- `colored` (with Color)
- `uri`, `hyperLink`
- `simplexLink`
- `mention`
- And more...

---

## Additional Resources

For the complete type definitions, see:
- [Official TYPES.md](https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/TYPES.md)
- [Commands Reference](./COMMANDS.md)
- [Events Reference](./EVENTS.md)
