// Copyright © 2021 The Things Industries B.V.

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
	if r[i].NetId < r[j].NetId {
		return true
	} else if r[i].NetId == r[j].NetId {
		if r[i].TenantId < r[j].TenantId {
			return true
		} else if r[i].TenantId == r[j].TenantId {
			return r[i].HomeNetworkClusterId < r[j].HomeNetworkClusterId
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
			return r[i].Id < r[j].Id
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
	RunE: func(cmd *cobra.Command, args []string) error {
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
				return err
			}
			devAddrRoutes = append(devAddrRoutes, res.Routes...)
			offset += uint32(len(res.Routes))
			if len(res.Routes) == 0 || offset >= res.Total {
				break
			}
		}
		sort.Sort(sortDevAddrRoutesByPrefix(devAddrRoutes))
		fmt.Fprintln(tabout, "DevAddr Prefix\tNetID\tTenant ID\tCluster ID\tTarget\t")
		for _, p := range devAddrRoutes {
			fmt.Fprintf(tabout,
				"%08X/%d\t%s\t%s\t%s\t%s\t\n",
				p.GetPrefix().GetValue(),
				p.GetPrefix().GetLength(),
				packetbroker.NetID(p.GetNetId()),
				p.GetTenantId(),
				p.GetHomeNetworkClusterId(),
				(*column.Target)(p.Target),
			)
		}
		fmt.Fprintln(tabout)

		offset = uint32(0)
		var joinEUIPrefixRoutes []*packetbroker.JoinEUIPrefixRoute
		for {
			res, err := client.ListJoinRequestRoutes(ctx, &routingpb.ListJoinRequestRoutesRequest{
				Offset: offset,
			})
			if err != nil {
				return err
			}
			joinEUIPrefixRoutes = append(joinEUIPrefixRoutes, res.Routes...)
			offset += uint32(len(res.Routes))
			if len(res.Routes) == 0 || offset >= res.Total {
				break
			}
		}
		sort.Sort(sortJoinEUIPrefixRoutesByPrefix(joinEUIPrefixRoutes))
		fmt.Fprintln(tabout, "JoinEUI Prefix\tJoin Server ID\tResolver\t")
		for _, p := range joinEUIPrefixRoutes {
			var resolver string
			if lookup := p.GetLookup(); lookup != nil {
				resolver = (*column.Target)(lookup).String()
			} else if fixed := p.GetFixed(); fixed != nil {
				resolver = (*column.JoinServerFixedEndpoint)(fixed).String()
			}
			fmt.Fprintf(tabout,
				"%016X/%d\t%14d\t%s\t\n",
				p.GetPrefix().GetValue(),
				p.GetPrefix().GetLength(),
				p.GetId(),
				resolver,
			)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(routeCmd)
}
