package models

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	InputMessage = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_input_send_message",
			Help: "Total incoming messages",
		},
		[]string{"bot_lang"},
	)

	OutputMessage = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_output_send_message",
			Help: "Total sent messages",
		},
		[]string{"bot_lang"},
	)

	SendToDevError = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "send_to_developers_error",
			Help: "Error then send msg to developers",
		},
		[]string{"developer_id", "error"},
	)
)
