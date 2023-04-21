package msgs

import (
	"github.com/bots-empire/base-bot/log"
	"github.com/bots-empire/base-bot/models"
)

type Service struct {
	Sender models.Sender

	logger     log.Logger
	Developers []int64
}

func NewService(sender models.Sender, logger log.Logger, developers []int64) *Service {
	return &Service{
		Sender:     sender,
		logger:     logger,
		Developers: developers,
	}
}
