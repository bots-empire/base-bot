package models

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	InputMessage = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_send_message",
			Help: "Total count send message",
		},
		[]string{"bot_lang", "input"},
	)

	OutputMessage = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_send_message",
			Help: "Total count send message",
		},
		[]string{"bot_lang", "output"},
	)
)
