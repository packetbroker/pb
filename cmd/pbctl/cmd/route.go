// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sortRoutesByPrefix []*packetbroker.DevAddrPrefixRoute

func (r sortRoutesByPrefix) Len() int {
	return len(r)
}

func (r sortRoutesByPrefix) Less(i, j int) bool {
	if r[i].GetPrefix().GetValue() < r[j].GetPrefix().GetValue() {
		return true
	} else if r[i].GetPrefix().GetValue() == r[j].GetPrefix().GetValue() {
		return r[i].GetPrefix().GetLength() < r[j].GetPrefix().GetLength()
	}
	return false
}

func (r sortRoutesByPrefix) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

var routeCmd = &cobra.Command{
	Use:          "route",
	Aliases:      []string{"routes", "ro"},
	Short:        "List Packet Broker routes",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			client        = routingpb.NewRoutesClient(conn)
			lastCreatedAt *timestamppb.Timestamp
			routes        []*packetbroker.DevAddrPrefixRoute
		)
		for {
			res, err := client.ListRoutes(ctx, &routingpb.ListRoutesRequest{
				CreatedSince: lastCreatedAt,
			})
			if err != nil {
				return err
			}
			if len(res.Routes) == 0 {
				break
			}
			lastCreatedAt = res.Routes[len(res.Routes)-1].GetCreatedAt()
			routes = append(routes, res.Routes...)
		}
		sort.Sort(sortRoutesByPrefix(routes))
		fmt.Fprintln(tabout, "DevAddr Prefix\tNetID\tTenant ID\tCluster ID\t")
		for _, p := range routes {
			fmt.Fprintf(tabout,
				"%08X/%d\t%s\t%s\t%s\t\n",
				p.GetPrefix().GetValue(),
				p.GetPrefix().GetLength(),
				packetbroker.NetID(p.GetNetId()),
				p.GetTenantId(),
				p.GetHomeNetworkClusterId(),
			)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(routeCmd)
}
