// Copyright © 2020 The Things Industries B.V.

package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config configures Client.
type Config struct {
	Address     string
	DialTimeout time.Duration

	CertFile,
	KeyFile,
	CAFile string
}

// DialContext dials a Packet Broker service using the given configuration.
func DialContext(ctx context.Context, logger *zap.Logger, config Config) (*grpc.ClientConn, error) {
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("client: failed to load X.509 certificate file %q and key file %q: %w",
			config.CertFile, config.KeyFile, err)
	}
	buf, err := ioutil.ReadFile(config.CAFile)
	if err != nil {
		return nil, fmt.Errorf("client: failed to read CA file %q: %w", config.CAFile, err)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(buf) {
		return nil, fmt.Errorf("client: failed to append CAs from %q", config.CAFile)
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootCAs,
	})

	timeout := config.DialTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return grpc.DialContext(ctx, config.Address,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithUserAgent(fmt.Sprintf("%s go/%s %s/%s",
			filepath.Base(os.Args[0]),
			strings.TrimPrefix(runtime.Version(), "go"),
			runtime.GOOS, runtime.GOARCH,
		)),
		grpc.WithChainStreamInterceptor(
			grpc_zap.StreamClientInterceptor(logger),
		),
		grpc.WithChainUnaryInterceptor(
			grpc_zap.UnaryClientInterceptor(logger),
		),
	)
}
