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
	network   = flag.String("network", "udp", "network to use")
	sloglevel slog.Level
)

func parseFlags() {
	flag.TextVar(&sloglevel, "sloglevel", slog.LevelInfo, "slog level")

	flag.Parse()
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

func getNTPServers() []string {
	if flag.NArg() == 0 {
		panic("no ntp server specified")
	}

	return flag.Args()
}

type dialerFunc = func(localAddress, remoteAddress string) (net.Conn, error)

func createDialer(logger *slog.Logger) dialerFunc {

	return func(localAddress, remoteAddress string) (net.Conn, error) {
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

		logger.Info("dialing",
			"network", network,
			"laddr", laddr,
			"raddr", raddr,
		)

		return net.DialUDP(*network, laddr, raddr)
	}
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

	parseFlags()

	setupSlog()

	for _, ntpServer := range getNTPServers() {
		logger := slog.Default().With(
			"network", *network,
			"ntpServer", ntpServer,
		)

		logger.Info("quering server")

		response, err := ntp.QueryWithOptions(
			ntpServer,
			ntp.QueryOptions{
				Dialer: createDialer(logger),
			},
		)

		if err != nil {
			panic(fmt.Errorf("ntp.QueryWithOptions error: %w", err))
		}

		logger.Info("server response",
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
