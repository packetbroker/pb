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

type targetProtocol struct {
	*packetbroker.TargetProtocol
}

func (p targetProtocol) String() string {
	if p.TargetProtocol == nil {
		return ""
	}
	return packetbroker.TargetProtocol_name[int32(*p.TargetProtocol)]
}

func (p *targetProtocol) Set(s string) error {
	if s == "" {
		*p = targetProtocol{}
		return nil
	}
	i, ok := packetbroker.TargetProtocol_value[s]
	if !ok {
		return fmt.Errorf("pbflag: invalid protocol: %s", s)
	}
	*p = targetProtocol{
		TargetProtocol: (*packetbroker.TargetProtocol)(&i),
	}
	return nil
}

func (p *targetProtocol) Type() string {
	return "targetProtocol"
}

// TargetProtocol returns flags for a target protocol.
func TargetProtocol() *flag.FlagSet {
	names := make([]string, 0, len(packetbroker.TargetProtocol_value))
	for k := range packetbroker.TargetProtocol_value {
		names = append(names, k)
	}
	flags := new(flag.FlagSet)
	flags.Var(new(targetProtocol), "target-protocol", fmt.Sprintf("target protocol (%s)", strings.Join(names, ",")))
	return flags
}

// GetTargetProtocol returns the target protocol from the flags.
func GetTargetProtocol(flags *flag.FlagSet) *packetbroker.TargetProtocol {
	return flags.Lookup("target-protocol").Value.(*targetProtocol).TargetProtocol
}

type apiKeyRightsValue []packetbroker.Right

func (p apiKeyRightsValue) String() string {
	rights := make([]string, 0, len(p))
	for _, v := range p {
		rights = append(rights, v.String())
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
		v, ok := packetbroker.Right_value[r]
		if !ok {
			return fmt.Errorf("pbflag: invalid right: %s", r)
		}
		res[i] = packetbroker.Right(v)
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
