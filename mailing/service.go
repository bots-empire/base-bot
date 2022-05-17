package mailing

import (
	"fmt"
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

func NewService(messages *msgs.Service, userPerIter int, lang string, initiatorID int64) *Service {
	service := &Service{
		messages:          messages,
		usersPerIteration: userPerIter,
	}

	service.messages.SendNotificationToDeveloper(
		fmt.Sprintf("%s // mailing continued", lang),
		false,
	)

	sendToUsers, _ := service.mailToUserWithPagination(initiatorID)
	service.sendRespMsgToMailingInitiator(initiatorID, "complete_mailing_text", sendToUsers)

	return service
}
