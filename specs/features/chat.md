# Chat (v3)

## Overview

Implement the ClickUp Chat v3 API — channels, messages, reactions, and replies. This is the largest v3 domain with 19 endpoints.

**Why**: Chat is ClickUp's built-in messaging system. CLI access enables automated notifications, channel provisioning, and message management.

**Requires**: v3 base URL support and workspace ID configuration (see architecture spec).

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /api/v3/workspaces/{}/chat/channels | Retrieve Channels | getChatChannels | v3 |
| GET | /api/v3/workspaces/{}/chat/channels/{} | Retrieve a Channel | getChatChannel | v3 |
| GET | /api/v3/workspaces/{}/chat/channels/{}/followers | Retrieve Channel followers | getChatChannelFollowers | v3 |
| GET | /api/v3/workspaces/{}/chat/channels/{}/members | Retrieve Channel members | getChatChannelMembers | v3 |
| GET | /api/v3/workspaces/{}/chat/channels/{}/messages | Retrieve Channel messages | getChatMessages | v3 |
| POST | /api/v3/workspaces/{}/chat/channels | Create a Channel | createChatChannel | v3 |
| POST | /api/v3/workspaces/{}/chat/channels/direct_message | Create a Direct Message | createDirectMessageChatChannel | v3 |
| POST | /api/v3/workspaces/{}/chat/channels/location | Create Channel on Location | createLocationChatChannel | v3 |
| PATCH | /api/v3/workspaces/{}/chat/channels/{} | Update a Channel | updateChatChannel | v3 |
| DELETE | /api/v3/workspaces/{}/chat/channels/{} | Delete a Channel | deleteChatChannel | v3 |
| POST | /api/v3/workspaces/{}/chat/channels/{}/messages | Send a message | createChatMessage | v3 |
| GET | /api/v3/workspaces/{}/chat/messages/{}/reactions | Retrieve message reactions | getChatMessageReactions | v3 |
| GET | /api/v3/workspaces/{}/chat/messages/{}/replies | Retrieve message replies | getChatMessageReplies | v3 |
| GET | /api/v3/workspaces/{}/chat/messages/{}/tagged_users | Retrieve tagged users | getChatMessageTaggedUsers | v3 |
| POST | /api/v3/workspaces/{}/chat/messages/{}/reactions | Create a reaction | createChatReaction | v3 |
| POST | /api/v3/workspaces/{}/chat/messages/{}/replies | Create a reply | createReplyMessage | v3 |
| PATCH | /api/v3/workspaces/{}/chat/messages/{} | Update a message | patchChatMessage | v3 |
| DELETE | /api/v3/workspaces/{}/chat/messages/{} | Delete a message | deleteChatMessage | v3 |
| DELETE | /api/v3/workspaces/{}/chat/messages/{}/reactions/{} | Delete a reaction | deleteChatReaction | v3 |

## User Stories

### US-001: List Channels

**CLI Command:** `clickup chat channels [--workspace <id>]`

**JSON Output:**
```json
{"channels": [{"id": "chan_123", "name": "General", "type": "public", "member_count": 10}]}
```

**Plain Output (TSV):** Headers: `ID	NAME	TYPE	MEMBER_COUNT`
```
chan_123	General	public	10
```

### US-002: Get Channel Details

**CLI Command:** `clickup chat channel <channel_id>`

### US-003: Channel Followers/Members

**CLI Commands:**
- `clickup chat channel-followers <channel_id>`
- `clickup chat channel-members <channel_id>`

**Plain Output (TSV):** Headers: `ID	USERNAME`

### US-004: Create Channel

**CLI Command:** `clickup chat create-channel [--workspace <id>] --name <name>`

### US-005: Create Direct Message

**CLI Command:** `clickup chat create-dm [--workspace <id>] --members <user_id1,user_id2>`

### US-006: Create Location Channel

**CLI Command:** `clickup chat create-location-channel [--workspace <id>] --parent-type <space|folder|list> --parent-id <id> --name <name>`

### US-007: Update/Delete Channel

**CLI Commands:**
- `clickup chat update-channel <channel_id> [--name "..."]`
- `clickup chat delete-channel <channel_id>`

### US-008: List Channel Messages

**CLI Command:** `clickup chat messages <channel_id> [--limit <n>] [--cursor <token>]`

**JSON Output:**
```json
{
  "data": [
    {"id": "msg_123", "content": "Hello!", "user_id": "1", "type": "message", "date_created": 1700000000000, "replies_count": 2}
  ],
  "pagination": {"next_page_token": "abc"}
}
```

