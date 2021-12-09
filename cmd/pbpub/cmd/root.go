// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
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

	ctx     = context.Background()
	logger  *zap.Logger
	conn    *grpc.ClientConn
	decoder *json.Decoder
)

var rootCmd = &cobra.Command{
	Use:          "pbpub",
	Short:        "pbpub can be used to publish uplink and downlink messages.",
	SilenceUsage: true,
	Example: `
  Publish uplink message as Forwarder:

    Publish as network:
      $ pbpub --forwarder-net-id 000013 < message.json

    Publish as tenant:
      $ pbpub --forwarder-net-id 000013 --forwarder-tenant-id tti < uplink.json

    Publish as named cluster:
      $ pbpub --forwarder-net-id 000013 --forwarder-cluster-id eu1 < uplink.json

    Publish as tenant in a named cluster:
      $ pbpub --forwarder-net-id 000013 --forwarder-tenant-id tti \
        --forwarder-cluster-id eu1 < uplink.json

  Publish downlink message as Home Network:

    Publish as network to a Forwarder network:
      $ pbpub --home-network-net-id 000009 \
        --forwarder-net-id 000013 < downlink.json

    Publish as tenant to a Forwarder network:
      $ pbpub --home-network-net-id 000013 \
        --home-network-tenant-id community \
        --forwarder-net-id 000009 < downlink.json

    Publish as tenant to a Forwarder tenant:
      $ pbpub --home-network-net-id 000013 \
        --home-network-tenant-id community \
        --forwarder-net-id 000013 \
        --forwarder-tenant-id tti < downlink.json

    Publish as named cluster to a named cluster in a Forwarder network:
      $ pbpub --home-network-net-id 000013 \
        --home-network-tenant-id community \
        --home-network-cluster-id eu1 \
        --forwarder-net-id 000013 \
        --forwarder-tenant-id community \
        --forwarder-cluster-id eu2 < downlink.json`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
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
			forwarder, forwarderOK     = pbflag.GetEndpoint(cmd.Flags(), "forwarder")
			homeNetwork, homeNetworkOK = pbflag.GetEndpoint(cmd.Flags(), "home-network")
		)
		switch {
		case forwarderOK:
			return asForwarder(cmd.Flags(), forwarder)
		case homeNetworkOK:
			return asHomeNetwork(cmd.Flags(), forwarder, homeNetwork)
		}
		return errors.New("no role specified")
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		logger.Sync()
		conn.Close()
	},
}

func asForwarder(flags *flag.FlagSet, forwarder packetbroker.Endpoint) error {
	client := routingpb.NewForwarderDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg := pbflag.NewForwarderMessage(flags)
		if err := protojson.Decode(decoder, msg); err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				return err
			}
			return nil
		}

		switch msg := msg.(type) {
		case *packetbroker.UplinkMessage:
			res, err := client.Publish(ctx, &routingpb.PublishUplinkMessageRequest{
				ForwarderNetId:     uint32(forwarder.NetID),
				ForwarderClusterId: forwarder.ClusterID,
				ForwarderTenantId:  forwarder.TenantID.ID,
				Message:            msg,
			})
			if err != nil {
				return err
			}
			logger.Info("Published uplink message", zap.String("id", res.Id))

		case *packetbroker.DownlinkMessageDeliveryStateChange:
			msg.ForwarderNetId = uint32(forwarder.NetID)
			msg.ForwarderClusterId = forwarder.ClusterID
			msg.ForwarderTenantId = forwarder.TenantID.ID
			_, err := client.ReportDownlinkMessageDeliveryState(ctx, &routingpb.DownlinkMessageDeliveryStateChangeRequest{
				StateChange: msg,
			})
			if err != nil {
				return err
			}
			logger.Info("Published uplink message delivery state change")
		}
	}
}

func asHomeNetwork(flags *flag.FlagSet, forwarder, homeNetwork packetbroker.Endpoint) error {
	client := routingpb.NewHomeNetworkDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg := pbflag.NewHomeNetworkMessage(flags)
		if err := protojson.Decode(decoder, msg); err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				return err
			}
			return nil
		}

		switch msg := msg.(type) {
		case *packetbroker.DownlinkMessage:
			res, err := client.Publish(ctx, &routingpb.PublishDownlinkMessageRequest{
				HomeNetworkNetId:     uint32(homeNetwork.NetID),
				HomeNetworkClusterId: homeNetwork.ClusterID,
				HomeNetworkTenantId:  homeNetwork.TenantID.ID,
				ForwarderNetId:       uint32(forwarder.NetID),
				ForwarderClusterId:   forwarder.ClusterID,
				ForwarderTenantId:    forwarder.TenantID.ID,
				Message:              msg,
			})
			if err != nil {
				return err
			}
			logger.Info("Published downlink message", zap.String("id", res.Id))

		case *packetbroker.UplinkMessageDeliveryStateChange:
			msg.HomeNetworkNetId = uint32(homeNetwork.NetID)
			msg.HomeNetworkClusterId = homeNetwork.ClusterID
			msg.HomeNetworkTenantId = homeNetwork.TenantID.ID
			msg.ForwarderNetId = uint32(forwarder.NetID)
			msg.ForwarderClusterId = forwarder.ClusterID
			msg.ForwarderTenantId = forwarder.TenantID.ID
			_, err := client.ReportUplinkMessageDeliveryState(ctx, &routingpb.UplinkMessageDeliveryStateChangeRequest{
				StateChange: msg,
			})
			if err != nil {
				return err
			}
			logger.Info("Published uplink message delivery state change")
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
	rootCmd.Flags().AddFlagSet(pbflag.MessageType())

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
	viper.ReadInConfig()
}
