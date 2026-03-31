package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	sxtypes "simplex-chat-bot/types"
)

// setupTestClient creates a test client with a mock server
func setupTestClient(t *testing.T, handler func(ws *websocket.Conn, msg sxtypes.Command)) (*Client, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			t.Fatalf("accept websocket: %v", err)
		}
		defer ws.Close(websocket.StatusNormalClosure, "")

		for {
			var msg sxtypes.Command
			err := wsjson.Read(r.Context(), ws, &msg)
			if err != nil {
				return
			}
			handler(ws, msg)
		}
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	c := New(wsURL, WithOperationTimeout(5*time.Second))

	ctx := context.Background()
	if err := c.Connect(ctx); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	return c, func() {
		c.Close()
		server.Close()
	}
}

// Test Profile API Methods
func TestShowActiveUser(t *testing.T) {
	c, cleanup := setupTestClient(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		if msg.Cmd == "/user" {
			resp := sxtypes.Response{
				CorrID: msg.CorrID,
				Resp: json.RawMessage(`{
					"userId": 1,
					"agentUserId": "agent1",
					"userContactId": 1,
					"localDisplayName": "testuser",
					"profile": {
						"profileId": 1,
						"displayName": "Test User",
						"fullName": "Test User Full",
						"localAlias": ""
					},
					"activeUser": true
				}`),
			}
			wsjson.Write(context.Background(), ws, resp)
		}
	})
	defer cleanup()

	user, err := c.ShowActiveUser(context.Background())
	if err != nil {
		t.Fatalf("ShowActiveUser failed: %v", err)
	}

	if user.UserID != 1 {
		t.Errorf("expected userId 1, got %d", user.UserID)
	}
	if user.LocalDisplayName != "testuser" {
		t.Errorf("expected localDisplayName 'testuser', got %s", user.LocalDisplayName)
	}
	if !user.ActiveUser {
		t.Error("expected activeUser to be true")
	}
}

func TestListUsers(t *testing.T) {
	c, cleanup := setupTestClient(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		if msg.Cmd == "/users" {
			resp := sxtypes.Response{
				CorrID: msg.CorrID,
				Resp: json.RawMessage(`[
					{
						"user": {
							"userId": 1,
							"agentUserId": "agent1",
							"userContactId": 1,
							"localDisplayName": "user1",
							"profile": {
								"profileId": 1,
								"displayName": "User One",
								"fullName": "User One",
								"localAlias": ""
							},
							"activeUser": true
						},
						"activeUser": true,
						"showNtfs": true,
						"autoAcceptAutomated": false
					}
				]`),
			}
			wsjson.Write(context.Background(), ws, resp)
		}
	})
	defer cleanup()

	users, err := c.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
	if users[0].User.LocalDisplayName != "user1" {
		t.Errorf("expected user1, got %s", users[0].User.LocalDisplayName)
	}
}

func TestCreateActiveUser(t *testing.T) {
	c, cleanup := setupTestClient(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		if strings.HasPrefix(msg.Cmd, "/_create user") {
			resp := sxtypes.Response{
				CorrID: msg.CorrID,
				Resp: json.RawMessage(`{
					"userId": 2,
					"agentUserId": "agent2",
					"userContactId": 2,
					"localDisplayName": "newuser",
					"profile": {
						"profileId": 2,
						"displayName": "New User",
						"fullName": "New User Full",
						"localAlias": ""
					},
					"activeUser": true
				}`),
			}
			wsjson.Write(context.Background(), ws, resp)
		}
	})
	defer cleanup()

	newUser := sxtypes.NewUser{
		Profile: sxtypes.Profile{
			DisplayName: "newuser",
			FullName:    "New User Full",
		},
		ShortName: "newuser",
	}

	user, err := c.CreateActiveUser(context.Background(), newUser)
	if err != nil {
		t.Fatalf("CreateActiveUser failed: %v", err)
	}

	if user.LocalDisplayName != "newuser" {
		t.Errorf("expected localDisplayName 'newuser', got %s", user.LocalDisplayName)
	}
}

