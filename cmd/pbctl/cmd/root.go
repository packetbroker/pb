// SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package cmd implements the pbctl commands.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/packetbroker/pb/cmd/internal/column"
	"github.com/packetbroker/pb/cmd/internal/config"
	"github.com/packetbroker/pb/cmd/internal/gen"
	"github.com/packetbroker/pb/cmd/internal/logging"
	"github.com/packetbroker/pb/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	cfgFile string
	debug   bool

	ctx    = context.Background()
	logger *zap.Logger
	iamConn,
	cpConn,
	reportsConn *grpc.ClientConn
	tabout = column.NewWriter(os.Stdout)
)

func prerunConnect(_ *cobra.Command, _ []string) error {
	iamClientConf, err := config.OAuth2Client(ctx, "iam", "networks")
	if err != nil {
		return fmt.Errorf("configure IAM client: %w", err)
	}
	iamConn, err = client.DialContext(ctx, logger, iamClientConf, 443)
	if err != nil {
		return fmt.Errorf("connect to IAM: %w", err)
	}

	cpClientConf, err := config.OAuth2Client(ctx, "controlplane", "networks")
	if err != nil {
		return fmt.Errorf("configure Control Plane client: %w", err)
	}
	cpConn, err = client.DialContext(ctx, logger, cpClientConf, 443)
	if err != nil {
		return fmt.Errorf("connect to Control Plane: %w", err)
	}

	reportsClientConf, err := config.OAuth2Client(ctx, "reports", "networks")
	if err != nil {
		return fmt.Errorf("configure Reporter client: %w", err)
	}
	reportsConn, err = client.DialContext(ctx, logger, reportsClientConf, 443)
	if err != nil {
		return fmt.Errorf("connect to Reporter: %w", err)
	}

	return nil
}

func postrunConnect(_ *cobra.Command, _ []string) {
	// The connections are no longer used; close errors are not actionable.
	_ = iamConn.Close()
	_ = cpConn.Close()
	_ = reportsConn.Close()
}

var rootCmd = &cobra.Command{
	Use:          "pbctl",
	Short:        "pbctl can be used to manage routing policies and list routes.",
	SilenceUsage: true,
}

// Execute runs pbctl.
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
	rootCmd.PersistentFlags().AddFlagSet(config.ClientFlags("controlplane", "cp.packetbroker.net:443"))
	rootCmd.PersistentFlags().AddFlagSet(config.ClientFlags("reports", "reports.packetbroker.net:443"))
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
	if err := config.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Warning:", err)
	}
}
