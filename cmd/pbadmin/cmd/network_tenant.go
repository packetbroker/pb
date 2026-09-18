// SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			var (
				netID, _        = pbflag.GetNetID(cmd.Flags(), "")
				offset          = uint32(0)
				idContains, _   = cmd.Flags().GetString("id-contains")
				nameContains, _ = cmd.Flags().GetString("name-contains")
			)
			tabout.Println("NetID\tTenant ID\tAuthority\tName\tDevAddr Blocks\tListed\tTarget\t")
			for {
				res, err := iampb.NewTenantRegistryClient(conn).ListTenants(ctx, &iampb.ListTenantsRequest{
					NetId:            uint32(netID),
					Offset:           offset,
					TenantIdContains: idContains,
					NameContains:     nameContains,
				})
				if err != nil {
					return fmt.Errorf("list tenants: %w", err)
				}
				for _, t := range res.GetTenants() {
					tabout.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
						packetbroker.NetID(t.GetNetId()),
						t.GetTenantId(),
						t.GetAuthority(),
						t.GetName(),
						column.DevAddrBlocks(t.GetDevAddrBlocks()),
						column.YesNo(t.GetListed()),
						(*column.Target)(t.GetTarget()),
					)
				}
				offset += uint32(len(res.GetTenants()))
				if len(res.GetTenants()) == 0 || offset >= res.GetTotal() {
					break
				}
			}
			return nil
		},
	}
	networkTenantCreateCmd = &cobra.Command{ //nolint:gosec // the example URL contains a placeholder password
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			name, _ := cmd.Flags().GetString("name")
			devAddrBlocks, _, _ := pbflag.GetDevAddrBlocks(cmd.Flags())
			adminContact := pbflag.GetContactInfo(cmd.Flags(), "admin")
			techContact := pbflag.GetContactInfo(cmd.Flags(), "tech")
			listed, _ := cmd.Flags().GetBool("listed")
			var target *packetbroker.Target
			if err := pbflag.ApplyToTarget(cmd.Flags(), "target", &target); err != nil {
				return fmt.Errorf("configure target: %w", err)
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
				return fmt.Errorf("create tenant: %w", err)
			}
			if err := column.WriteTenant(tabout, res.GetTenant(), false); err != nil {
				return fmt.Errorf("write tenant: %w", err)
			}
			return nil
		},
	}
	networkTenantGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a tenant",
		Example: `
  Get:
    $ pbadmin network tenant get --net-id 000013 --tenant-id tti`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			res, err := iampb.NewTenantRegistryClient(conn).GetTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			if err != nil {
				return fmt.Errorf("get tenant: %w", err)
			}
			verbose, _ := cmd.Flags().GetBool("verbose")
			if err := column.WriteTenant(tabout, res.GetTenant(), verbose); err != nil {
				return fmt.Errorf("write tenant: %w", err)
			}
			return nil
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			client := iampb.NewTenantRegistryClient(conn)
			tnt, err := client.GetTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			if err != nil {
				return fmt.Errorf("get tenant: %w", err)
			}
			if cmd.Flags().Changed("listed") {
				listed, _ := cmd.Flags().GetBool("listed")
				_, err := client.UpdateTenantListed(ctx, &iampb.UpdateTenantListedRequest{
					NetId:    uint32(tenantID.NetID),
					TenantId: tenantID.ID,
					Listed:   listed,
				})
				if err != nil {
					return fmt.Errorf("update tenant listed: %w", err)
				}
			}
			var changed bool
			req := &iampb.UpdateTenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			}
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = wrapperspb.String(name)
				changed = true
			}
			devAddrBlocksAll, devAddrBlocksAllAdd, devAddrBlocksAllRemove := pbflag.GetDevAddrBlocks(cmd.Flags())
			if cmd.Flags().Changed("dev-addr-blocks") {
				req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
					Value: devAddrBlocksAll,
				}
				changed = true
			} else if len(devAddrBlocksAllAdd) > 0 || len(devAddrBlocksAllRemove) > 0 {
				req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
					Value: mergeDevAddrBlocks(tnt.GetTenant().GetDevAddrBlocks(), devAddrBlocksAllAdd, devAddrBlocksAllRemove),
				}
				changed = true
			}
			if adminContact := pbflag.GetContactInfo(cmd.Flags(), "admin"); adminContact != nil {
				req.AdministrativeContact = &packetbroker.ContactInfoValue{
					Value: adminContact,
				}
				changed = true
			}
			if techContact := pbflag.GetContactInfo(cmd.Flags(), "tech"); techContact != nil {
				req.TechnicalContact = &packetbroker.ContactInfoValue{
					Value: techContact,
				}
				changed = true
			}
			if changed {
				if _, err := client.UpdateTenant(ctx, req); err != nil {
					return fmt.Errorf("update tenant: %w", err)
				}
			}
			return nil
		},
	}
	networkTenantUpdateTargetCmd = &cobra.Command{ //nolint:gosec // the example URL contains a placeholder password
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			client := iampb.NewTenantRegistryClient(conn)
			tnt, err := client.GetTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			if err != nil {
				return fmt.Errorf("get tenant: %w", err)
			}
			target := tnt.GetTenant().GetTarget()
			if err := pbflag.ApplyToTarget(cmd.Flags(), "", &target); err != nil {
				return fmt.Errorf("configure target: %w", err)
			}
			req := &iampb.UpdateTenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
				Target: &iampb.TargetValue{
					Value: target,
				},
			}
			_, err = client.UpdateTenant(ctx, req)
			if err != nil {
				return fmt.Errorf("update tenant: %w", err)
			}
			return nil
		},
	}
	networkTenantDeleteCmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm"},
		Short:   "Delete a tenant",
		Example: `
  Delete:
    $ pbadmin network tenant delete --net-id 000013 --tenant-id tti`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tenantID, _ := pbflag.GetTenantID(cmd.Flags(), "")
			_, err := iampb.NewTenantRegistryClient(conn).DeleteTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			if err != nil {
				return fmt.Errorf("delete tenant: %w", err)
			}
			return nil
		},
	}
	networkTenantDeleteTargetCmd = &cobra.Command{
		Use:   "target",
		Short: "Delete a tenant target",
		Example: `
  Delete a tenant target:
    $ pbadmin network delete target --net-id 000013 --tenant-id tti`,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
			if err != nil {
				return fmt.Errorf("delete tenant target: %w", err)
			}
			return nil
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
	networkTenantDeleteTargetCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantDeleteCmd.AddCommand(networkTenantDeleteTargetCmd)
	networkTenantCmd.AddCommand(networkTenantDeleteCmd)
}
