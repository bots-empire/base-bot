package mailing

import (
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	messages *msgs.Service

	messageConfigs     map[string]map[int]tgbotapi.MessageConfig
	photoMessageConfig map[string]map[int]tgbotapi.PhotoConfig
	videoMessageConfig map[string]map[int]tgbotapi.VideoConfig

	usersPerIteration int
}
