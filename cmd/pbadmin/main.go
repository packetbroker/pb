// Copyright © 2020 The Things Industries B.V.

// Command pbadmin configures Packet Broker IAM.
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

      Network registry:
      $ pbctl network list
      $ pbctl network create --net-id NETID [--name NAME]
      $ pbctl network get --net-id NETID
      $ pbctl network update --net-id NETID [--name NAME]
      $ pbctl network delete --net-id NETID

      Tenant registry:
      $ pbctl tenant list --net-id NETID
      $ pbctl tenant create --net-id NETID --tenant-id TENANTID [--dev-addr-prefixes PREFIX,PREFIX]
      $ pbctl tenant get --net-id NETID --tenant-id TENANTID
      $ pbctl tenant update --net-id NETID --tenant-id TENANTID [--dev-addr-prefixes PREFIX,PREFIX]
      $ pbctl tenant delete --net-id NETID --tenant-id TENANTID

Commands:
      network
      tenant

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
	case "network":
		runNetwork(ctx)
	case "tenant":
		runTenant(ctx)
	}
}
