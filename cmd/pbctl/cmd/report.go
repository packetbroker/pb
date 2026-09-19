// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	iampb "go.packetbroker.org/api/iam/v2"
	reportingpb "go.packetbroker.org/api/reporting"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/pbflag"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"go.packetbroker.org/pb/pkg/csv"
	"go.packetbroker.org/pb/pkg/graph"
)

var (
	reportCmd = &cobra.Command{
		Use:               "report",
		Aliases:           []string{"reports"},
		Short:             "Packet Broker report",
		PersistentPreRunE: prerunConnect,
		PersistentPostRun: postrunConnect,
	}
	reportRoutedMessagesCmd = &cobra.Command{
		Use:     "routed-messages",
		Aliases: []string{"routedmsgs"},
		Short:   "Report routed messages",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Query the routed messages for the selected period.
			// If a generic tenant ID is provided, request the routed messages both as Forwarder and Home Network.
			// Otherwise, the routed messages are either requested for the Forwarder or Home Network, or between the given
			// Forwarder and Home Network.
			var (
				records                     []*reportingpb.RoutedMessagesRecord
				format                      = *cmd.Flags().Lookup("format").Value.(*reportFormat)
				today, _                    = cmd.Flags().GetBool("today")
				last30Days, _               = cmd.Flags().GetBool("last-30d")
				fromMonth, fromYear, fromOK = pbflag.GetMonthYear(cmd.Flags(), "from")
				toMonth, toYear, toOK       = pbflag.GetMonthYear(cmd.Flags(), "to")
				anyRole                     bool
				highlight                   *packetbroker.TenantID
			)
			for _, actor := range []string{"forwarder", "home-network", ""} {
				if tenantID, ok := pbflag.GetTenantID(cmd.Flags(), actor); ok {
					if actor == "" && anyRole {
						return errors.New("specify either any role or (a) specific role(s)")
					}
					anyRole = true
					if highlight == nil {
						highlight = &tenantID
					} else {
						highlight = nil
					}
				}
			}
			for _, fillFn := range []func(req *reportingpb.GetRoutedMessagesRequest){
				func(req *reportingpb.GetRoutedMessagesRequest) {
					req.ForwarderNetId, req.ForwarderTenantId = pbflag.GetTenantIDWrappers(cmd.Flags(), "forwarder")
					req.HomeNetworkNetId, req.HomeNetworkTenantId = pbflag.GetTenantIDWrappers(cmd.Flags(), "home-network")
				},
				func(req *reportingpb.GetRoutedMessagesRequest) {
					req.ForwarderNetId, req.ForwarderTenantId = pbflag.GetTenantIDWrappers(cmd.Flags(), "")
				},
				func(req *reportingpb.GetRoutedMessagesRequest) {
					req.HomeNetworkNetId, req.HomeNetworkTenantId = pbflag.GetTenantIDWrappers(cmd.Flags(), "")
				},
			} {
				req := new(reportingpb.GetRoutedMessagesRequest)
				fillFn(req)
				if req.GetForwarderNetId() == nil && req.GetHomeNetworkNetId() == nil {
					continue
				}
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
					if format.isImage() && (fromMonth != toMonth || fromYear != toYear) {
						return errors.New("cannot produce image of period")
					}
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
				res, err := reportingpb.NewReporterClient(reportsConn).GetRoutedMessages(ctx, req)
				if err != nil {
					return fmt.Errorf("get routed messages: %w", err)
				}
			nextRecord:
				for _, record := range res.GetRecords() {
					fID, hnID := packetbroker.ForwarderTenantID(record), packetbroker.HomeNetworkTenantID(record)
					for _, existing := range records {
						// Skip duplicate records.
						if packetbroker.ForwarderTenantID(existing) == fID && packetbroker.HomeNetworkTenantID(existing) == hnID {
							continue nextRecord
						}
					}
					records = append(records, record)
				}
			}
			sort.Sort(byToForwarderHomeNetwork(records))

			// List the listed networks so we can put the names in the report.
			var (
				networks      []*packetbroker.NetworkOrTenant
				offset        uint32
				catalogClient = iampb.NewCatalogClient(iamConn)
			)
			for {
				res, err := catalogClient.ListNetworks(ctx, &iampb.ListNetworksRequest{
					Offset: offset,
				})
				if err != nil {
					return fmt.Errorf("list networks: %w", err)
				}
				networks = append(networks, res.GetNetworks()...)
				if len(networks) >= int(res.GetTotal()) {
					break
				}
				offset += uint32(len(res.GetNetworks()))
			}
			networkMap := make(map[packetbroker.TenantID]*packetbroker.NetworkOrTenant, len(networks))
			for _, n := range networks {
				switch nt := n.GetValue().(type) {
				case *packetbroker.NetworkOrTenant_Network:
					networkMap[packetbroker.TenantID{NetID: packetbroker.NetID(nt.Network.GetNetId())}] = n
				case *packetbroker.NetworkOrTenant_Tenant:
					networkMap[packetbroker.RequestTenantID(nt.Tenant)] = n
				}
			}

			// Determine the output: a (temporary) file or stdout.
			var (
				output    io.Writer = os.Stdout
				closeFunc func() error
			)
			if outputFile, _ := cmd.Flags().GetString("output-file"); outputFile != "" {
				file, err := os.Create(outputFile) //nolint:gosec // the output file path is provided by the user of this CLI
				if err != nil {
					return fmt.Errorf("create file: %w", err)
				}
				output, closeFunc = file, file.Close
			} else if format.isImage() {
				workDir, _ := os.Getwd()
				file, err := os.CreateTemp(workDir, fmt.Sprintf("pbreport-*%s", format.ext()))
				if err != nil {
					return fmt.Errorf("create temporary file: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Writing to %s\n", file.Name())
				output, closeFunc = file, file.Close
			}

			// Write to the output.
			if err := writeRoutedMessages(output, format, records, networkMap, highlight); err != nil {
				if closeFunc != nil {
					_ = closeFunc()
				}
				return err
			}
			if closeFunc != nil {
				if err := closeFunc(); err != nil {
					return fmt.Errorf("close output file: %w", err)
				}
			}
			return nil
		},
	}
)

