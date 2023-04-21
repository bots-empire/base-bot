package log

import "go.uber.org/zap"

type WrapLogger struct {
	logger Logger
}

func NewWrapLogger(logger Logger) *WrapLogger {
	return &WrapLogger{
		logger: logger,
	}
}

func (w *WrapLogger) Log(kv ...interface{}) error {
	fields := make([]zap.Field, 0, len(kv)/2)
	var logLevel string

	for i := 0; i < len(kv); i += 2 {
		k, ok := kv[i].(string)
		if !ok {
			continue
		}
		if i+1 > len(kv) {
			continue
		}

		if k == "level" {
			lvl, ok := kv[i].(string)
			if !ok {
				continue
			}

			logLevel = lvl
			continue
		}

		fields = append(fields, zap.Any(k, kv[i+1]))
	}
	if logLevel == "debug" {
		w.logger.Debug("Report", fields...)
	} else {
		w.logger.Info("Report", fields...)
	}

	return nil
}
