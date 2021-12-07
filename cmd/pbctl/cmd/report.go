// Copyright © 2021 The Things Industries B.V.

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
				any                         bool
				highlight                   *packetbroker.TenantID
			)
			for _, actor := range []string{"forwarder", "home-network", ""} {
				if id, ok := pbflag.GetTenantID(cmd.Flags(), actor); ok {
					if actor == "" && any {
						return errors.New("specify either any role or (a) specific role(s)")
					}
					any = true
					if highlight == nil {
						highlight = &id
					} else {
						highlight = nil
					}
				}
			}
			for _, fillFn := range []func(req *reportingpb.GetRoutedMessagesRequest) error{
				func(req *reportingpb.GetRoutedMessagesRequest) error {
					req.ForwarderNetId, req.ForwarderTenantId = pbflag.GetTenantIDWrappers(cmd.Flags(), "forwarder")
					req.HomeNetworkNetId, req.HomeNetworkTenantId = pbflag.GetTenantIDWrappers(cmd.Flags(), "home-network")
					return nil
				},
				func(req *reportingpb.GetRoutedMessagesRequest) error {
					req.ForwarderNetId, req.ForwarderTenantId = pbflag.GetTenantIDWrappers(cmd.Flags(), "")
					return nil
				},
				func(req *reportingpb.GetRoutedMessagesRequest) error {
					req.HomeNetworkNetId, req.HomeNetworkTenantId = pbflag.GetTenantIDWrappers(cmd.Flags(), "")
					return nil
				},
			} {
				req := new(reportingpb.GetRoutedMessagesRequest)
				if err := fillFn(req); err != nil {
					return err
				}
				if req.ForwarderNetId == nil && req.HomeNetworkNetId == nil {
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
					return err
				}
			nextRecord:
				for _, nr := range res.Records {
					if fID, hnID := packetbroker.ForwarderTenantID(nr), packetbroker.HomeNetworkTenantID(nr); fID == hnID {
						for _, er := range records {
							// Skip duplicate records.
							if packetbroker.ForwarderTenantID(er) == fID && packetbroker.HomeNetworkTenantID(er) == hnID {
								continue nextRecord
							}
						}
					}
					records = append(records, nr)
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
				networks = append(networks, res.Networks...)
				if len(networks) >= int(res.Total) {
					break
				}
				offset += uint32(len(res.Networks))
			}
			networkMap := make(map[packetbroker.TenantID]*packetbroker.NetworkOrTenant, len(networks))
			for _, n := range networks {
				switch nt := n.Value.(type) {
				case *packetbroker.NetworkOrTenant_Network:
					networkMap[packetbroker.TenantID{NetID: packetbroker.NetID(nt.Network.NetId)}] = n
				case *packetbroker.NetworkOrTenant_Tenant:
					networkMap[packetbroker.RequestTenantID(nt.Tenant)] = n
				}
			}

			// Determine the output: a (temporary) file or stdout.
			var output io.WriteCloser
			if outputFile, _ := cmd.Flags().GetString("output-file"); outputFile != "" {
				var err error
				output, err = os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("create file: %w", err)
				}
				defer output.Close()
			} else if format.isImage() {
				wd, _ := os.Getwd()
				f, err := os.CreateTemp(wd, fmt.Sprintf("pbreport-*%s", format.ext()))
				if err != nil {
					return fmt.Errorf("create temporary file: %w", err)
				}
				defer f.Close()
				fmt.Fprintf(os.Stderr, "Writing to %s\n", f.Name())
				output = f
			} else {
				output = os.Stdout
			}

			// Write to the output.
			switch format {
			case "json":
				for _, rec := range records {
					if err := protojson.Write(output, rec); err != nil {
						return err
					}
				}
				return nil
			case "csv":
				return csv.WriteRoutedMessages(output, records, networkMap)
			case "dot":
				return graph.WriteRoutedMessages(output, records, networkMap, highlight)
			case "svg", "png", "pdf", "ps":
				rd, w := io.Pipe()
				go func() {
					defer w.Close()
					graph.WriteRoutedMessages(w, records, networkMap, highlight)
				}()
				if err := graph.RunDot(ctx, rd, output, string(format)); err != nil {
					fmt.Fprintln(os.Stderr, "Running a Graphviz command failed. Is Graphviz installed?")
					fmt.Fprintln(os.Stderr, "Download and install from https://graphviz.org/download/")
					return err
				}
				return nil
			default:
				return errors.New("unsupported format")
			}
		},
	}
)

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
