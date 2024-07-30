package chat

import (
	"context"

	"buttonmania.win/db"
	"buttonmania.win/protocol"
)

// Chat represents the chat queu.
type Chat struct {
	db *db.DB
}

// NewChat creates a new chat instance.
func NewChat(
	ctx context.Context,
	db *db.DB,
) (*Chat, error) {
	return &Chat{
		db: db,
	}, nil
}

func (c *Chat) InitConsumerGroup(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) error {
	return c.db.InitChatConsumerGroup(
		clientId,
		roomId,
	)
}

func (c *Chat) AddConsumerToGroup(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userId protocol.UserID,
) error {
	return c.db.AddConsumerToGroup(
		clientId,
		roomId,
		userId,
	)
}

func (c *Chat) PushMessage(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	msg *protocol.ChatMessage,
) error {
	return c.db.PushChatMessage(
		clientId,
		roomId,
		*msg,
	)
}

func (c *Chat) PopMessage(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) (*protocol.ChatMessage, error) {
	msg, err := c.db.PopChatMessage(clientId, roomId, userID)
	return &msg, err
}
