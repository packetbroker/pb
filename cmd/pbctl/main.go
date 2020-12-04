// Copyright © 2020 The Things Industries B.V.

// Command pbctl configures Packet Broker Router.
package main

import (
	"context"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"go.packetbroker.org/pb/cmd/internal/logging"
	"go.packetbroker.org/pb/pkg/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"htdvisser.dev/exp/clicontext"
)

const usage = `Usage:

      Routing policy management:
      $ pbctl policy list [--forwarder-net-id NETID] [--defaults]
      $ pbctl policy get --forwarder-net-id NETID [--forwarder-tenant-id TENANTID] \
            [--defaults|--home-network-net-id NETID [--home-network-tenant-id TENANTID]]
      $ pbctl policy set --forwarder-net-id NETID [--forwarder-id ID] \
            [--defaults|--home-network-net-id NETID [--home-network-tenant-id TENANTID]] \
            --set-uplink JMASL --set-downlink JMA
      $ pbctl policy set --forwarder-net-id NETID [--forwarder-id ID] \
            [--defaults|--home-network-net-id NETID [--home-network-tenant-id TENANTID]] \
            --unset

      Routes:
      $ pbctl route list

Commands:
      policy
      route

Flags:`

var (
	ctx      = context.Background()
	logger   *zap.Logger
	conn     *grpc.ClientConn
	exitCode int
)

func main() {
	ctx := clicontext.WithExitCode(ctx, &exitCode)
	defer func() {
		os.Exit(exitCode)
	}()

	if invalid := !parseInput(); invalid || input.help {
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
		if invalid {
			exitCode = 1
			return
		}
		return
	}

	logger = logging.GetLogger(input.debug)
	defer logger.Sync()

	var err error
	conn, err = client.DialContext(ctx, logger, input.client, 1900)
	if err != nil {
		logger.Error("Failed to connect", zap.String("address", input.client.Address), zap.Error(err))
		exitCode = 1
		return
	}
	defer conn.Close()

	switch input.mode {
	case "policy":
		runPolicy(ctx)
	case "route":
		runRoute(ctx)
	}
}