**Plain Output (TSV):** Headers: `ID	USER_ID	TYPE	DATE	CONTENT	REPLIES`
```
msg_123	1	message	1700000000000	Hello!	2
```

### US-009: Send/Update/Delete Message

**CLI Commands:**
- `clickup chat send <channel_id> --text "..."`
- `clickup chat update-message <message_id> --text "..."`
- `clickup chat delete-message <message_id>`

### US-010: Reactions

**CLI Commands:**
- `clickup chat reactions <message_id>`
- `clickup chat react <message_id> --emoji <emoji>`
- `clickup chat unreact <message_id> <reaction_id>`

**Plain Output (TSV):** Headers: `ID	USER_ID	EMOJI	DATE`

### US-011: Replies

**CLI Commands:**
- `clickup chat replies <message_id>`
- `clickup chat reply <message_id> --text "..."`

### US-012: Tagged Users

**CLI Command:** `clickup chat tagged-users <message_id>`

**Plain Output (TSV):** Headers: `ID	USERNAME`

## Request/Response Types

```go
type ChatChannel struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Type        string `json:"type"` // public, private, direct_message
    MemberCount int    `json:"member_count,omitempty"`
}

type ChatMessage struct {
    ID            string      `json:"id"`
    Content       string      `json:"content"`
    UserID        string      `json:"user_id"`
    Type          string      `json:"type"` // message, post
    DateCreated   json.Number `json:"date_created"`
    DateUpdated   json.Number `json:"date_updated"`
    ParentChannel string      `json:"parent_channel"`
    ParentMessage string      `json:"parent_message,omitempty"`
    Resolved      bool        `json:"resolved"`
    RepliesCount  int         `json:"replies_count"`
}

type ChatReaction struct {
    ID          string      `json:"id"`
    MessageID   string      `json:"message_id"`
    UserID      string      `json:"user_id"`
    Reaction    string      `json:"reaction"`
    DateCreated json.Number `json:"date_created"`
}

type ChatChannelsResponse struct {
    Channels []ChatChannel `json:"channels"`
}

type ChatMessagesResponse struct {
    Data       []ChatMessage   `json:"data"`
    Pagination *ChatPagination `json:"pagination,omitempty"`
}

type ChatPagination struct {
    NextPageToken string `json:"next_page_token,omitempty"`
}

type CreateChatChannelRequest struct {
    Name string `json:"name"`
}

type CreateDMRequest struct {
    Members []string `json:"members"` // user IDs
}

type CreateLocationChannelRequest struct {
    Name       string `json:"name"`
    ParentType string `json:"parent_type"` // space, folder, list
    ParentID   string `json:"parent_id"`
}

type SendMessageRequest struct {
    Content string `json:"content"`
}

type CreateReactionRequest struct {
    Reaction string `json:"reaction"`
}

type UpdateChannelRequest struct {
    Name string `json:"name,omitempty"`
}

type UpdateMessageRequest struct {
    Content string `json:"content,omitempty"`
}
```

## Edge Cases

- v3 uses PATCH for updates (not PUT)
- Cursor-based pagination with `next_page_token`
- Message content may contain rich text/markdown
- Reactions use emoji strings, not codes
- DM channels require exactly 2 members
- Location channels are tied to a space/folder/list

## Feedback Loops

### Unit Tests
```go
func TestChatService_ListChannels(t *testing.T)     { /* channels list */ }
func TestChatService_GetChannel(t *testing.T)       { /* single channel */ }
func TestChatService_CreateChannel(t *testing.T)    { /* create */ }
func TestChatService_CreateDM(t *testing.T)         { /* direct message */ }
func TestChatService_UpdateChannel(t *testing.T)    { /* PATCH */ }
func TestChatService_DeleteChannel(t *testing.T)    { /* delete */ }
func TestChatService_ListMessages(t *testing.T)     { /* with pagination */ }
func TestChatService_SendMessage(t *testing.T)      { /* send */ }
func TestChatService_UpdateMessage(t *testing.T)    { /* PATCH */ }
func TestChatService_DeleteMessage(t *testing.T)    { /* delete */ }
func TestChatService_Reactions(t *testing.T)        { /* list/add/remove */ }
func TestChatService_Replies(t *testing.T)          { /* list/create */ }
func TestChatService_TaggedUsers(t *testing.T)      { /* tagged users */ }
```

## Technical Requirements

- New `ChatService` on `clickup.Client`
- All paths use `v3Path()` helper
- PATCH method required (see architecture spec)
- Cursor pagination (not page-based)
- Workspace ID required for all operations
