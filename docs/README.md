# SimpleX Chat Bot API Documentation

This documentation covers the SimpleX Chat Bot API for building bot clients.

## API Types Reference

- [API Commands and Responses](COMMANDS.md)
- [API Events](EVENTS.md)
- [API Types](TYPES.md)

## Overview

SimpleX Chat bots are participants in the SimpleX network that can:
- Send and receive messages and files
- Connect to addresses and join groups
- Manage user profiles
- Handle commands from users

## Key Concepts

### Bot Profile

To distinguish a SimpleX user profile as a bot, set its `peerType` property to `"bot"`. This can be done:

- Using CLI options `--create-bot-display-name` and `--create-bot-allow-files` when first starting CLI
- Using command `/create bot [files=on] [<display_name>]` for additional profiles
- Using API commands to configure bot commands

### Bot Commands

Bot commands are messages that start with `/` character. Configured bot commands are offered to users as a menu.

Configure commands with:
```
/set bot commands <commands>
```

Where `<commands>` follows this syntax:
```
commands = <commandOrMenu>[,<commandOrMenu>...]
commandOrMenu = command | menu
command = '<label>':/'<keyword>[ <params>]'
menu = '<label>':{<commands>}
```

### WebSocket Communication

SimpleX Chat CLI runs as a local WebSocket server:
```bash
simplex-chat -p 5225
```

Your bot connects via WebSocket on the chosen port. All communication happens via JSON-encoded WebSocket text messages.

#### Command Format
```json
{
  "corrId": "<any unique string>",
  "cmd": "<command string>"
}
```

#### Response Format
```json
{
  "corrId": "<corrId sent with command>",
  "resp": {
    "type": "<response record tag>",
    "...": null
  }
}
```

#### Event Format
```json
{
  "resp": {
    "type": "<event record tag>",
    "...": null
  }
}
```

## Security Considerations

- WebSockets API does not support authentication
- CLI binds only to localhost to prevent accidental public access
- Messages are not encrypted — do not send via public networks
- For remote bots, use a web proxy (Caddy/Nginx) with TLS termination and HTTP basic auth

## Use Cases

- Customer support (single or multi-agent)
- Information search and retrieval bots
- Moderation bots for groups and communities
- Broadcast bots for status updates
- Feedback bots
- P2P trading bots
