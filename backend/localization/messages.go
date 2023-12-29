package localization

import (
	"embed"
	"fmt"
	"math/rand"
	"strings"

	"buttonmania.win/protocol"
)

//go:embed en/messages/buttonmania/love.txt
//go:embed en/messages/buttonmania/fortune.txt
//go:embed en/messages/buttonmania/peace.txt
//go:embed en/messages/buttonmania/prestige.txt
//go:embed ru/messages/buttonmania/love.txt
//go:embed ru/messages/buttonmania/fortune.txt
//go:embed ru/messages/buttonmania/peace.txt
//go:embed ru/messages/buttonmania/prestige.txt
var fsMessagesLocalization embed.FS

// MessagesLocalization is responsible for loading and providing localized messages.
type MessagesLocalization struct {
	buttonType protocol.ButtonType
	messages   map[protocol.UserLocale][]string
}

// loadLocalizedMessages loads localized messages from the embedded filesystem.
func loadLocalizedMessages(
	fs embed.FS,
	locale protocol.UserLocale,
	clientId protocol.ClientID,
	buttonType protocol.ButtonType,
) ([]string, error) {
	filename := fmt.Sprintf(
		"%s/messages/%s/%s.txt",
		locale,
		clientId,
		buttonType,
	)
	content, err := fs.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

// NewMessagesLocalization creates a new MessagesLocalization instance.
func NewMessagesLocalization(clientId protocol.ClientID, buttonType protocol.ButtonType) (*MessagesLocalization, error) {
	messages := make(map[protocol.UserLocale][]string)
	for _, locale := range []protocol.UserLocale{protocol.EN, protocol.RU} {
		msgs, err := loadLocalizedMessages(
			fsMessagesLocalization,
			locale,
			clientId,
			buttonType,
		)
		if err != nil {
			return nil, err
		}
		messages[locale] = msgs
	}

	return &MessagesLocalization{
		buttonType,
		messages,
	}, nil
}

// RandomLocalizedMessage returns a random localized message.
func (s *MessagesLocalization) RandomLocalizedMessage(locale protocol.UserLocale) *string {
	messages := s.messages[locale]
	return &messages[rand.Intn(len(messages))]
}
