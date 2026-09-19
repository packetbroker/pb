// SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package cmd implements the pbadmin commands.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.packetbroker.org/pb/cmd/internal/column"
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
	tabout = column.NewWriter(os.Stdout)
)

func prerunConnect(_ *cobra.Command, _ []string) error {
	clientConf, err := config.AutomaticClient(ctx, "iam", config.BasicAuthIAM, "networks")
	if err != nil {
		return fmt.Errorf("configure IAM client: %w", err)
	}
	conn, err = client.DialContext(ctx, logger, clientConf, 443)
	if err != nil {
		return fmt.Errorf("connect to IAM: %w", err)
	}
	return nil
}

func postrunConnect(_ *cobra.Command, _ []string) {
	// The connection is no longer used; a close error is not actionable.
	_ = conn.Close()
}

var rootCmd = &cobra.Command{
	Use:          "pbadmin",
	Short:        "pbadmin can be used to manage networks, tenants and API keys.",
	SilenceUsage: true,
}

// Execute runs pbadmin.
func Execute() {
	logger = logging.GetLogger(debug)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := tabout.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Syncing stderr is not supported on all platforms; the error is not actionable.
	_ = logger.Sync()
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
	if err := config.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Warning:", err)
	}
}
