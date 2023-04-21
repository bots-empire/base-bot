package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func SecureField(key, val string) zapcore.Field {
	if zap.L().Core().Enabled(zapcore.DebugLevel) {
		return zapcore.Field{Key: key, Type: zapcore.StringType, String: val}
	}
	return zapcore.Field{Key: key, Type: zapcore.StringType, String: "***"}
}
