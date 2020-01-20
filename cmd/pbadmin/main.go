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

      Get policy:
      $ pbadmin policy get --forwarder-net-id NETID [--forwarder-id ID] [--defaults|--home-network-net-id NETID]

      Set policy:
      $ pbadmin policy set --forwarder-net-id NETID [--forwarder-id ID] [--defaults|--home-network-net-id NETID] \
            [--set-uplink JMASLD|--unset-uplink] [--set-downlink JMA|--unset-downlink]

Commands:
      policy

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
	conn, err = client.DialContext(ctx, logger, input.client, 1912)
	if err != nil {
		logger.Fatal("Failed to connect", zap.String("address", input.client.Address), zap.Error(err))
	}
	defer conn.Close()

	switch input.mode {
	case "policy":
		runPolicy()
	}
}
