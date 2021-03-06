// Copyright © 2020 The Things Industries B.V.

package pbflag

import (
	"fmt"
	"strings"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v3"
)

type netIDValue packetbroker.NetID

func (f *netIDValue) String() string {
	return packetbroker.NetID(*f).String()
}

func (f *netIDValue) Set(s string) error {
	var netID packetbroker.NetID
	if err := netID.UnmarshalText([]byte(s)); err != nil {
		return err
	}
	*f = netIDValue(netID)
	return nil
}

func (f *netIDValue) Type() string {
	return "netID"
}

func actorf(actor, name string) string {
	if actor != "" {
		return fmt.Sprintf("%s-%s", actor, name)
	}
	return name
}

// NetID returns flags for a NetID.
func NetID(actor string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(netIDValue), actorf(actor, "net-id"), "LoRa Alliance NetID")
	return flags
}

// GetNetID returns the NetID from the flags.
func GetNetID(flags *flag.FlagSet, actor string) packetbroker.NetID {
	return packetbroker.NetID(*flags.Lookup(actorf(actor, "net-id")).Value.(*netIDValue))
}

// TenantID returns flags for a TenantID.
func TenantID(actor string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(NetID(actor))
	flags.String(actorf(actor, "tenant-id"), "", "tenant ID")
	return flags
}

// GetTenantID returns the TenantID from the flags.
// The actor is used as prefix.
func GetTenantID(flags *flag.FlagSet, actor string) packetbroker.TenantID {
	netID := GetNetID(flags, actor)
	tenantID, _ := flags.GetString(actorf(actor, "tenant-id"))
	return packetbroker.TenantID{
		NetID: netID,
		ID:    tenantID,
	}
}

// Endpoint returns flags for an Endpoint.
func Endpoint(actor string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(TenantID(actor))
	flags.String(actorf(actor, "cluster-id"), "", "cluster ID")
	return flags
}

// GetEndpoint returns the Endpoint from the flags.
func GetEndpoint(flags *flag.FlagSet, actor string) packetbroker.Endpoint {
	tenantID := GetTenantID(flags, actor)
	clusterID, _ := flags.GetString(actorf(actor, "cluster-id"))
	return packetbroker.Endpoint{
		TenantID:  tenantID,
		ClusterID: clusterID,
	}
}

// HasEndpoint returns which endpoint flags are set.
func HasEndpoint(flags *flag.FlagSet, actor string) (hasNetID, hasTenantID, hasClusterID bool) {
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case actorf(actor, "net-id"):
			hasNetID = true
		case actorf(actor, "tenant-id"):
			hasTenantID = true
		case actorf(actor, "cluster-id"):
			hasClusterID = true
		}
	})
	return
}

type devAddrBlocksValue []*packetbroker.DevAddrBlock

func (f *devAddrBlocksValue) String() string {
	ss := make([]string, len(*f))
	for i, b := range *f {
		prefix, _ := b.Prefix.MarshalText()
		if b.HomeNetworkClusterId != "" {
			ss[i] = fmt.Sprintf("%s=%s", string(prefix), b.HomeNetworkClusterId)
		} else {
			ss[i] = string(prefix)
		}
	}
	return strings.Join(ss, ",")
}

func (f *devAddrBlocksValue) Set(s string) error {
	if s == "" {
		*f = []*packetbroker.DevAddrBlock{}
		return nil
	}
	blocks := strings.Split(s, ",")
	res := make([]*packetbroker.DevAddrBlock, len(blocks))
	for i, b := range blocks {
		parts := strings.SplitN(b, "=", 2)
		res[i] = &packetbroker.DevAddrBlock{
			Prefix: &packetbroker.DevAddrPrefix{},
		}
		if len(parts) == 2 {
			res[i].HomeNetworkClusterId = parts[1]
		}
		if err := res[i].Prefix.UnmarshalText([]byte(parts[0])); err != nil {
			return err
		}
	}
	*f = res
	return nil
}

func (f *devAddrBlocksValue) Type() string {
	return "devAddrBlocks"
}

// DevAddrBlocks returns flags for DevAddr blocks.
func DevAddrBlocks() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(devAddrBlocksValue), "dev-addr-blocks", "DevAddr blocks")
	return flags
}

// GetDevAddrBlocks returns the DevAddr blocks from the flags.
func GetDevAddrBlocks(flags *flag.FlagSet) []*packetbroker.DevAddrBlock {
	blocks := flags.Lookup("dev-addr-blocks").Value.(*devAddrBlocksValue)
	return []*packetbroker.DevAddrBlock(*blocks)
}

type apiKeyRightsValue []packetbroker.Right

