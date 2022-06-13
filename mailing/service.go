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

	startSignaller    chan interface{}
	usersPerIteration int
	debugMode         bool
}

func NewService(messages *msgs.Service, userPerIter int) *Service {
	return (&Service{
		messages:          messages,
		startSignaller:    make(chan interface{}, 1),
		usersPerIteration: userPerIter,
	}).init()
}

func (s *Service) init() *Service {
	go s.startSenderHandler()
	return s
}

func (s *Service) debugModeOn() {
	s.debugMode = true
}

func (s *Service) debugModeOff() {
	s.debugMode = false
}
