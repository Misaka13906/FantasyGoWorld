package logger

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/natefinch/lumberjack"
)

func InitLogger() {
	// logFilePath := viper.GetString("project.log_file_path")
	// if logFilePath == "" {
	// 	logFilePath = config.DefaultLogFilePath
	// }
	logFilePath := "logs/app.log"

	zap.ReplaceGlobals(zap.New(
		zapcore.NewCore(
			getEncoder(),
			getWriteSyncer(logFilePath),
			getLogLevel(),
		),
		zap.AddCaller(),
	))

}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogLevel() zapcore.Level {
	// if viper.GetString("project.mode") == "dev" {
	return zapcore.DebugLevel
	// } else {
	// 	return zapcore.ErrorLevel
	// }
}

func getWriteSyncer(logFilePath string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    5,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	// if viper.GetString("project.mode") == "dev" {
	return zapcore.AddSync(io.MultiWriter(os.Stdout, lumberJackLogger))
	// }
	// return zapcore.AddSync(lumberJackLogger)
}