// writeRoutedMessages writes the records of routed messages to the output in the given format.
func writeRoutedMessages(
	output io.Writer,
	format reportFormat,
	records []*reportingpb.RoutedMessagesRecord,
	networks map[packetbroker.TenantID]*packetbroker.NetworkOrTenant,
	highlight *packetbroker.TenantID,
) error {
	switch format {
	case "json":
		for _, record := range records {
			if err := protojson.Write(output, record); err != nil {
				return fmt.Errorf("write record: %w", err)
			}
		}
		return nil
	case "csv":
		if err := csv.WriteRoutedMessages(output, records, networks); err != nil {
			return fmt.Errorf("write CSV: %w", err)
		}
		return nil
	case "dot":
		if err := graph.WriteRoutedMessages(output, records, networks, highlight); err != nil {
			return fmt.Errorf("write graph: %w", err)
		}
		return nil
	case "svg", "png", "pdf", "ps":
		reader, writer := io.Pipe()
		go func() {
			// A write error is propagated to the reader.
			_ = writer.CloseWithError(graph.WriteRoutedMessages(writer, records, networks, highlight))
		}()
		if err := graph.RunDot(ctx, reader, output, string(format)); err != nil {
			fmt.Fprintln(os.Stderr, "Running a Graphviz command failed. Is Graphviz installed?")
			fmt.Fprintln(os.Stderr, "Download and install from https://graphviz.org/download/")
			return fmt.Errorf("convert graph: %w", err)
		}
		return nil
	default:
		return errors.New("unsupported format")
	}
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportRoutedMessagesCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	reportRoutedMessagesCmd.Flags().AddFlagSet(pbflag.TenantID("forwarder"))
	reportRoutedMessagesCmd.Flags().AddFlagSet(pbflag.TenantID("home-network"))
	reportRoutedMessagesCmd.Flags().AddFlagSet(pbflag.MonthYear("from"))
	reportRoutedMessagesCmd.Flags().AddFlagSet(pbflag.MonthYear("to"))
	reportRoutedMessagesCmd.Flags().Bool("today", false, "select today")
	reportRoutedMessagesCmd.Flags().Bool("last-30d", false, "select last 30 days")
	reportRoutedMessagesCmd.Flags().VarP(newReportFormat("json"), "format", "f",
		fmt.Sprintf("format (%s)", strings.Join(reportFormats[:], ", ")),
	)
	reportRoutedMessagesCmd.Flags().StringP("output-file", "o", "", "output file")
	reportCmd.AddCommand(reportRoutedMessagesCmd)
}
