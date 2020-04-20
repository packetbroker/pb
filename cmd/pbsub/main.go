// Copyright © 2020 The Things Industries B.V.

// Command pbsub subscribes to Packet Broker traffic.
package main

import (
	"context"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"go.packetbroker.org/pb/cmd/internal/logging"
	"go.packetbroker.org/pb/internal/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"htdvisser.dev/exp/clicontext"
)

const usage = `Usage:

      Subscribe as Forwarder:
      $ pbsub --forwarder-net-id NETID [--forwarder-id ID] \
          [--forwarder-tenant-id TENANTID]

      Subscribe as Home Network:
      $ pbsub --home-network-net-id NETID [--home-network-tenant-id TENANTID] \
          [--filter-forwarder-net-id NETID [--filter-forwarder-id ID]]

Flags:`

var (
	ctx    = context.Background()
	logger *zap.Logger
	conn   *grpc.ClientConn
)

func main() {
	ctx, exit := clicontext.WithInterruptAndExit(ctx)
	defer exit()

	if invalid := !parseInput(); invalid || input.help {
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
		if invalid {
			clicontext.SetExitCode(ctx, 1)
		}
		return
	}

	logger = logging.GetLogger(input.debug)
	defer logger.Sync()

	var err error
	conn, err = client.DialContext(ctx, logger, input.client, 1900)
	if err != nil {
		logger.Error("Failed to connect", zap.String("address", input.client.Address), zap.Error(err))
		clicontext.SetExitCode(ctx, 1)
		return
	}
	defer conn.Close()

	switch {
	case input.forwarderNetID != nil:
		runForwarder(ctx)
	case input.homeNetworkNetID != nil:
		runHomeNetwork(ctx)
	}
}
