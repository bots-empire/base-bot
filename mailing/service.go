package mailing

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/bots-empire/base-bot/msgs"
)

type Service struct {
	messages *msgs.Service

	messageConfigs     map[int]tgbotapi.MessageConfig
	photoMessageConfig map[int]tgbotapi.PhotoConfig
	videoMessageConfig map[int]tgbotapi.VideoConfig

	startSignaller    chan interface{}
	usersPerIteration int
	dbType            string
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

func (s *Service) DebugModeOn() {
	s.debugMode = true
}

func (s *Service) DebugModeOff() {
	s.debugMode = false
}

func (s *Service) setMySQLdb() {
	s.dbType = MySQL
}

func (s *Service) setPSQLdb() {
	s.dbType = PSQL
}
