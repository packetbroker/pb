// Copyright © 2021 The Things Industries B.V.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	iampbv2 "go.packetbroker.org/api/iam/v2"
	"go.packetbroker.org/pb/cmd/internal/column"
	pbflag "go.packetbroker.org/pb/cmd/internal/pbflag"
	"golang.org/x/term"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	clusterAPIKeyCmd = &cobra.Command{
		Use:     "apikey",
		Aliases: []string{"apikeys", "key", "keys"},
		Short:   "Manage Packet Broker API keys for clusters",
	}
	clusterAPIKeyListCmd = &cobra.Command{
		Use:          "list",
		Aliases:      []string{"ls"},
		Short:        "List API keys",
		SilenceUsage: true,
		Example: `
  List all API keys:
    $ pbadmin cluster apikey list

  List API keys of a named cluster:
    $ pbadmin cluster apikey list --cluster-id eu1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &iampbv2.ListClusterAPIKeysRequest{}
			cmd.Flags().Visit(func(f *pflag.Flag) {
				if f.Name == "cluster-id" {
					req.ClusterId = wrapperspb.String(f.Value.String())
				}
			})
			res, err := iampbv2.NewClusterAPIKeyVaultClient(conn).ListAPIKeys(ctx, req)
			if err != nil {
				return err
			}
			fmt.Fprintln(tabout, "Key ID\tClusterID\tRights\tState\tLast Used\t")
			for _, t := range res.Keys {
				fmt.Fprintf(tabout, "%s\t%s\t%s\t%s\t%s\t\n",
					t.GetKeyId(),
					t.GetClusterId(),
					column.Rights(t.GetRights()),
					t.GetState(),
					(*column.TimeSince)(t.GetAuthenticatedAt()),
				)
			}
			return nil
		},
	}
	clusterAPIKeyCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an API key",
		Long: `Create an API key for a named cluster.

The secret API key is returned only when creating the API key. You should store
the API key in a secure place, as it cannot be retrieved after create.`,
		Example: `
  Create API key:
    $ pbadmin cluster apikey create --cluster-id eu1

  Create API key with rights to read networks and tenants:
    $ pbadmin cluster apikey create --cluster-id eu1 --rights READ_NETWORK

Rights:
  READ_NETWORK          Read networks
  READ_NETWORK_CONTACT  Read network contact information
  READ_TENANT           Read tenants
  READ_TENANT_CONTACT   Read tenant contact information
  READ_ROUTING_POLICY   Read routing policies
  READ_ROUTE_TABLE      Read route table
  READ_TARGET_AUTH      Read target authentication information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := pbflag.GetEndpoint(cmd.Flags(), "")
			req := &iampbv2.CreateClusterAPIKeyRequest{
				ClusterId: endpoint.ClusterID,
				Rights:    pbflag.GetAPIKeyRights(cmd.Flags()),
			}
			if promptKey, _ := cmd.Flags().GetBool("prompt-key"); promptKey {
				fmt.Fprint(os.Stdout, "Secret key: ")
				keyBuf, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return err
				}
				req.Key = string(keyBuf)
			}
			res, err := iampbv2.NewClusterAPIKeyVaultClient(conn).CreateAPIKey(ctx, req)
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "Store the API key now in a secure place, as it cannot be retrieved later.")
			return column.WriteKV(tabout,
				"Key ID", res.Key.GetKeyId(),
				"Secret Key", res.Key.GetKey(),
				"Cluster ID", res.Key.GetClusterId(),
				"Rights", column.Rights(res.Key.GetRights()),
				"State", res.Key.GetState(),
			)
		},
	}
	clusterAPIKeyUpdateStateCmd = &cobra.Command{
		Use:          "update-state",
		Short:        "Update the API key state",
		SilenceUsage: true,
		Example: `
  Update the API key state to APPROVED:
    $ pbadmin cluster apikey update-state --key-id C5232IFFX4UKEELB --state APPROVED`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyID, _ := cmd.Flags().GetString("key-id")
			state := pbflag.GetAPIKeyState(cmd.Flags(), "state")
			_, err := iampbv2.NewClusterAPIKeyVaultClient(conn).UpdateAPIKeyState(ctx, &iampbv2.UpdateAPIKeyStateRequest{
				KeyId: keyID,
				State: state,
			})
			return err
		},
	}
	clusterAPIKeyDeleteCmd = &cobra.Command{
		Use:          "delete",
		Aliases:      []string{"rm"},
		Short:        "Delete an API key",
		SilenceUsage: true,
		Example: `
  Delete an API key:
    $ pbadmin cluster apikey delete --key-id C5232IFFX4UKEELB`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyID, _ := cmd.Flags().GetString("key-id")
			_, err := iampbv2.NewClusterAPIKeyVaultClient(conn).DeleteAPIKey(ctx, &iampbv2.APIKeyRequest{
				KeyId: keyID,
			})
			return err
		},
	}
)

func init() {
	clusterCmd.AddCommand(clusterAPIKeyCmd)

	clusterAPIKeyListCmd.Flags().AddFlagSet(pbflag.Endpoint(""))
	clusterAPIKeyCmd.AddCommand(clusterAPIKeyListCmd)

	clusterAPIKeyCreateCmd.Flags().AddFlagSet(pbflag.Endpoint(""))
	clusterAPIKeyCreateCmd.Flags().AddFlagSet(pbflag.APIKeyRights())
	clusterAPIKeyCreateCmd.Flags().Bool("prompt-key", false, "prompt custom secret key value")
	clusterAPIKeyCmd.AddCommand(clusterAPIKeyCreateCmd)

	clusterAPIKeyUpdateStateCmd.Flags().String("key-id", "", "API key ID")
	clusterAPIKeyUpdateStateCmd.Flags().AddFlagSet(pbflag.APIKeyState("state"))
	clusterAPIKeyCmd.AddCommand(clusterAPIKeyUpdateStateCmd)

	clusterAPIKeyDeleteCmd.Flags().String("key-id", "", "API key ID")
	clusterAPIKeyCmd.AddCommand(clusterAPIKeyDeleteCmd)
}
