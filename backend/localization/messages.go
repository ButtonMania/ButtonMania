package localization

import (
	"embed"
	"fmt"
	"math/rand"
	"strings"

	"buttonmania.win/protocol"
)

//go:embed en/messages/love.txt
//go:embed en/messages/fortune.txt
//go:embed en/messages/peace.txt
//go:embed en/messages/prestige.txt
//go:embed ru/messages/love.txt
//go:embed ru/messages/fortune.txt
//go:embed ru/messages/peace.txt
//go:embed ru/messages/prestige.txt
var fsMessagesLocalization embed.FS

// MessagesLocalization is responsible for loading and providing localized messages.
type MessagesLocalization struct {
	buttonType protocol.ButtonType
	messages   map[protocol.UserLocale][]string
}

// loadLocalizedMessages loads localized messages from the embedded filesystem.
func loadLocalizedMessages(fs embed.FS, locale protocol.UserLocale, buttonType protocol.ButtonType) ([]string, error) {
	filename := fmt.Sprintf("%s/messages/%s.txt", locale, buttonType)
	content, err := fs.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

// NewMessagesLocalization creates a new MessagesLocalization instance.
func NewMessagesLocalization(buttonType protocol.ButtonType) (*MessagesLocalization, error) {
	messages := make(map[protocol.UserLocale][]string)
	for _, locale := range []protocol.UserLocale{protocol.EN, protocol.RU} {
		msgs, err := loadLocalizedMessages(fsMessagesLocalization, locale, buttonType)
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
