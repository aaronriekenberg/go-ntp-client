package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
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

var network *string

func dialer(localAddress, remoteAddress string) (net.Conn, error) {
	var laddr *net.UDPAddr
	if localAddress != "" {
		var err error
		laddr, err = net.ResolveUDPAddr(*network, net.JoinHostPort(localAddress, "0"))
		if err != nil {
			return nil, err
		}
	}

	raddr, err := net.ResolveUDPAddr(*network, remoteAddress)
	if err != nil {
		return nil, err
	}

	slog.Info("dialing",
		"network", network,
		"laddr", laddr,
		"raddr", raddr,
	)

	return net.DialUDP(*network, laddr, raddr)
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

	logger := slog.Default()

	network = flag.String("network", "udp6", "network to use")

	flag.Parse()

	ntpServer := flag.Arg(0)
	if len(ntpServer) == 0 {
		panic("no server specified")
	}

	logger.Info("quering server",
		"network", *network,
		"ntpServer", ntpServer,
	)

	response, err := ntp.QueryWithOptions(
		ntpServer,
		ntp.QueryOptions{
			Dialer: dialer,
		},
	)
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
