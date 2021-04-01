// Copyright © 2021 The Things Industries B.V.

package cmd

import (
	"fmt"
	"net/url"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/pbflag"
)

func targetFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(pbflag.TargetProtocol())
	flags.String("target-address", "", "address (e.g. URL with HTTP basic authentication)")
	flags.String("target-authorization", "", "custom authorization value (e.g. HTTP Authorization header value)")
	return flags
}

func target(flags *flag.FlagSet) (*packetbroker.Target, error) {
	protocol := pbflag.GetTargetProtocol(flags)
	if protocol == nil {
		return nil, nil
	}
	switch *protocol {
	case packetbroker.TargetProtocol_TS002_V1_0, packetbroker.TargetProtocol_TS002_V1_1_0:
		address, _ := flags.GetString("target-address")
		url, err := url.Parse(address)
		if err != nil {
			return nil, err
		}
		res := &packetbroker.Target{
			Protocol: *protocol,
		}
		if url.User != nil {
			password, _ := url.User.Password()
			res.Authorization = &packetbroker.Target_BasicAuth_{
				BasicAuth: &packetbroker.Target_BasicAuth{
					Username: url.User.Username(),
					Password: password,
				},
			}
			url.User = nil
		} else if authorization, _ := flags.GetString("target-authorization"); authorization != "" {
			res.Authorization = &packetbroker.Target_CustomAuth_{
				CustomAuth: &packetbroker.Target_CustomAuth{
					Value: authorization,
				},
			}
		}
		res.Address = url.String()
		return res, nil

	default:
		return nil, fmt.Errorf("invalid protocol: %s", protocol)
	}
}
