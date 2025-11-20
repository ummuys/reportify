package logger

import (
	"io"
	"os"
	"time"

	"github.com/ummuys/reportify/internal/config"

	"github.com/rs/zerolog"
)

func InitLogger(path string) (*config.Loggers, error) {
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
	const consoleTimeFormat = "02 Jan 06 15:04 MST"

	cw := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: consoleTimeFormat,
	}

	// STD-OUT
	file := initLogFile(path)

	multiWriter := io.MultiWriter(file, cw)

	baseLog := zerolog.New(multiWriter).With().Timestamp().Logger()

	logLevels, err := config.ParseLogLevels()
	if err != nil {
		return nil, err
	}

	appLog := baseLog.With().Str("component", "app").Logger().Level(logLevels.AppLvl)
	srvLog := baseLog.With().Str("component", "srv").Logger().Level(logLevels.SrvLvl)
	svcLog := baseLog.With().Str("component", "svc").Logger().Level(logLevels.SvcLvl)
	dbLog := baseLog.With().Str("component", "db").Logger().Level(logLevels.DbLvl)
	chcLog := baseLog.With().Str("component", "chc").Logger().Level(logLevels.ChcLvl)

	return &config.Loggers{
		AppLog: &appLog,
		SrvLog: &srvLog,
		DbLog:  &dbLog,
		SvcLog: &svcLog,
		ChcLog: &chcLog,
	}, nil
}
