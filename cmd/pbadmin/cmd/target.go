// Copyright © 2021 The Things Industries B.V.

package cmd

import (
	"fmt"
	"io/ioutil"
	"net/url"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/pbflag"
)

func targetFlags(prefix string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(pbflag.TargetProtocol(prefix))
	if prefix != "" {
		prefix = prefix + "-"
	}
	flags.String(prefix+"address", "", "address (e.g. URL with HTTP basic authentication)")
	flags.String(prefix+"authorization", "", "custom authorization value (e.g. HTTP Authorization header value)")
	flags.String(prefix+"root-cas-file", "", "path to PEM encoded root CAs")
	flags.String(prefix+"tls-cert-file", "", "path to PEM encoded client certificate")
	flags.String(prefix+"tls-key-file", "", "path to PEM encoded private key")
	return flags
}

func target(flags *flag.FlagSet, prefix string) (*packetbroker.Target, error) {
	protocol := pbflag.GetTargetProtocol(flags, prefix)
	if protocol == nil {
		return nil, nil
	}
	res := &packetbroker.Target{
		Protocol: *protocol,
	}

	if prefix != "" {
		prefix = prefix + "-"
	}

	switch res.Protocol {
	case packetbroker.TargetProtocol_TS002_V1_0, packetbroker.TargetProtocol_TS002_V1_1_0:
		address, _ := flags.GetString(prefix + "address")
		url, err := url.Parse(address)
		if err != nil {
			return nil, err
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
		} else if authorization, _ := flags.GetString(prefix + "authorization"); authorization != "" {
			res.Authorization = &packetbroker.Target_CustomAuth_{
				CustomAuth: &packetbroker.Target_CustomAuth{
					Value: authorization,
				},
			}
		} else {
			tlsCertFile, _ := flags.GetString(prefix + "tls-cert-file")
			tlsKeyFile, _ := flags.GetString(prefix + "tls-key-file")
			if tlsCertFile != "" || tlsKeyFile != "" {
				tlsCert, err := ioutil.ReadFile(tlsCertFile)
				if err != nil {
					return nil, err
				}
				tlsKey, err := ioutil.ReadFile(tlsKeyFile)
				if err != nil {
					return nil, err
				}
				res.Authorization = &packetbroker.Target_TlsClientAuth{
					TlsClientAuth: &packetbroker.Target_TLSClientAuth{
						Cert: tlsCert,
						Key:  tlsKey,
					},
				}
			}
		}
		res.Address = url.String()

	default:
		return nil, fmt.Errorf("invalid protocol: %s", protocol)
	}

	if rootCAsFile, _ := flags.GetString(prefix + "root-cas-file"); rootCAsFile != "" {
		var err error
		res.RootCas, err = ioutil.ReadFile(rootCAsFile)
		if err != nil {
			return nil, fmt.Errorf("read root CAs file %q: %w", rootCAsFile, err)
		}
	}

	return res, nil
}
