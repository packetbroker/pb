// Copyright © 2020 The Things Industries B.V.

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	packetbroker "go.packetbroker.org/api/v1alpha1"
	"go.packetbroker.org/pb/cmd/internal/flags"
	"go.packetbroker.org/pb/internal/client"
)

type inputData struct {
	help, debug         bool
	client              client.Config
	forwarderNetIDHex   string
	forwarderNetID      *packetbroker.NetID
	forwarderID         string
	homeNetworkNetIDHex string
	homeNetworkNetID    *packetbroker.NetID
	mode                string
	policy              struct {
		defaults      bool
		setUplink     string
		unsetUplink   bool
		setDownlink   string
		unsetDownlink bool
	}
}

var input = new(inputData)

func parseInput() {
	flags.Common(&input.help, &input.debug)
	flags.Client(&input.client)

	flag.StringVar(&input.forwarderNetIDHex, "forwarder-net-id", "", "NetID of the Forwarder (hex)")
	flag.StringVar(&input.forwarderID, "forwarder-id", "", "ID of the Forwarder")
	flag.StringVar(&input.homeNetworkNetIDHex, "home-network-net-id", "", "NetID of the Home Network (hex)")

	if len(os.Args) < 2 {
		input.help = true
		flag.Parse()
	} else {
		switch input.mode = strings.ToLower(os.Args[1]); input.mode {
		case "policy":
			parsePolicyFlags()
		default:
			fmt.Fprintln(os.Stderr, "Invalid command")
			input.help = true
		}
	}

	if input.forwarderNetIDHex != "" {
		input.forwarderNetID = new(packetbroker.NetID)
		if err := input.forwarderNetID.UnmarshalText([]byte(input.forwarderNetIDHex)); err != nil {
			fmt.Fprintln(os.Stderr, "Invalid forwarder-net-id")
			input.help = true
		}
	}
	if input.homeNetworkNetIDHex != "" {
		input.homeNetworkNetID = new(packetbroker.NetID)
		if err := input.homeNetworkNetID.UnmarshalText([]byte(input.homeNetworkNetIDHex)); err != nil {
			fmt.Fprintln(os.Stderr, "Invalid home-network-net-id")
			input.help = true
		}
	}
}
