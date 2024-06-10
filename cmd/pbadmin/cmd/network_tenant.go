// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	iampb "go.packetbroker.org/api/iam"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
	pbflag "go.packetbroker.org/pb/cmd/internal/pbflag"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	networkTenantCmd = &cobra.Command{
		Use:     "tenant",
		Aliases: []string{"tenants", "tnt", "tnts", "t"},
		Short:   "Manage Packet Broker tenants",
	}
	networkTenantListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List tenants",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				netID, _        = pbflag.GetNetID(cmd.Flags(), "")
				offset          = uint32(0)
				idContains, _   = cmd.Flags().GetString("id-contains")
				nameContains, _ = cmd.Flags().GetString("name-contains")
			)
			fmt.Fprintln(tabout, "NetID\tTenant ID\tAuthority\tName\tDevAddr Blocks\tListed\tTarget\t")
			for {
				res, err := iampb.NewTenantRegistryClient(conn).ListTenants(ctx, &iampb.ListTenantsRequest{
					NetId:            uint32(netID),
					Offset:           offset,
					TenantIdContains: idContains,
					NameContains:     nameContains,
				})
				if err != nil {
					return err
				}
				for _, t := range res.Tenants {
					fmt.Fprintf(tabout, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
						packetbroker.NetID(t.GetNetId()),
						t.GetTenantId(),
						t.GetAuthority(),
						t.GetName(),
						column.DevAddrBlocks(t.GetDevAddrBlocks()),
						column.YesNo(t.GetListed()),
						(*column.Target)(t.GetTarget()),
					)
				}
				offset += uint32(len(res.Tenants))
				if len(res.Tenants) == 0 || offset >= res.Total {
					break
				}
			}
			return nil
		},
	}
	networkTenantCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a tenant",
		Example: `
  Create:
    $ pbadmin network tenant create --net-id 000013 --tenant-id tti

  Create with name and listed in the catalog:
    $ pbadmin network tenant create --net-id 000013 --tenant-id tti \
      --name "The Things Industries" --listed

  Define DevAddr blocks to named clusters:
    $ pbadmin network tenant create --net-id 000013 --tenant-id tti \
      --dev-addr-blocks 26011000/20=eu1,26012000=eu2

  Configure a LoRaWAN Backend Interfaces 1.1.0 target with HTTP basic auth:
    $ pbadmin network tenant create --net-id 000013 --tenant-id tti \
      --target-protocol TS002_V1_1 \
      --target-address https://user:pass@example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			name, _ := cmd.Flags().GetString("name")
			devAddrBlocks, _, _ := pbflag.GetDevAddrBlocks(cmd.Flags())
			adminContact := pbflag.GetContactInfo(cmd.Flags(), "admin")
			techContact := pbflag.GetContactInfo(cmd.Flags(), "tech")
			listed, _ := cmd.Flags().GetBool("listed")
			var target *packetbroker.Target
			if err := pbflag.ApplyToTarget(cmd.Flags(), "target", &target); err != nil {
				return err
			}
			res, err := iampb.NewTenantRegistryClient(conn).CreateTenant(ctx, &iampb.CreateTenantRequest{
				Tenant: &packetbroker.Tenant{
					NetId:                 uint32(tenantID.NetID),
					TenantId:              tenantID.ID,
					Name:                  name,
					DevAddrBlocks:         devAddrBlocks,
					AdministrativeContact: adminContact,
					TechnicalContact:      techContact,
					Listed:                listed,
					Target:                target,
				},
			})
			if err != nil {
				return err
			}
			return column.WriteTenant(tabout, res.Tenant, false)
		},
	}
	networkTenantGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a tenant",
		Example: `
  Get:
    $ pbadmin network tenant get --net-id 000013 --tenant-id tti`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			res, err := iampb.NewTenantRegistryClient(conn).GetTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			if err != nil {
				return err
			}
			verbose, _ := cmd.Flags().GetBool("verbose")
			return column.WriteTenant(tabout, res.Tenant, verbose)
		},
	}
	networkTenantUpdateCmd = &cobra.Command{
		Use:     "update",
		Aliases: []string{"up"},
		Short:   "Update a tenant",
		Example: `
  Update name:
    $ pbadmin network tenant update --net-id 000013 --tenant-id tti \
      --name "The Things Network"

  Define DevAddr blocks to named clusters:
    $ pbadmin network tenant update --net-id 000013 --tenant-id tti \
      --dev-addr-blocks 26011000/20=eu1,26012000=eu2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			client := iampb.NewTenantRegistryClient(conn)
			tnt, err := client.GetTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("listed") {
				listed, _ := cmd.Flags().GetBool("listed")
				_, err := client.UpdateTenantListed(ctx, &iampb.UpdateTenantListedRequest{
					NetId:    uint32(tenantID.NetID),
					TenantId: tenantID.ID,
					Listed:   listed,
				})
				if err != nil {
					return err
				}
			}
			var any bool
			req := &iampb.UpdateTenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			}
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = wrapperspb.String(name)
				any = true
			}
			devAddrBlocksAll, devAddrBlocksAllAdd, devAddrBlocksAllRemove := pbflag.GetDevAddrBlocks(cmd.Flags())
			if cmd.Flags().Changed("dev-addr-blocks") {
				req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
					Value: devAddrBlocksAll,
				}
				any = true
			} else if len(devAddrBlocksAllAdd) > 0 || len(devAddrBlocksAllRemove) > 0 {
				req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
					Value: mergeDevAddrBlocks(tnt.Tenant.DevAddrBlocks, devAddrBlocksAllAdd, devAddrBlocksAllRemove),
				}
				any = true
			}
			if adminContact := pbflag.GetContactInfo(cmd.Flags(), "admin"); adminContact != nil {
				req.AdministrativeContact = &packetbroker.ContactInfoValue{
					Value: adminContact,
				}
				any = true
			}
			if techContact := pbflag.GetContactInfo(cmd.Flags(), "tech"); techContact != nil {
				req.TechnicalContact = &packetbroker.ContactInfoValue{
					Value: techContact,
				}
				any = true
			}
			if any {
				_, err = client.UpdateTenant(ctx, req)
				return err
			}
			return nil
		},
	}
	networkTenantUpdateTargetCmd = &cobra.Command{
		Use:   "target",
		Short: "Update a tenant target",
		Example: `
  Configure a LoRaWAN Backend Interfaces 1.0 target with Packet Broker token
  authentication:
    $ pbadmin network tenant update target --net-id 000013 --tenant-id tti \
      --protocol TS002_V1_0 --address https://example.com --pb-token

  Configure a LoRaWAN Backend Interfaces 1.0 target with HTTP basic auth:
    $ pbadmin network tenant update target --net-id 000013 --tenant-id tti \
      --protocol TS002_V1_0 --address https://user:pass@example.com

  Configure a LoRaWAN Backend Interfaces 1.0 target with TLS:
    $ pbadmin network tenant update target --net-id 000013 --tenant-id tti \
      --protocol TS002_V1_0 --address https://example.com \
      --root-cas-file ca.pem --tls-cert-file key.pem --tls-key-file key.pem

  Configure a LoRaWAN Backend Interfaces 1.0 target with TLS and custom
  originating NetID:
    $ pbadmin network tenant update target --net-id 000013 --tenant-id tti \
      --origin-net-id 000013 \
      --root-cas-file ca.pem --tls-cert-file key.pem --tls-key-file key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			client := iampb.NewTenantRegistryClient(conn)
			tnt, err := client.GetTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			if err != nil {
				return err
			}
			target := tnt.Tenant.Target
			if err := pbflag.ApplyToTarget(cmd.Flags(), "", &target); err != nil {
				return err
			}
			req := &iampb.UpdateTenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
				Target: &iampb.TargetValue{
					Value: target,
				},
			}
			_, err = client.UpdateTenant(ctx, req)
			return err
		},
	}
	networkTenantDeleteCmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm"},
		Short:   "Delete a tenant",
		Example: `
  Delete:
    $ pbadmin network tenant delete --net-id 000013 --tenant-id tti`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			_, err := iampb.NewTenantRegistryClient(conn).DeleteTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			return err
		},
	}
	networkTenantDeleteTargetCmd = &cobra.Command{
		Use:   "target",
		Short: "Delete a tenant target",
		Example: `
  Delete a tenant target:
    $ pbadmin network delete target --net-id 000013 --tenant-id tti`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			client := iampb.NewTenantRegistryClient(conn)
			req := &iampb.UpdateTenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
				Target: &iampb.TargetValue{
					Value: nil,
				},
			}
			_, err := client.UpdateTenant(ctx, req)
			return err
		},
	}
)

func tenantSettingsFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.String("name", "", "tenant name")
	flags.Bool("listed", false, "list tenant in catalog")
	return flags
}

func init() {
	networkCmd.AddCommand(networkTenantCmd)

	networkTenantListCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkTenantListCmd.Flags().String("id-contains", "", "filter tenants by ID")
	networkTenantListCmd.Flags().String("name-contains", "", "filter tenants by name")
	networkTenantCmd.AddCommand(networkTenantListCmd)

	networkTenantCreateCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantCreateCmd.Flags().AddFlagSet(tenantSettingsFlags())
	networkTenantCreateCmd.Flags().AddFlagSet(pbflag.DevAddrBlocks(false))
	networkTenantCreateCmd.Flags().AddFlagSet(pbflag.Target("target"))
	networkTenantCreateCmd.Flags().AddFlagSet(pbflag.ContactInfo("admin"))
	networkTenantCreateCmd.Flags().AddFlagSet(pbflag.ContactInfo("tech"))
	networkTenantCmd.AddCommand(networkTenantCreateCmd)

	networkTenantGetCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantGetCmd.Flags().Bool("verbose", false, "verbose output")
	networkTenantCmd.AddCommand(networkTenantGetCmd)

	networkTenantUpdateCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantUpdateCmd.Flags().AddFlagSet(tenantSettingsFlags())
	networkTenantUpdateCmd.Flags().AddFlagSet(pbflag.DevAddrBlocks(true))
	networkTenantUpdateCmd.Flags().AddFlagSet(pbflag.ContactInfo("admin"))
	networkTenantUpdateCmd.Flags().AddFlagSet(pbflag.ContactInfo("tech"))
	networkTenantUpdateTargetCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantUpdateTargetCmd.Flags().AddFlagSet(pbflag.Target(""))
	networkTenantUpdateCmd.AddCommand(networkTenantUpdateTargetCmd)
	networkTenantCmd.AddCommand(networkTenantUpdateCmd)

	networkTenantDeleteCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantCmd.AddCommand(networkTenantDeleteCmd)
}
