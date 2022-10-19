package msgs

import (
	"github.com/bots-empire/base-bot/models"
)

type Service struct {
	Sender models.Sender

	Developers []int64
}

func NewService(sender models.Sender, developers []int64) *Service {
	return &Service{
		Sender:     sender,
		Developers: developers,
	}
}
