# API Commands and Responses

This file documents all available API commands for SimpleX Chat bots.

## Address Commands

### APICreateMyAddress

Create bot address.

**Network usage:** interactive

**Parameters:**
- `userId`: int64

**Syntax:**
```
/_address <userId>
```

**Responses:**
- `UserContactLinkCreated`: User contact address created
- `ChatCmdError`: Command error

---

### APIDeleteMyAddress

Delete bot address.

**Network usage:** background

**Parameters:**
- `userId`: int64

**Syntax:**
```
/_delete_address <userId>
```

**Responses:**
- `UserContactLinkDeleted`: User contact address deleted
- `ChatCmdError`: Command error

---

### APIShowMyAddress

Get bot address and settings.

**Network usage:** no

**Parameters:**
- `userId`: int64

**Syntax:**
```
/_show_address <userId>
```

**Responses:**
- `UserContactLink`: User contact address
- `ChatCmdError`: Command error

---

### APISetProfileAddress

Add address to bot profile.

**Network usage:** interactive

**Parameters:**
- `userId`: int64
- `enable`: bool

**Syntax:**
```
/_profile_address <userId> on|off
```

**Responses:**
- `UserProfileUpdated`: User profile updated
- `ChatCmdError`: Command error

---

### APISetAddressSettings

Set bot address settings.

**Network usage:** interactive

**Parameters:**
- `userId`: int64
- `settings`: AddressSettings

**Syntax:**
```
/_address_settings <userId> <json(settings)>
```

**Responses:**
- `UserContactLinkUpdated`: User contact address updated
- `ChatCmdError`: Command error

---

## Message Commands

### APISendMessages

Send messages.

**Network usage:** background

**Parameters:**
- `sendRef`: ChatRef
- `liveMessage`: bool
- `ttl`: int?
- `composedMessages`: [ComposedMessage]

**Syntax:**
```
/_send <str(sendRef)>[ live=on][ ttl=<ttl>] json <json(composedMessages)>
```

**Responses:**
- `NewChatItems`: New messages
- `ChatCmdError`: Command error

---

### APIUpdateChatItem

Update message.

**Network usage:** background

**Parameters:**
- `chatRef`: ChatRef
- `chatItemId`: int64
- `liveMessage`: bool
- `updatedMessage`: UpdatedMessage

**Syntax:**
```
/_update item <str(chatRef)> <chatItemId>[ live=on] json <json(updatedMessage)>
```

**Responses:**
- `ChatItemUpdated`: Message updated
- `ChatItemNotChanged`: Message not changed
- `ChatCmdError`: Command error

---

### APIDeleteChatItem

Delete message.

**Network usage:** background

**Parameters:**
- `chatRef`: ChatRef
- `chatItemIds`: [int64]
- `deleteMode`: CIDeleteMode (broadcast|internal|internalMark)

**Syntax:**
```
/_delete item <str(chatRef)> <chatItemIds[0]>[,<chatItemIds[1]>...] broadcast|internal|internalMark
```

**Responses:**
- `ChatItemsDeleted`: Messages deleted
- `ChatCmdError`: Command error

---

### APIDeleteMemberChatItem

