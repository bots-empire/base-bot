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
)

func init() {
	prometheus.MustRegister(InputMessage)
	prometheus.MustRegister(OutputMessage)
}