func (p apiKeyRightsValue) String() string {
	rights := make([]string, 0, len(p))
	for _, v := range p {
		switch v {
		case packetbroker.Right_READ_NETWORK:
			rights = append(rights, "r:network")
		case packetbroker.Right_READ_NETWORK_CONTACT:
			rights = append(rights, "r:network:contact")
		case packetbroker.Right_READ_TENANT:
			rights = append(rights, "r:tenant")
		case packetbroker.Right_READ_TENANT_CONTACT:
			rights = append(rights, "r:tenant:contact")
		case packetbroker.Right_READ_ROUTING_POLICY:
			rights = append(rights, "r:routing_policy")
		case packetbroker.Right_READ_ROUTE_TABLE:
			rights = append(rights, "r:route_table")
		}
	}
	return strings.Join(rights, ",")
}

func (p *apiKeyRightsValue) Set(s string) error {
	if s == "" {
		*p = []packetbroker.Right{}
		return nil
	}
	rights := strings.Split(s, ",")
	res := make([]packetbroker.Right, len(rights))
	for i, r := range rights {
		switch r {
		case "r:network":
			res[i] = packetbroker.Right_READ_NETWORK
		case "r:network:contact":
			res[i] = packetbroker.Right_READ_NETWORK_CONTACT
		case "r:tenant":
			res[i] = packetbroker.Right_READ_TENANT
		case "r:tenant:contact":
			res[i] = packetbroker.Right_READ_TENANT_CONTACT
		case "r:routing_policy":
			res[i] = packetbroker.Right_READ_ROUTING_POLICY
		case "r:route_table":
			res[i] = packetbroker.Right_READ_ROUTE_TABLE
		default:
			return fmt.Errorf("pbflag: invalid right: %s", r)
		}
	}
	*p = res
	return nil
}

func (p *apiKeyRightsValue) Type() string {
	return "apiKeyRightsValue"
}

// APIKeyRights returns flags for API key rights.
func APIKeyRights() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(apiKeyRightsValue), "rights", "API key rights")
	return flags
}

// GetAPIKeyRights returns the API key rights from the flags.
func GetAPIKeyRights(flags *flag.FlagSet) []packetbroker.Right {
	rights := flags.Lookup("rights").Value.(*apiKeyRightsValue)
	return []packetbroker.Right(*rights)
}

type uplinkPolicyValue packetbroker.RoutingPolicy_Uplink

func (p *uplinkPolicyValue) String() string {
	var res string
	if p.JoinRequest {
		res += "J"
	}
	if p.MacData {
		res += "M"
	}
	if p.ApplicationData {
		res += "A"
	}
	if p.SignalQuality {
		res += "S"
	}
	if p.Localization {
		res += "L"
	}
	return res
}

func (p *uplinkPolicyValue) Set(s string) error {
	*p = uplinkPolicyValue(packetbroker.RoutingPolicy_Uplink{
		JoinRequest:     strings.ContainsRune(s, 'J'),
		MacData:         strings.ContainsRune(s, 'M'),
		ApplicationData: strings.ContainsRune(s, 'A'),
		SignalQuality:   strings.ContainsRune(s, 'S'),
		Localization:    strings.ContainsRune(s, 'L'),
	})
	return nil
}

func (p *uplinkPolicyValue) Type() string {
	return "uplinkPolicy"
}

type downlinkPolicyValue packetbroker.RoutingPolicy_Downlink

func (p *downlinkPolicyValue) String() string {
	var res string
	if p.JoinAccept {
		res += "J"
	}
	if p.MacData {
		res += "M"
	}
	if p.ApplicationData {
		res += "A"
	}
	return res
}

func (p *downlinkPolicyValue) Set(s string) error {
	*p = downlinkPolicyValue(packetbroker.RoutingPolicy_Downlink{
		JoinAccept:      strings.ContainsRune(s, 'J'),
		MacData:         strings.ContainsRune(s, 'M'),
		ApplicationData: strings.ContainsRune(s, 'A'),
	})
	return nil
}

func (p *downlinkPolicyValue) Type() string {
	return "downlinkPolicy"
}

// RoutingPolicy returns flags for a routing policy.
func RoutingPolicy() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(uplinkPolicyValue), "set-uplink", "uplink policy, use letters J, M, A, S and L")
	flags.Var(new(downlinkPolicyValue), "set-downlink", "downlink policy, use letters J, M and A")
	return flags
}

// GetRoutingPolicy returns the routing policy from the flags.
func GetRoutingPolicy(flags *flag.FlagSet) (*packetbroker.RoutingPolicy_Uplink, *packetbroker.RoutingPolicy_Downlink) {
	return (*packetbroker.RoutingPolicy_Uplink)(flags.Lookup("set-uplink").Value.(*uplinkPolicyValue)),
		(*packetbroker.RoutingPolicy_Downlink)(flags.Lookup("set-downlink").Value.(*downlinkPolicyValue))
}
