// Copyright © 2020 The Things Industries B.V.

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
)

const usage = `Usage:

      Subscribe as Forwarder:
      $ pbsub --forwarder-net-id NETID [--forwarder-id ID]

      Subscribe as Home Network:
      $ pbsub --home-network-net-id NETID [--filter-forwarder-net-id NETID [--filter-forwarder-id ID]]

Flags:`

var (
	ctx    = context.Background()
	logger *zap.Logger
	conn   *grpc.ClientConn
)

func main() {
	if invalid := !parseInput(); invalid || input.help {
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
		if invalid {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	logger = logging.GetLogger(input.debug)
	defer logger.Sync()

	var err error
	conn, err = client.DialContext(ctx, logger, input.client, 1913)
	if err != nil {
		logger.Fatal("Failed to connect", zap.String("address", input.client.Address), zap.Error(err))
	}
	defer conn.Close()

	switch {
	case input.forwarderNetID != nil:
		runForwarder()
	case input.homeNetworkNetID != nil:
		runHomeNetwork()
	}
}
