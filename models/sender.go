package models

import (
	"database/sql"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Sender interface {
	GetBot() *tgbotapi.BotAPI
	GetDataBase() *sql.DB

	AvailableLang() []string
	GetCurrency() string
	GetTexts(lang string) map[string]string

	CheckAdmin(userID int64) bool
	AdminLang(userID int64) string
	AdminText(adminLang, key string) string

	UpdateBlockedUsers(channel int)

	GetAdvertURL(userLang string, channel int) string
	GetAdvertText(userLang string, channel int) string
	GetAdvertisingPhoto(lang string, channel int) string
	GetAdvertisingVideo(lang string, channel int) string
	ButtonUnderAdvert() bool
	AdvertisingChoice(channel int) string
}