// Test Contacts API Methods
func TestListContacts(t *testing.T) {
	c, cleanup := setupTestClient(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		if strings.HasPrefix(msg.Cmd, "/_contacts") {
			resp := sxtypes.Response{
				CorrID: msg.CorrID,
				Resp: json.RawMessage(`[
					{
						"contactId": 1,
						"localDisplayName": "alice",
						"profile": {
							"profileId": 1,
							"displayName": "Alice",
							"fullName": "Alice Smith",
							"localAlias": ""
						},
						"activeConn": null,
						"contactUsed": true,
						"contactStatus": "active",
						"chatSettings": {"enableNtfs": "all", "favorite": false},
						"userPreferences": {},
						"mergedPreferences": {},
						"createdAt": "2024-01-01T00:00:00Z",
						"updatedAt": "2024-01-01T00:00:00Z"
					}
				]`),
			}
			wsjson.Write(context.Background(), ws, resp)
		}
	})
	defer cleanup()

	contacts, err := c.ListContacts(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListContacts failed: %v", err)
	}

	if len(contacts) != 1 {
		t.Errorf("expected 1 contact, got %d", len(contacts))
	}
	if contacts[0].LocalDisplayName != "alice" {
		t.Errorf("expected 'alice', got %s", contacts[0].LocalDisplayName)
	}
}

func TestAPIAddContact(t *testing.T) {
	c, cleanup := setupTestClient(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		if strings.HasPrefix(msg.Cmd, "/_connect 1") && !strings.Contains(msg.Cmd, "plan") {
			resp := sxtypes.Response{
				CorrID: msg.CorrID,
				Resp: json.RawMessage(`{
					"connFullLink": "https://simplex.chat/contact#/?v=1&smp=smp%3A%2F%2F...",
					"connShortLink": null
				}`),
			}
			wsjson.Write(context.Background(), ws, resp)
		}
	})
	defer cleanup()

	link, err := c.AddContact(context.Background(), 1, false)
	if err != nil {
		t.Fatalf("APIAddContact failed: %v", err)
	}

	if link.ConnFullLink == "" {
		t.Error("expected connFullLink to be set")
	}
}

// Test Groups API Methods
func TestCreateGroup(t *testing.T) {
	c, cleanup := setupTestClient(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		if strings.HasPrefix(msg.Cmd, "/_group") {
			resp := sxtypes.Response{
				CorrID: msg.CorrID,
				Resp: json.RawMessage(`{
					"groupInfo": {
						"groupId": 1,
						"localDisplayName": "testgroup",
						"groupProfile": {
							"displayName": "Test Group",
							"fullName": "Test Group Full"
						},
						"localAlias": "",
						"fullGroupPreferences": {},
						"membership": {
							"groupMemberId": 1,
							"groupId": 1,
							"memberId": "member1",
							"memberRole": "owner",
							"memberCategory": "user",
							"memberStatus": "connected",
							"memberSettings": {"showMessages": true},
							"localDisplayName": "testuser",
							"memberProfile": {
								"profileId": 1,
								"displayName": "Test User",
								"fullName": "Test User",
								"localAlias": ""
							},
							"createdAt": "2024-01-01T00:00:00Z"
						},
						"chatSettings": {"enableNtfs": "all", "favorite": false},
						"createdAt": "2024-01-01T00:00:00Z",
						"updatedAt": "2024-01-01T00:00:00Z"
					}
				}`),
			}
			wsjson.Write(context.Background(), ws, resp)
		}
	})
	defer cleanup()

	groupProfile := sxtypes.GroupProfile{
		DisplayName: "Test Group",
		FullName:    "Test Group Full",
	}

	group, err := c.NewGroup(context.Background(), 1, false, groupProfile)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if group.LocalDisplayName != "testgroup" {
		t.Errorf("expected 'testgroup', got %s", group.LocalDisplayName)
	}
}

