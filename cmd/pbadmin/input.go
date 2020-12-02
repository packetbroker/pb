// Copyright © 2020 The Things Industries B.V.

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/config"
	"go.packetbroker.org/pb/pkg/client"
)

type inputData struct {
	help, debug bool
	client      *client.Config
	netIDHex    string
	netID       *packetbroker.NetID
	tenantID    string
	mode        string
	network     struct {
		hasName bool
		name    string
	}
	tenant struct {
		hasName          bool
		name             string
		hasDevAddrBlocks bool
		devAddrBlocksHex []string
		devAddrBlocks    []*packetbroker.DevAddrBlock
	}
	apiKey struct {
		clusterID string
		validFor  time.Duration
		keyID     string
	}
}

var input = new(inputData)

func parseInput() bool {
	config.CommonFlags(&input.help, &input.debug)
	config.BasicAuthClientFlags()

	flag.StringVar(&input.netIDHex, "net-id", "", "NetID (hex)")
	flag.StringVar(&input.tenantID, "tenant-id", "", "Tenant ID")

	if len(os.Args) < 2 {
		flag.Parse()
		return false
	}

	switch input.mode = strings.ToLower(os.Args[1]); input.mode {
	case "network":
		if !parseNetworkFlags() {
			return false
		}
	case "tenant":
		if !parseTenantFlags() {
			return false
		}
	case "apikey":
		if !parseAPIKeyFlags() {
			return false
		}
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	var err error
	input.client, err = config.BasicAuthClient()
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

func parseDevAddrBlock(s string) (*packetbroker.DevAddrBlock, error) {
	parts := strings.SplitN(s, "=", 2)
	prefix, err := parseDevAddrPrefix(parts[0])
	if err != nil {
		return nil, err
	}
	block := &packetbroker.DevAddrBlock{
		Prefix: prefix,
	}
	if len(parts) > 1 {
		block.HomeNetworkClusterId = parts[1]
	}
	return block, nil
}
