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
	conn   *grpc.ClientConn
	tabout = tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
)

func prerunConnect(cmd *cobra.Command, args []string) error {
	clientConf, err := config.AutomaticClient(ctx, "iam", config.BasicAuthIAM, "networks")
	if err != nil {
		return err
	}
	conn, err = client.DialContext(ctx, logger, clientConf, 443)
	if err != nil {
		return err
	}
	return nil
}

func postrunConnect(cmd *cobra.Command, args []string) {
	conn.Close()
}

var rootCmd = &cobra.Command{
	Use:          "pbadmin",
	Short:        "pbadmin can be used to manage networks, tenants and API keys.",
	SilenceUsage: true,
}

// Execute runs pbadmin.
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
	rootCmd.PersistentFlags().AddFlagSet(config.BasicAuthClientFlags(config.BasicAuthIAM))
	rootCmd.PersistentFlags().AddFlagSet(config.OAuth2ClientFlags())

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .pb.yaml, $HOME/.pb.yaml)")
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
