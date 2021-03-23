// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.packetbroker.org/pb/cmd/internal/config"
	"go.packetbroker.org/pb/cmd/internal/logging"
	"go.packetbroker.org/pb/pkg/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	cfgFile string
	debug   bool

	ctx    = context.Background()
	logger *zap.Logger
	iamConn,
	cpConn *grpc.ClientConn
	tabout = tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
)

var rootCmd = &cobra.Command{
	Use:   "pbctl",
	Short: "pbctl can be used to manage routing policies and list routes.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger = logging.GetLogger(debug)

		iamClientConf, err := config.OAuth2Client(ctx, "iam", "")
		if err != nil {
			return err
		}
		iamConn, err = client.DialContext(ctx, logger, iamClientConf, 443)
		if err != nil {
			return err
		}

		cpClientConf, err := config.OAuth2Client(ctx, "controlplane", "networks")
		if err != nil {
			return err
		}
		cpConn, err = client.DialContext(ctx, logger, cpClientConf, 443)
		if err != nil {
			return err
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		logger.Sync()
		tabout.Flush()
		iamConn.Close()
		cpConn.Close()
	},
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

	rootCmd.PersistentFlags().AddFlagSet(config.ClientFlags("iam", "iam.packetbroker.org:443"))
	rootCmd.PersistentFlags().AddFlagSet(config.ClientFlags("controlplane", "cp.packetbroker.org:443"))
	rootCmd.PersistentFlags().AddFlagSet(config.OAuth2ClientFlags())

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pb.yaml, .pb.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug mode")
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
