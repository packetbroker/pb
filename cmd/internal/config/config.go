// Copyright © 2020 The Things Industries B.V.

package config

import (
	"errors"
	"os"

	flag "github.com/spf13/pflag"
	"go.packetbroker.org/pb/internal/client"
)

// CommonFlags defines common flags.
func CommonFlags(help, debug *bool) {
	flag.BoolVar(help, "help", false, "show usage")
	flag.BoolVar(debug, "debug", false, "debug mode")
}

// ClientFlags defines flags used for Client configuration.
func ClientFlags() {
	flag.String("address", "", `address of the server "host[:port]" (default $PB_ADDRESS)`)
	flag.String("cert-file", "cert.pem", "path to a PEM encoded TLS server certificate file")
	flag.String("key-file", "key.pem", "path to a PEM encoded TLS server key file")
	flag.String("ca-file", "", "path to a file containing PEM encoded root certificate authorities")
}

// Client returns client configuration.
func Client() (*client.Config, error) {
	var (
		res client.Config
		err error
	)
	if res.Address, err = flag.CommandLine.GetString("address"); err != nil {
		return nil, err
	}
	if res.CertFile, err = flag.CommandLine.GetString("cert-file"); err != nil {
		return nil, err
	}
	if res.KeyFile, err = flag.CommandLine.GetString("key-file"); err != nil {
		return nil, err
	}
	if res.CAFile, err = flag.CommandLine.GetString("ca-file"); err != nil {
		return nil, err
	}
	if res.Address == "" {
		res.Address = os.Getenv("PB_ADDRESS")
	}
	if res.Address == "" {
		return nil, errors.New("missing server address")
	}
	if res.CertFile == "" || res.KeyFile == "" {
		return nil, errors.New("missing TLS client settings")
	}
	return &res, nil
}
