// Copyright © 2020 The Things Industries B.V.

package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/config"
	"go.packetbroker.org/pb/pkg/client"
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
	group                string
	homeNetworkFilters   struct {
		devAddrPrefixesText []string
		devAddrPrefixes     []*packetbroker.DevAddrPrefix
	}
}

var input = new(inputData)

func parseInput() bool {
	config.CommonFlags(&input.help, &input.debug)
	config.ClientFlags("")
	config.OAuth2ClientFlags()

	flag.StringVar(&input.forwarderNetIDHex, "forwarder-net-id", "", "NetID of the Forwarder (hex)")
	flag.StringVar(&input.forwarderClusterID, "forwarder-cluster-id", "", "Cluster ID of the Forwarder")
	flag.StringVar(&input.forwarderTenantID, "forwarder-tenant-id", "", "Tenant ID of the Forwarder")
	flag.StringVar(&input.homeNetworkNetIDHex, "home-network-net-id", "", "NetID of the Home Network (hex)")
	flag.StringVar(&input.homeNetworkClusterID, "home-network-cluster-id", "", "Cluster ID of the Home Network")
	flag.StringVar(&input.homeNetworkTenantID, "home-network-tenant-id", "", "Tenant ID of the Home Network")
	flag.StringVar(&input.group, "group", "", "group name of shared subscription")
	flag.StringSliceVar(&input.homeNetworkFilters.devAddrPrefixesText, "filter-dev-addr-prefixes", nil, "filter DevAddr prefixes (i.e. 00000000/3)")

	flag.Parse()

	if !input.help {
		var err error
		input.client, err = config.OAuth2Client(ctx, "networks")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid client settings:", err)
			return false
		}

		if (input.forwarderNetIDHex != "") == (input.homeNetworkNetIDHex != "") {
			fmt.Fprintln(os.Stderr, "Must set either forwarder-net-id or home-network-net-id")
			return false
		}

		if input.forwarderNetIDHex != "" {
			input.forwarderNetID = new(packetbroker.NetID)
			if err := input.forwarderNetID.UnmarshalText([]byte(input.forwarderNetIDHex)); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid forwarder-net-id:", err)
				return false
			}
		}
		if input.homeNetworkNetIDHex != "" {
			input.homeNetworkNetID = new(packetbroker.NetID)
			if err := input.homeNetworkNetID.UnmarshalText([]byte(input.homeNetworkNetIDHex)); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid home-network-net-id:", err)
				return false
			}
			if len(input.homeNetworkFilters.devAddrPrefixesText) > 0 {
				input.homeNetworkFilters.devAddrPrefixes = make([]*packetbroker.DevAddrPrefix, len(input.homeNetworkFilters.devAddrPrefixesText))
				for i, text := range input.homeNetworkFilters.devAddrPrefixesText {
					var prefix packetbroker.DevAddrPrefix
					if err := prefix.UnmarshalText([]byte(text)); err != nil {
						fmt.Fprintln(os.Stderr, "Invalid filter-dev-addr-prefixes:", err)
						return false
					}
					input.homeNetworkFilters.devAddrPrefixes[i] = &prefix
				}
			}
		}
	}

	return true
}
