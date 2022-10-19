package msgs

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/bots-empire/base-bot/models"
)

type Service struct {
	Sender models.Sender

	Developers []int64
}

func NewService(sender models.Sender, developers []int64) *Service {
	prometheus.MustRegister(models.InputMessage)
	prometheus.MustRegister(models.OutputMessage)
	return &Service{
		Sender:     sender,
		Developers: developers,
	}
}
