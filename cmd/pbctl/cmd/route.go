// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	routingpb "go.packetbroker.org/api/routing/v2"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
)

type sortRoutesByEndpoint []*packetbroker.DevAddrPrefixRoute

func (r sortRoutesByEndpoint) Len() int {
	return len(r)
}

func (r sortRoutesByEndpoint) Less(i, j int) bool {
	if r[i].GetNetId() < r[j].GetNetId() {
		return true
	} else if r[i].GetNetId() == r[j].GetNetId() {
		if r[i].GetTenantId() < r[j].GetTenantId() {
			return true
		} else if r[i].GetTenantId() == r[j].GetTenantId() {
			return r[i].GetHomeNetworkClusterId() < r[j].GetHomeNetworkClusterId()
		}
	}
	return false
}

func (r sortRoutesByEndpoint) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type sortDevAddrRoutesByPrefix []*packetbroker.DevAddrPrefixRoute

func (r sortDevAddrRoutesByPrefix) Len() int {
	return len(r)
}

func (r sortDevAddrRoutesByPrefix) Less(i, j int) bool {
	if r[i].GetPrefix().GetValue() < r[j].GetPrefix().GetValue() {
		return true
	} else if r[i].GetPrefix().GetValue() == r[j].GetPrefix().GetValue() {
		if r[i].GetPrefix().GetLength() < r[j].GetPrefix().GetLength() {
			return true
		} else if r[i].GetPrefix().GetLength() == r[j].GetPrefix().GetLength() {
			return sortRoutesByEndpoint(r).Less(i, j)
		}
	}
	return false
}

func (r sortDevAddrRoutesByPrefix) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type sortJoinEUIPrefixRoutesByPrefix []*packetbroker.JoinEUIPrefixRoute

func (r sortJoinEUIPrefixRoutesByPrefix) Len() int {
	return len(r)
}

func (r sortJoinEUIPrefixRoutesByPrefix) Less(i, j int) bool {
	if r[i].GetPrefix().GetValue() < r[j].GetPrefix().GetValue() {
		return true
	} else if r[i].GetPrefix().GetValue() == r[j].GetPrefix().GetValue() {
		if r[i].GetPrefix().GetLength() < r[j].GetPrefix().GetLength() {
			return true
		} else if r[i].GetPrefix().GetLength() == r[j].GetPrefix().GetLength() {
			return r[i].GetId() < r[j].GetId()
		}
	}
	return false
}

func (r sortJoinEUIPrefixRoutesByPrefix) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

var routeCmd = &cobra.Command{
	Use:               "route",
	Aliases:           []string{"routes", "ro"},
	Short:             "List Packet Broker routes",
	SilenceUsage:      true,
	PersistentPreRunE: prerunConnect,
	PersistentPostRun: postrunConnect,
	RunE: func(_ *cobra.Command, _ []string) error {
		var (
			client        = routingpb.NewRoutesClient(cpConn)
			offset        = uint32(0)
			devAddrRoutes []*packetbroker.DevAddrPrefixRoute
		)
		for {
			res, err := client.ListUplinkRoutes(ctx, &routingpb.ListUplinkRoutesRequest{
				Offset: offset,
			})
			if err != nil {
				return fmt.Errorf("list uplink routes: %w", err)
			}
			devAddrRoutes = append(devAddrRoutes, res.GetRoutes()...)
			offset += uint32(len(res.GetRoutes()))
			if len(res.GetRoutes()) == 0 || offset >= res.GetTotal() {
				break
			}
		}
		sort.Sort(sortDevAddrRoutesByPrefix(devAddrRoutes))
		tabout.Println("DevAddr Prefix\tNetID\tTenant ID\tCluster ID\tTarget\t")
		for _, route := range devAddrRoutes {
			tabout.Printf("%08X/%d\t%s\t%s\t%s\t%s\t\n",
				route.GetPrefix().GetValue(),
				route.GetPrefix().GetLength(),
				packetbroker.NetID(route.GetNetId()),
				route.GetTenantId(),
				route.GetHomeNetworkClusterId(),
				(*column.Target)(route.GetTarget()),
			)
		}
		tabout.Println("")

		offset = uint32(0)
		var joinEUIPrefixRoutes []*packetbroker.JoinEUIPrefixRoute
		for {
			res, err := client.ListJoinRequestRoutes(ctx, &routingpb.ListJoinRequestRoutesRequest{
				Offset: offset,
			})
			if err != nil {
				return fmt.Errorf("list join-request routes: %w", err)
			}
			joinEUIPrefixRoutes = append(joinEUIPrefixRoutes, res.GetRoutes()...)
			offset += uint32(len(res.GetRoutes()))
			if len(res.GetRoutes()) == 0 || offset >= res.GetTotal() {
				break
			}
		}
		sort.Sort(sortJoinEUIPrefixRoutesByPrefix(joinEUIPrefixRoutes))
		tabout.Println("JoinEUI Prefix\tJoin Server ID\tResolver\t")
		for _, route := range joinEUIPrefixRoutes {
			var resolver string
			if lookup := route.GetLookup(); lookup != nil {
				resolver = (*column.Target)(lookup).String()
			} else if fixed := route.GetFixed(); fixed != nil {
				resolver = (*column.JoinServerFixedEndpoint)(fixed).String()
			}
			tabout.Printf("%016X/%d\t%14d\t%s\t\n",
				route.GetPrefix().GetValue(),
				route.GetPrefix().GetLength(),
				route.GetId(),
				resolver,
			)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(routeCmd)
}
