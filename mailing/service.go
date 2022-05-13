package mailing

import (
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	messages *msgs.Service

	messageConfigs     map[int]tgbotapi.MessageConfig
	photoMessageConfig map[int]tgbotapi.PhotoConfig
	videoMessageConfig map[int]tgbotapi.VideoConfig

	usersPerIteration int
}

func NewService(messages *msgs.Service, userPerIter int) *Service {
	return &Service{
		messages:          messages,
		usersPerIteration: userPerIter,
	}
}
