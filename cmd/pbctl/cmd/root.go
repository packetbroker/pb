// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.packetbroker.org/pb/cmd/internal/config"
	"go.packetbroker.org/pb/cmd/internal/gen"
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

func prerunConnect(cmd *cobra.Command, args []string) error {
	iamClientConf, err := config.OAuth2Client(ctx, "iam", "networks")
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
}

func postrunConnect(cmd *cobra.Command, args []string) {
	iamConn.Close()
	cpConn.Close()
}

var rootCmd = &cobra.Command{
	Use:   "pbctl",
	Short: "pbctl can be used to manage routing policies and list routes.",
}

// Execute runs pbctl.
func Execute() {
	logger = logging.GetLogger(debug)
	defer logger.Sync()
	defer tabout.Flush()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().AddFlagSet(config.ClientFlags("iam", "iam.packetbroker.net:443"))
	rootCmd.PersistentFlags().AddFlagSet(config.ClientFlags("controlplane", "cp.packetbroker.net:443"))
	rootCmd.PersistentFlags().AddFlagSet(config.OAuth2ClientFlags())

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pb.yaml, .pb.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug mode")

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
