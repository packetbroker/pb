// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	flag "github.com/spf13/pflag"
	iampb "go.packetbroker.org/api/iam"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"htdvisser.dev/exp/clicontext"
)

func parseAPIKeyFlags() bool {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Missing command")
		return false
	}
	switch os.Args[2] {
	case "list":
	case "create":
		flag.CommandLine.DurationVar(&input.apiKey.validFor, "valid-for", 365*24*time.Hour, "valid for")
	case "delete":
		flag.CommandLine.StringVar(&input.apiKey.keyID, "key-id", "", "API Key ID")
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	flag.CommandLine.StringVar(&input.apiKey.clusterID, "cluster-id", "", "cluster ID")

	flag.CommandLine.Parse(os.Args[3:])

	if !input.help {
		switch os.Args[2] {
		case "list", "create":
			if input.netIDHex == "" {
				fmt.Fprintln(os.Stderr, "Must set net-id")
				return false
			}
		case "delete":
			if input.apiKey.keyID == "" {
				fmt.Fprintln(os.Stderr, "Must set key-id")
				return false
			}
		}
	}

	return true
}

func runAPIKey(ctx context.Context) {
	client := iampb.NewAPIKeyVaultClient(conn)
	switch os.Args[2] {
	case "list":
		res, err := client.ListAPIKeys(ctx, &iampb.ListAPIKeysRequest{
			NetId:     uint32(*input.netID),
			TenantId:  input.tenantID,
			ClusterId: input.apiKey.clusterID,
		})
		if err != nil {
			logger.Error("Failed to list API keys", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
		for _, k := range res.Keys {
			if err := protojson.Write(os.Stdout, k); err != nil {
				logger.Error("Failed to convert API key to JSON", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
				return
			}
		}

	case "create":
		notBefore := time.Now()
		notAfter := notBefore.Add(input.apiKey.validFor)
		res, err := client.CreateAPIKey(ctx, &iampb.CreateAPIKeyRequest{
			NetId:     uint32(*input.netID),
			TenantId:  input.tenantID,
			ClusterId: input.apiKey.clusterID,
			NotBefore: timestamppb.New(notBefore),
			NotAfter:  timestamppb.New(notAfter),
		})
		if err != nil {
			logger.Error("Failed to create API key", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
		fmt.Fprintln(os.Stderr, "Copy the secret value as secrets cannot be retrieved later")
		if err := protojson.Write(os.Stdout, res); err != nil {
			logger.Error("Failed to convert API key to JSON", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}

	case "delete":
		_, err := client.DeleteAPIKey(ctx, &iampb.APIKeyRequest{
			KeyId: input.apiKey.keyID,
		})
		if err != nil {
			logger.Error("Failed to delete API key", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
	}
}
