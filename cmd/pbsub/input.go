// Copyright © 2020 The Things Industries B.V.

package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v1beta3"
	"go.packetbroker.org/pb/cmd/internal/config"
	"go.packetbroker.org/pb/internal/client"
)

type inputData struct {
	help, debug         bool
	client              *client.Config
	forwarderNetIDHex   string
	forwarderNetID      *packetbroker.NetID
	forwarderID         string
	homeNetworkNetIDHex string
	homeNetworkNetID    *packetbroker.NetID
	group               string
	homeNetworkFilters  struct {
		forwarderNetIDHex string
		forwarderNetID    *packetbroker.NetID
		forwarderID       string
	}
}

var input = new(inputData)

func parseInput() bool {
	config.CommonFlags(&input.help, &input.debug)
	config.ClientFlags()

	flag.StringVar(&input.forwarderNetIDHex, "forwarder-net-id", "", "NetID of the Forwarder (hex)")
	flag.StringVar(&input.forwarderID, "forwarder-id", "", "ID of the Forwarder")
	flag.StringVar(&input.homeNetworkNetIDHex, "home-network-net-id", "", "NetID of the Home Network (hex)")
	flag.StringVar(&input.group, "group", "", "group name of shared subscription")
	flag.StringVar(&input.homeNetworkFilters.forwarderNetIDHex, "filter-forwarder-net-id", "", "filter Forwarder by NetID")
	flag.StringVar(&input.homeNetworkFilters.forwarderID, "filter-forwarder-id", "", "filter Forwarder by ID")

	flag.Parse()

	var err error
	input.client, err = config.Client()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid client settings:", err)
		return false
	}

	if !input.help {
		if (input.forwarderNetIDHex != "") == (input.homeNetworkNetIDHex != "") {
			fmt.Fprintln(os.Stderr, "Must set either forwarder-net-id or home-network-net-id")
			return false
		}

		if input.forwarderNetIDHex != "" {
			input.forwarderNetID = new(packetbroker.NetID)
			if err := input.forwarderNetID.UnmarshalText([]byte(input.forwarderNetIDHex)); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid forwarder-net-id")
				return false
			}
		}
		if input.homeNetworkNetIDHex != "" {
			input.homeNetworkNetID = new(packetbroker.NetID)
			if err := input.homeNetworkNetID.UnmarshalText([]byte(input.homeNetworkNetIDHex)); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid home-network-net-id")
				return false
			}
			if input.homeNetworkFilters.forwarderNetIDHex != "" {
				input.homeNetworkFilters.forwarderNetID = new(packetbroker.NetID)
				if err := input.homeNetworkFilters.forwarderNetID.UnmarshalText([]byte(input.homeNetworkFilters.forwarderNetIDHex)); err != nil {
					fmt.Fprintln(os.Stderr, "Invalid filter-forwarder-net-id")
					return false
				}
			}
		}
	}

	return true
}
