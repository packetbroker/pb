// Copyright © 2021 The Things Industries B.V.

package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	reportingpb "go.packetbroker.org/api/reporting"
	"go.packetbroker.org/pb/cmd/internal/pbflag"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	reportCmd = &cobra.Command{
		Use:               "report",
		Short:             "Packet Broker report",
		PersistentPreRunE: prerunConnect,
		PersistentPostRun: postrunConnect,
	}
	reportRoutedMessagesCmd = &cobra.Command{
		Use:          "routed-messages",
		Aliases:      []string{"routedmsgs"},
		Short:        "Report routed messages",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			req := new(reportingpb.GetRoutedMessagesRequest)
			today, _ := cmd.Flags().GetBool("today")
			last30Days, _ := cmd.Flags().GetBool("last-30d")
			fromMonth, fromYear, fromOK := pbflag.GetMonthYear(cmd.Flags(), "from")
			toMonth, toYear, toOK := pbflag.GetMonthYear(cmd.Flags(), "to")
			switch {
			case today && !last30Days && !fromOK && !toOK:
				req.Time = &reportingpb.GetRoutedMessagesRequest_Today{
					Today: &reportingpb.Today{},
				}
			case last30Days && !today && !fromOK && !toOK:
				req.Time = &reportingpb.GetRoutedMessagesRequest_Last_30Days{
					Last_30Days: &reportingpb.Last30Days{},
				}
			case fromOK && toOK && !today && !last30Days:
				req.Time = &reportingpb.GetRoutedMessagesRequest_Period{
					Period: &reportingpb.MonthPeriod{
						From: &reportingpb.MonthYear{
							Month: uint32(fromMonth),
							Year:  uint32(fromYear),
						},
						To: &reportingpb.MonthYear{
							Month: uint32(toMonth),
							Year:  uint32(toYear),
						},
					},
				}
			default:
				return errors.New("specify either today, last 30 days or a period")
			}

			if forwarderID, ok := pbflag.GetTenantID(cmd.Flags(), "forwarder"); ok {
				req.ForwarderNetId = &wrapperspb.UInt32Value{
					Value: uint32(forwarderID.NetID),
				}
				if pbflag.TenantIDChanged(cmd.Flags(), "forwarder") {
					req.ForwarderTenantId = wrapperspb.String(forwarderID.ID)
				}
			}
			if homeNetworkID, ok := pbflag.GetTenantID(cmd.Flags(), "home-network"); ok {
				req.HomeNetworkNetId = &wrapperspb.UInt32Value{
					Value: uint32(homeNetworkID.NetID),
				}
				if pbflag.TenantIDChanged(cmd.Flags(), "home-network") {
					req.HomeNetworkTenantId = wrapperspb.String(homeNetworkID.ID)
				}
			}

			res, err := reportingpb.NewReporterClient(reportsConn).GetRoutedMessages(ctx, req)
			if err != nil {
				return err
			}

			return protojson.Write(os.Stdout, res)
		},
	}
)

func init() {
	rootCmd.AddCommand(reportCmd)

	reportRoutedMessagesCmd.Flags().AddFlagSet(pbflag.TenantID("forwarder"))
	reportRoutedMessagesCmd.Flags().AddFlagSet(pbflag.TenantID("home-network"))
	reportRoutedMessagesCmd.Flags().AddFlagSet(pbflag.MonthYear("from"))
	reportRoutedMessagesCmd.Flags().AddFlagSet(pbflag.MonthYear("to"))
	reportRoutedMessagesCmd.Flags().Bool("today", false, "select today")
	reportRoutedMessagesCmd.Flags().Bool("last-30d", false, "select last 30 days")
	reportCmd.AddCommand(reportRoutedMessagesCmd)
}