Moderate message. Requires Moderator role (and higher than message author's).

**Network usage:** background

**Parameters:**
- `groupId`: int64
- `chatItemIds`: [int64]

**Syntax:**
```
/_delete member item #<groupId> <chatItemIds[0]>[,<chatItemIds[1]>...]
```

**Responses:**
- `ChatItemsDeleted`: Messages deleted
- `ChatCmdError`: Command error

---

### APIChatItemReaction

Add/remove message reaction.

**Network usage:** background

**Parameters:**
- `chatRef`: ChatRef
- `chatItemId`: int64
- `add`: bool
- `reaction`: MsgReaction

**Syntax:**
```
/_reaction <str(chatRef)> <chatItemId> on|off <json(reaction)>
```

**Responses:**
- `ChatItemReaction`: Message reaction
- `ChatCmdError`: Command error

---

## File Commands

### ReceiveFile

Receive file.

**Network usage:** no

**Parameters:**
- `fileId`: int64
- `userApprovedRelays`: bool
- `storeEncrypted`: bool?
- `fileInline`: bool?
- `filePath`: string?

**Syntax:**
```
/freceive <fileId>[ approved_relays=on][ encrypt=on|off][ inline=on|off][ <filePath>]
```

**Responses:**
- `RcvFileAccepted`: File accepted to be received
- `RcvFileAcceptedSndCancelled`: File accepted, but no longer sent
- `ChatCmdError`: Command error

---

### CancelFile

Cancel file.

**Network usage:** background

**Parameters:**
- `fileId`: int64

**Syntax:**
```
/fcancel <fileId>
```

**Responses:**
- `SndFileCancelled`: Cancelled sending file
- `RcvFileCancelled`: Cancelled receiving file
- `ChatCmdError`: Command error

---

## Group Commands

### APIAddMember

Add contact to group. Requires Admin role.

**Parameters:**
- `groupId`: int64
- `contactId`: int64
- `memberRole`: GroupMemberRole

**Syntax:**
```
/_add #<groupId> <contactId> observer|author|member|moderator|admin|owner
```

### APIJoinGroup

Join group.

**Parameters:**
- `groupId`: int64

**Syntax:**
```
/_join #<groupId>
```

### APIAcceptMember

Accept group member. Requires Admin role.

**Parameters:**
- `groupId`: int64
- `groupMemberId`: int64
- `memberRole`: GroupMemberRole

**Syntax:**
```
/_accept member #<groupId> <groupMemberId> observer|author|member|moderator|admin|owner
```

### APIMembersRole

Set members role. Requires Admin role.

**Parameters:**
- `groupId`: int64
- `groupMemberIds`: [int64]
- `memberRole`: GroupMemberRole

**Syntax:**
```
/_member role #<groupId> <groupMemberIds[0]>[,<groupMemberIds[1]>...] observer|author|member|moderator|admin|owner
```

### APIBlockMembersForAll

Block members. Requires Moderator role.

**Parameters:**
- `groupId`: int64
- `groupMemberIds`: [int64]
- `blocked`: bool

**Syntax:**
```
/_block #<groupId> <groupMemberIds[0]>[,<groupMemberIds[1]>...] blocked=on|off
```

### APIRemoveMembers

Remove members. Requires Admin role.

**Parameters:**
- `groupId`: int64
- `groupMemberIds`: [int64]
- `withMessages`: bool

**Syntax:**
```
/_remove #<groupId> <groupMemberIds[0]>[,<groupMemberIds[1]>...][ messages=on]
```

### APILeaveGroup

Leave group.

**Parameters:**
- `groupId`: int64

**Syntax:**
```
/_leave #<groupId>
```

### APIListMembers

Get group members.

**Parameters:**
- `groupId`: int64

**Syntax:**
```
/_members #<groupId>
```

### APINewGroup

Create group.

**Parameters:**
- `userId`: int64
- `incognito`: bool
- `groupProfile`: GroupProfile

**Syntax:**
```
/_group <userId>[ incognito=on] <json(groupProfile)>
```

### APIUpdateGroupProfile

Update group profile.

**Parameters:**
- `groupId`: int64
- `groupProfile`: GroupProfile

**Syntax:**
```
/_group_profile #<groupId> <json(groupProfile)>
```

---

## Group Link Commands

### APICreateGroupLink

Create group link.

**Parameters:**
- `groupId`: int64
- `memberRole`: GroupMemberRole

**Syntax:**
```
/_create link #<groupId> observer|author|member|moderator|admin|owner
```

### APIGroupLinkMemberRole

Set member role for group link.

**Parameters:**
- `groupId`: int64
- `memberRole`: GroupMemberRole

**Syntax:**
```
/_set link role #<groupId> observer|author|member|moderator|admin|owner
```

### APIDeleteGroupLink

Delete group link.

**Parameters:**
- `groupId`: int64

**Syntax:**
```
/_delete link #<groupId>
```

### APIGetGroupLink

Get group link.

**Parameters:**
- `groupId`: int64

**Syntax:**
```
/_get link #<groupId>
```

---

## Connection Commands

### APIAddContact

Create 1-time invitation link.

**Parameters:**
- `userId`: int64
- `incognito`: bool

**Syntax:**
```
/_connect <userId>[ incognito=on]
```

### APIConnectPlan

Determine SimpleX link type and if already connected.

**Parameters:**
- `userId`: int64
- `connectionLink`: string?

**Syntax:**
```
/_connect plan <userId> <connectionLink>
```

### APIConnect

Connect via prepared SimpleX link.

**Parameters:**
- `userId`: int64
- `incognito`: bool
- `preparedLink_`: CreatedConnLink?

**Syntax:**
```
/_connect <userId>[ <str(preparedLink_)>]
```

### Connect

Connect via SimpleX link as string.

**Parameters:**
- `incognito`: bool
- `connLink_`: string?

**Syntax:**
```
/connect[ <connLink_>]
```

### APIAcceptContact

Accept contact request.

**Parameters:**
- `contactReqId`: int64

**Syntax:**
```
/_accept <contactReqId>
```

### APIRejectContact

Reject contact request.

**Parameters:**
- `contactReqId`: int64

**Syntax:**
```
/_reject <contactReqId>
```

---

## Chat Commands

### APIListContacts

Get contacts.

**Parameters:**
- `userId`: int64

**Syntax:**
```
/_contacts <userId>
```

### APIListGroups

Get groups.

**Parameters:**
- `userId`: int64
- `contactId_`: int64?
- `search`: string?

**Syntax:**
```
/_groups <userId>[ @<contactId_>][ <search>]
```

### APIDeleteChat

Delete chat.

**Parameters:**
- `chatRef`: ChatRef
- `chatDeleteMode`: ChatDeleteMode

**Syntax:**
```
/_delete <str(chatRef)> <str(chatDeleteMode)>
```

---

## User Profile Commands

### ShowActiveUser

Get active user profile.

**Syntax:**
```
/user
```

### CreateActiveUser

Create new user profile.

**Parameters:**
- `newUser`: NewUser

**Syntax:**
```
/_create user <json(newUser)>
```

### ListUsers

Get all user profiles.

**Syntax:**
```
/users
```

### APISetActiveUser

Set active user profile.

**Parameters:**
- `userId`: int64
- `viewPwd`: string?

**Syntax:**
```
/_user <userId>[ <json(viewPwd)>]
```

### APIDeleteUser

Delete user profile.

**Parameters:**
- `userId`: int64
- `delSMPQueues`: bool
- `viewPwd`: string?

**Syntax:**
```
/_delete user <userId> del_smp=on|off[ <json(viewPwd)>]
```

### APIUpdateProfile

Update user profile.

**Parameters:**
- `userId`: int64
- `profile`: Profile

**Syntax:**
```
/_profile <userId> <json(profile)>
```

### APISetContactPrefs

Configure chat preference overrides for the contact.

**Parameters:**
- `contactId`: int64
- `preferences`: Preferences

**Syntax:**
```
/_set prefs @<contactId> <json(preferences)>
```
