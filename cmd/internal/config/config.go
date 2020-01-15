// Copyright © 2020 The Things Industries B.V.

package config

import (
	flag "github.com/spf13/pflag"
	"go.packetbroker.org/pb/internal/client"
)

// CommonFlags defines common flags.
func CommonFlags(help, debug *bool) {
	flag.BoolVar(help, "help", false, "show usage")
	flag.BoolVar(debug, "debug", false, "debug mode")
}

// ClientFlags defines flags used for Client configuration.
func ClientFlags(config *client.Config) {
	flag.StringVar(&config.Address, "address", "localhost:1900", "address of the server")
	flag.StringVar(&config.CertFile, "cert-file", "cert.pem", "path to a PEM encoded TLS server certificate file")
	flag.StringVar(&config.KeyFile, "key-file", "key.pem", "path to a PEM encoded TLS server key file")
	flag.StringVar(&config.CAFile, "ca-file", "ca.pem", "path to a file containing PEM encoded root certificate authorities")
}
