// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/config"
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

	ctx     = context.Background()
	logger  *zap.Logger
	conn    *grpc.ClientConn
	decoder *json.Decoder
)

var rootCmd = &cobra.Command{
	Use:   "pbsub",
	Short: "pbsub can be used to subscribe to uplink and downlink messages.",
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger = logging.GetLogger(debug)
		clientConf, err := config.OAuth2Client(ctx, "router", "networks")
		if err != nil {
			return err
		}
		conn, err = client.DialContext(ctx, logger, clientConf, 443)
		if err != nil {
			return err
		}
		decoder = json.NewDecoder(os.Stdin)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			forwarder   = pbflag.GetEndpoint(cmd.Flags(), "forwarder")
			homeNetwork = pbflag.GetEndpoint(cmd.Flags(), "home-network")
		)
		group, _ := cmd.Flags().GetString("group")
		if homeNetwork.IsEmpty() {
			return asForwarder(forwarder, group)
		}
		return asHomeNetwork(homeNetwork, group)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		logger.Sync()
		conn.Close()
	},
}

func asForwarder(forwarder packetbroker.Endpoint, group string) error {
	client := routingpb.NewForwarderDataClient(conn)
	stream, err := client.Subscribe(ctx, &routingpb.SubscribeForwarderRequest{
		ForwarderNetId:     uint32(forwarder.NetID),
		ForwarderClusterId: forwarder.ClusterID,
		ForwarderTenantId:  forwarder.TenantID.ID,
		Group:              group,
	})
	if err != nil {
		return err
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				return err
			}
			return nil
		}
		if err := protojson.Write(os.Stdout, msg); err != nil {
			return err
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
		HomeNetworkTenantId:  homeNetwork.TenantID.ID,
		Group:                group,
		Filters:              filters,
	})
	if err != nil {
		return err
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				return err
			}
			return nil
		}
		if err := protojson.Write(os.Stdout, msg); err != nil {
			return err
		}
	}
}

// Execute runs pbctl.
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
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".pb")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Read config file:", err)
	}
}
