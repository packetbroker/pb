// Copyright © 2020 The Things Industries B.V.

package main

import (
	"fmt"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/config"
	"go.packetbroker.org/pb/pkg/client"
)

type inputData struct {
	help, debug         bool
	client              *client.Config
	forwarderNetIDHex   string
	forwarderNetID      *packetbroker.NetID
	forwarderTenantID   string
	homeNetworkNetIDHex string
	homeNetworkTenantID string
	homeNetworkNetID    *packetbroker.NetID
	mode                string
	tenant              struct {
		devAddrPrefixes []string
	}
	policy struct {
		defaults    bool
		setUplink   string
		setDownlink string
		unset       bool
	}
}

var input = new(inputData)

func parseInput() bool {
	config.CommonFlags(&input.help, &input.debug)
	config.ClientFlags("cp.packetbroker.io:443")
	config.BasicAuthClientFlags(config.BasicAuthRouter)
	config.OAuth2ClientFlags()

	if len(os.Args) < 2 {
		flag.Parse()
		return false
	}

	switch input.mode = strings.ToLower(os.Args[1]); input.mode {
	case "policy":
		if !parsePolicyFlags() {
			return false
		}
	case "route":
		if !parseRouteFlags() {
			return false
		}
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	if !input.help {
		var err error
		input.client, err = config.AutomaticClient(ctx, config.BasicAuthRouter, "networks")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid client settings:", err)
			return false
		}
	}

	return true
}
