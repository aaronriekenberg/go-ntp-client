package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/beevik/ntp"
)

var (
	network   = flag.String("network", "udp6", "network to use")
	sloglevel slog.Level
)

func parseFlags() (ntpServers []string) {
	flag.TextVar(&sloglevel, "sloglevel", slog.LevelInfo, "slog level")

	flag.Parse()

	if flag.NArg() == 0 {
		panic("no ntp server specified")
	}

	ntpServers = flag.Args()
	return
}

func setupSlog() {
	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					Level: sloglevel,
				},
			),
		),
	)

	slog.Info("setupSlog",
		"sloglevel", sloglevel,
	)
}

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

	ntpServers := parseFlags()

	setupSlog()

	for _, ntpServer := range ntpServers {
		slog.Info("quering server",
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
			panic(fmt.Errorf("ntp.QueryWithOptions error: %w", err))
		}

		slog.Info("server response",
			"response", response,
			"clockOffset", response.ClockOffset.String(),
			"precision", response.Precision.String(),
			"rootDelay", response.RootDelay.String(),
			"rootDispersion", response.RootDispersion.String(),
			"rootDistance", response.RootDistance.String(),
			"rtt", response.RTT.String(),
		)
	}
}
