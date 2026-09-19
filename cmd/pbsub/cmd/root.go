// SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package cmd implements the pbsub command.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/config"
	"go.packetbroker.org/pb/cmd/internal/gen"
	"go.packetbroker.org/pb/cmd/internal/logging"
	"go.packetbroker.org/pb/cmd/internal/pbflag"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"go.packetbroker.org/pb/pkg/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	cfgFile string
	debug   bool

	ctx    = context.Background()
	logger *zap.Logger
	conn   *grpc.ClientConn
)

var rootCmd = &cobra.Command{
	Use:          "pbsub",
	Short:        "pbsub can be used to subscribe to uplink and downlink messages.",
	SilenceUsage: true,
	Example: `
  Subscribe as Forwarder:

    Subscribe as network:
      $ pbsub --forwarder-net-id 000013

    Subscribe as tenant:
      $ pbsub --forwarder-net-id 000013 --forwarder-tenant-id community

    Subscribe as named cluster in network:
      $ pbsub --forwarder-net-id 000013 --forwarder-cluster-id eu1

    Subscribe as named cluster in tenant:
      $ pbsub --forwarder-net-id 000013 --forwarder-tenant-id community \
        --forwarder-cluster-id eu1

  Subscribe as Home Network:

    Subscribe as network:
      $ pbsub --home-network-net-id 000013

    Subscribe as tenant:
      $ pbsub --home-network-net-id 000013 --home-network-tenant-id community

    Subscribe as named cluster in network:
      $ pbsub --home-network-net-id 000013 --home-network-cluster-id eu1

    Subscribe as named cluster in tenant:
      $ pbsub --home-network-net-id 000013 --home-network-tenant-id community \
        --home-network-cluster-id eu1`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		logger = logging.GetLogger(debug)
		clientConf, err := config.OAuth2Client(ctx, "router", "networks")
		if err != nil {
			return fmt.Errorf("configure Router client: %w", err)
		}
		conn, err = client.DialContext(ctx, logger, clientConf, 443)
		if err != nil {
			return fmt.Errorf("connect to Router: %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		var (
			forwarder, forwarderOK     = pbflag.GetEndpoint(cmd.Flags(), "forwarder")
			homeNetwork, homeNetworkOK = pbflag.GetEndpoint(cmd.Flags(), "home-network")
		)
		group, _ := cmd.Flags().GetString("group")
		switch {
		case forwarderOK:
			return asForwarder(forwarder, group)
		case homeNetworkOK:
			return asHomeNetwork(homeNetwork, group)
		}
		return errors.New("no role specified")
	},
	PostRun: func(_ *cobra.Command, _ []string) {
		// Syncing stderr is not supported on all platforms, and the connection is no longer used.
		_ = logger.Sync()
		_ = conn.Close()
	},
}

func asForwarder(forwarder packetbroker.Endpoint, group string) error {
	client := routingpb.NewForwarderDataClient(conn)
	stream, err := client.Subscribe(ctx, &routingpb.SubscribeForwarderRequest{
		ForwarderNetId:     uint32(forwarder.NetID),
		ForwarderClusterId: forwarder.ClusterID,
		ForwarderTenantId:  forwarder.ID,
		Group:              group,
	})
	if err != nil {
		return fmt.Errorf("subscribe as Forwarder: %w", err)
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				return fmt.Errorf("receive message: %w", err)
			}
			return nil
		}
		if err := protojson.Write(os.Stdout, msg); err != nil {
			return fmt.Errorf("write message: %w", err)
		}
	}
}

func asHomeNetwork(homeNetwork packetbroker.Endpoint, group string) error {
	// Subscribe to all MAC payload and join-requests.
	filters := []*packetbroker.RoutingFilter{
		{
			Message: &packetbroker.RoutingFilter_Mac{
				Mac: &packetbroker.RoutingFilter_MACPayload{},
			},
		},
		{
			Message: &packetbroker.RoutingFilter_JoinRequest_{
				JoinRequest: &packetbroker.RoutingFilter_JoinRequest{
					EuiPrefixes: []*packetbroker.RoutingFilter_JoinRequest_EUIPrefixes{{}},
				},
			},
		},
	}

	client := routingpb.NewHomeNetworkDataClient(conn)
	stream, err := client.Subscribe(ctx, &routingpb.SubscribeHomeNetworkRequest{
		HomeNetworkNetId:     uint32(homeNetwork.NetID),
		HomeNetworkClusterId: homeNetwork.ClusterID,
		HomeNetworkTenantId:  homeNetwork.ID,
		Group:                group,
		Filters:              filters,
	})
	if err != nil {
		return fmt.Errorf("subscribe as Home Network: %w", err)
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				return fmt.Errorf("receive message: %w", err)
			}
			return nil
		}
		if err := protojson.Write(os.Stdout, msg); err != nil {
			return fmt.Errorf("write message: %w", err)
		}
	}
}

// Execute runs pbsub.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().AddFlagSet(config.ClientFlags("router", ""))
	rootCmd.PersistentFlags().AddFlagSet(config.OAuth2ClientFlags())

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pb.yaml, .pb.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug mode")

	rootCmd.Flags().AddFlagSet(pbflag.Endpoint("forwarder"))
	rootCmd.Flags().AddFlagSet(pbflag.Endpoint("home-network"))
	rootCmd.Flags().String("group", "", "subscription group")

	rootCmd.AddCommand(gen.Cmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName(".pb")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("pb")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	if err := config.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Warning:", err)
	}
}
