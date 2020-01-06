// Copyright © 2020 The Things Industries B.V.

package flags

import (
	"flag"

	"go.packetbroker.org/pb/internal/client"
)

// Common defines common flags.
func Common(help, debug *bool) {
	flag.BoolVar(help, "help", false, "show usage")
	flag.BoolVar(debug, "debug", false, "debug mode")
}

// Client defines flags used for Client configuration.
func Client(config *client.Config) {
	flag.StringVar(&config.Address, "address", "localhost:1900", "address of the server")
	flag.StringVar(&config.CertFile, "cert-file", "cert.pem", "path to a PEM encoded TLS server certificate file")
	flag.StringVar(&config.KeyFile, "key-file", "key.pem", "path to a PEM encoded TLS server key file")
	flag.StringVar(&config.CAFile, "ca-file", "ca.pem", "path to a file containing PEM encoded root certificate authorities")
}
