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
	conn   *grpc.ClientConn
	tabout = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
)

var rootCmd = &cobra.Command{
	Use:   "pbadmin",
	Short: "pbadmin can be used to manage networks, tenants and API keys.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger = logging.GetLogger(debug)
		clientConf, err := config.AutomaticClient(ctx, "iam", config.BasicAuthIAM, "networks")
		if err != nil {
			return err
		}
		conn, err = client.DialContext(ctx, logger, clientConf, 443)
		if err != nil {
			return err
		}
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		logger.Sync()
		tabout.Flush()
		conn.Close()
	},
}

// Execute runs pbadmin.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().AddFlagSet(config.ClientFlags("iam", "iam.packetbroker.org:443"))
	rootCmd.PersistentFlags().AddFlagSet(config.BasicAuthClientFlags(config.BasicAuthIAM))
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
