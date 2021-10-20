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
	flags.String(prefix+"fns-path", "", "path for Forwarding Network Server (fNS)")
	flags.String(prefix+"sns-path", "", "path for Serving Network Server (sNS)")
	flags.String(prefix+"hns-path", "", "path for Home Network Server (hNS)")
	flags.Bool(prefix+"pb-token", false, "use Packet Broker token")
	flags.String(prefix+"authorization", "", "custom authorization value (e.g. HTTP Authorization header value)")
	flags.String(prefix+"root-cas-file", "", "path to PEM encoded root CAs")
	flags.String(prefix+"tls-cert-file", "", "path to PEM encoded client certificate")
	flags.String(prefix+"tls-key-file", "", "path to PEM encoded private key")
	flags.AddFlagSet(pbflag.NetID(prefix + "origin"))
	return flags
}

func mergeTarget(flags *flag.FlagSet, prefix string, target **packetbroker.Target) error {
	protocol := pbflag.GetTargetProtocol(flags, prefix)
	if *target == nil {
		if protocol == nil {
			return nil
		}
		*target = new(packetbroker.Target)
	}
	if protocol != nil {
		(*target).Protocol = *protocol
	}

	if prefix != "" {
		prefix = prefix + "-"
	}
	switch (*target).Protocol {
	case packetbroker.TargetProtocol_TS002_V1_0, packetbroker.TargetProtocol_TS002_V1_1:
		var url *url.URL
		if address, err := flags.GetString(prefix + "address"); err == nil && address != "" {
			url, err = url.Parse(address)
			if err != nil {
				return err
			}
		}

		pbToken, _ := flags.GetBool(prefix + "pb-token")
		authorization, _ := flags.GetString(prefix + "authorization")
		tlsCertFile, _ := flags.GetString(prefix + "tls-cert-file")
		tlsKeyFile, _ := flags.GetString(prefix + "tls-key-file")

		var authentication *packetbroker.Target_Authentication
		switch {
		// Packet Broker token authentication.
		case pbToken:
			authentication = &packetbroker.Target_Authentication{
				Value: &packetbroker.Target_Authentication_PbTokenAuth{
					PbTokenAuth: &packetbroker.Target_PacketBrokerTokenAuth{},
				},
			}
		// HTTP basic authentication.
		case url != nil && url.User != nil:
			password, _ := url.User.Password()
			authentication = &packetbroker.Target_Authentication{
				Value: &packetbroker.Target_Authentication_BasicAuth{
					BasicAuth: &packetbroker.Target_BasicAuth{
						Username: url.User.Username(),
						Password: password,
					},
				},
			}
			url.User = nil
		// Custom HTTP authorization value.
		case authorization != "":
			authentication = &packetbroker.Target_Authentication{
				Value: &packetbroker.Target_Authentication_CustomAuth{
					CustomAuth: &packetbroker.Target_CustomAuth{
						Value: authorization,
					},
				},
			}
		// TLS client authentication.
		case tlsCertFile != "" || tlsKeyFile != "":
			tlsCert, err := ioutil.ReadFile(tlsCertFile)
			if err != nil {
				return err
			}
			tlsKey, err := ioutil.ReadFile(tlsKeyFile)
			if err != nil {
				return err
			}
			authentication = &packetbroker.Target_Authentication{
				Value: &packetbroker.Target_Authentication_TlsClientAuth{
					TlsClientAuth: &packetbroker.Target_TLSClientAuth{
						Cert: tlsCert,
						Key:  tlsKey,
					},
				},
			}
		}

		if pbflag.NetIDChanged(flags, prefix+"origin") {
			netID := pbflag.GetNetID(flags, prefix+"origin")
			if (*target).OriginNetIdAuthentication == nil {
				(*target).OriginNetIdAuthentication = make(map[uint32]*packetbroker.Target_Authentication)
			}
			if authentication == nil {
				delete((*target).OriginNetIdAuthentication, uint32(netID))
			} else {
				(*target).OriginNetIdAuthentication[uint32(netID)] = authentication
			}
		} else if authentication != nil {
			switch auth := authentication.GetValue().(type) {
			case *packetbroker.Target_Authentication_PbTokenAuth:
				(*target).DefaultAuthentication = &packetbroker.Target_PbTokenAuth{
					PbTokenAuth: auth.PbTokenAuth,
				}
			case *packetbroker.Target_Authentication_BasicAuth:
				(*target).DefaultAuthentication = &packetbroker.Target_BasicAuth_{
					BasicAuth: auth.BasicAuth,
				}
			case *packetbroker.Target_Authentication_CustomAuth:
				(*target).DefaultAuthentication = &packetbroker.Target_CustomAuth_{
					CustomAuth: auth.CustomAuth,
				}
			case *packetbroker.Target_Authentication_TlsClientAuth:
				(*target).DefaultAuthentication = &packetbroker.Target_TlsClientAuth{
					TlsClientAuth: auth.TlsClientAuth,
				}
			}
		}

		if url != nil {
			(*target).Address = url.String()
		}
		for _, p := range []struct {
			target *string
			flag   string
		}{
			{&(*target).FNsPath, prefix + "fns-path"},
			{&(*target).SNsPath, prefix + "sns-path"},
			{&(*target).HNsPath, prefix + "hns-path"},
		} {
			if flags.Changed(p.flag) {
				*p.target, _ = flags.GetString(p.flag)
			}
		}

	default:
		return fmt.Errorf("invalid protocol: %s", protocol)
	}

	if rootCAsFile, err := flags.GetString(prefix + "root-cas-file"); err == nil && rootCAsFile != "" {
		var err error
		(*target).RootCas, err = ioutil.ReadFile(rootCAsFile)
		if err != nil {
			return fmt.Errorf("read root CAs file %q: %w", rootCAsFile, err)
		}
	}

	return nil
}
