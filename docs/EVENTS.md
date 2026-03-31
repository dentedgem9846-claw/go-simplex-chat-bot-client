# API Events

This file documents all events that SimpleX Chat CLI sends to bots.

## Contact Connection Events

### ContactConnected

**Type:** `contactConnected`

Sent after a user connects via bot SimpleX address (not a business address).

**Record type:**
```json
{
  "type": "contactConnected",
  "user": User,
  "contact": Contact,
  "userCustomProfile": Profile?
}
```

---

### ContactUpdated

**Type:** `contactUpdated`

Contact profile of another user is updated.

**Record type:**
```json
{
  "type": "contactUpdated",
  "user": User,
  "fromContact": Contact,
  "toContact": Contact
}
```

---

### ContactDeletedByContact

**Type:** `contactDeletedByContact`

Bot user's connection with another contact is deleted (conversation is kept).

**Record type:**
```json
{
  "type": "contactDeletedByContact",
  "user": User,
  "contact": Contact
}
```

---

### ReceivedContactRequest

**Type:** `receivedContactRequest`

Contact request received. This event is only sent when auto-accept is disabled.

**Record type:**
```json
{
  "type": "receivedContactRequest",
  "user": User,
  "contactRequest": UserContactRequest,
  "chat_": AChat?
}
```

---

### NewMemberContactReceivedInv

**Type:** `newMemberContactReceivedInv`

Received invitation to connect directly with a group member.

**Record type:**
```json
{
  "type": "newMemberContactReceivedInv",
  "user": User,
  "contact": Contact,
  "groupInfo": GroupInfo,
  "member": GroupMember
}
```

---

### ContactSndReady

**Type:** `contactSndReady`

Connecting via 1-time invitation or after accepting contact request. After this event bot can send messages to this contact.

**Record type:**
```json
{
  "type": "contactSndReady",
  "user": User,
  "contact": Contact
}
```

---

## Message Events

### NewChatItems

**Type:** `newChatItems`

Received message(s). This is the main event bots use to process received messages.

**Record type:**
```json
{
  "type": "newChatItems",
  "user": User,
  "chatItems": [AChatItem]
}
```

---

### ChatItemReaction

**Type:** `chatItemReaction`

Received message reaction.

**Record type:**
```json
{
  "type": "chatItemReaction",
  "user": User,
  "added": bool,
  "reaction": ACIReaction
}
```

---

### ChatItemsDeleted

**Type:** `chatItemsDeleted`

Message was deleted by another user.

**Record type:**
```json
{
  "type": "chatItemsDeleted",
  "user": User,
  "chatItemDeletions": [ChatItemDeletion],
  "byUser": bool,
  "timed": bool
}
```

---

### ChatItemUpdated

**Type:** `chatItemUpdated`

Message was updated by another user.

**Record type:**
```json
{
  "type": "chatItemUpdated",
  "user": User,
  "chatItem": AChatItem
}
```

---

### GroupChatItemsDeleted

**Type:** `groupChatItemsDeleted`

Group messages are deleted or moderated.

**Record type:**
```json
{
  "type": "groupChatItemsDeleted",
  "user": User,
  "groupInfo": GroupInfo,
  "chatItemIDs": [int64],
  "byUser": bool,
  "member_": GroupMember?
}
```

---

### ChatItemsStatusesUpdated

**Type:** `chatItemsStatusesUpdated`

Message delivery status updates.

**Record type:**
```json
{
  "type": "chatItemsStatusesUpdated",
  "user": User,
  "chatItems": [AChatItem]
}
```

---

## Group Events

### ReceivedGroupInvitation

**Type:** `receivedGroupInvitation`

Received group invitation.

**Record type:**
```json
{
  "type": "receivedGroupInvitation",
  "user": User,
  "groupInfo": GroupInfo,
  "contact": Contact,
  "fromMemberRole": GroupMemberRole,
  "memberRole": GroupMemberRole
}
```

---

### UserJoinedGroup

**Type:** `userJoinedGroup`

Bot user joined group. Received when connection via group link completes.

**Record type:**
```json
{
  "type": "userJoinedGroup",
  "user": User,
  "groupInfo": GroupInfo,
  "hostMember": GroupMember
}
```

---

### GroupUpdated

**Type:** `groupUpdated`

Group profile or preferences updated.

**Record type:**
```json
{
  "type": "groupUpdated",
  "user": User,
  "fromGroup": GroupInfo,
  "toGroup": GroupInfo,
  "member_": GroupMember?
}
```

