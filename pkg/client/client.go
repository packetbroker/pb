// Copyright © 2020 The Things Industries B.V.

package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Config configures Client.
type Config struct {
	Address     string
	DialTimeout time.Duration
	Insecure    bool
	Credentials credentials.PerRPCCredentials
}

func appendDefaultPort(target string, port int) (string, error) {
	i := strings.LastIndexByte(target, ':')
	if i < 0 {
		return fmt.Sprintf("%s:%d", target, port), nil
	}
	// Check if target is an IPv6 host, i.e. [::1]:1912.
	if target[0] == '[' {
		end := strings.IndexByte(target, ']')
		if end < 0 || end+1 != i {
			return "", fmt.Errorf("client: invalid address %q", target)
		}
		return target, nil
	}
	// No IPv6 hostport, so target with colon must be a hostport or IPv6.
	ip := net.ParseIP(target)
	if len(ip) == net.IPv6len {
		return fmt.Sprintf("[%s]:%d", ip.String(), port), nil
	}
	return target, nil
}

// DialContext dials a Packet Broker service using the given configuration.
func DialContext(ctx context.Context, logger *zap.Logger, config *Config, defaultPort int) (*grpc.ClientConn, error) {
	timeout := config.DialTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	address, err := appendDefaultPort(config.Address, defaultPort)
	if err != nil {
		return nil, err
	}

	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                5 * time.Minute,
			Timeout:             20 * time.Second,
			PermitWithoutStream: false,
		}),
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
	}

	if config.Insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	if config.Credentials != nil {
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(config.Credentials))
	}

	return grpc.DialContext(ctx, address, dialOpts...)
}
