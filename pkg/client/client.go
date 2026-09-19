// SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package client provides gRPC clients for Packet Broker services.
package client

import (
	"context"
	"errors"
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
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Config configures Client.
type Config struct {
	Address     string
	DialTimeout time.Duration
	Insecure    bool
	Credentials credentials.PerRPCCredentials
}

var (
	errConnectionFailed   = errors.New("connection failed")
	errConnectionShutdown = errors.New("connection shut down")
)

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
//
// The connection is established before returning, so that connection errors are reported before the first RPC.
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
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                5 * time.Minute,
			Timeout:             20 * time.Second,
			PermitWithoutStream: false,
		}),
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
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	if config.Credentials != nil {
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(config.Credentials))
	}

	conn, err := grpc.NewClient(address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("client: dial %s: %w", address, err)
	}
	// grpc.NewClient connects lazily on the first RPC. Connect explicitly so that a CLI command fails before doing
	// any work if the service is unreachable.
	if err := waitForReady(ctx, conn); err != nil {
		// The connection is not usable; a close error is not actionable.
		_ = conn.Close()
		return nil, fmt.Errorf("client: dial %s: %w", address, err)
	}
	return conn, nil
}

// waitForReady connects conn and waits until the connection is ready or ctx is done.
//
// It fails fast when the connection enters transient failure, i.e. when all resolved addresses failed to connect,
// for example because the connection is refused or the TLS handshake fails. Otherwise, gRPC would keep retrying
// with backoff until ctx is done. The cause of the failure is not exposed by *grpc.ClientConn.
func waitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	for {
		state := conn.GetState()
		switch state {
		case connectivity.Idle:
			conn.Connect()
		case connectivity.Connecting:
			// Wait for the connection attempt to complete.
		case connectivity.Ready:
			return nil
		case connectivity.TransientFailure:
			return errConnectionFailed
		case connectivity.Shutdown:
			return errConnectionShutdown
		}
		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}
	}
}