---

### JoinedGroupMember

**Type:** `joinedGroupMember`

Another member joined group.

**Record type:**
```json
{
  "type": "joinedGroupMember",
  "user": User,
  "groupInfo": GroupInfo,
  "member": GroupMember
}
```

---

### MemberRole

**Type:** `memberRole`

Member (or bot user's) group role changed.

**Record type:**
```json
{
  "type": "memberRole",
  "user": User,
  "groupInfo": GroupInfo,
  "byMember": GroupMember,
  "member": GroupMember,
  "fromRole": GroupMemberRole,
  "toRole": GroupMemberRole
}
```

---

### DeletedMember

**Type:** `deletedMember`

Another member is removed from the group.

**Record type:**
```json
{
  "type": "deletedMember",
  "user": User,
  "groupInfo": GroupInfo,
  "byMember": GroupMember,
  "deletedMember": GroupMember,
  "withMessages": bool
}
```

---

### LeftMember

**Type:** `leftMember`

Another member left the group.

**Record type:**
```json
{
  "type": "leftMember",
  "user": User,
  "groupInfo": GroupInfo,
  "member": GroupMember
}
```

---

### DeletedMemberUser

**Type:** `deletedMemberUser`

Bot user was removed from the group.

**Record type:**
```json
{
  "type": "deletedMemberUser",
  "user": User,
  "groupInfo": GroupInfo,
  "member": GroupMember,
  "withMessages": bool
}
```

---

### GroupDeleted

**Type:** `groupDeleted`

Group was deleted by the owner (not bot user).

**Record type:**
```json
{
  "type": "groupDeleted",
  "user": User,
  "groupInfo": GroupInfo,
  "member": GroupMember
}
```

---

### ConnectedToGroupMember

**Type:** `connectedToGroupMember`

Connected to another group member.

**Record type:**
```json
{
  "type": "connectedToGroupMember",
  "user": User,
  "groupInfo": GroupInfo,
  "member": GroupMember,
  "memberContact": Contact?
}
```

---

### MemberAcceptedByOther

**Type:** `memberAcceptedByOther`

Another group owner, admin or moderator accepted member to the group after review.

**Record type:**
```json
{
  "type": "memberAcceptedByOther",
  "user": User,
  "groupInfo": GroupInfo,
  "acceptingMember": GroupMember,
  "member": GroupMember
}
```

---

### MemberBlockedForAll

**Type:** `memberBlockedForAll`

Another member blocked for all members.

**Record type:**
```json
{
  "type": "memberBlockedForAll",
  "user": User,
  "groupInfo": GroupInfo,
  "byMember": GroupMember,
  "member": GroupMember,
  "blocked": bool
}
```

---

### GroupMemberUpdated

**Type:** `groupMemberUpdated`

Another group member profile updated.

**Record type:**
```json
{
  "type": "groupMemberUpdated",
  "user": User,
  "groupInfo": GroupInfo,
  "fromMember": GroupMember,
  "toMember": GroupMember
}
```

---

## File Events

### RcvFileDescrReady

**Type:** `rcvFileDescrReady`

File is ready to be received. This event is useful for processing sender file servers and monitoring file reception progress.

**Record type:**
```json
{
  "type": "rcvFileDescrReady",
  "user": User,
  "chatItem": AChatItem,
  "rcvFileTransfer": RcvFileTransfer,
  "rcvFileDescr": RcvFileDescr
}
```

---

### RcvFileComplete

**Type:** `rcvFileComplete`

File reception is completed.

**Record type:**
```json
{
  "type": "rcvFileComplete",
  "user": User,
  "chatItem": AChatItem
}
```

---

### SndFileCompleteXFTP

**Type:** `sndFileCompleteXFTP`

File upload is completed.

**Record type:**
```json
{
  "type": "sndFileCompleteXFTP",
  "user": User,
  "chatItem": AChatItem,
  "fileTransferMeta": FileTransferMeta
}
```

---

### RcvFileStart

**Type:** `rcvFileStart`

File reception started. This event will be sent after `CEvtRcvFileDescrReady` event.

**Record type:**
```json
{
  "type": "rcvFileStart",
  "user": User,
  "chatItem": AChatItem
}
```

---

### RcvFileSndCancelled

**Type:** `rcvFileSndCancelled`

File was cancelled by the sender.

**Record type:**
```json
{
  "type": "rcvFileSndCancelled",
  "user": User,
  "chatItem": AChatItem,
  "rcvFileTransfer": RcvFileTransfer
}
```

---

### RcvFileAccepted

**Type:** `rcvFileAccepted`

This event will be sent when file is automatically accepted because of CLI option.

**Record type:**
```json
{
  "type": "rcvFileAccepted",
  "user": User,
  "chatItem": AChatItem
}
```

---

### RcvFileError

**Type:** `rcvFileError`

Error receiving file.

**Record type:**
```json
{
  "type": "rcvFileError",
  "user": User,
  "chatItem_": AChatItem?,
  "agentError": AgentErrorType,
  "rcvFileTransfer": RcvFileTransfer
}
```

---

### RcvFileWarning

**Type:** `rcvFileWarning`

Warning when receiving file.

**Record type:**
```json
{
  "type": "rcvFileWarning",
  "user": User,
  "chatItem_": AChatItem?,
  "agentError": AgentErrorType,
  "rcvFileTransfer": RcvFileTransfer
}
```

---

### SndFileError

**Type:** `sndFileError`

Error sending file.

**Record type:**
```json
{
  "type": "sndFileError",
  "user": User,
  "chatItem_": AChatItem?,
  "fileTransferMeta": FileTransferMeta,
  "errorMessage": string
}
```

---

### SndFileWarning

**Type:** `sndFileWarning`

Warning when sending file.

**Record type:**
```json
{
  "type": "sndFileWarning",
  "user": User,
  "chatItem_": AChatItem?,
  "fileTransferMeta": FileTransferMeta,
  "errorMessage": string
}
```

---

## Connection Progress Events

### AcceptingContactRequest

**Type:** `acceptingContactRequest`

Automatically accepting contact request via bot's SimpleX address with auto-accept enabled.

**Record type:**
```json
{
  "type": "acceptingContactRequest",
  "user": User,
  "contact": Contact
}
```

---

### AcceptingBusinessRequest

**Type:** `acceptingBusinessRequest`

Automatically accepting contact request via bot's business address.

**Record type:**
```json
{
  "type": "acceptingBusinessRequest",
  "user": User,
  "groupInfo": GroupInfo
}
```

---

### ContactConnecting

**Type:** `contactConnecting`

Contact confirmed connection. Sent when contact started connecting via bot's 1-time invitation link or when bot connects to another SimpleX address.

**Record type:**
```json
{
  "type": "contactConnecting",
  "user": User,
  "contact": Contact
}
```

---

### BusinessLinkConnecting

**Type:** `businessLinkConnecting`

Contact confirmed connection. Sent when bot connects to another business address.

**Record type:**
```json
{
  "type": "businessLinkConnecting",
  "user": User,
  "groupInfo": GroupInfo,
  "hostMember": GroupMember,
  "fromContact": Contact
}
```

---

### JoinedGroupMemberConnecting

**Type:** `joinedGroupMemberConnecting`

Group member is announced to the group and will be connecting to bot.

**Record type:**
```json
{
  "type": "joinedGroupMemberConnecting",
  "user": User,
  "groupInfo": GroupInfo,
  "hostMember": GroupMember,
  "member": GroupMember
}
```

---

### SentGroupInvitation

**Type:** `sentGroupInvitation`

Sent when another user joins group via bot's link.

**Record type:**
```json
{
  "type": "sentGroupInvitation",
  "user": User,
  "groupInfo": GroupInfo,
  "contact": Contact,
  "member": GroupMember
}
```

---

### GroupLinkConnecting

**Type:** `groupLinkConnecting`

Sent when bot joins group via another user link.

**Record type:**
```json
{
  "type": "groupLinkConnecting",
  "user": User,
  "groupInfo": GroupInfo,
  "hostMember": GroupMember
}
```

---

## Error Events

### MessageError

**Type:** `messageError`

**Record type:**
```json
{
  "type": "messageError",
  "user": User,
  "severity": string,
  "errorMessage": string
}
```

---

### ChatError

**Type:** `chatError`

**Record type:**
```json
{
  "type": "chatError",
  "chatError": ChatError
}
```

---

### ChatErrors

**Type:** `chatErrors`

**Record type:**
```json
{
  "type": "chatErrors",
  "chatErrors": [ChatError]
}
```

---

## Important Notes

1. **Must allow unknown events:** Your bot must allow and ignore all events it does not process. It should not fail when encountering undocumented event types.

2. **Must allow additional properties:** Your bot's JSON parser must allow additional properties in all types.

3. **Must allow unknown union tags:** Your bot must allow and ignore records with unknown union tags and unknown enum strings.
