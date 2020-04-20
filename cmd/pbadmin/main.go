// Copyright © 2020 The Things Industries B.V.

// Command pbadmin configures Packet Broker.
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

      Tenant management:
      $ pbadmin tenant list --net-id NETID
      $ pbadmin tenant get --net-id NETID [--tenant-id TENANTID]
      $ pbadmin tenant set --net-id NETID [--tenant-id TENANTID] [--dev-addr-prefixes PREFIX,PREFIX]
      $ pbadmin tenant delete --net-id NETID [--tenant-id TENANTID]

      Routing policy management:
      $ pbadmin policy get --forwarder-net-id NETID [--forwarder-tenant-id TENANTID] \
            [--defaults|--home-network-net-id NETID [--home-network-tenant-id TENANTID]]
      $ pbadmin policy set --forwarder-net-id NETID [--forwarder-id ID] \
            [--defaults|--home-network-net-id NETID [--home-network-tenant-id TENANTID]] \
            --set-uplink JMASL --set-downlink JMA
      $ pbadmin policy set --forwarder-net-id NETID [--forwarder-id ID] \
            [--defaults|--home-network-net-id NETID [--home-network-tenant-id TENANTID]] \
            --unset

Commands:
      tenant
      policy

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
	case "tenant":
		runTenant(ctx)
	case "policy":
		runPolicy(ctx)
	}
}