func TestListGroups(t *testing.T) {
	c, cleanup := setupTestClient(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		if strings.HasPrefix(msg.Cmd, "/_groups") {
			resp := sxtypes.Response{
				CorrID: msg.CorrID,
				Resp: json.RawMessage(`[
					{
						"groupInfo": {
							"groupId": 1,
							"localDisplayName": "group1",
							"groupProfile": {
								"displayName": "Group One",
								"fullName": "Group One"
							},
							"localAlias": "",
							"fullGroupPreferences": {},
							"membership": {
								"groupMemberId": 1,
								"groupId": 1,
								"memberId": "member1",
								"memberRole": "owner",
								"memberCategory": "user",
								"memberStatus": "connected",
								"memberSettings": {"showMessages": true},
								"localDisplayName": "user1",
								"memberProfile": {
									"profileId": 1,
									"displayName": "User One",
									"fullName": "User One",
									"localAlias": ""
								},
								"createdAt": "2024-01-01T00:00:00Z"
							},
							"chatSettings": {"enableNtfs": "all", "favorite": false},
							"createdAt": "2024-01-01T00:00:00Z",
							"updatedAt": "2024-01-01T00:00:00Z"
						}
					}
				]`),
			}
			wsjson.Write(context.Background(), ws, resp)
		}
	})
	defer cleanup()

	groups, err := c.ListGroups(context.Background(), 1, nil, nil)
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if groups[0].LocalDisplayName != "group1" {
		t.Errorf("expected 'group1', got %s", groups[0].LocalDisplayName)
	}
}

// Test Messages API Methods
func TestSendMessages(t *testing.T) {
	c, cleanup := setupTestClient(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		if strings.HasPrefix(msg.Cmd, "/_send") {
			resp := sxtypes.Response{
				CorrID: msg.CorrID,
				Resp: json.RawMessage(`{
					"chatItems": [
						{
							"chatInfo": {
								"type": "direct",
								"contact": {
									"contactId": 1,
									"localDisplayName": "alice",
									"profile": {
										"profileId": 1,
										"displayName": "Alice",
										"fullName": "Alice Smith",
										"localAlias": ""
									},
									"activeConn": null,
									"contactUsed": true,
									"contactStatus": "active",
									"chatSettings": {"enableNtfs": "all", "favorite": false},
									"userPreferences": {},
									"mergedPreferences": {},
									"createdAt": "2024-01-01T00:00:00Z",
									"updatedAt": "2024-01-01T00:00:00Z"
								}
							},
							"chatItem": {
								"chatDir": {"directSnd": {}},
								"meta": {
									"itemId": 1,
									"itemTs": "2024-01-01T00:00:00Z",
									"itemText": "Hello!",
									"itemStatus": "sndNew",
									"itemEdited": false,
									"createdAt": "2024-01-01T00:00:00Z"
								},
								"content": {
									"type": "sndMsgContent",
									"msgContent": {
										"type": "text",
										"text": "Hello!"
									}
								},
								"mentions": {},
								"formattedText": null,
								"quotedItem": null,
								"reactions": [],
								"file": null
							}
						}
					]
				}`),
			}
			wsjson.Write(context.Background(), ws, resp)
		}
	})
	defer cleanup()

	chatRef := sxtypes.ChatRef{
		ChatType: "direct",
		ChatID:   1,
	}

	messages := []sxtypes.ComposedMessage{
		{
			MsgContent: sxtypes.MsgContent{
				Type: "text",
				Text: "Hello!",
			},
		},
	}

	chatItems, err := c.SendMessages(context.Background(), chatRef, false, nil, messages)
	if err != nil {
		t.Fatalf("SendMessages failed: %v", err)
	}

	if len(chatItems) != 1 {
		t.Errorf("expected 1 chat item, got %d", len(chatItems))
	}
}

// Test Address API Methods
func TestCreateMyAddress(t *testing.T) {
	c, cleanup := setupTestClient(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		if msg.Cmd == "/_address 1" {
			resp := sxtypes.Response{
				CorrID: msg.CorrID,
				Resp: json.RawMessage(`{
					"userContactLink": {
						"connReqContact": "https://simplex.chat/contact#/?v=1&smp=smp%3A%2F%2F...",
						"autoAccept": null
					}
				}`),
			}
			wsjson.Write(context.Background(), ws, resp)
		}
	})
	defer cleanup()

	link, err := c.CreateMyAddress(context.Background(), 1)
	if err != nil {
		t.Fatalf("CreateMyAddress failed: %v", err)
	}

	if link.ConnReqContact == "" {
		t.Error("expected connReqContact to be set")
	}
}
