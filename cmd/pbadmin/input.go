// Copyright © 2020 The Things Industries B.V.

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v2"
	"go.packetbroker.org/pb/cmd/internal/config"
	"go.packetbroker.org/pb/internal/client"
)

type inputData struct {
	help, debug         bool
	client              *client.Config
	netIDHex            string
	netID               *packetbroker.NetID
	tenantID            string
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
	config.ClientFlags()

	flag.StringVar(&input.netIDHex, "net-id", "", "NetID (hex)")
	flag.StringVar(&input.tenantID, "tenant-id", "", "Tenant ID")
	flag.StringVar(&input.forwarderNetIDHex, "forwarder-net-id", "", "NetID of the Forwarder (hex)")
	flag.StringVar(&input.forwarderTenantID, "forwarder-tenant-id", "", "Tenant ID of the Forwarder")
	flag.StringVar(&input.homeNetworkNetIDHex, "home-network-net-id", "", "NetID of the Home Network (hex)")
	flag.StringVar(&input.homeNetworkTenantID, "home-network-tenant-id", "", "Tenant ID of the Home Network")

	if len(os.Args) < 2 {
		flag.Parse()
		return false
	}

	switch input.mode = strings.ToLower(os.Args[1]); input.mode {
	case "tenant":
		if !parseTenantFlags() {
			return false
		}
	case "policy":
		if !parsePolicyFlags() {
			return false
		}
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	var err error
	input.client, err = config.Client()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid client settings:", err)
		return false
	}

	if !input.help {
		if input.netIDHex != "" {
			input.netID = new(packetbroker.NetID)
			if err := input.netID.UnmarshalText([]byte(input.netIDHex)); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid net-id:", err)
				return false
			}
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
		}
	}

	return true
}

func parseDevAddrPrefix(s string) (*packetbroker.DevAddrPrefix, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid format")
	}
	res := &packetbroker.DevAddrPrefix{}
	if v, err := strconv.ParseUint(parts[0], 16, 32); err == nil {
		res.Value = uint32(v)
	} else {
		return nil, fmt.Errorf("invalid value: %w", err)
	}
	if v, err := strconv.ParseUint(parts[1], 10, 8); err == nil && v <= 32 {
		res.Length = uint32(v)
	} else if err != nil {
		return nil, fmt.Errorf("invalid length: %w", err)
	} else {
		return nil, errors.New("invalid length: must be at most 32 bits")
	}
	return res, nil
}
