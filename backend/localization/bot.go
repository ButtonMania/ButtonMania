package localization

import (
	"embed"
	"fmt"

	"buttonmania.win/protocol"
)

//go:embed en/bot/start.txt
//go:embed en/bot/donate.txt
//go:embed ru/bot/start.txt
//go:embed ru/bot/donate.txt
var fsBot embed.FS

// Bot is responsible for loading and providing localized strings for bot commands.
type BotLocalization struct {
	localization map[protocol.UserLocale]map[string]string
}

// loadLocalizedString loads localized strings from the embedded filesystem.
func loadLocalizedStrings(fs embed.FS, locale protocol.UserLocale, command string) (string, error) {
	filename := fmt.Sprintf("%s/bot/%s.txt", locale, command)
	content, err := fs.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// NewMessagesLocalizationLocalization creates a new BotLocalization instance.
func NewBotLocalization() (*BotLocalization, error) {
	localization := make(map[protocol.UserLocale]map[string]string)
	for _, locale := range []protocol.UserLocale{protocol.EN, protocol.RU} {
		strings := make(map[string]string)
		for _, command := range []string{"start", "donate"} {
			str, err := loadLocalizedStrings(fsBot, locale, command)
			if err != nil {
				return nil, err
			}
			strings[command] = str
		}
		localization[locale] = strings
	}

	return &BotLocalization{
		localization,
	}, nil
}

// LocalizedCommandString returns a localized string for bot command.
func (b *BotLocalization) LocalizedCommandString(locale protocol.UserLocale, command string) string {
	strings := b.localization[locale]
	return strings[command]
}
