package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/beevik/ntp"
)

func setupSlog() {
	level := slog.LevelInfo

	if levelString, ok := os.LookupEnv("LOG_LEVEL"); ok {
		err := level.UnmarshalText([]byte(levelString))
		if err != nil {
			panic(fmt.Errorf("level.UnmarshalText error %w", err))
		}
	}

	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				os.Stdout, &slog.HandlerOptions{
					Level: level,
				},
			),
		),
	)

	slog.Info("setupSlog",
		"configuredLevel", level,
	)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic in main",
				"error", err,
			)
			os.Exit(1)
		}
	}()

	setupSlog()

	if len(os.Args) != 2 {
		panic("ntp server required as command line arument")
	}

	ntpServer := os.Args[1]

	logger := slog.Default()

	logger.Info("quering server",
		"ntpServer", ntpServer,
	)

	response, err := ntp.Query(ntpServer)
	if err != nil {
		logger.Warn("error quering ntp server",
			"error", err,
		)
		os.Exit(1)
	}

	rttMilliseconds := response.RTT.Milliseconds()

	logger.Info("server response",
		"response", response,
		"rttMilliseconds", rttMilliseconds,
	)
}
