// Copyright © 2020 The Things Industries B.V.

package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/config"
	"go.packetbroker.org/pb/internal/client"
)

type inputData struct {
	help, debug          bool
	client               *client.Config
	forwarderNetIDHex    string
	forwarderNetID       *packetbroker.NetID
	forwarderClusterID   string
	forwarderTenantID    string
	homeNetworkNetIDHex  string
	homeNetworkNetID     *packetbroker.NetID
	homeNetworkClusterID string
	homeNetworkTenantID  string
}

var input = new(inputData)

func parseInput() bool {
	config.CommonFlags(&input.help, &input.debug)
	config.ClientFlags()

	flag.StringVar(&input.forwarderNetIDHex, "forwarder-net-id", "", "NetID of the Forwarder (hex)")
	flag.StringVar(&input.forwarderClusterID, "forwarder-cluster-id", "", "Cluster ID of the Forwarder")
	flag.StringVar(&input.forwarderTenantID, "forwarder-tenant-id", "", "Tenant ID of the Forwarder")
	flag.StringVar(&input.homeNetworkNetIDHex, "home-network-net-id", "", "NetID of the Home Network (hex)")
	flag.StringVar(&input.homeNetworkClusterID, "home-network-cluster-id", "", "Cluster ID of the Home Network")
	flag.StringVar(&input.homeNetworkTenantID, "home-network-tenant-id", "", "Tenant ID of the Home Network")

	flag.Parse()

	var err error
	input.client, err = config.Client()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid client settings:", err)
		return false
	}

	if !input.help {
		if input.forwarderNetIDHex == "" {
			fmt.Fprintln(os.Stderr, "Must set forwarder-net-id")
			return false
		}
		input.forwarderNetID = new(packetbroker.NetID)
		if err := input.forwarderNetID.UnmarshalText([]byte(input.forwarderNetIDHex)); err != nil {
			fmt.Fprintln(os.Stderr, "Invalid forwarder-net-id:", err)
			return false
		}

		if input.homeNetworkNetIDHex != "" {
			input.homeNetworkNetID = new(packetbroker.NetID)
			if err := input.homeNetworkNetID.UnmarshalText([]byte(input.homeNetworkNetIDHex)); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid home-network-net-id:", err)
				return false
			}
		}
	}

	return true
}
